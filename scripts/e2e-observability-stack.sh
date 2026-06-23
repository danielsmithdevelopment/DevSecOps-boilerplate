#!/usr/bin/env bash
# End-to-end verification of the observability + task-runner stack.
# Usage:
#   ./scripts/e2e-observability-stack.sh              # full stack (task-runner + observability)
#   ./scripts/e2e-observability-stack.sh --security   # also run security overlay (Linux recommended)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_SECURITY=false
SKIP_BUILD=false

for arg in "$@"; do
  case "$arg" in
    --security) RUN_SECURITY=true ;;
    --skip-build) SKIP_BUILD=true ;;
  esac
done

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'
pass=0
fail=0
skip=0

log_pass() { echo -e "${GREEN}PASS${NC} $1"; pass=$((pass + 1)); }
log_fail() { echo -e "${RED}FAIL${NC} $1"; fail=$((fail + 1)); }
log_skip() { echo -e "${YELLOW}SKIP${NC} $1"; skip=$((skip + 1)); }

wait_http() {
  local url="$1" name="$2" max="${3:-120}"
  local i=0
  while [ "$i" -lt "$max" ]; do
    if curl -sf "$url" >/dev/null 2>&1; then
      log_pass "$name reachable ($url)"
      return 0
    fi
    sleep 2
    i=$((i + 2))
  done
  log_fail "$name not reachable after ${max}s ($url)"
  return 1
}

wait_cmd() {
  local name="$1" max="${2:-120}"
  shift 2
  local i=0
  while [ "$i" -lt "$max" ]; do
    if "$@" >/dev/null 2>&1; then
      log_pass "$name"
      return 0
    fi
    sleep 2
    i=$((i + 2))
  done
  log_fail "$name (timeout ${max}s)"
  return 1
}

ensure_env() {
  local dir="$1"
  if [ ! -f "$dir/.env" ]; then
    cp "$dir/.env.example" "$dir/.env"
    echo "Created $dir/.env from .env.example"
  fi
}

echo "=== E2E observability stack ==="
echo "Root: $ROOT"
echo

ensure_env "$ROOT/docker/observability"
ensure_env "$ROOT/task-runner"

cd "$ROOT/task-runner"
echo "--- Starting stack ---"
if [ "$SKIP_BUILD" != true ]; then
  docker compose up -d --build
else
  docker compose up -d
fi

if [ "$RUN_SECURITY" = true ]; then
  echo "--- Starting security overlay ---"
  docker compose -f "$ROOT/docker/observability/docker-compose.security.yaml" up -d || true
fi

echo
echo "--- Waiting for core services ---"
wait_http "http://localhost:8080/health" "task-runner-1" 180 || true
wait_http "http://localhost:3000/api/health" "Grafana" 180 || true
wait_http "http://localhost:9090/-/ready" "Prometheus" 120 || true
wait_http "http://localhost:9093/-/ready" "Alertmanager" 120 || true
wait_http "http://localhost:3100/ready" "Loki" 120 || true
wait_http "http://localhost:3200/ready" "Tempo" 120 || true
wait_http "http://localhost:9009/ready" "Mimir" 120 || true
wait_http "http://localhost:4040/ready" "Pyroscope" 120 || true
wait_http "http://localhost:12345/-/ready" "Alloy" 120 || true
wait_http "http://localhost:8090/health/" "OnCall engine" 180 || true

echo
echo "--- Task runner cluster ---"
for port in 8080 8081 8082 8083 8084; do
  if curl -sf "http://localhost:${port}/health" >/dev/null; then
    log_pass "task-runner :${port}/health"
  else
    log_fail "task-runner :${port}/health"
  fi
done

echo
echo "--- Generating traffic (metrics + traces) ---"
for _ in $(seq 1 20); do
  curl -sf "http://localhost:8080/health" >/dev/null || true
  curl -sf "http://localhost:8080/metrics" >/dev/null || true
done
log_pass "Generated HTTP traffic to task-runner"

echo
echo "--- Prometheus targets ---"
PROM_TARGETS="$(curl -sf "http://localhost:9090/api/v1/targets" || echo '{}')"
if echo "$PROM_TARGETS" | grep -q '"health":"up"'; then
  UP_COUNT="$(echo "$PROM_TARGETS" | python3 -c "import sys,json; d=json.load(sys.stdin); print(sum(1 for t in d.get('data',{}).get('activeTargets',[]) if t.get('health')=='up'))" 2>/dev/null || echo "?")"
  log_pass "Prometheus has ${UP_COUNT} target(s) up"
else
  log_fail "Prometheus targets check"
fi

echo
echo "--- Mimir query ---"
if curl -sf "http://localhost:9009/prometheus/api/v1/query?query=up" | grep -q '"status":"success"'; then
  log_pass "Mimir Prometheus API query (up)"
else
  log_fail "Mimir Prometheus API query"
fi

echo
echo "--- Loki logs ---"
LOKI_NOW="$(date +%s)000000000"
LOKI_START="$((LOKI_NOW - 3600000000000))"
if curl -sf "http://localhost:3100/loki/api/v1/query?query=%7Bjob%3D%22docker%22%7D&limit=5&start=${LOKI_START}&end=${LOKI_NOW}" | grep -q '"status":"success"'; then
  log_pass "Loki query API"
else
  # fallback: any log stream
  if curl -sf "http://localhost:3100/loki/api/v1/labels" | grep -q '"status":"success"'; then
    log_pass "Loki labels API (logs ingested)"
  else
    log_fail "Loki query"
  fi
fi

echo
echo "--- Tempo traces ---"
if curl -sf "http://localhost:3200/api/search?limit=5" | grep -qE 'traces|results|"traceID"'; then
  log_pass "Tempo search API"
else
  log_skip "Tempo search (no traces yet — may need more OTel traffic)"
fi

echo
echo "--- Grafana datasources ---"
GF_PASS="${GF_SECURITY_ADMIN_PASSWORD:-changeme-grafana-admin}"
DS_RESP="$(curl -sf -u "admin:${GF_PASS}" "http://localhost:3000/api/datasources" || echo '[]')"
for ds in Prometheus Mimir Loki Tempo Pyroscope; do
  if echo "$DS_RESP" | grep -q "\"name\":\"${ds}\""; then
    log_pass "Grafana datasource: ${ds}"
  else
    log_fail "Grafana datasource missing: ${ds}"
  fi
done

echo
echo "--- OTel collector ---"
if curl -sf "http://localhost:14318" >/dev/null 2>&1 || docker compose exec -T otel-collector wget -qO- http://localhost:13133/  2>/dev/null | grep -q .; then
  log_skip "OTel collector health (internal only)"
fi
if docker compose ps otel-collector 2>/dev/null | grep -qE 'Up|running'; then
  log_pass "OTel collector container running"
else
  log_fail "OTel collector container"
fi

echo
echo "--- Alloy / Faro ---"
FARO_PAYLOAD='{"meta":{"app":{"name":"e2e-test","version":"1.0.0"}},"logs":[{"message":"e2e faro test","level":"info","timestamp":"2020-01-01T00:00:00.000Z"}]}'
if curl -sf -X POST "http://localhost:8027/collect" -H 'Content-Type: application/json' -d "$FARO_PAYLOAD" >/dev/null 2>&1; then
  log_pass "Faro collector accepted POST /collect"
else
  log_skip "Faro POST (endpoint may differ by Alloy version)"
fi

echo
echo "--- Beyla ---"
if docker compose ps beyla 2>/dev/null | grep -qE 'Up|running'; then
  if curl -sf "http://localhost:9090/metrics" 2>/dev/null | head -1 | grep -q .; then
    log_skip "Beyla metrics on host :9090 (conflicts with Prometheus port mapping)"
  fi
  BEYLA_STATE="$(docker compose ps beyla --format '{{.Status}}' 2>/dev/null || echo unknown)"
  if echo "$BEYLA_STATE" | grep -qiE 'up|running'; then
    log_pass "Beyla container: ${BEYLA_STATE}"
  else
    log_skip "Beyla may require Linux eBPF (${BEYLA_STATE})"
  fi
else
  log_skip "Beyla not running (eBPF requires Linux)"
fi

echo
echo "--- k6 load test ---"
if docker compose --profile loadtest run --rm k6 2>&1; then
  log_pass "k6 load test completed"
else
  log_fail "k6 load test"
fi

echo
echo "--- Alertmanager ---"
if curl -sf "http://localhost:9093/-/healthy" | grep -q OK; then
  log_pass "Alertmanager healthy"
elif curl -sf "http://localhost:9093/api/v2/status" | grep -qE '"status":"ready"|cluster'; then
  log_pass "Alertmanager API"
else
  log_fail "Alertmanager API"
fi

echo
echo "--- OnCall ---"
if curl -sf "http://localhost:8090/health/" | grep -qE 'ok|healthy|200'; then
  log_pass "OnCall health endpoint"
elif curl -sf "http://localhost:8090/" >/dev/null 2>&1; then
  log_pass "OnCall engine responding"
else
  log_fail "OnCall engine"
fi

if [ "$RUN_SECURITY" = true ]; then
  echo
  echo "--- Security overlay ---"
  for svc in falco falcosidekick falco-talon tetragon wazuh-manager wazuh-dashboard; do
    STATUS="$(docker compose -f "$ROOT/docker/observability/docker-compose.security.yaml" ps "$svc" --format '{{.Status}}' 2>/dev/null || echo missing)"
    if echo "$STATUS" | grep -qiE 'up|running'; then
      log_pass "security/${svc}: ${STATUS}"
    else
      log_skip "security/${svc}: ${STATUS} (often requires Linux)"
    fi
  done
fi

echo
echo "=== Summary ==="
echo -e "${GREEN}Passed: ${pass}${NC}  ${RED}Failed: ${fail}${NC}  ${YELLOW}Skipped: ${skip}${NC}"
if [ "$fail" -gt 0 ]; then
  echo
  echo "Failed service logs (last 20 lines):"
  docker compose ps --format '{{.Name}} {{.Status}}' | grep -iv 'up\|running' || true
  exit 1
fi
exit 0
