#!/usr/bin/env bash
#
# Container end-to-end test for Anywhere.
#
# It builds Linux binaries, bakes the server/agent images, starts three agents
# plus HTTP/UDP/MySQL targets, creates proxy configs through the authenticated
# HTTP API and verifies every data path.
#
# Usage:
#   ./test/e2e/run.sh              # full run
#   KEEP=1 ./test/e2e/run.sh       # keep containers on failure for debugging
#   SKIP_UI=1 ./test/e2e/run.sh    # reuse an existing ./static build
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

PROJECT="${E2E_PROJECT:-anywhere-e2e}"
COMPOSE_FILE="test/composefiles/docker-compose.yml"
NETWORK="${PROJECT}_anywhere_test_net"
TIMEOUT="${E2E_TIMEOUT:-180}"
COOKIES="$(mktemp)"
TMPDIR="$(mktemp -d)"

compose() {
  docker compose -p "$PROJECT" -f "$COMPOSE_FILE" "$@"
}

cleanup() {
  local rc=$?
  trap - EXIT
  if [[ $rc -ne 0 ]]; then
    echo "--- compose ps ---" >&2
    compose ps >&2 || true
    echo "--- compose logs (tail 120) ---" >&2
    compose logs --tail=120 >&2 || true
  fi
  if [[ "${KEEP:-0}" == "1" ]]; then
    echo "KEEP=1: leaving project ${PROJECT} running" >&2
  else
    compose down -v --remove-orphans >/dev/null 2>&1 || true
  fi
  rm -f "$COOKIES"
  rm -rf "$TMPDIR"
  exit "$rc"
}
trap cleanup EXIT

fail() {
  echo "E2E FAILED: $*" >&2
  exit 1
}

http_code() {
  curl -sk -o /dev/null -w '%{http_code}' --max-time 3 "$1" || true
}

wait_server() {
  local deadline=$((SECONDS + TIMEOUT))
  while ((SECONDS < deadline)); do
    if [[ "$(http_code https://127.0.0.1:1114/user/login)" == "200" ]]; then
      return 0
    fi
    sleep 2
  done
  return 1
}

agents_online() {
  local body
  body="$(curl -sk -b "$COOKIES" https://127.0.0.1:1114/api/v1/agent/list || true)"
  AGENT_COUNT="$(printf '%s' "$body" | python3 -c 'import json,sys
try:
    data = json.load(sys.stdin)
    print(len(data) if isinstance(data, list) else 0)
except Exception:
    print(0)')"
  [[ "$AGENT_COUNT" -ge 3 ]]
}

wait_agents() {
  local deadline=$((SECONDS + TIMEOUT))
  while ((SECONDS < deadline)); do
    if agents_online; then
      return 0
    fi
    sleep 2
  done
  return 1
}

add_proxy() {
  local port="$1" target="$2" proto="$3" code
  code="$(curl -sk -b "$COOKIES" -o "$TMPDIR/add-$port.json" -w '%{http_code}' \
    --max-time 10 -X POST \
    --data-urlencode "user_name=admin" \
    --data-urlencode "zone_name=asia-shanghai" \
    --data-urlencode "remote_port=$port" \
    --data-urlencode "local_addr=$target" \
    --data-urlencode "white_list_enable=false" \
    --data-urlencode "listen_type=$proto" \
    https://127.0.0.1:1114/api/v1/proxy/add)"
  if [[ "$code" != "200" ]]; then
    cat "$TMPDIR/add-$port.json" >&2 || true
    fail "could not add proxy $port -> $target ($proto), http $code"
  fi
}

wait_http_target() {
  local deadline=$((SECONDS + 60))
  while ((SECONDS < deadline)); do
    if [[ "$(curl -fsS --max-time 3 http://127.0.0.1:18080/ 2>/dev/null || true)" == "anywhere-http-ok" ]]; then
      return 0
    fi
    sleep 2
  done
  return 1
}

wait_mysql_target() {
  local deadline=$((SECONDS + TIMEOUT))
  while ((SECONDS < deadline)); do
    if compose exec -T mysql mysqladmin ping -h 127.0.0.1 -uroot -p"${MYSQL_ROOT_PASSWORD:-root}" --silent >/dev/null 2>&1; then
      return 0
    fi
    sleep 3
  done
  return 1
}

echo "==> preflight"
for bin in docker curl python3 make; do
  command -v "$bin" >/dev/null 2>&1 || fail "$bin is required"
done
docker info >/dev/null 2>&1 || fail "docker daemon is not reachable"

echo "==> building Linux binaries"
make build_linux

if [[ "${SKIP_UI:-0}" != "1" ]]; then
  echo "==> building frontend"
  make ui
fi
[[ -f static/index.html ]] || fail "static/index.html is missing (run make ui first)"

echo "==> building images"
docker build -t anywhered-test-image:latest -f docker-build/Dockerfile.server . || fail "server image build"
docker build -t anywhere-test-image:latest -f docker-build/Dockerfile.agent . || fail "agent image build"

echo "==> starting compose project ${PROJECT}"
compose down -v --remove-orphans >/dev/null 2>&1 || true
compose up -d

echo "==> waiting for the console"
wait_server || fail "console did not become ready"

echo "==> logging in"
curl -sk -c "$COOKIES" -f --max-time 10 -X POST \
  --data-urlencode "username=admin" \
  --data-urlencode "password=admin" \
  --data-urlencode "otpcode=" \
  https://127.0.0.1:1114/user_login >/dev/null || fail "login failed"

echo "==> waiting for 3 agents"
wait_agents || fail "expected 3 agents, got ${AGENT_COUNT:-0}"

echo "==> creating proxy configs"
add_proxy 18080 "http-target:8080" tcp
add_proxy 18090 "udp-target:9090" udp
add_proxy 13306 "mysql:3306" tcp

echo "==> testing HTTP tunnel"
wait_http_target || fail "HTTP tunnel did not answer through port 18080"

echo "==> testing UDP tunnel"
python3 - <<'PY' || fail "UDP tunnel did not echo"
import socket, time
last = None
for _ in range(10):
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    s.settimeout(2)
    try:
        s.sendto(b"anywhere-udp-ok", ("127.0.0.1", 18090))
        data, _ = s.recvfrom(1024)
        if data == b"anywhere-udp-ok":
            print("udp echo ok")
            raise SystemExit(0)
        last = data
    except (socket.timeout, OSError) as exc:
        last = exc
    finally:
        s.close()
    time.sleep(1)
print(f"no udp echo: {last}", flush=True)
raise SystemExit(1)
PY

echo "==> testing MySQL tunnel"
wait_mysql_target || fail "mysql target did not become healthy"
mysql_out="$(docker run --rm --network "$NETWORK" mysql:8.0 \
  mysql --protocol=TCP -h anywhered -P 13306 -uroot -p"${MYSQL_ROOT_PASSWORD:-root}" \
  -N -e "select 42" 2>"$TMPDIR/mysql.err")" || {
  cat "$TMPDIR/mysql.err" >&2
  fail "mysql client through the tunnel failed"
}
[[ "$(echo "$mysql_out" | tr -d '[:space:]')" == "42" ]] || fail "unexpected mysql output: $mysql_out"
echo "mysql tunnel ok"

echo "==> testing the console UI"
html="$(curl -sk -b "$COOKIES" --max-time 10 https://127.0.0.1:1114/)"
grep -q "Anywhere" <<<"$html" || fail "UI index is missing the app shell"
asset="$(grep -o '/assets/[^"]*\.js' <<<"$html" | head -1)"
[[ -n "$asset" ]] || fail "UI did not reference a hashed asset"
[[ "$(http_code "https://127.0.0.1:1114$asset")" == "200" ]] || fail "UI asset $asset is not served"

echo "==> all container E2E checks passed"
