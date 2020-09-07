#!/bin/bash
set -e

timestamp=$(date +%Y%m%d%H%M%S)

## Stopping Services
kill "$(pgrep anywhere)"||true
ssh aliyun "kill \$(pgrep anywhere)||true"

## Upgrading Services
# Agent
mv /usr/local/anywhere  /usr/local/anywhere_"$timestamp"
mkdir -p /usr/local/anywhere
cd /usr/local/anywhere
wget ftp://ftp:ftp@10.0.0.2/ci/anywhere/anywhere-latest.tar.gz
tar -xvf anywhere-latest.tar.gz
rm -rf bin/anywhered bin/test
nohup ./bin/anywhere --user admin --pass admin -s 47.103.62.227 > anywhere.log 2>&1 &

# Server
# shellcheck disable=SC2029
ssh aliyun "mv /usr/local/anywhere  /usr/local/anywhere_$timestamp"
ssh aliyun "mkdir -p /usr/local/anywhere"
scp anywhere-latest.tar.gz aliyun:/usr/local/anywhere
ssh aliyun "cd /usr/local/anywhere && tar -xvf anywhere-latest.tar.gz && rm -rf anywhere-latest.tar.gz"
ssh aliyun "cd /usr/local/anywhere && rm -rf bin/anywhere && rm -rf bin/test"
# shellcheck disable=SC2029
ssh aliyun "cp /usr/local/anywhere_$timestamp/proxy.json /usr/local/anywhere/proxy.json"
# shellcheck disable=SC2029
ssh aliyun "cp /usr/local/anywhere_$timestamp/anywhered.json /usr/local/anywhere/anywhered.json"
ssh aliyun 'cd /usr/local/anywhere; nohup ./bin/anywhered start > anywhered.log 2>&1 &'