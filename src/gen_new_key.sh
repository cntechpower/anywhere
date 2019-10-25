#!/bin/bash
openssl genrsa -out credential/$1.key 2048
openssl req -new -key credential/$1.key -out credential/$1.csr -subj "/CN=cntechpower_anywhere"
openssl x509 -req -in credential/$1.csr -CA credential/ca.crt -CAkey credential/ca.key -CAcreateserial -out credential/$1.crt -days 365