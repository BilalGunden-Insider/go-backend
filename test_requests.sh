#!/bin/bash

BASE="http://localhost:8080/api/v1"
ADMIN_TOKEN=""
USER1_TOKEN=""
USER2_TOKEN=""
USER1_ID=""
USER2_ID=""

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

CURL="curl -s --max-time 5 --connect-timeout 3"

pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
fail() { echo -e "${RED}[FAIL]${NC} $1: $2"; }
info() { echo -e "${YELLOW}[INFO]${NC} $1"; }
separator() { echo ""; echo "========== $1 =========="; sleep 0.5; }

separator "1. AUTH TESTS"

info "Register admin user"
RES=$($CURL -X POST "$BASE/auth/register" -H "Content-Type: application/json" \
  -d '{"username":"admin1","email":"admin1@test.com","password":"admin123","role":"admin"}')
echo "$RES" | grep -q '"id"' && pass "Admin registered" || fail "Admin register" "$RES"

info "Register user1"
RES=$($CURL -X POST "$BASE/auth/register" -H "Content-Type: application/json" \
  -d '{"username":"user1","email":"user1@test.com","password":"user123"}')
USER1_ID=$(echo "$RES" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
[ -n "$USER1_ID" ] && pass "User1 registered (ID: $USER1_ID)" || fail "User1 register" "$RES"

info "Register user2"
RES=$($CURL -X POST "$BASE/auth/register" -H "Content-Type: application/json" \
  -d '{"username":"user2","email":"user2@test.com","password":"user123"}')
USER2_ID=$(echo "$RES" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
[ -n "$USER2_ID" ] && pass "User2 registered (ID: $USER2_ID)" || fail "User2 register" "$RES"

info "Register duplicate username (should fail)"
RES=$($CURL -X POST "$BASE/auth/register" -H "Content-Type: application/json" \
  -d '{"username":"user1","email":"other@test.com","password":"user123"}')
echo "$RES" | grep -q '"error"' && pass "Duplicate username rejected" || fail "Duplicate check" "$RES"

info "Login admin"
RES=$($CURL -X POST "$BASE/auth/login" -H "Content-Type: application/json" \
  -d '{"email":"admin1@test.com","password":"admin123"}')
ADMIN_TOKEN=$(echo "$RES" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
[ -n "$ADMIN_TOKEN" ] && pass "Admin login OK" || fail "Admin login" "$RES"

info "Login user1"
RES=$($CURL -X POST "$BASE/auth/login" -H "Content-Type: application/json" \
  -d '{"email":"user1@test.com","password":"user123"}')
USER1_TOKEN=$(echo "$RES" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
[ -n "$USER1_TOKEN" ] && pass "User1 login OK" || fail "User1 login" "$RES"

info "Login user2"
RES=$($CURL -X POST "$BASE/auth/login" -H "Content-Type: application/json" \
  -d '{"email":"user2@test.com","password":"user123"}')
USER2_TOKEN=$(echo "$RES" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
[ -n "$USER2_TOKEN" ] && pass "User2 login OK" || fail "User2 login" "$RES"

info "Login with wrong password (should fail)"
RES=$($CURL -X POST "$BASE/auth/login" -H "Content-Type: application/json" \
  -d '{"email":"user1@test.com","password":"wrongpass"}')
echo "$RES" | grep -q '"error"' && pass "Wrong password rejected" || fail "Wrong password" "$RES"

separator "2. USER TESTS"

info "List users (admin)"
RES=$($CURL "$BASE/users" -H "Authorization: Bearer $ADMIN_TOKEN")
echo "$RES" | grep -q '"username"' && pass "List users OK" || fail "List users" "$RES"

info "List users (non-admin, should fail)"
RES=$($CURL "$BASE/users" -H "Authorization: Bearer $USER1_TOKEN")
echo "$RES" | grep -q '"error"' && pass "Non-admin list rejected" || fail "Auth check" "$RES"

info "Get own profile (user1)"
RES=$($CURL "$BASE/users/$USER1_ID" -H "Authorization: Bearer $USER1_TOKEN")
echo "$RES" | grep -q '"username":"user1"' && pass "Get own profile OK" || fail "Get profile" "$RES"

info "Access without token (should fail)"
RES=$($CURL "$BASE/users/$USER1_ID")
echo "$RES" | grep -q '"error"' && pass "No token rejected" || fail "No token" "$RES"

separator "3. CREDIT / DEBIT TESTS"

info "Credit 5000 to user1 (admin)"
RES=$($CURL -X POST "$BASE/transactions/credit" -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" -d "{\"user_id\":\"$USER1_ID\",\"amount\":\"5000\"}")
echo "$RES" | grep -q '"completed"' && pass "Credit 5000 to user1 OK" || fail "Credit" "$RES"

info "Credit 3000 to user2 (admin)"
RES=$($CURL -X POST "$BASE/transactions/credit" -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" -d "{\"user_id\":\"$USER2_ID\",\"amount\":\"3000\"}")
echo "$RES" | grep -q '"completed"' && pass "Credit 3000 to user2 OK" || fail "Credit" "$RES"

info "Credit negative amount (should fail)"
RES=$($CURL -X POST "$BASE/transactions/credit" -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" -d "{\"user_id\":\"$USER1_ID\",\"amount\":\"-100\"}")
echo "$RES" | grep -q '"error"' && pass "Negative amount rejected" || fail "Negative check" "$RES"

info "Debit 1000 from user1 (admin)"
RES=$($CURL -X POST "$BASE/transactions/debit" -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" -d "{\"user_id\":\"$USER1_ID\",\"amount\":\"1000\"}")
echo "$RES" | grep -q '"completed"' && pass "Debit 1000 from user1 OK" || fail "Debit" "$RES"

info "Debit more than balance (should fail)"
RES=$($CURL -X POST "$BASE/transactions/debit" -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" -d "{\"user_id\":\"$USER1_ID\",\"amount\":\"99999\"}")
echo "$RES" | grep -q '"error"' && pass "Insufficient funds rejected" || fail "Insufficient funds" "$RES"

separator "4. BALANCE & CACHE TESTS"

info "Get user1 balance (should be 4000)"
RES=$($CURL "$BASE/balances/$USER1_ID" -H "Authorization: Bearer $USER1_TOKEN")
echo "$RES" | grep -q '"4000"' && pass "User1 balance = 4000" || fail "Balance check" "$RES"

info "Get user2 balance (should be 3000)"
RES=$($CURL "$BASE/balances/$USER2_ID" -H "Authorization: Bearer $USER2_TOKEN")
echo "$RES" | grep -q '"3000"' && pass "User2 balance = 3000" || fail "Balance check" "$RES"

info "Verify Redis cache (direct redis check)"
REDIS_VAL=$(docker exec $(docker ps -qf "ancestor=redis:7-alpine") redis-cli GET "balance:$USER1_ID" 2>/dev/null)
[ "$REDIS_VAL" = "4000" ] && pass "Redis cache = 4000" || info "Redis check skipped or value: $REDIS_VAL"

separator "5. TRANSFER TESTS"

info "Transfer 500 from user1 to user2"
RES=$($CURL -X POST "$BASE/transactions/transfer" -H "Authorization: Bearer $USER1_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"from_user_id\":\"$USER1_ID\",\"to_user_id\":\"$USER2_ID\",\"amount\":\"500\"}")
echo "$RES" | grep -q '"completed"' && pass "Transfer 500 OK" || fail "Transfer" "$RES"

info "Check user1 balance after transfer (should be 3500)"
RES=$($CURL "$BASE/balances/$USER1_ID" -H "Authorization: Bearer $USER1_TOKEN")
echo "$RES" | grep -q '"3500"' && pass "User1 balance = 3500" || fail "Balance" "$RES"

info "Check user2 balance after transfer (should be 3500)"
RES=$($CURL "$BASE/balances/$USER2_ID" -H "Authorization: Bearer $USER2_TOKEN")
echo "$RES" | grep -q '"3500"' && pass "User2 balance = 3500" || fail "Balance" "$RES"

separator "6. TRANSACTION LIMIT TESTS"

info "Credit 15000 to user1 (exceeds per-transaction limit of 10000, should fail)"
RES=$($CURL -X POST "$BASE/transactions/credit" -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" -d "{\"user_id\":\"$USER1_ID\",\"amount\":\"15000\"}")
echo "$RES" | grep -q "exceeds per-transaction limit" && pass "Per-transaction limit enforced" || fail "Limit check" "$RES"

info "Credit 9000 to user1 (within per-txn limit)"
RES=$($CURL -X POST "$BASE/transactions/credit" -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" -d "{\"user_id\":\"$USER1_ID\",\"amount\":\"9000\"}")
echo "$RES" | grep -q '"completed"' && pass "Credit 9000 OK" || fail "Credit" "$RES"

separator "7. TRANSACTION LIST & DETAIL TESTS"

info "List user1 transactions"
RES=$($CURL "$BASE/transactions?limit=5" -H "Authorization: Bearer $USER1_TOKEN")
echo "$RES" | grep -q '"id"' && pass "List transactions OK" || fail "List transactions" "$RES"
TX_ID=$(echo "$RES" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
info "First transaction ID: $TX_ID"

info "Get transaction detail"
RES=$($CURL "$BASE/transactions/$TX_ID" -H "Authorization: Bearer $USER1_TOKEN")
echo "$RES" | grep -q '"amount"' && pass "Get transaction detail OK" || fail "Detail" "$RES"

info "List with pagination (offset=2, limit=2)"
RES=$($CURL "$BASE/transactions?limit=2&offset=2" -H "Authorization: Bearer $USER1_TOKEN")
echo "$RES" | grep -q '"id"' && pass "Pagination OK" || info "No more transactions at offset=2"

separator "8. ROLLBACK TEST"

info "Rollback last transaction (admin)"
HTTP_CODE=$($CURL -o /dev/null -w "%{http_code}" -X POST "$BASE/transactions/$TX_ID/rollback" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
[ "$HTTP_CODE" = "204" ] && pass "Rollback OK (HTTP 204)" || fail "Rollback" "HTTP $HTTP_CODE"

info "Rollback already rolled back (should fail)"
RES=$($CURL -X POST "$BASE/transactions/$TX_ID/rollback" -H "Authorization: Bearer $ADMIN_TOKEN")
echo "$RES" | grep -q '"error"' && pass "Double rollback rejected" || fail "Double rollback" "$RES"

separator "9. SCHEDULED TRANSACTION TESTS"

FUTURE=$(date -u -v+2M +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+2 minutes" +"%Y-%m-%dT%H:%M:%SZ")
info "Schedule a transfer for $FUTURE"
RES=$($CURL -X POST "$BASE/scheduled-transactions" -H "Authorization: Bearer $USER1_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"from_user_id\":\"$USER1_ID\",\"to_user_id\":\"$USER2_ID\",\"amount\":\"100\",\"type\":\"transfer\",\"scheduled_at\":\"$FUTURE\"}")
echo "$RES" | grep -q '"pending"' && pass "Scheduled transaction created" || fail "Schedule" "$RES"
SCHED_ID=$(echo "$RES" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
info "Scheduled ID: $SCHED_ID"

info "Schedule with past date (should fail)"
RES=$($CURL -X POST "$BASE/scheduled-transactions" -H "Authorization: Bearer $USER1_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"from_user_id\":\"$USER1_ID\",\"to_user_id\":\"$USER2_ID\",\"amount\":\"100\",\"type\":\"transfer\",\"scheduled_at\":\"2020-01-01T00:00:00Z\"}")
echo "$RES" | grep -q '"error"' && pass "Past date rejected" || fail "Past date" "$RES"

info "Schedule with invalid type (should fail)"
RES=$($CURL -X POST "$BASE/scheduled-transactions" -H "Authorization: Bearer $USER1_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"from_user_id\":\"$USER1_ID\",\"to_user_id\":\"$USER2_ID\",\"amount\":\"100\",\"type\":\"invalid\",\"scheduled_at\":\"$FUTURE\"}")
echo "$RES" | grep -q '"error"' && pass "Invalid type rejected" || fail "Invalid type" "$RES"

info "List scheduled transactions"
RES=$($CURL "$BASE/scheduled-transactions" -H "Authorization: Bearer $USER1_TOKEN")
echo "$RES" | grep -q '"id"' && pass "List scheduled OK" || fail "List scheduled" "$RES"

info "Get scheduled transaction detail"
RES=$($CURL "$BASE/scheduled-transactions/$SCHED_ID" -H "Authorization: Bearer $USER1_TOKEN")
echo "$RES" | grep -q '"pending"' && pass "Get scheduled detail OK" || fail "Get scheduled" "$RES"

info "Cancel scheduled transaction"
HTTP_CODE=$($CURL -o /dev/null -w "%{http_code}" -X POST "$BASE/scheduled-transactions/$SCHED_ID/cancel" \
  -H "Authorization: Bearer $USER1_TOKEN")
[ "$HTTP_CODE" = "204" ] && pass "Cancel scheduled OK" || fail "Cancel scheduled" "HTTP $HTTP_CODE"

info "Cancel already cancelled (should fail)"
RES=$($CURL -X POST "$BASE/scheduled-transactions/$SCHED_ID/cancel" -H "Authorization: Bearer $USER1_TOKEN")
echo "$RES" | grep -q '"error"' && pass "Double cancel rejected" || fail "Double cancel" "$RES"

separator "10. BATCH TRANSACTION TESTS"

info "Batch: 3 valid credits"
RES=$($CURL -X POST "$BASE/transactions/batch" -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" -d "{\"transactions\":[
    {\"type\":\"credit\",\"to_user_id\":\"$USER1_ID\",\"amount\":\"100\"},
    {\"type\":\"credit\",\"to_user_id\":\"$USER2_ID\",\"amount\":\"200\"},
    {\"type\":\"credit\",\"to_user_id\":\"$USER1_ID\",\"amount\":\"300\"}
  ]}")
echo "$RES" | grep -q '"accepted":3' && pass "Batch 3/3 accepted" || fail "Batch 3 valid" "$RES"

info "Batch: mix of valid and invalid"
RES=$($CURL -X POST "$BASE/transactions/batch" -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" -d "{\"transactions\":[
    {\"type\":\"credit\",\"to_user_id\":\"$USER1_ID\",\"amount\":\"100\"},
    {\"type\":\"credit\",\"to_user_id\":\"$USER1_ID\",\"amount\":\"-50\"},
    {\"type\":\"invalid\",\"to_user_id\":\"$USER1_ID\",\"amount\":\"100\"}
  ]}")
echo "$RES" | grep -q '"rejected"' && pass "Batch partial reject OK" || fail "Batch mix" "$RES"
info "Batch result: $RES"
sleep 0.3

info "Batch: empty list (should fail)"
RES=$($CURL -X POST "$BASE/transactions/batch" -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" -d '{"transactions":[]}')
echo "$RES" | grep -q '"error"' && pass "Empty batch rejected" || fail "Empty batch" "$RES"
sleep 0.3

info "Batch: non-admin (should fail)"
RES=$($CURL -X POST "$BASE/transactions/batch" -H "Authorization: Bearer $USER1_TOKEN" \
  -H "Content-Type: application/json" -d "{\"transactions\":[
    {\"type\":\"credit\",\"to_user_id\":\"$USER1_ID\",\"amount\":\"100\"}
  ]}")
echo "$RES" | grep -q '"error"' && pass "Non-admin batch rejected" || fail "Auth batch" "$RES"

separator "11. WORKER STATS TEST"

info "Get worker stats (admin)"
RES=$($CURL "http://localhost:8080/api/v1/workers/stats" -H "Authorization: Bearer $ADMIN_TOKEN")
echo "$RES" | grep -q '"processed"' && pass "Worker stats OK: $RES" || fail "Worker stats" "$RES"

info "Get worker stats (non-admin, should fail)"
RES=$($CURL "http://localhost:8080/api/v1/workers/stats" -H "Authorization: Bearer $USER1_TOKEN")
echo "$RES" | grep -q '"error"' && pass "Non-admin worker stats rejected" || fail "Auth stats" "$RES"

separator "12. BALANCE AT (HISTORICAL) TEST"

NOW=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
info "Get user1 balance at now"
RES=$($CURL "$BASE/balances/$USER1_ID/at?time=$NOW" -H "Authorization: Bearer $USER1_TOKEN")
echo "$RES" | grep -q '"balance"' && pass "Historical balance OK: $RES" || fail "Historical balance" "$RES"

separator "13. EDGE CASES"

info "Invalid UUID in path"
RES=$($CURL "$BASE/users/not-a-uuid" -H "Authorization: Bearer $ADMIN_TOKEN")
echo "$RES" | grep -q '"error"' && pass "Invalid UUID rejected" || fail "UUID check" "$RES"

info "Missing required fields"
RES=$($CURL -X POST "$BASE/transactions/credit" -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" -d '{}')
echo "$RES" | grep -q '"error"' && pass "Missing fields rejected" || fail "Missing fields" "$RES"

info "Invalid JSON body"
RES=$($CURL -X POST "$BASE/auth/login" -H "Content-Type: application/json" -d 'not json')
echo "$RES" | grep -q '"error"' && pass "Invalid JSON rejected" || fail "Invalid JSON" "$RES"

info "Expired/invalid token"
RES=$($CURL "$BASE/users" -H "Authorization: Bearer invalidtoken123")
echo "$RES" | grep -q '"error"' && pass "Invalid token rejected" || fail "Token check" "$RES"

separator "DONE"
echo ""
echo "Test completed. Review results above."
echo "Note: Batch transactions are async - worker stats may take a moment to update."
echo "Note: Scheduled transactions execute every 30s - wait and check balances to verify."
