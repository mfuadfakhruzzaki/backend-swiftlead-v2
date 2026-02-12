# SwiftLead API Documentation

**Base URL:** `https://api.swiftlead.fuadfakhruz.com`  
**API Version:** `v1`  
**Prefix:** `/api/v1`

---

## Table of Contents

1. [Overview](#1-overview)
2. [Authentication](#2-authentication)
3. [Auth Endpoints](#3-auth-endpoints)
4. [User Endpoints](#4-user-endpoints)
5. [RBW Endpoints](#5-rbw-rumah-burung-walet-endpoints)
6. [Node Endpoints](#6-node-iot-device-endpoints)
7. [Sensor Endpoints](#7-sensor-endpoints)
8. [Alert Endpoints](#8-alert-endpoints)
9. [Harvest Endpoints](#9-harvest-endpoints)
10. [Service Request Endpoints](#10-service-request-endpoints)
11. [Transaction Endpoints](#11-transaction-endpoints)
12. [Transaction Category Endpoints](#12-transaction-category-endpoints)
13. [Financial Statement Endpoints](#13-financial-statement-endpoints)
14. [AI Engine Endpoints](#14-ai-engine-endpoints)
15. [Upload Endpoints](#15-upload-endpoints)
16. [WebSocket Endpoints](#16-websocket-endpoints)
17. [Health & Metrics](#17-health--metrics)

---

## 1. Overview

### Standard Response Format

All API responses follow this structure:

```json
{
  "success": true,
  "message": "Optional message",
  "data": {},
  "meta": {
    "total": 100,
    "page": 1,
    "limit": 20,
    "total_pages": 5
  }
}
```

### Error Response Format

```json
{
  "success": false,
  "message": "Error description",
  "error": "error_code"
}
```

### Error Codes

| HTTP Status | Error Code           | Description                                  |
| ----------- | -------------------- | -------------------------------------------- |
| 400         | `bad_request`        | Invalid request body or parameters           |
| 401         | `unauthorized`       | Missing or invalid authentication            |
| 403         | `forbidden`          | Insufficient permissions                     |
| 404         | `not_found`          | Resource not found                           |
| 405         | `method_not_allowed` | HTTP method not allowed                      |
| 409         | `conflict`           | Resource already exists                      |
| 422         | `validation_error`   | Validation failed (includes `details` field) |
| 500         | `internal_error`     | Internal server error                        |
| 503         | `ai_disabled`        | AI Engine is disabled                        |

### Pagination

Endpoints that return lists support pagination:

| Query Param | Type | Default | Description    |
| ----------- | ---- | ------- | -------------- |
| `page`      | int  | 1       | Page number    |
| `limit`     | int  | 20      | Items per page |

### User Roles

| Role         | Description                                  |
| ------------ | -------------------------------------------- |
| `admin`      | Full access, can manage users and categories |
| `technician` | Technical operations                         |
| `farmer`     | Default role for public registration         |

### Date Format

All dates use **ISO 8601** format: `2026-02-09T08:00:00Z`

---

## 2. Authentication

All protected endpoints require a JWT Bearer token in the `Authorization` header:

```
Authorization: Bearer <token>
```

Token is obtained from the login endpoint. Include this header in all requests to protected endpoints.

---

## 3. Auth Endpoints

### POST `/api/v1/auth/register` — Public Registration

Register a new user (defaults to `farmer` role).

**Auth:** None

**Request Body:**

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "SecurePass123!@",
  "phone": "081234567890"
}
```

| Field      | Type   | Required | Validation                                                       |
| ---------- | ------ | -------- | ---------------------------------------------------------------- |
| `name`     | string | ✅       | 2-100 characters                                                 |
| `email`    | string | ✅       | Valid email                                                      |
| `password` | string | ✅       | Min 8 chars, must have uppercase, lowercase, digit, special char |
| `phone`    | string | ❌       | -                                                                |

**Response:** `201 Created`

```json
{
  "success": true,
  "message": "Registration successful",
  "data": {
    "id": "uuid",
    "name": "John Doe",
    "email": "john@example.com",
    "role": "farmer",
    "phone": "081234567890",
    "created_at": "2026-02-09T10:00:00Z",
    "updated_at": "2026-02-09T10:00:00Z"
  }
}
```

**Errors:**

- `400` — Weak password or invalid body
- `409` — Email already registered

---

### POST `/api/v1/auth/login` — Login

**Auth:** None

**Request Body:**

```json
{
  "email": "john@example.com",
  "password": "SecurePass123!@"
}
```

| Field      | Type   | Required |
| ---------- | ------ | -------- |
| `email`    | string | ✅       |
| `password` | string | ✅       |

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "uuid",
      "name": "John Doe",
      "email": "john@example.com",
      "role": "farmer",
      "phone": "081234567890",
      "avatar_url": null,
      "created_at": "2026-02-09T10:00:00Z",
      "updated_at": "2026-02-09T10:00:00Z"
    }
  }
}
```

**Errors:**

- `400` — Invalid request body
- `401` — Invalid email or password

---

### POST `/api/v1/auth/change-password` — Change Password

**Auth:** 🔒 Bearer Token

**Request Body:**

```json
{
  "old_password": "OldPass123!@",
  "new_password": "NewPass456!@"
}
```

| Field          | Type   | Required | Validation                                           |
| -------------- | ------ | -------- | ---------------------------------------------------- |
| `old_password` | string | ✅       | -                                                    |
| `new_password` | string | ✅       | Min 8 chars, uppercase + lowercase + digit + special |

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "Password changed successfully"
}
```

**Errors:**

- `400` — Same password / weak password
- `401` — Old password incorrect

---

### POST `/api/v1/auth/forgot-password` — Admin Reset Password

Reset a user's password (generates temporary password).

**Auth:** 🔒 Bearer Token (Admin only)

**Request Body:**

```json
{
  "email": "user@example.com"
}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "Password has been reset",
  "data": {
    "temporary_password": "aB3$xYz9!qW2",
    "message": "Please change the password after login"
  }
}
```

**Errors:**

- `404` — User not found

---

## 4. User Endpoints

### GET `/api/v1/users/me` — Get My Profile

**Auth:** 🔒 Bearer Token

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "John Doe",
    "email": "john@example.com",
    "role": "farmer",
    "phone": "081234567890",
    "avatar_url": "https://...",
    "created_at": "2026-02-09T10:00:00Z",
    "updated_at": "2026-02-09T10:00:00Z"
  }
}
```

---

### PATCH `/api/v1/users/me` — Update My Profile

**Auth:** 🔒 Bearer Token

**Request Body** (all optional):

```json
{
  "name": "Updated Name",
  "phone": "089876543210",
  "avatar_url": "https://example.com/avatar.jpg"
}
```

| Field        | Type   | Required | Validation        |
| ------------ | ------ | -------- | ----------------- |
| `name`       | string | ❌       | 2-100 characters  |
| `phone`      | string | ❌       | -                 |
| `avatar_url` | string | ❌       | Must be valid URL |

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "Profile updated",
  "data": { "...user object..." }
}
```

---

### GET `/api/v1/users` — List Users (Admin)

**Auth:** 🔒 Bearer Token (Admin only)

**Query Parameters:**

| Param   | Type   | Default | Description                                     |
| ------- | ------ | ------- | ----------------------------------------------- |
| `page`  | int    | 1       | Page number                                     |
| `limit` | int    | 20      | Items per page                                  |
| `role`  | string | -       | Filter by role: `admin`, `technician`, `farmer` |

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "John Doe",
      "email": "john@example.com",
      "role": "farmer",
      "phone": "081234567890",
      "avatar_url": null,
      "created_at": "2026-02-09T10:00:00Z",
      "updated_at": "2026-02-09T10:00:00Z"
    }
  ],
  "meta": {
    "total": 50,
    "page": 1,
    "limit": 20,
    "total_pages": 3
  }
}
```

---

### POST `/api/v1/users` — Create User (Admin)

**Auth:** 🔒 Bearer Token (Admin only)

**Request Body:**

```json
{
  "name": "New Technician",
  "email": "tech@example.com",
  "password": "TechPass123!@",
  "role": "technician",
  "phone": "081234567890"
}
```

| Field      | Type   | Required | Validation                         |
| ---------- | ------ | -------- | ---------------------------------- |
| `name`     | string | ✅       | 2-100 characters                   |
| `email`    | string | ✅       | Valid email                        |
| `password` | string | ✅       | Min 8 characters                   |
| `role`     | string | ✅       | `admin`, `technician`, or `farmer` |
| `phone`    | string | ❌       | -                                  |

**Response:** `201 Created`

---

## 5. RBW (Rumah Burung Walet) Endpoints

### POST `/api/v1/rbw` — Create RBW

**Auth:** 🔒 Bearer Token

**Request Body:**

```json
{
  "code": "RBW-001",
  "name": "Rumah Walet Jakarta",
  "address": "Jl. Merdeka No. 1",
  "latitude": -6.2088,
  "longitude": 106.8456,
  "total_floors": 3,
  "description": "Walet house description"
}
```

| Field          | Type   | Required | Description      |
| -------------- | ------ | -------- | ---------------- |
| `code`         | string | ✅       | Unique RBW code  |
| `name`         | string | ✅       | RBW name         |
| `address`      | string | ❌       | Physical address |
| `latitude`     | float  | ❌       | GPS latitude     |
| `longitude`    | float  | ❌       | GPS longitude    |
| `total_floors` | int    | ✅       | Number of floors |
| `description`  | string | ❌       | Description      |

**Response:** `201 Created`

```json
{
  "success": true,
  "message": "RBW created",
  "data": {
    "id": "uuid",
    "owner_id": "uuid",
    "code": "RBW-001",
    "name": "Rumah Walet Jakarta",
    "address": "Jl. Merdeka No. 1",
    "latitude": -6.2088,
    "longitude": 106.8456,
    "total_floors": 3,
    "description": "Walet house description",
    "photo_url": null,
    "created_at": "2026-02-09T10:00:00Z",
    "updated_at": "2026-02-09T10:00:00Z"
  }
}
```

---

### GET `/api/v1/rbw` — List RBWs

**Auth:** 🔒 Bearer Token

**Query:** `?page=1&limit=20`

**Response:** `200 OK` — Array of RBW objects with pagination meta.

---

### GET `/api/v1/rbw/{rbw_id}` — Get RBW

**Auth:** 🔒 Bearer Token

**Response:** `200 OK` — Single RBW object.

**Errors:** `404` — RBW not found.

---

### PATCH `/api/v1/rbw/{rbw_id}` — Update RBW

**Auth:** 🔒 Bearer Token

**Request Body** (all optional):

```json
{
  "name": "Updated Name",
  "address": "New Address",
  "latitude": -6.21,
  "longitude": 106.85,
  "total_floors": 4,
  "description": "Updated description",
  "photo_url": "https://..."
}
```

**Response:** `200 OK` — Updated RBW object.

---

### DELETE `/api/v1/rbw/{rbw_id}` — Delete RBW

**Auth:** 🔒 Bearer Token

**Response:** `200 OK`

---

## 6. Node (IoT Device) Endpoints

### POST `/api/v1/rbw/{rbw_id}/nodes` — Create Node

**Auth:** 🔒 Bearer Token

**Request Body:**

```json
{
  "node_type": "gateway",
  "node_code": "GW-001",
  "esp32_uid": "AA:BB:CC:DD:EE:FF",
  "has_audio": true,
  "has_pump": true
}
```

| Field       | Type   | Required | Values                           |
| ----------- | ------ | -------- | -------------------------------- |
| `node_type` | string | ✅       | `gateway`, `nest`, `lmb`, `pump` |
| `node_code` | string | ✅       | Unique node code                 |
| `esp32_uid` | string | ❌       | ESP32 hardware UID               |
| `has_audio` | bool   | ✅       | Has audio capability             |
| `has_pump`  | bool   | ✅       | Has pump capability              |

**Response:** `201 Created`

```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "rbw_id": "uuid",
    "node_type": "gateway",
    "node_code": "GW-001",
    "esp32_uid": "AA:BB:CC:DD:EE:FF",
    "status_node": "offline",
    "last_seen": null,
    "has_audio": true,
    "state_audio_lmb": null,
    "state_audio_nest": null,
    "has_pump": true,
    "state_pump": null,
    "installed_at": null,
    "uninstalled_at": null,
    "created_at": "2026-02-09T10:00:00Z",
    "updated_at": "2026-02-09T10:00:00Z"
  }
}
```

---

### GET `/api/v1/rbw/{rbw_id}/nodes` — List Nodes by RBW

**Auth:** 🔒 Bearer Token

**Response:** `200 OK` — Array of Node objects.

---

### GET `/api/v1/nodes/{node_id}` — Get Node

**Auth:** 🔒 Bearer Token

**Response:** `200 OK` — Single Node object.

---

### PATCH `/api/v1/nodes/{node_id}` — Update Node

**Auth:** 🔒 Bearer Token

**Request Body** (all optional):

```json
{
  "node_code": "GW-UPDATED",
  "esp32_uid": "11:22:33:44:55:66",
  "has_audio": false,
  "has_pump": false
}
```

**Response:** `200 OK` — Updated Node object.

---

### DELETE `/api/v1/nodes/{node_id}` — Delete Node

**Auth:** 🔒 Bearer Token

**Response:** `200 OK`

---

### GET `/api/v1/nodes/{node_id}/audio` — Get Audio State

**Auth:** 🔒 Bearer Token

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "has_audio": true,
    "state_audio_lmb": true,
    "state_audio_nest": false
  }
}
```

---

### PATCH `/api/v1/nodes/{node_id}/audio` — Control Audio

**Auth:** 🔒 Bearer Token

**Request Body:**

```json
{
  "action": "audio_set_lmb",
  "value": 1
}
```

| Field    | Type   | Values                                         |
| -------- | ------ | ---------------------------------------------- |
| `action` | string | `audio_set_lmb`, `audio_set_nest`, `call_bird` |
| `value`  | int    | `0` (off) or `1` (on)                          |

**Response:** `200 OK`

---

### PATCH `/api/v1/nodes/{node_id}/pump` — Control Pump

**Auth:** 🔒 Bearer Token

**Request Body:**

```json
{
  "action": "sprayer_set",
  "value": 1
}
```

| Field    | Type   | Values                |
| -------- | ------ | --------------------- |
| `action` | string | `sprayer_set`         |
| `value`  | int    | `0` (off) or `1` (on) |

**Response:** `200 OK`

---

## 7. Sensor Endpoints

### POST `/api/v1/nodes/{node_id}/sensors` — Create Sensor

**Auth:** 🔒 Bearer Token

**Request Body:**

```json
{
  "sensor_type": "temp",
  "sensor_name": "Temperature Sensor 1",
  "unit": "°C"
}
```

| Field         | Type   | Required | Values                     |
| ------------- | ------ | -------- | -------------------------- |
| `sensor_type` | string | ✅       | `temp`, `humid`, `ammonia` |
| `sensor_name` | string | ❌       | Display name               |
| `unit`        | string | ❌       | Unit of measurement        |

**Response:** `201 Created`

```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "node_id": "uuid",
    "sensor_type": "temp",
    "sensor_name": "Temperature Sensor 1",
    "unit": "°C",
    "is_active": true,
    "created_at": "2026-02-09T10:00:00Z",
    "updated_at": "2026-02-09T10:00:00Z"
  }
}
```

---

### GET `/api/v1/nodes/{node_id}/sensors` — List Sensors by Node

**Auth:** 🔒 Bearer Token

**Response:** `200 OK` — Array of Sensor objects.

---

### GET `/api/v1/sensors/{sensor_id}` — Get Sensor

**Auth:** 🔒 Bearer Token

**Response:** `200 OK` — Single Sensor object.

---

### PATCH `/api/v1/sensors/{sensor_id}` — Update Sensor

**Auth:** 🔒 Bearer Token

**Request Body** (all optional):

```json
{
  "sensor_name": "Updated Sensor Name",
  "unit": "°F",
  "is_active": false
}
```

**Response:** `200 OK` — Updated Sensor object.

---

### POST `/api/v1/sensors/{sensor_id}/readings` — Create Sensor Reading

**Auth:** 🔒 Bearer Token

**Request Body:**

```json
{
  "value": 28.5,
  "recorded_at": "2026-02-09T10:00:00Z"
}
```

| Field         | Type     | Required | Description                 |
| ------------- | -------- | -------- | --------------------------- |
| `value`       | float    | ✅       | Sensor reading value        |
| `recorded_at` | datetime | ❌       | Timestamp (defaults to now) |

**Response:** `201 Created`

```json
{
  "success": true,
  "data": {
    "id": 1,
    "sensor_id": "uuid",
    "recorded_at": "2026-02-09T10:00:00Z",
    "value": 28.5,
    "is_anomaly": false,
    "created_at": "2026-02-09T10:00:00Z"
  }
}
```

---

### GET `/api/v1/sensors/{sensor_id}/readings` — Get Sensor Readings

**Auth:** 🔒 Bearer Token

**Query Parameters:**

| Param   | Type     | Default | Description       |
| ------- | -------- | ------- | ----------------- |
| `from`  | datetime | -       | Start time filter |
| `to`    | datetime | -       | End time filter   |
| `limit` | int      | 100     | Max readings      |

**Response:** `200 OK` — Array of SensorReading objects.

---

### GET `/api/v1/sensors/{sensor_id}/trend` — Get Sensor Trend

**Auth:** 🔒 Bearer Token

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "sensor_id": "uuid",
    "sensor_type": "temp",
    "direction": "rising",
    "slope": 0.25,
    "avg_value": 28.5,
    "min_value": 27.0,
    "max_value": 30.0,
    "data_points": 50,
    "period": "24h0m0s"
  }
}
```

| Field         | Type   | Description                   |
| ------------- | ------ | ----------------------------- |
| `direction`   | string | `rising`, `falling`, `stable` |
| `slope`       | float  | Rate of change                |
| `avg_value`   | float  | Average value in period       |
| `min_value`   | float  | Minimum value in period       |
| `max_value`   | float  | Maximum value in period       |
| `data_points` | int    | Number of readings analyzed   |

---

## 8. Alert Endpoints

### GET `/api/v1/alerts` — List All Alerts

**Auth:** 🔒 Bearer Token

**Query Parameters:**

| Param      | Type   | Description              |
| ---------- | ------ | ------------------------ |
| `page`     | int    | Page number              |
| `limit`    | int    | Items per page           |
| `rbw_id`   | string | Filter by RBW            |
| `severity` | int    | Filter by severity level |
| `is_read`  | bool   | Filter by read status    |

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "rbw_id": "uuid",
      "node_id": "uuid",
      "sensor_id": "uuid",
      "alert_type": "temp_high",
      "severity": 3,
      "message": "Temperature exceeds 35°C",
      "is_read": false,
      "resolved_at": null,
      "resolved_by": null,
      "created_at": "2026-02-09T10:00:00Z"
    }
  ],
  "meta": { "..." }
}
```

### Alert Types

| Type           | Description            |
| -------------- | ---------------------- |
| `temp_high`    | Temperature too high   |
| `temp_low`     | Temperature too low    |
| `humid_high`   | Humidity too high      |
| `humid_low`    | Humidity too low       |
| `ammonia_high` | Ammonia level too high |
| `node_offline` | Node went offline      |
| `ai_anomaly`   | AI detected anomaly    |

---

### GET `/api/v1/rbw/{rbw_id}/alerts` — List Alerts by RBW

**Auth:** 🔒 Bearer Token

**Response:** `200 OK` — Array of Alert objects for specific RBW.

---

### PATCH `/api/v1/alerts/{alert_id}/read` — Mark Alert as Read

**Auth:** 🔒 Bearer Token

**Response:** `200 OK`

**Errors:** `404` — Alert not found.

---

### PATCH `/api/v1/alerts/{alert_id}/resolve` — Resolve Alert

**Auth:** 🔒 Bearer Token

**Response:** `200 OK`

**Errors:** `404` — Alert not found.

---

## 9. Harvest Endpoints

### POST `/api/v1/harvests` — Create Harvest

**Auth:** 🔒 Bearer Token

**Request Body:**

```json
{
  "rbw_id": "uuid",
  "node_id": "uuid",
  "floor_no": 1,
  "harvested_at": "2026-02-09T08:00:00Z",
  "nests_count": 50,
  "weight_kg": 2.5,
  "grade": "good",
  "notes": "Test harvest"
}
```

| Field          | Type     | Required | Validation               |
| -------------- | -------- | -------- | ------------------------ |
| `rbw_id`       | string   | ✅       | Valid RBW UUID           |
| `node_id`      | string   | ❌       | Valid Node UUID          |
| `floor_no`     | int      | ✅       | Min 1                    |
| `harvested_at` | datetime | ✅       | ISO 8601                 |
| `nests_count`  | int      | ❌       | Min 0                    |
| `weight_kg`    | float    | ❌       | Weight in kg             |
| `grade`        | string   | ❌       | `good`, `medium`, `poor` |
| `notes`        | string   | ❌       | -                        |

**Response:** `201 Created`

```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "rbw_id": "uuid",
    "node_id": null,
    "floor_no": 1,
    "harvested_at": "2026-02-09T08:00:00Z",
    "nests_count": 50,
    "weight_kg": 2.5,
    "grade": "good",
    "notes": "Test harvest",
    "created_by": "uuid",
    "cycle_days": 45,
    "created_at": "2026-02-09T10:00:00Z",
    "updated_at": "2026-02-09T10:00:00Z"
  }
}
```

---

### GET `/api/v1/harvests` — List All Harvests

**Auth:** 🔒 Bearer Token

**Query:** `?page=1&limit=20&rbw_id=uuid`

**Response:** `200 OK` — Array of Harvest objects with pagination meta.

---

### GET `/api/v1/harvests/{id}` — Get Harvest

**Auth:** 🔒 Bearer Token

**Response:** `200 OK` — Single Harvest object.

---

### PATCH `/api/v1/harvests/{id}` — Update Harvest

**Auth:** 🔒 Bearer Token

**Request Body** (all optional):

```json
{
  "floor_no": 2,
  "nests_count": 55,
  "weight_kg": 2.8,
  "grade": "good",
  "notes": "Updated notes"
}
```

**Response:** `200 OK` — Updated Harvest object.

---

### DELETE `/api/v1/harvests/{id}` — Delete Harvest

**Auth:** 🔒 Bearer Token

**Response:** `200 OK`

---

### GET `/api/v1/harvests/stats` — Get Harvest Statistics

**Auth:** 🔒 Bearer Token

**Query:** `?rbw_id=uuid` (optional filter)

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "total_harvests": 25,
    "total_nests": 1250,
    "total_weight_kg": 62.5,
    "avg_nests_per_harvest": 50.0,
    "avg_weight_kg": 2.5,
    "avg_cycle_days": 42.0
  }
}
```

---

### GET `/api/v1/rbw/{rbw_id}/harvests` — List Harvests by RBW

**Auth:** 🔒 Bearer Token

**Query:** `?page=1&limit=20`

**Response:** `200 OK` — Array of Harvest objects for specific RBW.

---

## 10. Service Request Endpoints

### POST `/api/v1/service-requests` — Create Service Request

**Auth:** 🔒 Bearer Token

**Request Body:**

```json
{
  "rbw_id": "uuid",
  "node_id": "uuid",
  "type": "maintenance",
  "issue": "Sensor malfunction on floor 2"
}
```

| Field     | Type   | Required | Values                                     |
| --------- | ------ | -------- | ------------------------------------------ |
| `rbw_id`  | string | ✅       | Valid RBW UUID                             |
| `node_id` | string | ❌       | Specific node                              |
| `type`    | string | ✅       | `installation`, `maintenance`, `uninstall` |
| `issue`   | string | ❌       | Issue description                          |

**Response:** `201 Created`

```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "rbw_id": "uuid",
    "node_id": null,
    "request_by": "uuid",
    "assigned_to": null,
    "approved_by": null,
    "type": "maintenance",
    "status": "draft",
    "request_date": "2026-02-09T10:00:00Z",
    "schedule_date": null,
    "uninstall_date": null,
    "issue": "Sensor malfunction on floor 2",
    "resolution": null,
    "notes": null,
    "created_at": "2026-02-09T10:00:00Z",
    "updated_at": "2026-02-09T10:00:00Z"
  }
}
```

### Service Request Status Flow

```
draft → pending → approved → assigned → in_progress → resolved
                → rejected
                                                      → cancelled
```

| Status        | Description            |
| ------------- | ---------------------- |
| `draft`       | Initial state          |
| `pending`     | Awaiting approval      |
| `approved`    | Approved by admin      |
| `rejected`    | Rejected               |
| `assigned`    | Assigned to technician |
| `in_progress` | Work in progress       |
| `resolved`    | Completed              |
| `cancelled`   | Cancelled              |

---

### GET `/api/v1/service-requests` — List Service Requests

**Auth:** 🔒 Bearer Token

**Query Parameters:**

| Param         | Type   | Description         |
| ------------- | ------ | ------------------- |
| `page`        | int    | Page number         |
| `limit`       | int    | Items per page      |
| `rbw_id`      | string | Filter by RBW       |
| `status`      | string | Filter by status    |
| `request_by`  | string | Filter by requester |
| `assigned_to` | string | Filter by assignee  |

**Response:** `200 OK` — Array of ServiceRequest objects with pagination meta.

---

### GET `/api/v1/service-requests/{id}` — Get Service Request

**Auth:** 🔒 Bearer Token

**Response:** `200 OK` — Single ServiceRequest object.

---

### PATCH `/api/v1/service-requests/{id}` — Update Service Request

**Auth:** 🔒 Bearer Token

**Request Body** (all optional):

```json
{
  "status": "pending",
  "assigned_to": "technician-uuid",
  "schedule_date": "2026-02-15T09:00:00Z",
  "resolution": "Sensor replaced",
  "notes": "Escalated for review"
}
```

**Response:** `200 OK` — Updated ServiceRequest object.

---

## 11. Transaction Endpoints

### POST `/api/v1/transactions` — Create Transaction

**Auth:** 🔒 Bearer Token

**Request Body:**

```json
{
  "rbw_id": "uuid",
  "category_id": "uuid",
  "amount": 500000,
  "type": "income",
  "description": "Nest sale batch #12",
  "transaction_date": "2026-02-09T00:00:00Z"
}
```

| Field              | Type     | Required | Values              |
| ------------------ | -------- | -------- | ------------------- |
| `rbw_id`           | string   | ✅       | Valid RBW UUID      |
| `category_id`      | string   | ✅       | Valid Category UUID |
| `amount`           | float    | ✅       | Transaction amount  |
| `type`             | string   | ✅       | `income`, `expense` |
| `description`      | string   | ❌       | Description         |
| `transaction_date` | datetime | ✅       | ISO 8601            |

**Response:** `201 Created`

```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "rbw_id": "uuid",
    "category_id": "uuid",
    "amount": 500000,
    "type": "income",
    "description": "Nest sale batch #12",
    "transaction_date": "2026-02-09T00:00:00Z",
    "created_by": "uuid",
    "created_at": "2026-02-09T10:00:00Z",
    "updated_at": "2026-02-09T10:00:00Z"
  }
}
```

---

### GET `/api/v1/transactions/{id}` — Get Transaction

**Auth:** 🔒 Bearer Token

**Response:** `200 OK` — Single Transaction object.

---

### PATCH `/api/v1/transactions/{id}` — Update Transaction

**Auth:** 🔒 Bearer Token

**Request Body** (all optional):

```json
{
  "category_id": "uuid",
  "amount": 600000,
  "description": "Updated description"
}
```

**Response:** `200 OK` — Updated Transaction object.

---

### DELETE `/api/v1/transactions/{id}` — Delete Transaction

**Auth:** 🔒 Bearer Token

**Response:** `200 OK`

---

### GET `/api/v1/rbw/{rbw_id}/transactions` — List Transactions by RBW

**Auth:** 🔒 Bearer Token

**Query Parameters:**

| Param        | Type   | Description                 |
| ------------ | ------ | --------------------------- |
| `page`       | int    | Page number                 |
| `limit`      | int    | Items per page              |
| `type`       | string | Filter: `income`, `expense` |
| `start_date` | string | Start date filter           |
| `end_date`   | string | End date filter             |

**Response:** `200 OK` — Array of Transaction objects with pagination meta.

---

## 12. Transaction Category Endpoints

### GET `/api/v1/transaction-categories` — List Categories

**Auth:** 🔒 Bearer Token

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Penjualan Sarang",
      "type": "income",
      "description": "Revenue from nest sales",
      "created_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

---

### POST `/api/v1/transaction-categories` — Create Category (Admin)

**Auth:** 🔒 Bearer Token (Admin only)

**Request Body:**

```json
{
  "name": "Penjualan Sarang",
  "type": "income",
  "description": "Revenue from nest sales"
}
```

| Field         | Type   | Required | Values              |
| ------------- | ------ | -------- | ------------------- |
| `name`        | string | ✅       | Category name       |
| `type`        | string | ✅       | `income`, `expense` |
| `description` | string | ❌       | Description         |

**Response:** `201 Created`

---

### PATCH `/api/v1/transaction-categories/{id}` — Update Category (Admin)

**Auth:** 🔒 Bearer Token (Admin only)

**Request Body** (all optional):

```json
{
  "name": "Updated Category Name",
  "description": "Updated description"
}
```

**Response:** `200 OK`

---

### DELETE `/api/v1/transaction-categories/{id}` — Delete Category (Admin)

**Auth:** 🔒 Bearer Token (Admin only)

**Response:** `200 OK`

---

## 13. Financial Statement Endpoints

### POST `/api/v1/financial-statements` — Generate Financial Statement

**Auth:** 🔒 Bearer Token

**Request Body:**

```json
{
  "rbw_id": "uuid",
  "start_date": "2026-01-01",
  "end_date": "2026-12-31"
}
```

| Field        | Type   | Required | Format       |
| ------------ | ------ | -------- | ------------ |
| `rbw_id`     | string | ✅       | UUID         |
| `start_date` | string | ✅       | `YYYY-MM-DD` |
| `end_date`   | string | ✅       | `YYYY-MM-DD` |

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "rbw_id": "uuid",
    "start_date": "2026-01-01T00:00:00Z",
    "end_date": "2026-12-31T00:00:00Z",
    "total_income": 15000000,
    "total_expense": 5000000,
    "balance": 10000000,
    "incomes": [
      {
        "id": "uuid",
        "rbw_id": "uuid",
        "category_id": "uuid",
        "amount": 500000,
        "type": "income",
        "description": "Nest sale",
        "transaction_date": "2026-02-09T00:00:00Z",
        "created_by": "uuid",
        "created_at": "2026-02-09T10:00:00Z",
        "updated_at": "2026-02-09T10:00:00Z"
      }
    ],
    "expenses": []
  }
}
```

---

## 14. AI Engine Endpoints

> **Note:** AI Engine may be disabled in production. When disabled, all prediction endpoints return `503 Service Unavailable`.

### GET `/api/v1/ai/health` — AI Health Check

**Auth:** 🔒 Bearer Token

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "version": "1.0.0"
  }
}
```

---

### POST `/api/v1/ai/predict-grade` — Predict Harvest Grade

**Auth:** 🔒 Bearer Token

**Request Body:**

```json
{
  "temperature": 28.5,
  "humidity": 75.0,
  "ammonia": 15.0,
  "rbw_id": "uuid",
  "node_id": "uuid"
}
```

| Field         | Type   | Required | Description         |
| ------------- | ------ | -------- | ------------------- |
| `temperature` | float  | ✅       | Temperature (°C)    |
| `humidity`    | float  | ✅       | Humidity (%)        |
| `ammonia`     | float  | ✅       | Ammonia level (ppm) |
| `rbw_id`      | string | ❌       | Context RBW         |
| `node_id`     | string | ❌       | Context Node        |

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "grade": "good",
    "confidence": 0.92,
    "probabilities": {
      "good": 0.92,
      "medium": 0.06,
      "poor": 0.02
    }
  }
}
```

---

### POST `/api/v1/ai/predict-pump` — Predict Pump Action

**Auth:** 🔒 Bearer Token

**Request Body:**

```json
{
  "temperature": 32.0,
  "humidity": 65.0,
  "ammonia": 20.0,
  "rbw_id": "uuid",
  "node_id": "uuid"
}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "pump_state": "on",
    "confidence": 0.88,
    "duration_minutes": 15.0
  }
}
```

---

### POST `/api/v1/ai/analyze` — Comprehensive Analysis

**Auth:** 🔒 Bearer Token

**Request Body:**

```json
{
  "temperature": 28.5,
  "humidity": 75.0,
  "ammonia": 15.0,
  "rbw_id": "uuid",
  "node_id": "uuid"
}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "overall_health_score": 85.5,
    "sensors": {
      "temperature": {
        "value": 28.5,
        "unit": "°C",
        "status": "normal",
        "health_score": 90.0
      },
      "humidity": {
        "value": 75.0,
        "unit": "%",
        "status": "normal",
        "health_score": 85.0
      },
      "ammonia": {
        "value": 15.0,
        "unit": "ppm",
        "status": "warning",
        "health_score": 70.0
      }
    },
    "grade_prediction": {
      "grade": "good",
      "confidence": 0.92,
      "probabilities": { "good": 0.92, "medium": 0.06, "poor": 0.02 }
    },
    "pump_recommendation": {
      "pump_state": "off",
      "confidence": 0.95,
      "duration_minutes": 0
    },
    "recommendations": [
      {
        "priority": "medium",
        "type": "ventilation",
        "message": "Consider increasing ventilation to reduce ammonia levels"
      }
    ]
  }
}
```

---

### POST `/api/v1/ai/anomaly-detect` — Anomaly Detection

**Auth:** 🔒 Bearer Token

**Request Body:**

```json
{
  "sensor_id": "uuid",
  "sensor_type": "temp",
  "rbw_id": "uuid",
  "node_id": "uuid",
  "recorded_at": "2026-02-09T10:00:00Z",
  "value": 45.0
}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "is_anomaly": true,
    "score": 0.95,
    "reason": "Temperature abnormally high"
  }
}
```

---

## 15. Upload Endpoints

### POST `/api/v1/uploads/avatar` — Upload Avatar

**Auth:** 🔒 Bearer Token

**Content-Type:** `multipart/form-data`

| Field  | Type | Description            |
| ------ | ---- | ---------------------- |
| `file` | file | Image file (JPEG, PNG) |

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "url": "https://storage.example.com/avatars/uuid.jpg"
  }
}
```

**Errors:** `503` — Storage not available

---

### POST `/api/v1/uploads/rbw/{rbw_id}/photo` — Upload RBW Photo

**Auth:** 🔒 Bearer Token

**Content-Type:** `multipart/form-data`

| Field  | Type | Description            |
| ------ | ---- | ---------------------- |
| `file` | file | Image file (JPEG, PNG) |

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "url": "https://storage.example.com/rbw/uuid.jpg"
  }
}
```

**Errors:** `503` — Storage not available

---

## 16. WebSocket Endpoints

### GET `/api/v1/ws` — WebSocket Connection

**Auth:** 🔒 Bearer Token (via query param `?token=xxx`)

Connects to real-time data stream for live sensor data, alerts, and node status updates.

**Connection:**

```
wss://api.swiftlead.fuadfakhruz.com/api/v1/ws?token=<jwt_token>
```

---

### GET `/api/v1/ws/stats` — WebSocket Stats

**Auth:** 🔒 Bearer Token

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "connected_clients": 5,
    "total_messages": 1234
  }
}
```

---

## 17. Health & Metrics

### GET `/health` — Health Check

**Auth:** None

**Response:** `200 OK` — Plain text `OK`

---

### GET `/metrics` — Prometheus Metrics

**Auth:** None

**Response:** `200 OK` — Prometheus-formatted metrics.

---

## Quick Reference — All Endpoints

| Method   | Endpoint                               | Auth | Role  | Description              |
| -------- | -------------------------------------- | ---- | ----- | ------------------------ |
| `POST`   | `/api/v1/auth/register`                | ❌   | -     | Public registration      |
| `POST`   | `/api/v1/auth/login`                   | ❌   | -     | Login                    |
| `POST`   | `/api/v1/auth/change-password`         | 🔒   | Any   | Change password          |
| `POST`   | `/api/v1/auth/forgot-password`         | 🔒   | Admin | Reset user password      |
| `POST`   | `/api/v1/auth/admin/register`          | 🔒   | Admin | Admin register user      |
| `GET`    | `/api/v1/users/me`                     | 🔒   | Any   | Get my profile           |
| `PATCH`  | `/api/v1/users/me`                     | 🔒   | Any   | Update my profile        |
| `GET`    | `/api/v1/users`                        | 🔒   | Admin | List users               |
| `POST`   | `/api/v1/users`                        | 🔒   | Admin | Create user              |
| `GET`    | `/api/v1/rbw`                          | 🔒   | Any   | List RBWs                |
| `POST`   | `/api/v1/rbw`                          | 🔒   | Any   | Create RBW               |
| `GET`    | `/api/v1/rbw/{rbw_id}`                 | 🔒   | Any   | Get RBW                  |
| `PATCH`  | `/api/v1/rbw/{rbw_id}`                 | 🔒   | Any   | Update RBW               |
| `DELETE` | `/api/v1/rbw/{rbw_id}`                 | 🔒   | Any   | Delete RBW               |
| `GET`    | `/api/v1/rbw/{rbw_id}/nodes`           | 🔒   | Any   | List nodes by RBW        |
| `POST`   | `/api/v1/rbw/{rbw_id}/nodes`           | 🔒   | Any   | Create node              |
| `GET`    | `/api/v1/rbw/{rbw_id}/alerts`          | 🔒   | Any   | List alerts by RBW       |
| `GET`    | `/api/v1/rbw/{rbw_id}/harvests`        | 🔒   | Any   | List harvests by RBW     |
| `GET`    | `/api/v1/rbw/{rbw_id}/transactions`    | 🔒   | Any   | List transactions by RBW |
| `GET`    | `/api/v1/nodes/{node_id}`              | 🔒   | Any   | Get node                 |
| `PATCH`  | `/api/v1/nodes/{node_id}`              | 🔒   | Any   | Update node              |
| `DELETE` | `/api/v1/nodes/{node_id}`              | 🔒   | Any   | Delete node              |
| `GET`    | `/api/v1/nodes/{node_id}/sensors`      | 🔒   | Any   | List sensors by node     |
| `POST`   | `/api/v1/nodes/{node_id}/sensors`      | 🔒   | Any   | Create sensor            |
| `GET`    | `/api/v1/nodes/{node_id}/audio`        | 🔒   | Any   | Get audio state          |
| `PATCH`  | `/api/v1/nodes/{node_id}/audio`        | 🔒   | Any   | Control audio            |
| `PATCH`  | `/api/v1/nodes/{node_id}/pump`         | 🔒   | Any   | Control pump             |
| `GET`    | `/api/v1/sensors/{sensor_id}`          | 🔒   | Any   | Get sensor               |
| `PATCH`  | `/api/v1/sensors/{sensor_id}`          | 🔒   | Any   | Update sensor            |
| `GET`    | `/api/v1/sensors/{sensor_id}/readings` | 🔒   | Any   | Get readings             |
| `POST`   | `/api/v1/sensors/{sensor_id}/readings` | 🔒   | Any   | Create reading           |
| `GET`    | `/api/v1/sensors/{sensor_id}/trend`    | 🔒   | Any   | Get trend                |
| `GET`    | `/api/v1/alerts`                       | 🔒   | Any   | List alerts              |
| `PATCH`  | `/api/v1/alerts/{alert_id}/read`       | 🔒   | Any   | Mark alert read          |
| `PATCH`  | `/api/v1/alerts/{alert_id}/resolve`    | 🔒   | Any   | Resolve alert            |
| `GET`    | `/api/v1/harvests`                     | 🔒   | Any   | List harvests            |
| `POST`   | `/api/v1/harvests`                     | 🔒   | Any   | Create harvest           |
| `GET`    | `/api/v1/harvests/stats`               | 🔒   | Any   | Get harvest stats        |
| `GET`    | `/api/v1/harvests/{id}`                | 🔒   | Any   | Get harvest              |
| `PATCH`  | `/api/v1/harvests/{id}`                | 🔒   | Any   | Update harvest           |
| `DELETE` | `/api/v1/harvests/{id}`                | 🔒   | Any   | Delete harvest           |
| `GET`    | `/api/v1/service-requests`             | 🔒   | Any   | List service requests    |
| `POST`   | `/api/v1/service-requests`             | 🔒   | Any   | Create service request   |
| `GET`    | `/api/v1/service-requests/{id}`        | 🔒   | Any   | Get service request      |
| `PATCH`  | `/api/v1/service-requests/{id}`        | 🔒   | Any   | Update service request   |
| `POST`   | `/api/v1/transactions`                 | 🔒   | Any   | Create transaction       |
| `GET`    | `/api/v1/transactions/{id}`            | 🔒   | Any   | Get transaction          |
| `PATCH`  | `/api/v1/transactions/{id}`            | 🔒   | Any   | Update transaction       |
| `DELETE` | `/api/v1/transactions/{id}`            | 🔒   | Any   | Delete transaction       |
| `GET`    | `/api/v1/transaction-categories`       | 🔒   | Any   | List categories          |
| `POST`   | `/api/v1/transaction-categories`       | 🔒   | Admin | Create category          |
| `PATCH`  | `/api/v1/transaction-categories/{id}`  | 🔒   | Admin | Update category          |
| `DELETE` | `/api/v1/transaction-categories/{id}`  | 🔒   | Admin | Delete category          |
| `POST`   | `/api/v1/financial-statements`         | 🔒   | Any   | Generate statement       |
| `GET`    | `/api/v1/ai/health`                    | 🔒   | Any   | AI health check          |
| `POST`   | `/api/v1/ai/predict-grade`             | 🔒   | Any   | Predict grade            |
| `POST`   | `/api/v1/ai/predict-pump`              | 🔒   | Any   | Predict pump action      |
| `POST`   | `/api/v1/ai/analyze`                   | 🔒   | Any   | Comprehensive analysis   |
| `POST`   | `/api/v1/ai/anomaly-detect`            | 🔒   | Any   | Anomaly detection        |
| `POST`   | `/api/v1/uploads/avatar`               | 🔒   | Any   | Upload avatar            |
| `POST`   | `/api/v1/uploads/rbw/{rbw_id}/photo`   | 🔒   | Any   | Upload RBW photo         |
| `GET`    | `/api/v1/ws`                           | 🔒   | Any   | WebSocket connection     |
| `GET`    | `/api/v1/ws/stats`                     | 🔒   | Any   | WebSocket stats          |
| `GET`    | `/health`                              | ❌   | -     | Health check             |
| `GET`    | `/metrics`                             | ❌   | -     | Prometheus metrics       |
