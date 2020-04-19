#!/bin/bash
set -e
docker run  -t --rm --network composefiles_anywhere_test_net 10.0.0.2:5000/mysql/mysql_client:8.0.19 -h172.90.101.11 -P4444 -proot -e "select @@version"
docker run  -t --rm --network composefiles_anywhere_test_net 10.0.0.2:5000/mysql/mysql_client:5.7.28 -h172.90.101.11 -P4445 -proot -e "select @@version"