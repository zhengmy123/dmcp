#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:18080}"
VAUTH_KEY="${VAUTH_KEY:-user-service}"
ENDPOINT="${BASE_URL}/mcp/${VAUTH_KEY}"

echo "MCP endpoint: ${ENDPOINT}"

INIT_HEADERS="$(mktemp)"
INIT_BODY="$(mktemp)"
TOOLS_LIST_BODY="$(mktemp)"
TOOLS_CALL_BODY="$(mktemp)"
trap 'rm -f "${INIT_HEADERS}" "${INIT_BODY}" "${TOOLS_LIST_BODY}" "${TOOLS_CALL_BODY}"' EXIT

echo "1) initialize"
curl -sS -D "${INIT_HEADERS}" -o "${INIT_BODY}" \
  -X POST "${ENDPOINT}" \
  -H "Content-Type: application/json" \
  --data-binary '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {
        "name": "curl-validator",
        "version": "0.1.0"
      }
    }
  }'

SESSION_ID="$(
  awk 'BEGIN{IGNORECASE=1} /^mcp-session-id:/{gsub(/\r$/,"",$2); print $2}' "${INIT_HEADERS}"
)"
if [[ -n "${SESSION_ID}" ]]; then
  echo "session id: ${SESSION_ID}"
else
  echo "session id not returned (stateless mode), continue without header"
fi
python -m json.tool "${INIT_BODY}"

echo "2) notifications/initialized"
if [[ -n "${SESSION_ID}" ]]; then
  curl -sS \
    -X POST "${ENDPOINT}" \
    -H "Content-Type: application/json" \
    -H "Mcp-Session-Id: ${SESSION_ID}" \
    --data-binary '{
      "jsonrpc": "2.0",
      "method": "notifications/initialized"
    }' >/dev/null
else
  curl -sS \
    -X POST "${ENDPOINT}" \
    -H "Content-Type: application/json" \
    --data-binary '{
      "jsonrpc": "2.0",
      "method": "notifications/initialized"
    }' >/dev/null
fi

echo "3) tools/list"
if [[ -n "${SESSION_ID}" ]]; then
  curl -sS -o "${TOOLS_LIST_BODY}" \
    -X POST "${ENDPOINT}" \
    -H "Content-Type: application/json" \
    -H "Mcp-Session-Id: ${SESSION_ID}" \
    --data-binary '{
      "jsonrpc": "2.0",
      "id": 2,
      "method": "tools/list",
      "params": {}
    }'
else
  curl -sS -o "${TOOLS_LIST_BODY}" \
    -X POST "${ENDPOINT}" \
    -H "Content-Type: application/json" \
    --data-binary '{
      "jsonrpc": "2.0",
      "id": 2,
      "method": "tools/list",
      "params": {}
    }'
fi
python -m json.tool "${TOOLS_LIST_BODY}"

echo "4) tools/call search_users"
if [[ -n "${SESSION_ID}" ]]; then
  curl -sS -o "${TOOLS_CALL_BODY}" \
    -X POST "${ENDPOINT}" \
    -H "Content-Type: application/json" \
    -H "Mcp-Session-Id: ${SESSION_ID}" \
    --data-binary '{
      "jsonrpc": "2.0",
      "id": 3,
      "method": "tools/call",
      "params": {
        "name": "search_users",
        "arguments": {
          "query": "alice",
          "limit": 3
        }
      }
    }'
else
  curl -sS -o "${TOOLS_CALL_BODY}" \
    -X POST "${ENDPOINT}" \
    -H "Content-Type: application/json" \
    --data-binary '{
      "jsonrpc": "2.0",
      "id": 3,
      "method": "tools/call",
      "params": {
        "name": "search_users",
        "arguments": {
          "query": "alice",
          "limit": 3
        }
      }
    }'
fi
python -m json.tool "${TOOLS_CALL_BODY}"

echo "Validation completed."
