# 🐛 Bug Report: WebSocket 401 Unauthorized

**Date:** 2026-02-15  
**Priority:** High  
**Affected Endpoint:** `GET /api/v1/ws?token=...`

---

## Symptom

Frontend WebSocket connection selalu gagal dengan **401 Unauthorized**:

```
"GET /api/v1/ws?token=eyJhbGci... HTTP/1.1" → 401 82B in 71.38µs
```

Token JWT valid dan belum expired, tapi koneksi tetap ditolak.

---

## Root Cause

Ada **2 bug** di backend:

### Bug 1: Auth middleware tidak support `?token=` query parameter

📄 **File:** `internal/auth/middleware.go` — line 22-26

```go
authHeader := r.Header.Get("Authorization")
if authHeader == "" {
    response.Unauthorized(w, "Missing authorization header") // ← 401 di sini
    return
}
```

**Masalah:** Middleware hanya membaca token dari `Authorization` header. Untuk WebSocket, browser **tidak bisa** mengirim custom header saat handshake — standar industri adalah mengirim token via **query parameter** (`?token=...`). Middleware langsung reject sebelum request sampai ke WS handler.

### Bug 2: WS handler salah membaca context key

📄 **File:** `internal/websocket/handler.go` — line 56-60

```go
// Handler mencari key "claims" dengan type map[string]interface{}
if claims, ok := r.Context().Value("claims").(map[string]interface{}); ok {
    if id, ok := claims["user_id"].(string); ok {
        userID = id
    }
}
```

**Masalah:** Auth middleware menyimpan claims ke context dengan:
- Key: `auth.UserContextKey` = `"user"` (bukan `"claims"`)
- Type: `*auth.Claims` (bukan `map[string]interface{}`)

Bahkan jika Bug 1 diperbaiki, handler tetap tidak akan mendapatkan `userID`.

---

## Proposed Fix

### Fix 1: `internal/auth/middleware.go`

Tambahkan fallback ke query parameter `token`:

```diff
 func Middleware(secret string) func(http.Handler) http.Handler {
     return func(next http.Handler) http.Handler {
         return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
             authHeader := r.Header.Get("Authorization")
             if authHeader == "" {
-                response.Unauthorized(w, "Missing authorization header")
-                return
+                // Fallback: check ?token= query param (required for WebSocket)
+                if tokenParam := r.URL.Query().Get("token"); tokenParam != "" {
+                    authHeader = "Bearer " + tokenParam
+                } else {
+                    response.Unauthorized(w, "Missing authorization header")
+                    return
+                }
             }

             parts := strings.Split(authHeader, " ")
```

### Fix 2: `internal/websocket/handler.go`

Gunakan context key dan type yang benar:

```diff
+import "github.com/swiftlead/backend-swiftlet/internal/auth"

 func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
     // ...

     userID := ""
-    if claims, ok := r.Context().Value("claims").(map[string]interface{}); ok {
-        if id, ok := claims["user_id"].(string); ok {
-            userID = id
-        }
-    }
+    if claims := auth.GetUserFromContext(r.Context()); claims != nil {
+        userID = claims.UserID
+    }

     // Upgrade HTTP connection to WebSocket
     conn, err := h.upgrader.Upgrade(w, r, nil)
```

> **Note:** `auth.GetUserFromContext()` sudah ada dan siap pakai (line 91-97 di middleware.go).

---

## How to Verify

1. Deploy fix ke staging
2. Buka frontend dashboard → cek Console, WebSocket harus konek tanpa 401
3. Cek backend log: `[WS] Incoming headers: ...` harus muncul (artinya request lolos middleware)
4. Sensor data harus muncul live di dashboard

---

## References

- **Frontend WS hook:** `src/lib/hooks/use-websocket.ts` — kirim token via `?token=` query param
- **Auth middleware:** `internal/auth/middleware.go`
- **WS handler:** `internal/websocket/handler.go`
- **Context helper:** `auth.GetUserFromContext()` di `internal/auth/middleware.go:91`
