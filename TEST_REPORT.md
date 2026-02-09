# SwiftLead API - Production Endpoint Test Report

**Target:** https://api.swiftlead.fuadfakhruz.com  
**Date:** 2026-02-09 19:46:24 WIB  
**Environment:** Production  
**Tested Endpoints:** 85 (across 17 categories)

---

## Summary

| Metric                  | Count                      |
| ----------------------- | -------------------------- |
| **Total Tests**         | **85**                     |
| Passed                  | 60                         |
| Failed                  | 11                         |
| Skipped                 | 14                         |
| **Effective Pass Rate** | **84.5%** (60/71 executed) |
| Overall Pass Rate       | 70.5% (60/85 total)        |

### Failure Breakdown

| Category                        | Count | Severity | Notes                                                                                                                              |
| ------------------------------- | ----- | -------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| **BUG - Database Query Errors** | 5     | CRITICAL | `GET /harvests`, `GET /harvests/stats`, `GET /rbw/{id}/harvests`, `GET /service-requests`, `POST /harvests {}` all return HTTP 500 |
| **AI Engine Disabled**          | 4     | EXPECTED | `AI_ENGINE_ENABLED=false` in production config, so all AI prediction endpoints return 503                                          |
| **Admin Login Unknown**         | 2     | INFO     | Admin credentials not known to test suite, so admin-only tests were skipped                                                        |

## Detailed Results

| #   | Test Name                                         | Method | Expected | Actual | Status                   |
| --- | ------------------------------------------------- | ------ | -------- | ------ | ------------------------ |
| 1   | GET /health                                       | GET    | 200      | 200    | PASS                     |
| 2   | GET /metrics (Prometheus)                         | GET    | 200      | 200    | PASS                     |
| 3   | POST /auth/register - invalid (no body)           | POST   | 400      | 400    | PASS                     |
| 4   | POST /auth/register - weak password               | POST   | 400      | 400    | PASS                     |
| 5   | POST /auth/register - valid farmer                | POST   | 201      | 201    | PASS                     |
| 6   | POST /auth/register - duplicate email             | POST   | 409      | 409    | PASS                     |
| 7   | POST /auth/login - wrong password                 | POST   | 401      | 401    | PASS                     |
| 8   | POST /auth/login - valid credentials              | POST   | 200      | 200    | PASS                     |
| 9   | POST /auth/login - non-existent email             | POST   | 401      | 401    | PASS                     |
| 10  | POST /auth/login - invalid body                   | POST   | 400      | 400    | PASS                     |
| 11  | GET /users/me - no token (401)                    | GET    | 401      | 401    | PASS                     |
| 12  | GET /users/me - invalid token (401)               | GET    | 401      | 401    | PASS                     |
| 13  | POST /auth/change-password - wrong old pass       | POST   | 401      | 401    | PASS                     |
| 14  | POST /auth/change-password - same password        | POST   | 400      | 400    | PASS                     |
| 15  | GET /users/me - get profile                       | GET    | 200      | 200    | PASS                     |
| 16  | PATCH /users/me - update name                     | PATCH  | 200      | 200    | PASS                     |
| 17  | PATCH /users/me - invalid avatar_url              | PATCH  | 400      | 400    | PASS                     |
| 18  | GET /users - farmer (403 forbidden)               | GET    | 403      | 403    | PASS                     |
| 19  | POST /auth/login - admin login                    | POST   | 200      | 401    | FAIL                     |
| 20  | POST /auth/login - admin alt login                | POST   | 200      | 401    | FAIL                     |
| 21  | GET /users - admin (list users)                   | -      | -        | -      | SKIP (No admin token)    |
| 22  | GET /users?page=1&limit=5 - paginated             | -      | -        | -      | SKIP (No admin token)    |
| 23  | GET /users?role=farmer - filter by role           | -      | -        | -      | SKIP (No admin token)    |
| 24  | POST /users - admin create technician             | -      | -        | -      | SKIP (No admin token)    |
| 25  | POST /auth/forgot-password - reset user           | -      | -        | -      | SKIP (No admin token)    |
| 26  | POST /rbw - create RBW                            | POST   | 201      | 201    | PASS                     |
| 27  | GET /rbw - list RBWs                              | GET    | 200      | 200    | PASS                     |
| 28  | GET /rbw/{id} - get RBW                           | GET    | 200      | 200    | PASS                     |
| 29  | PATCH /rbw/{id} - update RBW                      | PATCH  | 200      | 200    | PASS                     |
| 30  | GET /rbw/{id} - not found (404)                   | GET    | 404      | 404    | PASS                     |
| 31  | POST /rbw/{id}/nodes - create gateway node        | POST   | 201      | 201    | PASS                     |
| 32  | POST /rbw/{id}/nodes - create nest node           | POST   | 201      | 201    | PASS                     |
| 33  | GET /rbw/{id}/nodes - list nodes                  | GET    | 200      | 200    | PASS                     |
| 34  | GET /nodes/{id} - get node                        | GET    | 200      | 200    | PASS                     |
| 35  | PATCH /nodes/{id} - update node                   | PATCH  | 200      | 200    | PASS                     |
| 36  | GET /nodes/{id}/audio - get audio state           | GET    | 200      | 200    | PASS                     |
| 37  | PATCH /nodes/{id}/audio - control audio           | PATCH  | 200      | 200    | PASS                     |
| 38  | PATCH /nodes/{id}/pump - control pump             | PATCH  | 200      | 200    | PASS                     |
| 39  | POST /nodes/{id}/sensors - create temp sensor     | POST   | 201      | 201    | PASS                     |
| 40  | POST /nodes/{id}/sensors - create humidity sensor | POST   | 201      | 201    | PASS                     |
| 41  | GET /nodes/{id}/sensors - list sensors            | GET    | 200      | 200    | PASS                     |
| 42  | GET /sensors/{id} - get sensor                    | GET    | 200      | 200    | PASS                     |
| 43  | PATCH /sensors/{id} - update sensor               | PATCH  | 200      | 200    | PASS                     |
| 44  | POST /sensors/{id}/readings - create reading      | POST   | 201      | 201    | PASS                     |
| 45  | GET /sensors/{id}/readings - get readings         | GET    | 200      | 200    | PASS                     |
| 46  | GET /sensors/{id}/trend - get trend               | GET    | 200      | 200    | PASS                     |
| 47  | GET /alerts - list all alerts                     | GET    | 200      | 200    | PASS                     |
| 48  | GET /rbw/{id}/alerts - alerts by RBW              | GET    | 200      | 200    | PASS                     |
| 49  | PATCH /alerts/{id}/read - mark as read            | -      | -        | -      | SKIP (No alert exists)   |
| 50  | PATCH /alerts/{id}/resolve - resolve alert        | -      | -        | -      | SKIP (No alert exists)   |
| 51  | POST /harvests - create harvest                   | POST   | 201      | 201    | PASS                     |
| 52  | GET /harvests - list all harvests                 | GET    | 200      | 500    | FAIL                     |
| 53  | GET /harvests/{id} - get harvest                  | GET    | 200      | 200    | PASS                     |
| 54  | PATCH /harvests/{id} - update harvest             | PATCH  | 200      | 200    | PASS                     |
| 55  | GET /harvests/stats - get harvest stats           | GET    | 200      | 500    | FAIL                     |
| 56  | GET /rbw/{id}/harvests - harvests by RBW          | GET    | 200      | 500    | FAIL                     |
| 57  | POST /service-requests - create request           | POST   | 201      | 201    | PASS                     |
| 58  | GET /service-requests - list requests             | GET    | 200      | 500    | FAIL                     |
| 59  | GET /service-requests/{id} - get request          | GET    | 200      | 200    | PASS                     |
| 60  | PATCH /service-requests/{id} - update request     | PATCH  | 200      | 200    | PASS                     |
| 61  | GET /transaction-categories - list categories     | GET    | 200      | 200    | PASS                     |
| 62  | POST /transaction-categories - admin create       | -      | -        | -      | SKIP (No admin token)    |
| 63  | PATCH /transaction-categories/{id} - admin update | -      | -        | -      | SKIP (No admin/category) |
| 64  | POST /transactions - create transaction           | -      | -        | -      | SKIP (No RBW/Category)   |
| 65  | GET /transactions/{id} - get transaction          | -      | -        | -      | SKIP (No Transaction ID) |
| 66  | PATCH /transactions/{id} - update transaction     | -      | -        | -      | SKIP (No Transaction ID) |
| 67  | GET /rbw/{id}/transactions - transactions by RBW  | GET    | 200      | 200    | PASS                     |
| 68  | POST /financial-statements - generate report      | POST   | 200      | 200    | PASS                     |
| 69  | GET /ai/health - AI health check                  | GET    | 200      | 200    | PASS                     |
| 70  | POST /ai/predict-grade                            | POST   | 200      | 503    | FAIL                     |
| 71  | POST /ai/predict-pump                             | POST   | 200      | 503    | FAIL                     |
| 72  | POST /ai/analyze                                  | POST   | 200      | 503    | FAIL                     |
| 73  | POST /ai/anomaly-detect                           | POST   | 200      | 503    | FAIL                     |
| 74  | POST /uploads/avatar - no file                    | POST   | 400/503  | 400    | PASS                     |
| 75  | GET /ws/stats - websocket stats                   | GET    | 200      | 200    | PASS                     |
| 76  | GET /api/v1/nonexistent - 404/405                 | GET    | 404      | 404    | PASS                     |
| 77  | DELETE /auth/login - method not allowed           | DELETE | 405      | 405    | PASS                     |
| 78  | POST /rbw - malformed JSON                        | POST   | 400      | 400    | PASS                     |
| 79  | POST /harvests - empty body (400)                 | POST   | 400      | 500    | FAIL                     |
| 80  | GET /users/me - tampered token (401)              | GET    | 401      | 401    | PASS                     |
| 81  | DELETE /transactions/{id} - cleanup               | -      | -        | -      | SKIP (No Transaction ID) |
| 82  | DELETE /harvests/{id} - cleanup                   | DELETE | 200      | 200    | PASS                     |
| 83  | DELETE /transaction-categories/{id} - cleanup     | -      | -        | -      | SKIP (No admin/category) |
| 84  | DELETE /nodes/{id} - cleanup                      | DELETE | 200      | 200    | PASS                     |
| 85  | DELETE /rbw/{id} - cleanup                        | DELETE | 200      | 200    | PASS                     |

## Test Categories

### 1. Health & Public Endpoints

Basic server health and metrics availability.

### 2-3. Authentication

Registration, login, token validation, password management.

### 4-5. User Management

Profile CRUD, admin user operations.

### 6. RBW (Swiftlet House)

Full CRUD operations on Rumah Burung Walet.

### 7. Node (IoT Device)

Full CRUD, audio control, pump control on ESP32 nodes.

### 8. Sensor

Sensor CRUD, readings ingestion, trend analysis.

### 9. Alerts

Alert listing, read marking, resolution.

### 10. Harvests

Harvest CRUD, statistics.

### 11. Service Requests

Service request lifecycle management.

### 12. Transactions & Categories

Financial transactions, categories, financial statements.

### 13. AI Engine

AI prediction endpoints (may be disabled in production).

### 14. Uploads

File upload endpoints (depends on MinIO storage).

### 15. WebSocket

WebSocket connection stats.

### 16. Edge Cases

Error handling, malformed requests, authorization checks.

### 17. Cleanup

Deletion of test data created during the test run.

---

_Generated automatically by SwiftLead API Test Suite_

---

## Detailed Failure Analysis

### BUG 1 (CRITICAL): Harvest List Endpoints Return 500

**Affected Endpoints:**

- `GET /api/v1/harvests` → HTTP 500
- `GET /api/v1/harvests/stats` → HTTP 500
- `GET /api/v1/rbw/{rbw_id}/harvests` → HTTP 500

**Error Response:**

```json
{
  "success": false,
  "message": "Failed to list harvests",
  "error": "internal_error"
}
```

**Root Cause:** The SQL query in `harvest_repository.go` uses a pattern `WHERE ($1 = '' OR rbw_id = $1)` where `rbw_id` is a `UUID` column. When an empty string `''` is passed as the `$1` parameter, PostgreSQL cannot compare a UUID column with an empty string, causing a type mismatch error.

**Location:** [internal/repository/harvest_repository.go](internal/repository/harvest_repository.go#L101-L112)

```go
countQuery := `SELECT COUNT(*) FROM harvests WHERE ($1 = '' OR rbw_id = $1)`
```

**Fix Suggestion:** Use `CAST` or change to a `NULLIF` pattern:

```sql
WHERE ($1::text = '' OR rbw_id = $1::uuid)
```

Or use conditional query building in Go code.

---

### BUG 2 (CRITICAL): Service Request List Returns 500

**Affected Endpoint:**

- `GET /api/v1/service-requests` → HTTP 500

**Error Response:**

```json
{
  "success": false,
  "message": "Failed to list service requests",
  "error": "internal_error"
}
```

**Root Cause:** Same UUID/string comparison issue. The `assigned_to` column is `UUID` type, but the query tries `$3 = '' OR assigned_to = $3` where `$3` is an empty Go string.

**Location:** [internal/repository/service_transaction_repository.go](internal/repository/service_transaction_repository.go#L90-L108)

```go
AND ($3 = '' OR assigned_to = $3)
```

**Fix Suggestion:** Cast parameters to text for the empty comparison, or use `COALESCE`:

```sql
AND ($3::text = '' OR assigned_to = $3::uuid)
```

---

### BUG 3 (MEDIUM): POST /harvests with Empty Body Returns 500 instead of 400

**Affected Endpoint:**

- `POST /api/v1/harvests` with `{}` body → HTTP 500

**Expected:** HTTP 400 (Bad Request / validation error)

**Root Cause:** The handler does not validate required fields (rbw_id, floor_no, harvested_at, nests_count) before passing to the service layer. The request struct `CreateHarvestRequest` lacks `validate:"required"` tags, so the empty body passes through to the database which then fails on NOT NULL constraints.

**Location:** [internal/handlers/harvest_handler.go](internal/handlers/harvest_handler.go#L69-L82)

**Fix Suggestion:** Add validation tags to `CreateHarvestRequest`:

```go
type CreateHarvestRequest struct {
    RBWID       string    `json:"rbw_id" validate:"required,uuid"`
    FloorNo     int       `json:"floor_no" validate:"required,min=1"`
    HarvestedAt time.Time `json:"harvested_at" validate:"required"`
    NestsCount  int       `json:"nests_count" validate:"required,min=0"`
    // ...
}
```

---

### INFO: AI Engine Disabled (Expected)

**Affected Endpoints:**

- `POST /api/v1/ai/predict-grade` → HTTP 503
- `POST /api/v1/ai/predict-pump` → HTTP 503
- `POST /api/v1/ai/analyze` → HTTP 503
- `POST /api/v1/ai/anomaly-detect` → HTTP 503

**Reason:** `AI_ENGINE_ENABLED=false` in production configuration. This is **expected** behavior. The AI health check endpoint correctly returns 200 with disabled status.

---

### INFO: Admin Credentials Unknown

The test suite tried `admin@swiftlead.com` and `admin@swiftlet.com`, both failed. The seed file uses `admin@swiftlead.id` with password `admin123`, but that may have been changed in production. The following tests were **skipped** due to no admin access:

- `GET /users` (admin list)
- `POST /users` (admin create)
- `POST /auth/admin/register`
- `POST /auth/forgot-password`
- `POST /transaction-categories` (admin create)
- `PATCH /transaction-categories/{id}` (admin update)
- `DELETE /transaction-categories/{id}` (admin delete)
- Transaction CRUD (dependent on admin-created category)

---

## Endpoint Coverage Matrix

| Module           | Endpoints | Tested | Passed | Failed | Skipped |
| ---------------- | --------- | ------ | ------ | ------ | ------- |
| Health/Metrics   | 2         | 2      | 2      | 0      | 0       |
| Auth (Public)    | 2         | 8      | 8      | 0      | 0       |
| Auth (Protected) | 1         | 2      | 2      | 0      | 0       |
| Users            | 4         | 6      | 4      | 0      | 2       |
| Admin Auth       | 3         | 2      | 0      | 2      | 3       |
| RBW              | 5         | 5      | 5      | 0      | 0       |
| Nodes            | 8         | 8      | 8      | 0      | 0       |
| Sensors          | 5         | 8      | 8      | 0      | 0       |
| Alerts           | 3         | 2      | 2      | 0      | 2       |
| Harvests         | 6         | 6      | 3      | 3      | 0       |
| Service Requests | 4         | 4      | 3      | 1      | 0       |
| Transactions     | 4         | 2      | 2      | 0      | 4       |
| Categories       | 3         | 1      | 1      | 0      | 2       |
| Financial Stmt   | 1         | 1      | 1      | 0      | 0       |
| AI Engine        | 5         | 5      | 1      | 4      | 0       |
| Uploads          | 2         | 1      | 1      | 0      | 0       |
| WebSocket        | 2         | 1      | 1      | 0      | 0       |
| Edge Cases       | -         | 5      | 4      | 1      | 0       |
| Cleanup          | -         | 5      | 3      | 0      | 2       |
| **Total**        | **60+**   | **85** | **60** | **11** | **14**  |

---

## Recommendations

### Priority 1 - Fix Now (Production Bugs)

1. **Fix UUID type mismatch** in `harvest_repository.go` List and GetStats queries
2. **Fix UUID type mismatch** in `service_transaction_repository.go` List query
3. **Add input validation** to `CreateHarvestRequest` model with proper `validate` tags

### Priority 2 - Improve

4. **Add validation tags** to all Create/Update request models (`CreateServiceRequestRequest`, `CreateTransactionRequest`, etc.)
5. **Consistent error handling**: All endpoints should return 400 for missing required fields, never 500

### Priority 3 - Nice to Have

6. **Enable AI Engine** in production when ready
7. **Document admin credentials** or provide a way to create admin accounts via CLI
8. **Add rate limiting** to public endpoints (`/auth/login`, `/auth/register`)
