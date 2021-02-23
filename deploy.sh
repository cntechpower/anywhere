#!/bin/bash
set -e

timestamp=$(date +%Y%m%d%H%M%S)
curr_version=$(git rev-parse --short HEAD)

## Stopping Services
kill "$(pgrep anywhere)"||true
ssh aliyun "kill \$(pgrep anywhere)||true"
ssh pi "kill \$(pgrep anywhere)||true"

## Upgrading Services
# Agent
kubectl -n private set image deployment/anywhere-agent anywhere-agent=10.0.0.2:5000/cntechpower/anywhere-agent:$curr_version

echo "K8s Agent Upgrade Success"

# Agent Pi
ssh pi "mv /usr/local/anywhere  /usr/local/anywhere_$timestamp"
ssh pi "mkdir -p /usr/local/anywhere"
rm -rf anywhere-latest-arm.tar.gz
wget -q ftp://ftp:ftp@10.0.0.2/ci/anywhere/anywhere-latest-arm.tar.gz
scp anywhere-latest-arm.tar.gz pi:/usr/local/anywhere
ssh pi "cd /usr/local/anywhere && tar -xf anywhere-latest-arm.tar.gz && rm -rf anywhere-latest-arm.tar.gz"
ssh pi "cd /usr/local/anywhere && rm -rf bin/anywhered bin/test"
ssh pi "cd /usr/local/anywhere; nohup ./bin/anywhere -z asia-shanghai -i anywhere-agent-pi --user admin --pass admin -s 47.103.62.227 > anywhere.log 2>&1 &"

echo "Agent Pi Upgrade Success"

# Server
# shellcheck disable=SC2029
ssh aliyun "mv /usr/local/anywhere  /usr/local/anywhere_$timestamp"
ssh aliyun "mkdir -p /usr/local/anywhere"
wget -q ftp://ftp:ftp@10.0.0.2/ci/anywhere/anywhere-latest.tar.gz
scp anywhere-latest.tar.gz aliyun:/usr/local/anywhere
ssh aliyun "cd /usr/local/anywhere && tar -xf anywhere-latest.tar.gz && rm -rf anywhere-latest.tar.gz"
ssh aliyun "cd /usr/local/anywhere && rm -rf bin/anywhere && rm -rf bin/test"
# shellcheck disable=SC2029
ssh aliyun "cp /usr/local/anywhere_$timestamp/proxy.json /usr/local/anywhere/proxy.json"
# shellcheck disable=SC2029
ssh aliyun "cp /usr/local/anywhere_$timestamp/anywhered.json /usr/local/anywhere/anywhered.json"
ssh aliyun 'cd /usr/local/anywhere; nohup ./bin/anywhered start > anywhered.log 2>&1 &'

echo "Server Upgrade Success"
# sleep 5 to wait agent
sleep 5
ssh aliyun "cd /usr/local/anywhere; ./bin/anywhered agent list"
ssh aliyun "cd /usr/local/anywhere; ./bin/anywhered proxy list"
