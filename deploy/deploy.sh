#!/bin/bash
# Deploy a built artifact to a remote host and roll the Kubernetes agent image.
#
# Required environment:
#   DEPLOY_HOST    ssh target, e.g. root@server.example.com
#   ARTIFACT       local path to anywhere-latest.tar.gz
#   AGENT_IMAGE    full image reference, e.g. registry.example.com/anywhere-agent:abc123
# Optional:
#   REMOTE_DIR     install directory on the host (default /usr/local/anywhere)
#   K8S_NAMESPACE  namespace of the agent deployment (default private)
set -euo pipefail

: "${DEPLOY_HOST:?set DEPLOY_HOST}"
: "${ARTIFACT:?set ARTIFACT}"
: "${AGENT_IMAGE:?set AGENT_IMAGE}"

REMOTE_DIR=${REMOTE_DIR:-/usr/local/anywhere}
K8S_NAMESPACE=${K8S_NAMESPACE:-private}
timestamp=$(date +%Y%m%d%H%M%S)

echo "Rolling agent image to ${AGENT_IMAGE}"
kubectl -n "${K8S_NAMESPACE}" set image deployment/anywhere-agent anywhere-agent="${AGENT_IMAGE}"

echo "Uploading ${ARTIFACT} to ${DEPLOY_HOST}"
scp "${ARTIFACT}" "${DEPLOY_HOST}:/tmp/anywhere-latest.tar.gz"

echo "Installing on ${DEPLOY_HOST}:${REMOTE_DIR}"
ssh "${DEPLOY_HOST}" REMOTE_DIR="${REMOTE_DIR}" TIMESTAMP="${timestamp}" 'bash -s' <<'EOF'
set -euo pipefail
if [ -d "${REMOTE_DIR}" ]; then
  mv "${REMOTE_DIR}" "${REMOTE_DIR}_${TIMESTAMP}"
  mkdir -p "${REMOTE_DIR}"
  # Keep runtime state across upgrades.
  cp "${REMOTE_DIR}_${TIMESTAMP}/anywhered.json" "${REMOTE_DIR}/anywhered.json" 2>/dev/null || true
  cp "${REMOTE_DIR}_${TIMESTAMP}/proxy.db" "${REMOTE_DIR}/proxy.db" 2>/dev/null || true
fi
mkdir -p "${REMOTE_DIR}"
tar -xf /tmp/anywhere-latest.tar.gz -C "${REMOTE_DIR}"
rm -f /tmp/anywhere-latest.tar.gz
EOF

echo "Restarting the server"
ssh "${DEPLOY_HOST}" REMOTE_DIR="${REMOTE_DIR}" 'bash -s' <<'EOF'
set -euo pipefail
cd "${REMOTE_DIR}"
pkill -f "bin/anywhered start" || true
nohup ./bin/anywhered start > anywhered.log 2>&1 &
EOF

echo "Waiting for the agent to reconnect"
sleep 5
ssh "${DEPLOY_HOST}" "cd '${REMOTE_DIR}' && ./bin/anywhered agent list"
