# 🚀 SwiftLead Backend - Local Deployment Guide

Panduan untuk menjalankan SwiftLead Backend secara lokal menggunakan Docker.

## 📋 Prasyarat

- [Docker](https://docs.docker.com/get-docker/) (v20.10+)
- [Docker Compose](https://docs.docker.com/compose/install/) (v2.0+)
- Minimal 4GB RAM tersedia untuk Docker

## 📁 Struktur File

```
.
├── docker-compose.local.yml   # Docker Compose config
├── .env.example               # Template environment variables
├── mosquitto/
│   └── mosquitto.conf         # MQTT broker config
├── prometheus/
│   └── prometheus.yml         # Prometheus monitoring config
└── README.md                  # File ini
```

## ⚡ Quick Start

### 1. Siapkan Environment Variables

```bash
cp .env.example .env
```

Edit `.env` sesuai kebutuhan Anda. Minimal yang perlu diganti:
- `JWT_SECRET` — ganti dengan string random yang kuat
- `DB_PASSWORD` — password database
- `MINIO_ACCESS_KEY` & `MINIO_SECRET_KEY` — kredensial object storage

### 2. Jalankan Semua Services

```bash
docker compose -f docker-compose.local.yml up -d
```

### 3. Verifikasi

```bash
# Cek semua container running
docker compose -f docker-compose.local.yml ps

# Cek health endpoint
curl http://localhost:8080/health
```

## 🌐 Service Endpoints

| Service | URL | Keterangan |
|---------|-----|------------|
| **Backend API** | http://localhost:8080 | REST API utama |
| **MinIO Console** | http://localhost:9001 | Object storage dashboard |
| **MinIO API** | http://localhost:9000 | S3-compatible API |
| **Mosquitto MQTT** | tcp://localhost:1883 | MQTT broker |
| **Prometheus** | http://localhost:9090 | Metrics dashboard |
| **Grafana** | http://localhost:3030 | Monitoring dashboard |

> **Grafana login:** admin / `GRAFANA_ADMIN_PASSWORD` (default: `swiftlead_admin`)

## ⚙️ Konfigurasi

Semua konfigurasi dilakukan melalui file `.env`. Berikut kategori yang bisa dikustomisasi:

### Database (PostgreSQL + TimescaleDB)
```env
DB_USER=swiftlet
DB_PASSWORD=your_password
DB_NAME=swiftlet
DB_PORT=5432
```
> Untuk menggunakan database external, ubah `POSTGRES_DSN` ke connection string lengkap dan hapus/comment service `db` di docker-compose.

### MQTT Broker
```env
MQTT_BROKER=tcp://mosquitto:1883
MQTT_USERNAME=
MQTT_PASSWORD=
```
> Untuk menggunakan MQTT broker external, ubah `MQTT_BROKER` dan hapus service `mosquitto` di docker-compose. Untuk TLS, set `MQTT_USE_TLS=true` dan `MQTT_BROKER=ssl://mosquitto:8883`.

### MinIO Object Storage
```env
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=swiftlet
```

### Sensor Thresholds
```env
TEMP_HIGH_THRESHOLD=35.0
TEMP_LOW_THRESHOLD=20.0
HUMID_HIGH_THRESHOLD=85.0
HUMID_LOW_THRESHOLD=60.0
AMMONIA_HIGH_THRESHOLD=25.0
```

### Port Mapping
Semua port bisa diubah jika ada konflik:
```env
PORT=8080               # Backend API
DB_PORT=5432            # PostgreSQL
MINIO_PORT=9000         # MinIO API
MINIO_CONSOLE_PORT=9001 # MinIO Console
MQTT_PORT=1883          # MQTT
PROMETHEUS_PORT=9090    # Prometheus
GRAFANA_PORT=3030       # Grafana
```

## 🔧 Management

### Start / Stop / Restart
```bash
# Start
docker compose -f docker-compose.local.yml up -d

# Stop (data tetap tersimpan)
docker compose -f docker-compose.local.yml down

# Restart satu service
docker compose -f docker-compose.local.yml restart backend

# Lihat logs
docker compose -f docker-compose.local.yml logs -f backend
```

### Reset Data
```bash
# Hapus semua data (database, files, dll)
docker compose -f docker-compose.local.yml down -v
```

### Update Backend
```bash
# Pull image terbaru
docker compose -f docker-compose.local.yml pull backend

# Restart dengan image baru
docker compose -f docker-compose.local.yml up -d backend
```

## ❓ Troubleshooting

### Backend tidak bisa connect ke database
- Pastikan service `db` sudah healthy: `docker compose -f docker-compose.local.yml ps`
- Cek logs: `docker compose -f docker-compose.local.yml logs db`

### Port sudah digunakan
- Ubah port di `.env`, contoh: `DB_PORT=5433`

### Container out of memory
- Tambah resource limit Docker (Settings → Resources → Memory)

## 📖 API Documentation

Setelah backend berjalan, API documentation tersedia di endpoint `/health` dan format lengkapnya bisa dilihat di koleksi Postman yang disediakan terpisah.
