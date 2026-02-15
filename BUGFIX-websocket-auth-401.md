# 🐛 Bug Report: WebSocket Auth 401 + Missing Broadcast

**Date:** 2026-02-15  
**Priority:** High  
**Affected:** Realtime sensor data di dashboard (WebSocket)

---

## Symptoms

1. WebSocket connection gagal 401 Unauthorized
2. Setelah auth fix, WS konek tapi **tidak ada data sensor yang dikirim**

---

## Bug 1: Auth middleware tidak support `?token=` query param

📄 `internal/auth/middleware.go` — line 22-26

```go
authHeader := r.Header.Get("Authorization")
if authHeader == "" {
    response.Unauthorized(w, "Missing authorization header") // ← 401
    return
}
```

**Masalah:** Browser tidak bisa kirim custom header saat WebSocket handshake. Frontend kirim token via `?token=...` query param (standar industri), tapi middleware hanya baca `Authorization` header.

**Fix:**

```diff
 authHeader := r.Header.Get("Authorization")
 if authHeader == "" {
-    response.Unauthorized(w, "Missing authorization header")
-    return
+    if tokenParam := r.URL.Query().Get("token"); tokenParam != "" {
+        authHeader = "Bearer " + tokenParam
+    } else {
+        response.Unauthorized(w, "Missing authorization header")
+        return
+    }
 }
```

---

## Bug 2: WS handler salah baca context key

📄 `internal/websocket/handler.go` — line 56-60

```go
if claims, ok := r.Context().Value("claims").(map[string]interface{}); ok {
```

**Masalah:** Auth middleware simpan dengan key `auth.UserContextKey` = `"user"` (type `*auth.Claims`), tapi handler cari key `"claims"` (type `map[string]interface{}`).

**Fix:**

```diff
+import "github.com/swiftlead/backend-swiftlet/internal/auth"

 userID := ""
-if claims, ok := r.Context().Value("claims").(map[string]interface{}); ok {
-    if id, ok := claims["user_id"].(string); ok {
-        userID = id
-    }
-}
+if claims := auth.GetUserFromContext(r.Context()); claims != nil {
+    userID = claims.UserID
+}
```

---

## Bug 3 (UTAMA): Sensor data tidak di-broadcast ke WebSocket

📄 `internal/services/telemetry_service.go`

**Masalah:** `ProcessSensorPayload()` menyimpan reading ke database dan membuat alert, tapi **TIDAK PERNAH** memanggil `hub.BroadcastAll()` atau `hub.PublishToTopic()`. Datanya masuk ke DB tapi tidak pernah dikirim ke WS clients.

`TelemetryService` juga **tidak punya referensi ke `websocket.Hub`** — tidak di-inject lewat constructor.

**Fix — Step 1:** Inject Hub ke TelemetryService

```diff
+import "github.com/swiftlead/backend-swiftlet/internal/websocket"

 type TelemetryService struct {
     nodeRepo      repository.NodeRepository
     sensorRepo    repository.SensorRepository
     telemetryRepo repository.TelemetryRepository
     alertRepo     repository.AlertRepository
     aiClient      *ai.Client
     cfg           *config.Config
+    wsHub         *websocket.Hub
 }

 func NewTelemetryService(
     nodeRepo repository.NodeRepository,
     sensorRepo repository.SensorRepository,
     telemetryRepo repository.TelemetryRepository,
     alertRepo repository.AlertRepository,
     aiClient *ai.Client,
     cfg *config.Config,
+    wsHub *websocket.Hub,
 ) *TelemetryService {
     return &TelemetryService{
         nodeRepo:      nodeRepo,
         sensorRepo:    sensorRepo,
         telemetryRepo: telemetryRepo,
         alertRepo:     alertRepo,
         aiClient:      aiClient,
         cfg:           cfg,
+        wsHub:         wsHub,
     }
 }
```

**Fix — Step 2:** Broadcast setelah save reading (di `ProcessSensorPayload`, setelah line 139)

```diff
     if err := s.telemetryRepo.CreateReading(ctx, reading); err != nil {
         return err
     }

+    // Broadcast ke WebSocket clients
+    if s.wsHub != nil {
+        sensorData, _ := json.Marshal(map[string]interface{}{
+            "sensor_id":   sensor.ID,
+            "node_id":     node.ID,
+            "rbw_id":      node.RBWID,
+            "sensor_type": sensor.SensorType,
+            "value":       value,
+            "unit":        sensor.Unit,
+            "is_anomaly":  isAnomaly,
+            "recorded_at": recordedAt.Format(time.RFC3339),
+        })
+        s.wsHub.BroadcastAll(websocket.Message{
+            Type:      "sensor_reading",
+            Data:      sensorData,
+            Timestamp: time.Now(),
+        })
+    }
```

**Fix — Step 3:** Update `container.go` untuk inject Hub

```diff
-c.TelemetryService = services.NewTelemetryService(
-    c.NodeRepo, c.SensorRepo, c.TelemetryRepo, c.AlertRepo, c.AIClient, c.Config,
-)
+c.TelemetryService = services.NewTelemetryService(
+    c.NodeRepo, c.SensorRepo, c.TelemetryRepo, c.AlertRepo, c.AIClient, c.Config, c.WSHub.Hub,
+)
```

---

## How to Verify

1. Deploy semua fix
2. Buka frontend dashboard → cek Console, WS harus konek
3. Pastikan MQTT sensor data masuk → cek backend log ada `[WS] BroadcastAll` atau similar
4. Dashboard gauge cards + chart harus menampilkan data realtime

---

## Frontend Event Format yang Diharapkan

```json
{
  "type": "sensor_reading",
  "data": {
    "sensor_id": "uuid",
    "node_id": "uuid",
    "rbw_id": "uuid",
    "sensor_type": "temp" | "humid" | "ammonia",
    "value": 25.85,
    "unit": "°C",
    "is_anomaly": false,
    "recorded_at": "2026-02-15T13:54:05Z"
  },
  "timestamp": "2026-02-15T13:54:05Z"
}
```
