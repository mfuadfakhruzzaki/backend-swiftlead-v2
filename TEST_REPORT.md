# SwiftLead API - Production Endpoint Test Report

**Target:** https://api.swiftlead.fuadfakhruz.com  
**Date:** 2026-02-09 20:02:55 WIB  
**Environment:** Production

## Summary

| Metric | Count |
|--------|-------|
| Total Tests | 84 |
| Passed | 84 |
| Failed | 0 |
| Pass Rate | 100.0% |

## Detailed Results

| # | Test Name | Method | Expected | Actual | Status |
|---|-----------|--------|----------|--------|--------|
| 1 | GET /health | GET | 200 | 200 | PASS |
| 2 | GET /metrics (Prometheus) | GET | 200 | 200 | PASS |
| 3 | POST /auth/register - invalid (no body) | POST | 400 | 400 | PASS |
| 4 | POST /auth/register - weak password | POST | 400 | 400 | PASS |
| 5 | POST /auth/register - valid farmer | POST | 201 | 201 | PASS |
| 6 | POST /auth/register - duplicate email | POST | 409 | 409 | PASS |
| 7 | POST /auth/login - wrong password | POST | 401 | 401 | PASS |
| 8 | POST /auth/login - valid credentials | POST | 200 | 200 | PASS |
| 9 | POST /auth/login - non-existent email | POST | 401 | 401 | PASS |
| 10 | POST /auth/login - invalid body | POST | 400 | 400 | PASS |
| 11 | GET /users/me - no token (401) | GET | 401 | 401 | PASS |
| 12 | GET /users/me - invalid token (401) | GET | 401 | 401 | PASS |
| 13 | POST /auth/change-password - wrong old pass | POST | 401 | 401 | PASS |
| 14 | POST /auth/change-password - same password | POST | 400 | 400 | PASS |
| 15 | GET /users/me - get profile | GET | 200 | 200 | PASS |
| 16 | PATCH /users/me - update name | PATCH | 200 | 200 | PASS |
| 17 | PATCH /users/me - invalid avatar_url | PATCH | 400 | 400 | PASS |
| 18 | GET /users - farmer (403 forbidden) | GET | 403 | 403 | PASS |
| 19 | POST /auth/login - admin login | POST | 200 | 200 | PASS |
| 20 | GET /users - admin (list users) | GET | 200 | 200 | PASS |
| 21 | GET /users?page=1&limit=5 - paginated | GET | 200 | 200 | PASS |
| 22 | GET /users?role=farmer - filter by role | GET | 200 | 200 | PASS |
| 23 | POST /users - admin create technician | POST | 201 | 201 | PASS |
| 24 | POST /auth/forgot-password - reset user | POST | 200 | 200 | PASS |
| 25 | POST /rbw - create RBW | POST | 201 | 201 | PASS |
| 26 | GET /rbw - list RBWs | GET | 200 | 200 | PASS |
| 27 | GET /rbw/{id} - get RBW | GET | 200 | 200 | PASS |
| 28 | PATCH /rbw/{id} - update RBW | PATCH | 200 | 200 | PASS |
| 29 | GET /rbw/{id} - not found (404) | GET | 404 | 404 | PASS |
| 30 | POST /rbw/{id}/nodes - create gateway node | POST | 201 | 201 | PASS |
| 31 | POST /rbw/{id}/nodes - create nest node | POST | 201 | 201 | PASS |
| 32 | GET /rbw/{id}/nodes - list nodes | GET | 200 | 200 | PASS |
| 33 | GET /nodes/{id} - get node | GET | 200 | 200 | PASS |
| 34 | PATCH /nodes/{id} - update node | PATCH | 200 | 200 | PASS |
| 35 | GET /nodes/{id}/audio - get audio state | GET | 200 | 200 | PASS |
| 36 | PATCH /nodes/{id}/audio - control audio | PATCH | 200 | 200 | PASS |
| 37 | PATCH /nodes/{id}/pump - control pump | PATCH | 200 | 200 | PASS |
| 38 | POST /nodes/{id}/sensors - create temp sensor | POST | 201 | 201 | PASS |
| 39 | POST /nodes/{id}/sensors - create humidity sensor | POST | 201 | 201 | PASS |
| 40 | GET /nodes/{id}/sensors - list sensors | GET | 200 | 200 | PASS |
| 41 | GET /sensors/{id} - get sensor | GET | 200 | 200 | PASS |
| 42 | PATCH /sensors/{id} - update sensor | PATCH | 200 | 200 | PASS |
| 43 | POST /sensors/{id}/readings - create reading | POST | 201 | 201 | PASS |
| 44 | GET /sensors/{id}/readings - get readings | GET | 200 | 200 | PASS |
| 45 | GET /sensors/{id}/trend - get trend | GET | 200 | 200 | PASS |
| 46 | GET /alerts - list all alerts | GET | 200 | 200 | PASS |
| 47 | GET /rbw/{id}/alerts - alerts by RBW | GET | 200 | 200 | PASS |
| 48 | PATCH /alerts/{id}/read - not found (404) | PATCH | 404 | 404 | PASS |
| 49 | PATCH /alerts/{id}/resolve - not found (404) | PATCH | 404 | 404 | PASS |
| 50 | POST /harvests - create harvest | POST | 201 | 201 | PASS |
| 51 | GET /harvests - list all harvests | GET | 200 | 200 | PASS |
| 52 | GET /harvests/{id} - get harvest | GET | 200 | 200 | PASS |
| 53 | PATCH /harvests/{id} - update harvest | PATCH | 200 | 200 | PASS |
| 54 | GET /harvests/stats - get harvest stats | GET | 200 | 200 | PASS |
| 55 | GET /rbw/{id}/harvests - harvests by RBW | GET | 200 | 200 | PASS |
| 56 | POST /service-requests - create request | POST | 201 | 201 | PASS |
| 57 | GET /service-requests - list requests | GET | 200 | 200 | PASS |
| 58 | GET /service-requests/{id} - get request | GET | 200 | 200 | PASS |
| 59 | PATCH /service-requests/{id} - update request | PATCH | 200 | 200 | PASS |
| 60 | GET /transaction-categories - list categories | GET | 200 | 200 | PASS |
| 61 | POST /transaction-categories - admin create | POST | 201 | 201 | PASS |
| 62 | PATCH /transaction-categories/{id} - admin update | PATCH | 200 | 200 | PASS |
| 63 | POST /transactions - create transaction | POST | 201 | 201 | PASS |
| 64 | GET /transactions/{id} - get transaction | GET | 200 | 200 | PASS |
| 65 | PATCH /transactions/{id} - update transaction | PATCH | 200 | 200 | PASS |
| 66 | GET /rbw/{id}/transactions - transactions by RBW | GET | 200 | 200 | PASS |
| 67 | POST /financial-statements - generate report | POST | 200 | 200 | PASS |
| 68 | GET /ai/health - AI health check | GET | 200|503 | 200 | PASS |
| 69 | POST /ai/predict-grade | POST | 200|503 | 503 | PASS |
| 70 | POST /ai/predict-pump | POST | 200|503 | 503 | PASS |
| 71 | POST /ai/analyze | POST | 200|503 | 503 | PASS |
| 72 | POST /ai/anomaly-detect | POST | 200|503 | 503 | PASS |
| 73 | POST /uploads/avatar - no file | POST | 400/500/503 | 503 | PASS |
| 74 | GET /ws/stats - websocket stats | GET | 200 | 200 | PASS |
| 75 | GET /api/v1/nonexistent - 404/405 | GET | 404|405 | 404 | PASS |
| 76 | DELETE /auth/login - method not allowed | DELETE | 405 | 405 | PASS |
| 77 | POST /rbw - malformed JSON | POST | 400 | 400 | PASS |
| 78 | POST /harvests - empty body (400) | POST | 400 | 400 | PASS |
| 79 | GET /users/me - tampered token (401) | GET | 401 | 401 | PASS |
| 80 | DELETE /transactions/{id} - cleanup | DELETE | 200 | 200 | PASS |
| 81 | DELETE /harvests/{id} - cleanup | DELETE | 200 | 200 | PASS |
| 82 | DELETE /transaction-categories/{id} - cleanup | DELETE | 200 | 200 | PASS |
| 83 | DELETE /nodes/{id} - cleanup | DELETE | 200 | 200 | PASS |
| 84 | DELETE /rbw/{id} - cleanup | DELETE | 200 | 200 | PASS |

## Notes

- AI Engine tests accept both 200 (enabled) and 503 (disabled) as PASS
- Alert mark/resolve tests use non-existent UUID when no real alerts exist (expect 404)
- Admin credentials sourced from seed migration (admin@swiftlead.id)
- Upload test expects error response since no file is attached

---
*Generated automatically by SwiftLead API Test Suite*
