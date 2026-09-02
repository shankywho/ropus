#!/usr/bin/env bash
# ==============================================================================
# ROPUS — Razorpay AI Buildathon (Track 02: AI Risk Manager)
# Canonical Signed Webhook & Risk Evaluation Demonstration Script
# ==============================================================================
set -euo pipefail

BOLD="\033[1m"
GREEN="\033[32m"
RED="\033[31m"
YELLOW="\033[33m"
CYAN="\033[36m"
BLUE="\033[34m"
RESET="\033[0m"

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
WEBHOOK_SECRET="${RAZORPAY_WEBHOOK_SECRET:-rzp_test_sec_buildathon_2026}"
WEBHOOK_ENDPOINT="${API_BASE_URL}/v1/webhooks/razorpay"
AUDIT_ENDPOINT="${API_BASE_URL}/v1/audit/verify"

print_header() {
    echo -e "\n${BOLD}${CYAN}==============================================================================${RESET}"
    echo -e "${BOLD}${CYAN}  ROPUS — AI Risk Manager: Track 02 Buildathon Live Demonstration${RESET}"
    echo -e "${BOLD}${CYAN}  Target: ${WEBHOOK_ENDPOINT}${RESET}"
    echo -e "${BOLD}${CYAN}==============================================================================${RESET}\n"
}

compute_signature() {
    local payload="$1"
    local secret="$2"
    python3 -c "
import hmac, hashlib, sys
payload = '''$payload'''.encode('utf-8')
secret = '''$secret'''.encode('utf-8')
print(hmac.new(secret, payload, hashlib.sha256).hexdigest())
"
}

print_header

# ------------------------------------------------------------------------------
# STEP 1: Health / Reachability Check
# ------------------------------------------------------------------------------
echo -e "${BOLD}${BLUE}[STEP 1/6] Checking stack reachability at ${API_BASE_URL}...${RESET}"
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 2 "${API_BASE_URL}/health" 2>/dev/null || echo "000")

if [ "$HTTP_STATUS" -ne 200 ]; then
    echo -e "${RED}✗ Stack is not reachable at ${API_BASE_URL} (HTTP Status: ${HTTP_STATUS}).${RESET}"
    echo -e "${YELLOW}  Please ensure the service is running (e.g. via 'make up' or 'docker compose up -d') before running this demo.${RESET}"
    exit 1
fi
HEALTH_RESP=$(curl -s "${API_BASE_URL}/health")
echo -e "${GREEN}✓ Stack is reachable (HTTP 200). Health: ${HEALTH_RESP}${RESET}"

# ------------------------------------------------------------------------------
# STEP 2: Scenario A — Low-Risk Domestic Payment (Valid Signature)
# ------------------------------------------------------------------------------
echo -e "\n${BOLD}${BLUE}[STEP 2/6] Scenario A: Normal Domestic Payment (Valid Signed Webhook)${RESET}"
LOW_RISK_PAYLOAD='{
  "entity": "event",
  "account_id": "acc_RazorpayMerchant01",
  "event": "payment.authorized",
  "created_at": 1725200000,
  "payload": {
    "payment": {
      "entity": {
        "id": "pay_DemoLowRisk_001",
        "entity": "payment",
        "amount": 48000,
        "currency": "INR",
        "status": "authorized",
        "order_id": "order_Norm8391823",
        "international": false,
        "method": "card",
        "email": "gaurav.kumar@example.com",
        "contact": "+919876543210",
        "card": {
          "id": "card_98127391",
          "name": "Gaurav Kumar",
          "last4": "4321",
          "network": "Visa",
          "type": "debit",
          "issuer": "HDFC"
        }
      }
    }
  }
}'

LOW_SIG=$(compute_signature "$LOW_RISK_PAYLOAD" "$WEBHOOK_SECRET")

echo -e "  Sending 'payment.authorized' (₹480.00 domestic card)..."
LOW_RAW=$(curl -s -w "\n%{http_code}" -X POST "${WEBHOOK_ENDPOINT}" \
    -H "Content-Type: application/json" \
    -H "X-Razorpay-Signature: ${LOW_SIG}" \
    -d "$LOW_RISK_PAYLOAD")

HTTP_CODE=$(echo "$LOW_RAW" | tail -n 1)
BODY=$(echo "$LOW_RAW" | sed '$d')

if [ "$HTTP_CODE" -ne 200 ]; then
    echo -e "${RED}✗ Expected HTTP 200, got ${HTTP_CODE}${RESET}: $BODY"
    exit 1
fi

ACTION=$(python3 -c "import json; print(json.loads('''$BODY''').get('recommended_action', ''))")
SCORE=$(python3 -c "import json; print(json.loads('''$BODY''').get('risk_score', 0))")
DECISION_ID=$(python3 -c "import json; print(json.loads('''$BODY''').get('decision_id', ''))")

echo -e "${GREEN}✓ Scenario A Processed: HTTP 200 OK${RESET}"
echo -e "  • Decision ID:        ${BOLD}${DECISION_ID}${RESET}"
echo -e "  • Recommended Action: ${BOLD}${GREEN}${ACTION}${RESET}"
echo -e "  • Risk Score:         ${BOLD}${SCORE} / 100${RESET}"

if [[ "$ACTION" != *"ALLOW"* ]]; then
    echo -e "${RED}✗ Scenario A failed: Expected ALLOW action for domestic payment, got ${ACTION}${RESET}"
    exit 1
fi

# ------------------------------------------------------------------------------
# STEP 3: Scenario B — High-Risk Cross-Border Ingress (Valid Signature)
# ------------------------------------------------------------------------------
echo -e "\n${BOLD}${BLUE}[STEP 3/6] Scenario B: High-Risk Attack Ingress (Valid Signed Webhook)${RESET}"
HIGH_RISK_PAYLOAD='{
  "entity": "event",
  "account_id": "acc_RazorpayMerchant01",
  "event": "payment.authorized",
  "created_at": 1725200000,
  "payload": {
    "payment": {
      "entity": {
        "id": "pay_DemoAttackWhale_002",
        "entity": "payment",
        "amount": 15000000,
        "currency": "INR",
        "status": "authorized",
        "order_id": "order_Atk998811",
        "international": true,
        "method": "card",
        "email": "bot_attacker_01@guerrillamail.com",
        "contact": "+12025550199",
        "notes": {
          "ip_address": "185.220.101.5",
          "device_id": "dev_headless_puppet_v4"
        },
        "card": {
          "id": "card_99998888",
          "name": "Stolen Cardholder",
          "last4": "9999",
          "network": "Visa",
          "type": "credit",
          "issuer": "UNKNOWN",
          "international": true
        }
      }
    }
  }
}'

HIGH_SIG=$(compute_signature "$HIGH_RISK_PAYLOAD" "$WEBHOOK_SECRET")

echo -e "  Sending 'payment.authorized' (₹1,50,000.00 international card, disposable domain)..."
HIGH_RAW=$(curl -s -w "\n%{http_code}" -X POST "${WEBHOOK_ENDPOINT}" \
    -H "Content-Type: application/json" \
    -H "X-Razorpay-Signature: ${HIGH_SIG}" \
    -d "$HIGH_RISK_PAYLOAD")

HTTP_CODE=$(echo "$HIGH_RAW" | tail -n 1)
BODY=$(echo "$HIGH_RAW" | sed '$d')

if [ "$HTTP_CODE" -ne 200 ]; then
    echo -e "${RED}✗ Expected HTTP 200, got ${HTTP_CODE}${RESET}: $BODY"
    exit 1
fi

HIGH_ACTION=$(python3 -c "import json; print(json.loads('''$BODY''').get('recommended_action', ''))")
HIGH_SCORE=$(python3 -c "import json; print(json.loads('''$BODY''').get('risk_score', 0))")
HIGH_REASONS=$(python3 -c "import json; print(json.loads('''$BODY''').get('reason_codes', []))")

echo -e "${GREEN}✓ Scenario B Processed: HTTP 200 OK${RESET}"
echo -e "  • Recommended Action: ${BOLD}${RED}${HIGH_ACTION}${RESET}"
echo -e "  • Risk Score:         ${BOLD}${HIGH_SCORE} / 100${RESET}"
echo -e "  • Reasons Extracted:  ${BOLD}${HIGH_REASONS}${RESET}"

# High risk transaction must NOT produce an unsafe ALLOW
if [[ "$HIGH_ACTION" == "ALLOW" || "$HIGH_ACTION" == "ALLOW_RECOMMENDATION" ]]; then
    echo -e "${RED}✗ CRITICAL FAILURE: High-risk attack was unsafely allowed!${RESET}"
    exit 1
fi
echo -e "  • Defensive Verdict:  ${BOLD}Safely intercepted with ${HIGH_ACTION}${RESET}"

# ------------------------------------------------------------------------------
# STEP 4: Scenario C — Forged Signature Rejection
# ------------------------------------------------------------------------------
echo -e "\n${BOLD}${BLUE}[STEP 4/6] Scenario C: Forged / Invalid Webhook Signature Rejection${RESET}"
echo -e "  Sending webhook payload with forged signature 'deadbeef_tampered_signature_9988'..."

TAMPER_RAW=$(curl -s -w "\n%{http_code}" -X POST "${WEBHOOK_ENDPOINT}" \
    -H "Content-Type: application/json" \
    -H "X-Razorpay-Signature: deadbeef_tampered_signature_9988" \
    -d "$LOW_RISK_PAYLOAD")

HTTP_CODE=$(echo "$TAMPER_RAW" | tail -n 1)
BODY=$(echo "$TAMPER_RAW" | sed '$d')

if [ "$HTTP_CODE" -eq 401 ]; then
    echo -e "${GREEN}✓ Rejection Verified: HTTP 401 Unauthorized (Signature Validation Enforced)${RESET}"
    echo -e "  Response: $BODY"
else
    echo -e "${RED}✗ Expected HTTP 401 for forged signature, got ${HTTP_CODE}${RESET}: $BODY"
    exit 1
fi

# ------------------------------------------------------------------------------
# STEP 5: Scenario D — Idempotent Webhook Replay Protection
# ------------------------------------------------------------------------------
echo -e "\n${BOLD}${BLUE}[STEP 5/6] Scenario D: Idempotent Webhook Replay Protection${RESET}"
echo -e "  Re-sending identical Scenario A webhook..."

REPLAY_HEADER_RESP=$(curl -s -i -X POST "${WEBHOOK_ENDPOINT}" \
    -H "Content-Type: application/json" \
    -H "X-Razorpay-Signature: ${LOW_SIG}" \
    -d "$LOW_RISK_PAYLOAD")

HTTP_STATUS=$(echo "$REPLAY_HEADER_RESP" | head -n 1 | awk '{print $2}')
if [ "$HTTP_STATUS" -ne 200 ]; then
    echo -e "${RED}✗ Expected HTTP 200 on idempotent replay, got ${HTTP_STATUS}${RESET}"
    exit 1
fi

if echo "$REPLAY_HEADER_RESP" | grep -qi "X-ROPUS-Idempotent-Replayed"; then
    echo -e "${GREEN}✓ Idempotent Replay Verified: 'X-ROPUS-Idempotent-Replayed: true' header detected.${RESET}"
else
    echo -e "${RED}✗ Expected 'X-ROPUS-Idempotent-Replayed: true' header on replay!${RESET}"
    exit 1
fi

# ------------------------------------------------------------------------------
# STEP 6: Scenario E — Cryptographic SHA-256 Decision Audit Trail Verification
# ------------------------------------------------------------------------------
echo -e "\n${BOLD}${BLUE}[STEP 6/6] Scenario E: SHA-256 Audit Trail Cryptographic Verification${RESET}"
echo -e "  Querying ${AUDIT_ENDPOINT}..."

AUDIT_RAW=$(curl -s -w "\n%{http_code}" "${AUDIT_ENDPOINT}")
HTTP_CODE=$(echo "$AUDIT_RAW" | tail -n 1)
AUDIT_BODY=$(echo "$AUDIT_RAW" | sed '$d')

if [ "$HTTP_CODE" -ne 200 ]; then
    echo -e "${RED}✗ Expected HTTP 200 from audit endpoint, got ${HTTP_CODE}${RESET}: $AUDIT_BODY"
    exit 1
fi

AUDIT_STATUS=$(python3 -c "import json; print(json.loads('''$AUDIT_BODY''').get('status', ''))")
INTEGRITY=$(python3 -c "import json; print(json.loads('''$AUDIT_BODY''').get('integrity_verified', False))")
HEAD_HASH=$(python3 -c "import json; print(json.loads('''$AUDIT_BODY''').get('head_hash', ''))")
TOTAL_AUDITED=$(python3 -c "import json; print(json.loads('''$AUDIT_BODY''').get('total_decisions_audited', 0))")

echo -e "  • Audit Status:       ${BOLD}${AUDIT_STATUS}${RESET}"
echo -e "  • Integrity Verified: ${BOLD}${INTEGRITY}${RESET}"
echo -e "  • Decisions Audited:  ${BOLD}${TOTAL_AUDITED}${RESET}"
echo -e "  • Head Merkle Hash:   ${BOLD}${HEAD_HASH}${RESET}"

if [ "$INTEGRITY" != "True" ] && [ "$INTEGRITY" != "true" ]; then
    echo -e "${RED}✗ Audit integrity verification failed! (integrity_verified: ${INTEGRITY})${RESET}"
    exit 1
fi
echo -e "${GREEN}✓ Cryptographic SHA-256 Audit Ledger verified successfully.${RESET}"

echo -e "\n${BOLD}${GREEN}==============================================================================${RESET}"
echo -e "${BOLD}${GREEN}  ✓ ALL BUILDATHON DEMONSTRATION CRITERIA PASSED AND VERIFIED!${RESET}"
echo -e "${BOLD}${GREEN}  Defense Loop: Ingress Auth → Feature Normalization → Calibrated BMR → Audit Chain${RESET}"
echo -e "${BOLD}${GREEN}==============================================================================${RESET}\n"
