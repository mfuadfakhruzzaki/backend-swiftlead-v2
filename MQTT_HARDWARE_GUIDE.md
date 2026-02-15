# SwiftLead — Panduan Koneksi MQTT untuk Tim Hardware

**Versi:** 1.0  
**Tanggal:** 14 Februari 2026  
**Protokol:** MQTT v3.1.1  
**Broker:** Mosquitto

---

## Daftar Isi

1. [Informasi Koneksi](#1-informasi-koneksi)
2. [Arsitektur Komunikasi](#2-arsitektur-komunikasi)
3. [Topik MQTT](#3-topik-mqtt)
4. [Mengirim Data Sensor (Telemetri)](#4-mengirim-data-sensor-telemetri)
5. [Menerima Perintah dari Server](#5-menerima-perintah-dari-server)
6. [Mekanisme Status Node](#6-mekanisme-status-node)
7. [Contoh Kode ESP32 (Arduino)](#7-contoh-kode-esp32-arduino)
8. [Testing & Debugging](#8-testing--debugging)
9. [FAQ](#9-faq)

---

## 1. Informasi Koneksi

### Development (Lokal)

| Parameter     | Nilai                    |
| ------------- | ------------------------ |
| **Broker**    | `tcp://localhost:1883`   |
| **Username**  | _(kosong)_               |
| **Password**  | _(kosong)_               |
| **Client ID** | `esp32-<MAC_ADDRESS>`    |
| **Keep Alive**| `30` detik               |

### Production (Server)

| Parameter     | Nilai                              |
| ------------- | ---------------------------------- |
| **Broker**    | `tcp://<SERVER_IP>:1883`           |
| **Broker TLS**| `ssl://<SERVER_IP>:8883` (opsional)|
| **Username**  | _(sesuai konfigurasi server)_      |
| **Password**  | _(sesuai konfigurasi server)_      |
| **Client ID** | `esp32-<MAC_ADDRESS>`              |
| **QoS**       | `1` (At least once)                |

> **Catatan:** Hubungi tim backend untuk mendapatkan alamat IP dan kredensial server production.

---

## 2. Arsitektur Komunikasi

```
┌──────────────┐          MQTT           ┌──────────────────┐
│              │  ──── Telemetri ──────► │                  │
│   ESP32      │   swiftlead/tel/<UID>   │   Backend        │
│   Node       │                         │   Server         │
│              │  ◄──── Perintah ──────  │                  │
│              │   swiftlead/cmd/...     │                  │
└──────────────┘                         └──────────────────┘
       │                                          │
       │            MQTT Broker                   │
       └──────── (Mosquitto) ─────────────────────┘
```

**Alur Data:**
1. **ESP32 → Server:** ESP32 mengirim data sensor ke topik `swiftlead/tel/<ESP32_UID>`
2. **Server → ESP32:** Server mengirim perintah kontrol ke topik `swiftlead/cmd/...`

---

## 3. Topik MQTT

### Topik Telemetri (ESP32 → Server)

| Topik                           | Arah             | Deskripsi                          |
| ------------------------------- | ---------------- | ---------------------------------- |
| `swiftlead/tel/<ESP32_UID>`     | ESP32 → Server   | Kirim data sensor (suhu, kelembaban, amonia) |

- `<ESP32_UID>` adalah **MAC Address** atau **Unique ID** dari ESP32.
- Contoh: `swiftlead/tel/AA:BB:CC:DD:EE:FF`

### Topik Perintah (Server → ESP32)

| Topik                       | Arah             | Deskripsi                       |
| --------------------------- | ---------------- | ------------------------------- |
| `swiftlead/cmd/lmb/set`    | Server → ESP32   | Kontrol audio LMB / Nest / Call Bird |
| `swiftlead/cmd/pump/set`   | Server → ESP32   | Kontrol pompa sprayer           |

> **Penting:** ESP32 harus **subscribe** ke topik `swiftlead/cmd/#` untuk menerima semua perintah.

---

## 4. Mengirim Data Sensor (Telemetri)

### Format Payload JSON

ESP32 harus mengirim data sensor dalam format JSON berikut ke topik `swiftlead/tel/<ESP32_UID>`:

```json
{
  "esp32_uid": "AA:BB:CC:DD:EE:FF",
  "temp": 28.5,
  "rh": 75.0,
  "nh3": 15.0,
  "rssi": -60,
  "timestamp": 1707465600,
  "seq": 1
}
```

### Detail Field

| Field        | Tipe    | Wajib | Deskripsi                                              |
| ------------ | ------- | ----- | ------------------------------------------------------ |
| `esp32_uid`  | string  | ✅    | MAC Address / UID ESP32. **Harus sama** dengan yang terdaftar di server |
| `temp`       | float   | ✅    | Suhu dalam **°C** (contoh: `28.5`)                     |
| `rh`         | float   | ✅    | Kelembaban relatif dalam **%** (contoh: `75.0`)        |
| `nh3`        | float   | ✅    | Kadar amonia dalam **ppm** (contoh: `15.0`)            |
| `rssi`       | int     | ❌    | Kekuatan sinyal WiFi dalam dBm (contoh: `-60`)         |
| `timestamp`  | int     | ❌    | Unix timestamp (detik). Jika `0` atau tidak ada, server menggunakan waktu sekarang |
| `seq`        | int     | ❌    | Nomor urut paket, untuk debugging                      |

### Interval Pengiriman

- **Rekomendasi:** Kirim data setiap **5–10 detik**
- **Minimum:** Setiap 15 menit (agar node tidak dianggap offline)

### Threshold Alert di Server

Server akan otomatis membuat alert jika nilai sensor melebihi batas:

| Sensor     | Batas Bawah | Batas Atas |
| ---------- | ----------- | ---------- |
| Suhu (°C)  | 20.0        | 35.0       |
| Kelembaban (%) | 60.0    | 85.0       |
| Amonia (ppm) | -         | 25.0       |

---

## 5. Menerima Perintah dari Server

ESP32 harus **subscribe** ke topik `swiftlead/cmd/#` untuk menerima perintah dari server.

### 5.1 Perintah Audio

**Topik:** `swiftlead/cmd/lmb/set`

**Payload:**

```json
{
  "action": "audio_set_lmb",
  "value": 1
}
```

| Field    | Tipe   | Nilai yang Mungkin                            | Deskripsi              |
| -------- | ------ | --------------------------------------------- | ---------------------- |
| `action` | string | `audio_set_lmb`, `audio_set_nest`, `call_bird`| Jenis aksi audio       |
| `value`  | int    | `0` = Mati, `1` = Hidup                       | State on/off           |

**Aksi yang didukung:**

| Action           | Deskripsi                                      |
| ---------------- | ---------------------------------------------- |
| `audio_set_lmb`  | Nyalakan/matikan audio LMB (tweeter lantai)    |
| `audio_set_nest` | Nyalakan/matikan audio Nest (tweeter sarang)   |
| `call_bird`      | Panggil burung (momentary, otomatis mati)      |

### 5.2 Perintah Pompa

**Topik:** `swiftlead/cmd/pump/set`

**Payload:**

```json
{
  "action": "sprayer_set",
  "value": 1
}
```

| Field    | Tipe   | Nilai yang Mungkin | Deskripsi              |
| -------- | ------ | ------------------ | ---------------------- |
| `action` | string | `sprayer_set`      | Kontrol pompa sprayer  |
| `value`  | int    | `0` = Mati, `1` = Hidup | State on/off       |

---

## 6. Mekanisme Status Node

Status node ditentukan **otomatis** oleh server:

| Status    | Kondisi                                                  |
| --------- | -------------------------------------------------------- |
| `online`  | Server menerima data telemetri dari node                 |
| `offline` | Node tidak mengirim data selama > 15 menit (konfigurasi) |
| `error`   | Diset manual oleh admin jika terjadi error               |

**Cara kerja:**
1. Setiap kali data sensor diterima, server otomatis mengupdate `status_node` = `online` dan `last_seen` = waktu sekarang.
2. Pastikan ESP32 mengirim data secara berkala (minimal setiap 15 menit) agar status tetap `online`.

---

## 7. Contoh Kode ESP32 (Arduino)

### Dependencies

Tambahkan library berikut di Arduino IDE atau PlatformIO:
- **PubSubClient** by Nick O'Leary (untuk MQTT)
- **ArduinoJson** by Benoit Blanchon (untuk JSON)
- **WiFi** (bawaan ESP32)

### Kode Lengkap

```cpp
#include <WiFi.h>
#include <PubSubClient.h>
#include <ArduinoJson.h>

// ============================================
// KONFIGURASI - SESUAIKAN DENGAN KEBUTUHAN
// ============================================
const char* WIFI_SSID     = "NAMA_WIFI";
const char* WIFI_PASSWORD  = "PASSWORD_WIFI";

const char* MQTT_BROKER    = "IP_SERVER";       // Ganti dengan IP server
const int   MQTT_PORT      = 1883;
const char* MQTT_USERNAME  = "";                // Kosong jika tidak pakai auth
const char* MQTT_PASSWORD  = "";                // Kosong jika tidak pakai auth

// ESP32 UID - HARUS SAMA dengan yang terdaftar di server
// Gunakan MAC Address ESP32 (format AA:BB:CC:DD:EE:FF)
String ESP32_UID = "";  // Akan diisi otomatis dari MAC Address

// Interval pengiriman data (ms)
const unsigned long SEND_INTERVAL = 5000;  // 5 detik

// ============================================
// TOPIK MQTT
// ============================================
String TOPIC_TELEMETRY = "";  // Akan diisi setelah UID diketahui
const char* TOPIC_CMD_AUDIO = "swiftlead/cmd/lmb/set";
const char* TOPIC_CMD_PUMP  = "swiftlead/cmd/pump/set";

// ============================================
// VARIABEL GLOBAL
// ============================================
WiFiClient espClient;
PubSubClient mqttClient(espClient);

unsigned long lastSendTime = 0;
unsigned long packetSeq = 0;

// Pin sensor (sesuaikan dengan wiring)
// const int PIN_TEMP    = 34;
// const int PIN_HUMID   = 35;
// const int PIN_AMMONIA = 32;

// Pin aktuator (sesuaikan dengan wiring)
const int PIN_AUDIO_LMB  = 25;
const int PIN_AUDIO_NEST = 26;
const int PIN_PUMP       = 27;

// ============================================
// FUNGSI SETUP
// ============================================
void setup() {
  Serial.begin(115200);
  delay(1000);

  // Setup pin aktuator
  pinMode(PIN_AUDIO_LMB, OUTPUT);
  pinMode(PIN_AUDIO_NEST, OUTPUT);
  pinMode(PIN_PUMP, OUTPUT);
  digitalWrite(PIN_AUDIO_LMB, LOW);
  digitalWrite(PIN_AUDIO_NEST, LOW);
  digitalWrite(PIN_PUMP, LOW);

  // Dapatkan MAC Address sebagai UID
  ESP32_UID = WiFi.macAddress();
  TOPIC_TELEMETRY = "swiftlead/tel/" + ESP32_UID;

  Serial.println("========================================");
  Serial.println("SwiftLead ESP32 Node");
  Serial.printf("ESP32 UID: %s\n", ESP32_UID.c_str());
  Serial.printf("Telemetry Topic: %s\n", TOPIC_TELEMETRY.c_str());
  Serial.println("========================================");

  // Koneksi WiFi
  connectWiFi();

  // Setup MQTT
  mqttClient.setServer(MQTT_BROKER, MQTT_PORT);
  mqttClient.setCallback(onMqttMessage);
  mqttClient.setBufferSize(512);

  connectMQTT();
}

// ============================================
// FUNGSI LOOP
// ============================================
void loop() {
  // Pastikan koneksi MQTT tetap aktif
  if (!mqttClient.connected()) {
    connectMQTT();
  }
  mqttClient.loop();

  // Kirim data sensor secara berkala
  unsigned long now = millis();
  if (now - lastSendTime >= SEND_INTERVAL) {
    lastSendTime = now;
    sendTelemetry();
  }
}

// ============================================
// KONEKSI WIFI
// ============================================
void connectWiFi() {
  Serial.printf("Connecting to WiFi: %s", WIFI_SSID);
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);

  int attempts = 0;
  while (WiFi.status() != WL_CONNECTED && attempts < 30) {
    delay(500);
    Serial.print(".");
    attempts++;
  }

  if (WiFi.status() == WL_CONNECTED) {
    Serial.printf("\nWiFi connected! IP: %s\n", WiFi.localIP().toString().c_str());
  } else {
    Serial.println("\nWiFi connection FAILED! Restarting...");
    ESP.restart();
  }
}

// ============================================
// KONEKSI MQTT
// ============================================
void connectMQTT() {
  String clientId = "esp32-" + ESP32_UID;

  while (!mqttClient.connected()) {
    Serial.printf("Connecting to MQTT broker %s:%d...\n", MQTT_BROKER, MQTT_PORT);

    bool connected;
    if (strlen(MQTT_USERNAME) > 0) {
      connected = mqttClient.connect(clientId.c_str(), MQTT_USERNAME, MQTT_PASSWORD);
    } else {
      connected = mqttClient.connect(clientId.c_str());
    }

    if (connected) {
      Serial.println("MQTT connected!");

      // Subscribe ke topik perintah
      mqttClient.subscribe("swiftlead/cmd/#", 1);
      Serial.println("Subscribed to: swiftlead/cmd/#");
    } else {
      Serial.printf("MQTT connection failed, rc=%d. Retrying in 5s...\n", mqttClient.state());
      delay(5000);
    }
  }
}

// ============================================
// KIRIM DATA SENSOR (TELEMETRI)
// ============================================
void sendTelemetry() {
  // Baca sensor (GANTI dengan pembacaan sensor asli)
  float temperature = readTemperature();
  float humidity    = readHumidity();
  float ammonia     = readAmmonia();
  int   rssi        = WiFi.RSSI();

  packetSeq++;

  // Buat JSON payload
  StaticJsonDocument<256> doc;
  doc["esp32_uid"] = ESP32_UID;
  doc["temp"]      = temperature;
  doc["rh"]        = humidity;
  doc["nh3"]       = ammonia;
  doc["rssi"]      = rssi;
  doc["timestamp"] = (unsigned long)(millis() / 1000);  // Atau gunakan NTP
  doc["seq"]       = packetSeq;

  char payload[256];
  serializeJson(doc, payload, sizeof(payload));

  // Publish ke broker
  if (mqttClient.publish(TOPIC_TELEMETRY.c_str(), payload, false)) {
    Serial.printf("[TX] seq=%lu temp=%.1f rh=%.1f nh3=%.1f rssi=%d\n",
                  packetSeq, temperature, humidity, ammonia, rssi);
  } else {
    Serial.println("[TX] FAILED to publish!");
  }
}

// ============================================
// CALLBACK PERINTAH DARI SERVER
// ============================================
void onMqttMessage(char* topic, byte* payload, unsigned int length) {
  // Parse payload
  char message[length + 1];
  memcpy(message, payload, length);
  message[length] = '\0';

  Serial.printf("[RX] Topic: %s | Payload: %s\n", topic, message);

  StaticJsonDocument<128> doc;
  DeserializationError error = deserializeJson(doc, message);
  if (error) {
    Serial.printf("[RX] JSON parse error: %s\n", error.c_str());
    return;
  }

  const char* action = doc["action"];
  int value = doc["value"];

  // Handle perintah audio
  if (strcmp(topic, TOPIC_CMD_AUDIO) == 0) {
    if (strcmp(action, "audio_set_lmb") == 0) {
      digitalWrite(PIN_AUDIO_LMB, value ? HIGH : LOW);
      Serial.printf("[CMD] Audio LMB: %s\n", value ? "ON" : "OFF");
    }
    else if (strcmp(action, "audio_set_nest") == 0) {
      digitalWrite(PIN_AUDIO_NEST, value ? HIGH : LOW);
      Serial.printf("[CMD] Audio Nest: %s\n", value ? "ON" : "OFF");
    }
    else if (strcmp(action, "call_bird") == 0) {
      digitalWrite(PIN_AUDIO_LMB, value ? HIGH : LOW);
      digitalWrite(PIN_AUDIO_NEST, value ? HIGH : LOW);
      Serial.printf("[CMD] Call Bird: %s\n", value ? "ON" : "OFF");
    }
  }
  // Handle perintah pompa
  else if (strcmp(topic, TOPIC_CMD_PUMP) == 0) {
    if (strcmp(action, "sprayer_set") == 0) {
      digitalWrite(PIN_PUMP, value ? HIGH : LOW);
      Serial.printf("[CMD] Pump: %s\n", value ? "ON" : "OFF");
    }
  }
}

// ============================================
// FUNGSI BACA SENSOR (PLACEHOLDER)
// Ganti dengan implementasi sensor asli
// ============================================
float readTemperature() {
  // TODO: Implementasi pembacaan sensor suhu (DHT22, DS18B20, dll)
  return 28.0 + random(-10, 10) / 10.0;
}

float readHumidity() {
  // TODO: Implementasi pembacaan sensor kelembaban (DHT22, SHT31, dll)
  return 70.0 + random(-20, 20) / 10.0;
}

float readAmmonia() {
  // TODO: Implementasi pembacaan sensor amonia (MQ-137, dll)
  return 10.0 + random(-20, 20) / 10.0;
}
```

---

## 8. Testing & Debugging

### Menggunakan `mosquitto_pub` (CLI)

Kirim data sensor test dari terminal:

```bash
# Kirim data telemetri
mosquitto_pub -h localhost -p 1883 \
  -t "swiftlead/tel/AA:BB:CC:DD:EE:FF" \
  -m '{"esp32_uid":"AA:BB:CC:DD:EE:FF","temp":28.5,"rh":75.0,"nh3":15.0,"rssi":-60,"timestamp":0,"seq":1}'
```

### Monitor Semua Topik

```bash
# Subscribe ke semua topik swiftlead
mosquitto_sub -h localhost -p 1883 -t "swiftlead/#" -v
```

### Monitor Perintah dari Server

```bash
# Subscribe ke topik perintah saja
mosquitto_sub -h localhost -p 1883 -t "swiftlead/cmd/#" -v
```

### Menggunakan Python (Script Simulasi)

Tersedia script simulasi di repository:

```bash
pip install paho-mqtt
python3 scripts/simulate_node_telemetry.py --esp32-uid "AA:BB:CC:DD:EE:FF" --loop
```

---

## 9. FAQ

### Q: Bagaimana cara mendaftarkan ESP32 UID di server?

Gunakan REST API untuk membuat node baru:

```
POST /api/v1/rbw/{rbw_id}/nodes
Authorization: Bearer <token>

{
  "node_type": "gateway",
  "node_code": "GW-001",
  "esp32_uid": "AA:BB:CC:DD:EE:FF",
  "has_audio": true,
  "has_pump": true
}
```

### Q: Apakah format MAC Address harus pakai titik dua?

Server mendukung kedua format:
- `AA:BB:CC:DD:EE:FF` ✅
- `AABBCCDDEEFF` ✅

### Q: Apa yang terjadi jika ESP32 mengirim data tapi UID belum terdaftar?

Server akan mengabaikan data tersebut dan mencatat error di log. Pastikan UID sudah terdaftar terlebih dahulu.

### Q: Berapa interval minimum pengiriman data?

Tidak ada batasan minimum, tapi rekomendasi adalah **5–10 detik**. Jangan lebih dari **15 menit** agar status node tidak berubah menjadi offline.

### Q: Apakah perlu NTP untuk timestamp?

Opsional. Jika `timestamp` bernilai `0` atau tidak dikirim, server akan menggunakan waktu saat data diterima. Namun, untuk akurasi data yang lebih baik, disarankan menggunakan NTP.

### Q: Bagaimana jika koneksi MQTT terputus?

Kode contoh sudah menangani reconnect otomatis. ESP32 akan terus mencoba reconnect setiap 5 detik.

---

## Ringkasan Cepat

```
📤 KIRIM DATA SENSOR:
   Topik  : swiftlead/tel/<ESP32_UID>
   Format : {"esp32_uid":"...", "temp":28.5, "rh":75.0, "nh3":15.0, "rssi":-60, "timestamp":0, "seq":1}

📥 TERIMA PERINTAH AUDIO:
   Topik  : swiftlead/cmd/lmb/set
   Format : {"action":"audio_set_lmb|audio_set_nest|call_bird", "value":0|1}

📥 TERIMA PERINTAH POMPA:
   Topik  : swiftlead/cmd/pump/set
   Format : {"action":"sprayer_set", "value":0|1}

🔗 SUBSCRIBE: swiftlead/cmd/#
```
