#!/bin/bash
set -e

timestamp=$(date +%Y%m%d%H%M%S)

## Stopping Services
kill "$(pgrep anywhere)"||true
ssh aliyun "kill \$(pgrep anywhere)||true"
ssh pi "kill \$(pgrep anywhere)||true"

## Upgrading Services
# Agent
mv /usr/local/anywhere  /usr/local/anywhere_"$timestamp"
mkdir -p /usr/local/anywhere
cd /usr/local/anywhere
rm -rf anywhere-latest.tar.gz
wget ftp://ftp:ftp@10.0.0.2/ci/anywhere/anywhere-latest.tar.gz
tar -xf anywhere-latest.tar.gz
rm -rf bin/anywhered bin/test
nohup ./bin/anywhere --user admin --pass admin -s 47.103.62.227 > anywhere.log 2>&1 &

echo "Agent 10-0-0-2 Upgrade Success"

# Agent Pi
ssh pi "mv /usr/local/anywhere  /usr/local/anywhere_$timestamp"
ssh pi "mkdir -p /usr/local/anywhere"
rm -rf anywhere-latest-arm.tar.gz
wget ftp://ftp:ftp@10.0.0.2/ci/anywhere/anywhere-latest-arm.tar.gz
scp anywhere-latest-arm.tar.gz pi:/usr/local/anywhere
ssh pi "cd /usr/local/anywhere && tar -xf anywhere-latest-arm.tar.gz && rm -rf anywhere-latest-arm.tar.gz"
ssh pi "cd /usr/local/anywhere && rm -rf bin/anywhered bin/test"
ssh pi "cd /usr/local/anywhere; nohup ./bin/anywhere  -i anywhere-agent-pi --user admin --pass admin -s 47.103.62.227 > anywhere.log 2>&1 &"

echo "Agent Pi Upgrade Success"

# Server
# shellcheck disable=SC2029
ssh aliyun "mv /usr/local/anywhere  /usr/local/anywhere_$timestamp"
ssh aliyun "mkdir -p /usr/local/anywhere"
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
ssh aliyun "cd /usr/local/anywhere; bash /data/anywhere.sh"
ssh aliyun "cd /usr/local/anywhere; ./bin/anywhered proxy list"
