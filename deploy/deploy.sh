#!/bin/bash
set -e

timestamp=$(date +%Y%m%d%H%M%S)
curr_version=$(git rev-parse --short HEAD)

## Stopping Services
ssh aliyun "kill \$(pgrep anywhere)||true"

## Upgrading Services
# Agent
kubectl -n private set image deployment/anywhere-agent anywhere-agent=10.0.0.2:5000/cntechpower/anywhere-agent:$curr_version

echo "K8s Agent Upgrade Success"


# Server
# shellcheck disable=SC2029
ssh aliyun "mv /usr/local/anywhere  /usr/local/anywhere_$timestamp"
ssh aliyun "mkdir -p /usr/local/anywhere"
wget -q ftp://ftp:ftp@10.0.0.2/ci/anywhere/anywhere-latest.tar.gz
scp anywhere-latest.tar.gz aliyun:/usr/local/anywhere
ssh aliyun "cd /usr/local/anywhere && tar -xf anywhere-latest.tar.gz && rm -rf anywhere-latest.tar.gz"
ssh aliyun "cd /usr/local/anywhere && rm -rf bin/anywhere && rm -rf bin/test"
# shellcheck disable=SC2029
ssh aliyun "cp /usr/local/anywhere_$timestamp/proxy.db /usr/local/anywhere/proxy.db"
# shellcheck disable=SC2029
ssh aliyun "cp /usr/local/anywhere_$timestamp/anywhered.json /usr/local/anywhere/anywhered.json"
ssh aliyun 'cd /usr/local/anywhere; nohup ./bin/anywhered start > anywhered.log 2>&1 &'

echo "Server Upgrade Success"
# sleep 5 to wait agent
sleep 5
ssh aliyun "cd /usr/local/anywhere; ./bin/anywhered agent list"
