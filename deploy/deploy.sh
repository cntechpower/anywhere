#!/bin/bash
set -e

timestamp=$(date +%Y%m%d%H%M%S)
curr_version=$(git rev-parse --short HEAD)

## Stopping Services
ssh aliyun "kill \$(pgrep anywhere)||true"

## Upgrading Services
# Aliyun Agent
kubectl -n private set image deployment/anywhere-agent anywhere-agent=10.0.0.2:5000/cntechpower/anywhere-agent:$curr_version

# QingCloud Agent
kubectl -n private set image deployment/anywhere-agent-qc anywhere-agent=10.0.0.2:5000/cntechpower/anywhere-agent:$curr_version
echo "K8s Agent Upgrade Success"


# Aliyun
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
ssh aliyun 'cd /usr/local/anywhere; ES_ADDR=http://127.0.0.1:9200 TRACE_ADDR=127.0.0.1:6831 nohup ./bin/anywhered start > anywhered.log 2>&1 &'
echo "Aliyun Upgrade Success"

# QingCloud
# shellcheck disable=SC2029
ssh qc1 "mv /usr/local/anywhere  /usr/local/anywhere_$timestamp"
ssh qc1 "mkdir -p /usr/local/anywhere"
wget -q ftp://ftp:ftp@10.0.0.2/ci/anywhere/anywhere-latest.tar.gz
scp anywhere-latest.tar.gz qc1:/usr/local/anywhere
ssh qc1 "cd /usr/local/anywhere && tar -xf anywhere-latest.tar.gz && rm -rf anywhere-latest.tar.gz"
ssh qc1 "cd /usr/local/anywhere && rm -rf bin/anywhere && rm -rf bin/test"
# shellcheck disable=SC2029
ssh qc1 "cp /usr/local/anywhere_$timestamp/proxy.db /usr/local/anywhere/proxy.db"
# shellcheck disable=SC2029
ssh qc1 "cp /usr/local/anywhere_$timestamp/anywhered.json /usr/local/anywhere/anywhered.json"
ssh qc1 'cd /usr/local/anywhere; nohup ./bin/anywhered start > anywhered.log 2>&1 &'

echo "QingCloud Upgrade Success"
# sleep 5 to wait agent
sleep 5
ssh aliyun "cd /usr/local/anywhere; ./bin/anywhered agent list"
ssh qc1 "cd /usr/local/anywhere; ./bin/anywhered agent list"
