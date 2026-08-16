#!/bin/bash
# ================================================================
# SwiftLead API Production Endpoint Test Suite
# Target: https://api.swiftlead.fuadfakhruz.com
# ================================================================

set -euo pipefail

BASE_URL="https://api.swiftlead.fuadfakhruz.com"
API="$BASE_URL/api/v1"

# Test counters
TOTAL=0
PASSED=0
FAILED=0
REPORT=""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

# Storage for IDs created during tests
TOKEN=""
ADMIN_TOKEN=""
USER_ID=""
RBW_ID=""
NODE_ID=""
SENSOR_ID=""
ALERT_ID=""
HARVEST_ID=""
SERVICE_REQUEST_ID=""
TRANSACTION_ID=""
CATEGORY_ID=""
NEW_CAT_ID=""
AI_DISABLED=false
TEST_EMAIL="testuser_$(date +%s)@swiftlead.test"
TEST_PASSWORD="TestPass123!@"

# Admin credentials (from seed migration 010_seed_admin.sql)
ADMIN_EMAIL="admin@swiftlead.id"
ADMIN_PASSWORD="admin123"

# ================================================================
# Helper Functions
# ================================================================

log_section() {
    echo ""
    echo -e "${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BOLD}${CYAN}  $1${NC}"
    echo -e "${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# run_test: run a single test with one expected status
run_test() {
    local test_name="$1"
    local method="$2"
    local url="$3"
    local expected_status="$4"
    local data="${5:-}"
    local auth_token="${6:-}"

    TOTAL=$((TOTAL + 1))

    local curl_cmd="curl -s -o /tmp/test_response.json -w '%{http_code}' -X $method"
    curl_cmd="$curl_cmd -H 'Content-Type: application/json'"

    if [ -n "$auth_token" ]; then
        curl_cmd="$curl_cmd -H 'Authorization: Bearer $auth_token'"
    fi

    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -d '$data'"
    fi

    curl_cmd="$curl_cmd '$url'"

    local status_code
    status_code=$(eval $curl_cmd 2>/dev/null) || status_code="000"
    local response_body
    response_body=$(cat /tmp/test_response.json 2>/dev/null) || response_body="{}"

    local result=""
    local detail=""

    if [ "$status_code" = "$expected_status" ]; then
        PASSED=$((PASSED + 1))
        result="${GREEN}PASS${NC}"
        detail="HTTP $status_code"
    else
        FAILED=$((FAILED + 1))
        result="${RED}FAIL${NC}"
        detail="Expected $expected_status, Got $status_code"
    fi

    printf "  %-4s %-60s [%b] %s\n" "$TOTAL." "$test_name" "$result" "$detail"

    local status_text
    if [ "$status_code" = "$expected_status" ]; then
        status_text="PASS"
    else
        status_text="FAIL"
    fi
    REPORT="$REPORT\n| $TOTAL | $test_name | $method | $expected_status | $status_code | $status_text |"

    echo "$response_body" > /tmp/last_test_response.json
    echo "$status_code" > /tmp/last_test_status.txt
}

# run_test_multi: run a test accepting multiple valid status codes
# expected_statuses is a pipe-separated string: "200|503"
run_test_multi() {
    local test_name="$1"
    local method="$2"
    local url="$3"
    local expected_statuses="$4"
    local data="${5:-}"
    local auth_token="${6:-}"

    TOTAL=$((TOTAL + 1))

    local curl_cmd="curl -s -o /tmp/test_response.json -w '%{http_code}' -X $method"
    curl_cmd="$curl_cmd -H 'Content-Type: application/json'"

    if [ -n "$auth_token" ]; then
        curl_cmd="$curl_cmd -H 'Authorization: Bearer $auth_token'"
    fi

    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -d '$data'"
    fi

    curl_cmd="$curl_cmd '$url'"

    local status_code
    status_code=$(eval $curl_cmd 2>/dev/null) || status_code="000"
    local response_body
    response_body=$(cat /tmp/test_response.json 2>/dev/null) || response_body="{}"

    local result=""
    local detail=""
    local matched=false

    IFS='|' read -ra STATUSES <<< "$expected_statuses"
    for s in "${STATUSES[@]}"; do
        if [ "$status_code" = "$s" ]; then
            matched=true
            break
        fi
    done

    if $matched; then
        PASSED=$((PASSED + 1))
        result="${GREEN}PASS${NC}"
        detail="HTTP $status_code (accepted: $expected_statuses)"
    else
        FAILED=$((FAILED + 1))
        result="${RED}FAIL${NC}"
        detail="Expected $expected_statuses, Got $status_code"
    fi

    printf "  %-4s %-60s [%b] %s\n" "$TOTAL." "$test_name" "$result" "$detail"

    local status_text
    if $matched; then
        status_text="PASS"
    else
        status_text="FAIL"
    fi
    REPORT="$REPORT\n| $TOTAL | $test_name | $method | $expected_statuses | $status_code | $status_text |"

    echo "$response_body" > /tmp/last_test_response.json
    echo "$status_code" > /tmp/last_test_status.txt
}

extract_json() {
    local field="$1"
    local file="${2:-/tmp/last_test_response.json}"
    python3 -c "
import json, sys
try:
    d = json.load(open('$file'))
    keys = '$field'.split('.')
    for k in keys:
        if isinstance(d, dict):
            d = d.get(k, '')
        elif isinstance(d, list):
            try:
                d = d[int(k)]
            except:
                d = ''
                break
        else:
            d = ''
            break
    print(d if d else '')
except:
    print('')
" 2>/dev/null
}

# ================================================================
# 0. PRE-FLIGHT CHECK
# ================================================================
log_section "0. PRE-FLIGHT CHECKS"

echo -e "  Target: ${BOLD}$BASE_URL${NC}"
echo -e "  Date:   $(date '+%Y-%m-%d %H:%M:%S %Z')"
echo ""

if ! curl -s --max-time 10 "$BASE_URL" > /dev/null 2>&1; then
    echo -e "  ${RED}ERROR: Server not reachable at $BASE_URL${NC}"
    exit 1
fi
echo -e "  ${GREEN}Server is reachable${NC}"
echo ""

# ================================================================
# 1. HEALTH & PUBLIC ENDPOINTS
# ================================================================
log_section "1. HEALTH & PUBLIC ENDPOINTS"

run_test "GET /health" \
    "GET" "$BASE_URL/health" "200"

run_test "GET /metrics (Prometheus)" \
    "GET" "$BASE_URL/metrics" "200"

# ================================================================
# 2. AUTH - REGISTRATION & LOGIN
# ================================================================
log_section "2. AUTH - REGISTRATION & LOGIN"

run_test "POST /auth/register - invalid (no body)" \
    "POST" "$API/auth/register" "400" '{}'

run_test "POST /auth/register - weak password" \
    "POST" "$API/auth/register" "400" \
    '{"name":"Test User","email":"'"$TEST_EMAIL"'","password":"weak"}'

run_test "POST /auth/register - valid farmer" \
    "POST" "$API/auth/register" "201" \
    '{"name":"Test API User","email":"'"$TEST_EMAIL"'","password":"'"$TEST_PASSWORD"'","phone":"081234567890"}'

USER_ID=$(extract_json "data.id")
echo "       -> Created user ID: $USER_ID"

run_test "POST /auth/register - duplicate email" \
    "POST" "$API/auth/register" "409" \
    '{"name":"Test API User","email":"'"$TEST_EMAIL"'","password":"'"$TEST_PASSWORD"'"}'

run_test "POST /auth/login - wrong password" \
    "POST" "$API/auth/login" "401" \
    '{"email":"'"$TEST_EMAIL"'","password":"WrongPass123!"}'

run_test "POST /auth/login - valid credentials" \
    "POST" "$API/auth/login" "200" \
    '{"email":"'"$TEST_EMAIL"'","password":"'"$TEST_PASSWORD"'"}'

TOKEN=$(extract_json "data.token")
if [ -n "$TOKEN" ]; then
    echo "       -> Got auth token: ${TOKEN:0:20}..."
else
    echo -e "       ${RED}-> CRITICAL: No token received. Protected tests will fail.${NC}"
fi

run_test "POST /auth/login - non-existent email" \
    "POST" "$API/auth/login" "401" \
    '{"email":"nonexistent@test.com","password":"SomePass123!"}'

run_test "POST /auth/login - invalid body" \
    "POST" "$API/auth/login" "400" \
    '{"email":"not-an-email"}'

# ================================================================
# 3. AUTH - PROTECTED ENDPOINTS
# ================================================================
log_section "3. AUTH - PROTECTED ENDPOINTS"

run_test "GET /users/me - no token (401)" \
    "GET" "$API/users/me" "401"

run_test "GET /users/me - invalid token (401)" \
    "GET" "$API/users/me" "401" "" "invalid_token_here"

run_test "POST /auth/change-password - wrong old pass" \
    "POST" "$API/auth/change-password" "401" \
    '{"old_password":"WrongOldPass1!","new_password":"NewPass123!@"}' \
    "$TOKEN"

run_test "POST /auth/change-password - same password" \
    "POST" "$API/auth/change-password" "400" \
    '{"old_password":"'"$TEST_PASSWORD"'","new_password":"'"$TEST_PASSWORD"'"}' \
    "$TOKEN"

# ================================================================
# 4. USER ENDPOINTS
# ================================================================
log_section "4. USER ENDPOINTS"

run_test "GET /users/me - get profile" \
    "GET" "$API/users/me" "200" "" "$TOKEN"

run_test "PATCH /users/me - update name" \
    "PATCH" "$API/users/me" "200" \
    '{"name":"Updated Test User"}' \
    "$TOKEN"

run_test "PATCH /users/me - invalid avatar_url" \
    "PATCH" "$API/users/me" "400" \
    '{"avatar_url":"not-a-url"}' \
    "$TOKEN"

run_test "GET /users - farmer (403 forbidden)" \
    "GET" "$API/users" "403" "" "$TOKEN"

# ================================================================
# 5. ADMIN ACCESS
# ================================================================
log_section "5. ADMIN ACCESS"

run_test "POST /auth/login - admin login" \
    "POST" "$API/auth/login" "200" \
    '{"email":"'"$ADMIN_EMAIL"'","password":"'"$ADMIN_PASSWORD"'"}'

ADMIN_TOKEN=$(extract_json "data.token")
if [ -n "$ADMIN_TOKEN" ]; then
    echo "       -> Got admin token: ${ADMIN_TOKEN:0:20}..."
else
    echo -e "       ${RED}-> Admin login failed. Some tests may fail.${NC}"
fi

# Use admin token as main auth for subsequent tests
AUTH="${ADMIN_TOKEN:-$TOKEN}"

run_test "GET /users - admin (list users)" \
    "GET" "$API/users" "200" "" "$ADMIN_TOKEN"

run_test "GET /users?page=1&limit=5 - paginated" \
    "GET" "$API/users?page=1&limit=5" "200" "" "$ADMIN_TOKEN"

run_test "GET /users?role=farmer - filter by role" \
    "GET" "$API/users?role=farmer" "200" "" "$ADMIN_TOKEN"

TECH_EMAIL="tech_$(date +%s)@swiftlead.test"
run_test "POST /users - admin create technician" \
    "POST" "$API/users" "201" \
    '{"name":"Test Technician","email":"'"$TECH_EMAIL"'","password":"TechPass123!@","role":"technician"}' \
    "$ADMIN_TOKEN"

run_test "POST /auth/forgot-password - reset user" \
    "POST" "$API/auth/forgot-password" "200" \
    '{"email":"'"$TEST_EMAIL"'"}' \
    "$ADMIN_TOKEN"

# ================================================================
# 6. RBW (Rumah Burung Walet) ENDPOINTS
# ================================================================
log_section "6. RBW ENDPOINTS"

run_test "POST /rbw - create RBW" \
    "POST" "$API/rbw" "201" \
    '{"code":"RBW-TEST-'$(date +%s)'","name":"Test Swiftlet House","address":"Jl. Test No. 1","latitude":-6.2088,"longitude":106.8456,"total_floors":3,"description":"Test RBW for API testing"}' \
    "$AUTH"

RBW_ID=$(extract_json "data.id")
echo "       -> Created RBW ID: $RBW_ID"

run_test "GET /rbw - list RBWs" \
    "GET" "$API/rbw" "200" "" "$AUTH"

run_test "GET /rbw/{id} - get RBW" \
    "GET" "$API/rbw/$RBW_ID" "200" "" "$AUTH"

run_test "PATCH /rbw/{id} - update RBW" \
    "PATCH" "$API/rbw/$RBW_ID" "200" \
    '{"name":"Updated Test Swiftlet House","total_floors":4}' \
    "$AUTH"

run_test "GET /rbw/{id} - not found (404)" \
    "GET" "$API/rbw/00000000-0000-0000-0000-000000000000" "404" "" "$AUTH"

# ================================================================
# 7. NODE ENDPOINTS
# ================================================================
log_section "7. NODE ENDPOINTS"

run_test "POST /rbw/{id}/nodes - create gateway node" \
    "POST" "$API/rbw/$RBW_ID/nodes" "201" \
    '{"node_type":"gateway","node_code":"GW-TEST-001","has_audio":true,"has_pump":true}' \
    "$AUTH"

NODE_ID=$(extract_json "data.id")
echo "       -> Created Node ID: $NODE_ID"

run_test "POST /rbw/{id}/nodes - create nest node" \
    "POST" "$API/rbw/$RBW_ID/nodes" "201" \
    '{"node_type":"nest","node_code":"NEST-TEST-001","has_audio":true,"has_pump":false}' \
    "$AUTH"

run_test "GET /rbw/{id}/nodes - list nodes" \
    "GET" "$API/rbw/$RBW_ID/nodes" "200" "" "$AUTH"

run_test "GET /nodes/{id} - get node" \
    "GET" "$API/nodes/$NODE_ID" "200" "" "$AUTH"

run_test "PATCH /nodes/{id} - update node" \
    "PATCH" "$API/nodes/$NODE_ID" "200" \
    '{"node_code":"GW-TEST-UPDATED"}' \
    "$AUTH"

run_test "GET /nodes/{id}/audio - get audio state" \
    "GET" "$API/nodes/$NODE_ID/audio" "200" "" "$AUTH"

run_test "PATCH /nodes/{id}/audio - control audio" \
    "PATCH" "$API/nodes/$NODE_ID/audio" "200" \
    '{"action":"audio_set_lmb","value":1}' \
    "$AUTH"

run_test "PATCH /nodes/{id}/pump - control pump" \
    "PATCH" "$API/nodes/$NODE_ID/pump" "200" \
    '{"action":"sprayer_set","value":1}' \
    "$AUTH"

# ================================================================
# 8. SENSOR ENDPOINTS
# ================================================================
log_section "8. SENSOR ENDPOINTS"

run_test "POST /nodes/{id}/sensors - create temp sensor" \
    "POST" "$API/nodes/$NODE_ID/sensors" "201" \
    '{"sensor_type":"temp","sensor_name":"Temperature Sensor 1","unit":"°C"}' \
    "$AUTH"

SENSOR_ID=$(extract_json "data.id")
echo "       -> Created Sensor ID: $SENSOR_ID"

run_test "POST /nodes/{id}/sensors - create humidity sensor" \
    "POST" "$API/nodes/$NODE_ID/sensors" "201" \
    '{"sensor_type":"humid","sensor_name":"Humidity Sensor 1","unit":"%"}' \
    "$AUTH"

run_test "GET /nodes/{id}/sensors - list sensors" \
    "GET" "$API/nodes/$NODE_ID/sensors" "200" "" "$AUTH"

run_test "GET /sensors/{id} - get sensor" \
    "GET" "$API/sensors/$SENSOR_ID" "200" "" "$AUTH"

run_test "PATCH /sensors/{id} - update sensor" \
    "PATCH" "$API/sensors/$SENSOR_ID" "200" \
    '{"sensor_name":"Updated Temp Sensor"}' \
    "$AUTH"

run_test "POST /sensors/{id}/readings - create reading" \
    "POST" "$API/sensors/$SENSOR_ID/readings" "201" \
    '{"value":28.5,"recorded_at":"2026-02-09T10:00:00Z"}' \
    "$AUTH"

# Add extra readings for trend
for i in 1 2 3 4; do
    TEMP_VAL=$(echo "28 + $i * 0.5" | bc)
    HOUR=$((10 + i))
    curl -s -X POST "$API/sensors/$SENSOR_ID/readings" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $AUTH" \
        -d '{"value":'"$TEMP_VAL"',"recorded_at":"2026-02-09T'"$HOUR"':00:00Z"}' > /dev/null 2>&1
done
echo "       -> Inserted 4 additional readings for trend analysis"

run_test "GET /sensors/{id}/readings - get readings" \
    "GET" "$API/sensors/$SENSOR_ID/readings" "200" "" "$AUTH"

run_test "GET /sensors/{id}/trend - get trend" \
    "GET" "$API/sensors/$SENSOR_ID/trend" "200" "" "$AUTH"

# ================================================================
# 9. ALERT ENDPOINTS
# ================================================================
log_section "9. ALERT ENDPOINTS"

run_test "GET /alerts - list all alerts" \
    "GET" "$API/alerts" "200" "" "$AUTH"

ALERT_ID=$(extract_json "data.0.id" 2>/dev/null || echo "")

run_test "GET /rbw/{id}/alerts - alerts by RBW" \
    "GET" "$API/rbw/$RBW_ID/alerts" "200" "" "$AUTH"

# Mark alert as read - use real alert if exists, or test with non-existent UUID (expect 404)
if [ -n "$ALERT_ID" ] && [ "$ALERT_ID" != "" ]; then
    run_test "PATCH /alerts/{id}/read - mark as read" \
        "PATCH" "$API/alerts/$ALERT_ID/read" "200" "" "$AUTH"
else
    run_test "PATCH /alerts/{id}/read - not found (404)" \
        "PATCH" "$API/alerts/00000000-0000-0000-0000-000000000099/read" "404" "" "$AUTH"
fi

# Resolve alert
if [ -n "$ALERT_ID" ] && [ "$ALERT_ID" != "" ]; then
    run_test "PATCH /alerts/{id}/resolve - resolve alert" \
        "PATCH" "$API/alerts/$ALERT_ID/resolve" "200" "" "$AUTH"
else
    run_test "PATCH /alerts/{id}/resolve - not found (404)" \
        "PATCH" "$API/alerts/00000000-0000-0000-0000-000000000099/resolve" "404" "" "$AUTH"
fi

# ================================================================
# 10. HARVEST ENDPOINTS
# ================================================================
log_section "10. HARVEST ENDPOINTS"

run_test "POST /harvests - create harvest" \
    "POST" "$API/harvests" "201" \
    '{"rbw_id":"'"$RBW_ID"'","floor_no":1,"harvested_at":"2026-02-09T08:00:00Z","nests_count":50,"weight_kg":2.5,"grade":"good","notes":"Test harvest"}' \
    "$AUTH"

HARVEST_ID=$(extract_json "data.id")
echo "       -> Created Harvest ID: $HARVEST_ID"

run_test "GET /harvests - list all harvests" \
    "GET" "$API/harvests" "200" "" "$AUTH"

run_test "GET /harvests/{id} - get harvest" \
    "GET" "$API/harvests/$HARVEST_ID" "200" "" "$AUTH"

run_test "PATCH /harvests/{id} - update harvest" \
    "PATCH" "$API/harvests/$HARVEST_ID" "200" \
    '{"nests_count":55,"weight_kg":2.8,"grade":"good"}' \
    "$AUTH"

run_test "GET /harvests/stats - get harvest stats" \
    "GET" "$API/harvests/stats" "200" "" "$AUTH"

run_test "GET /rbw/{id}/harvests - harvests by RBW" \
    "GET" "$API/rbw/$RBW_ID/harvests" "200" "" "$AUTH"

# ================================================================
# 11. SERVICE REQUEST ENDPOINTS
# ================================================================
log_section "11. SERVICE REQUEST ENDPOINTS"

run_test "POST /service-requests - create request" \
    "POST" "$API/service-requests" "201" \
    '{"rbw_id":"'"$RBW_ID"'","type":"maintenance","issue":"Sensor malfunction on floor 2"}' \
    "$AUTH"

SERVICE_REQUEST_ID=$(extract_json "data.id")
echo "       -> Created Service Request ID: $SERVICE_REQUEST_ID"

run_test "GET /service-requests - list requests" \
    "GET" "$API/service-requests" "200" "" "$AUTH"

run_test "GET /service-requests/{id} - get request" \
    "GET" "$API/service-requests/$SERVICE_REQUEST_ID" "200" "" "$AUTH"

run_test "PATCH /service-requests/{id} - update request" \
    "PATCH" "$API/service-requests/$SERVICE_REQUEST_ID" "200" \
    '{"status":"pending","notes":"Escalated for review"}' \
    "$AUTH"

# ================================================================
# 12. TRANSACTION & CATEGORY ENDPOINTS
# ================================================================
log_section "12. TRANSACTION & CATEGORY ENDPOINTS"

run_test "GET /transaction-categories - list categories" \
    "GET" "$API/transaction-categories" "200" "" "$AUTH"

CATEGORY_ID=$(extract_json "data.0.id" 2>/dev/null || echo "")
if [ -n "$CATEGORY_ID" ]; then
    echo "       -> Found existing category ID: $CATEGORY_ID"
fi

run_test "POST /transaction-categories - admin create" \
    "POST" "$API/transaction-categories" "201" \
    '{"name":"Test Category '$(date +%s)'","type":"income","description":"Test category for API testing"}' \
    "$ADMIN_TOKEN"

NEW_CAT_ID=$(extract_json "data.id")
if [ -n "$NEW_CAT_ID" ]; then
    CATEGORY_ID="$NEW_CAT_ID"
    echo "       -> Created category ID: $CATEGORY_ID"
fi

run_test "PATCH /transaction-categories/{id} - admin update" \
    "PATCH" "$API/transaction-categories/$CATEGORY_ID" "200" \
    '{"name":"Updated Test Category"}' \
    "$ADMIN_TOKEN"

run_test "POST /transactions - create transaction" \
    "POST" "$API/transactions" "201" \
    '{"rbw_id":"'"$RBW_ID"'","category_id":"'"$CATEGORY_ID"'","amount":500000,"type":"income","description":"Test nest sale","transaction_date":"2026-02-09T00:00:00Z"}' \
    "$AUTH"

TRANSACTION_ID=$(extract_json "data.id")
echo "       -> Created Transaction ID: $TRANSACTION_ID"

run_test "GET /transactions/{id} - get transaction" \
    "GET" "$API/transactions/$TRANSACTION_ID" "200" "" "$AUTH"

run_test "PATCH /transactions/{id} - update transaction" \
    "PATCH" "$API/transactions/$TRANSACTION_ID" "200" \
    '{"amount":600000,"description":"Updated test sale"}' \
    "$AUTH"

run_test "GET /rbw/{id}/transactions - transactions by RBW" \
    "GET" "$API/rbw/$RBW_ID/transactions" "200" "" "$AUTH"

run_test "POST /financial-statements - generate report" \
    "POST" "$API/financial-statements" "200" \
    '{"rbw_id":"'"$RBW_ID"'","start_date":"2026-01-01","end_date":"2026-12-31"}' \
    "$AUTH"

# ================================================================
# 13. AI ENDPOINTS (may be disabled - accept 200 or 503)
# ================================================================
log_section "13. AI ENDPOINTS"

run_test_multi "GET /ai/health - AI health check" \
    "GET" "$API/ai/health" "200|503" "" "$AUTH"

AI_STATUS=$(cat /tmp/last_test_status.txt)
if [ "$AI_STATUS" = "503" ]; then
    AI_DISABLED=true
    echo -e "       ${YELLOW}-> AI Engine is disabled. AI tests will accept 503 as PASS.${NC}"
fi

run_test_multi "POST /ai/predict-grade" \
    "POST" "$API/ai/predict-grade" "200|503" \
    '{"temperature":28.5,"humidity":75.0,"ammonia":15.0}' \
    "$AUTH"

run_test_multi "POST /ai/predict-pump" \
    "POST" "$API/ai/predict-pump" "200|503" \
    '{"temperature":32.0,"humidity":65.0,"ammonia":20.0}' \
    "$AUTH"

run_test_multi "POST /ai/analyze" \
    "POST" "$API/ai/analyze" "200|503" \
    '{"sensor_type":"temp","values":[28.0,28.5,29.0,29.5,30.0]}' \
    "$AUTH"

run_test_multi "POST /ai/anomaly-detect" \
    "POST" "$API/ai/anomaly-detect" "200|503" \
    '{"sensor_id":"test","sensor_type":"temp","rbw_id":"test","node_id":"test","recorded_at":"2026-02-09T10:00:00Z","value":45.0}' \
    "$AUTH"

# ================================================================
# 14. UPLOAD ENDPOINTS
# ================================================================
log_section "14. UPLOAD ENDPOINTS"

TOTAL=$((TOTAL + 1))
STATUS=$(curl -s -o /tmp/test_response.json -w '%{http_code}' -X POST \
    -H "Authorization: Bearer $AUTH" \
    "$API/uploads/avatar" 2>/dev/null) || STATUS="000"

if [ "$STATUS" = "400" ] || [ "$STATUS" = "500" ] || [ "$STATUS" = "503" ]; then
    PASSED=$((PASSED + 1))
    printf "  %-4s %-60s [${GREEN}PASS${NC}] HTTP %s (expected error)\n" "$TOTAL." "POST /uploads/avatar - no file" "$STATUS"
    REPORT="$REPORT\n| $TOTAL | POST /uploads/avatar - no file | POST | 400/500/503 | $STATUS | PASS |"
else
    FAILED=$((FAILED + 1))
    printf "  %-4s %-60s [${RED}FAIL${NC}] HTTP %s\n" "$TOTAL." "POST /uploads/avatar - no file" "$STATUS"
    REPORT="$REPORT\n| $TOTAL | POST /uploads/avatar - no file | POST | 400/500/503 | $STATUS | FAIL |"
fi

# ================================================================
# 15. WEBSOCKET ENDPOINTS
# ================================================================
log_section "15. WEBSOCKET ENDPOINTS"

run_test "GET /ws/stats - websocket stats" \
    "GET" "$API/ws/stats" "200" "" "$AUTH"

# ================================================================
# 16. EDGE CASES & ERROR HANDLING
# ================================================================
log_section "16. EDGE CASES & ERROR HANDLING"

run_test_multi "GET /api/v1/nonexistent - 404/405" \
    "GET" "$API/nonexistent" "404|405"

run_test "DELETE /auth/login - method not allowed" \
    "DELETE" "$API/auth/login" "405"

# Malformed JSON
TOTAL=$((TOTAL + 1))
STATUS=$(curl -s -o /tmp/test_response.json -w '%{http_code}' -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $AUTH" \
    -d '{invalid json}' \
    "$API/rbw" 2>/dev/null) || STATUS="000"

if [ "$STATUS" = "400" ]; then
    PASSED=$((PASSED + 1))
    printf "  %-4s %-60s [${GREEN}PASS${NC}] HTTP %s\n" "$TOTAL." "POST /rbw - malformed JSON" "$STATUS"
    REPORT="$REPORT\n| $TOTAL | POST /rbw - malformed JSON | POST | 400 | $STATUS | PASS |"
else
    FAILED=$((FAILED + 1))
    printf "  %-4s %-60s [${RED}FAIL${NC}] Expected 400, Got %s\n" "$TOTAL." "POST /rbw - malformed JSON" "$STATUS"
    REPORT="$REPORT\n| $TOTAL | POST /rbw - malformed JSON | POST | 400 | $STATUS | FAIL |"
fi

run_test "POST /harvests - empty body (400)" \
    "POST" "$API/harvests" "400" '{}' "$AUTH"

run_test "GET /users/me - tampered token (401)" \
    "GET" "$API/users/me" "401" "" \
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZmFrZSIsInJvbGUiOiJhZG1pbiJ9.fake"

# ================================================================
# 17. CLEANUP - DELETE TEST DATA
# ================================================================
log_section "17. CLEANUP - DELETE TEST DATA"

if [ -n "$TRANSACTION_ID" ]; then
    run_test "DELETE /transactions/{id} - cleanup" \
        "DELETE" "$API/transactions/$TRANSACTION_ID" "200" "" "$AUTH"
else
    run_test "DELETE /transactions/{id} - nothing to delete" \
        "DELETE" "$API/transactions/00000000-0000-0000-0000-000000000000" "404" "" "$AUTH"
fi

if [ -n "$HARVEST_ID" ]; then
    run_test "DELETE /harvests/{id} - cleanup" \
        "DELETE" "$API/harvests/$HARVEST_ID" "200" "" "$AUTH"
else
    run_test "DELETE /harvests/{id} - nothing to delete" \
        "DELETE" "$API/harvests/00000000-0000-0000-0000-000000000000" "404" "" "$AUTH"
fi

if [ -n "$NEW_CAT_ID" ]; then
    run_test "DELETE /transaction-categories/{id} - cleanup" \
        "DELETE" "$API/transaction-categories/$NEW_CAT_ID" "200" "" "$ADMIN_TOKEN"
else
    run_test "DELETE /transaction-categories/{id} - nothing" \
        "DELETE" "$API/transaction-categories/00000000-0000-0000-0000-000000000000" "404" "" "$ADMIN_TOKEN"
fi

if [ -n "$NODE_ID" ]; then
    run_test "DELETE /nodes/{id} - cleanup" \
        "DELETE" "$API/nodes/$NODE_ID" "200" "" "$AUTH"
else
    run_test "DELETE /nodes/{id} - nothing to delete" \
        "DELETE" "$API/nodes/00000000-0000-0000-0000-000000000000" "404" "" "$AUTH"
fi

if [ -n "$RBW_ID" ]; then
    run_test "DELETE /rbw/{id} - cleanup" \
        "DELETE" "$API/rbw/$RBW_ID" "200" "" "$AUTH"
else
    run_test "DELETE /rbw/{id} - nothing to delete" \
        "DELETE" "$API/rbw/00000000-0000-0000-0000-000000000000" "404" "" "$AUTH"
fi

# ================================================================
# FINAL REPORT
# ================================================================
log_section "TEST RESULTS SUMMARY"

echo ""
echo -e "  ${BOLD}Total Tests:${NC}  $TOTAL"
echo -e "  ${GREEN}Passed:${NC}       $PASSED"
echo -e "  ${RED}Failed:${NC}       $FAILED"
echo ""

PASS_RATE=0
if [ $TOTAL -gt 0 ]; then
    PASS_RATE=$(echo "scale=1; $PASSED * 100 / $TOTAL" | bc)
    echo -e "  ${BOLD}Pass Rate:${NC}    ${PASS_RATE}%"
fi

if [ $FAILED -eq 0 ]; then
    echo -e "\n  ${GREEN}${BOLD}ALL $TOTAL TESTS PASSED!${NC}"
else
    echo -e "\n  ${RED}${BOLD}SOME TESTS FAILED ($FAILED / $TOTAL)${NC}"
fi

# Write detailed report
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REPORT_FILE="${REPORT_FILE:-$PROJECT_ROOT/TEST_REPORT.md}"
cat > "$REPORT_FILE" << ENDOFFILE
# SwiftLead API - Production Endpoint Test Report

**Target:** https://api.swiftlead.fuadfakhruz.com  
**Date:** $(date '+%Y-%m-%d %H:%M:%S %Z')  
**Environment:** Production

## Summary

| Metric | Count |
|--------|-------|
| Total Tests | $TOTAL |
| Passed | $PASSED |
| Failed | $FAILED |
| Pass Rate | ${PASS_RATE}% |

## Detailed Results

| # | Test Name | Method | Expected | Actual | Status |
|---|-----------|--------|----------|--------|--------|$(echo -e "$REPORT")

## Notes

- AI Engine tests accept both 200 (enabled) and 503 (disabled) as PASS
- Alert mark/resolve tests use non-existent UUID when no real alerts exist (expect 404)
- Admin credentials sourced from seed migration (admin@swiftlead.id)
- Upload test expects error response since no file is attached

---
*Generated automatically by SwiftLead API Test Suite*
ENDOFFILE

echo ""
echo -e "  ${BOLD}Report saved to:${NC} TEST_REPORT.md"
echo ""
