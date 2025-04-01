#!/bin/bash

set -e

echo "🔐 GoSight TLS Certificate Generator"

read -p "Enter server hostname (e.g. localhost or gosight.internal): " HOSTNAME

mkdir -p certs
cd certs

# Generate CA
openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.crt -subj "/CN=GoSightCA"

# Generate Server Key and CSR
openssl genrsa -out server.key 4096
openssl req -new -key server.key -out server.csr -config server.cnf
openssl x509 -req  -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial  -out server.crt -days 365 -sha256 -extensions req_ext -extfile server.cnf

# Generate Client Key and CSR
openssl genrsa -out client.key 4096
openssl req -new -key client.key -out client.csr -subj "/CN=agent"
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365 -sha256

echo "✅ Certificates generated in ./certs/"
