#!/bin/bash
set -e

echo "🔐 GoSight TLS Cert Generator (with SANs)"
read -p "Enter the hostname or IP for the server certificate SAN (e.g., gosight.local or 192.168.0.40): " HOSTNAME

mkdir -p certs && cd certs

# === 1. Generate CA ===
openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.crt -subj "/CN=GoSightCA"

# === 2. Create server SAN config ===
cat > server.cnf <<EOF
[ req ]
default_bits       = 2048
prompt             = no
default_md         = sha256
req_extensions     = req_ext
distinguished_name = dn

[ dn ]
CN = ${HOSTNAME}

[ req_ext ]
subjectAltName = @alt_names

[ alt_names ]
DNS.1 = localhost
IP.1  = 127.0.0.1
DNS.2 = ${HOSTNAME}
EOF

# === 3. Generate server cert ===
openssl genrsa -out server.key 4096
openssl req -new -key server.key -out server.csr -config server.cnf
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt -days 365 -sha256 -extensions req_ext -extfile server.cnf
cat server.crt ca.crt > server-cert.pem
cp server.key server-key.pem

# === 4. Create client SAN config ===
cat > client.cnf <<EOF
[ req ]
default_bits       = 2048
prompt             = no
default_md         = sha256
req_extensions     = req_ext
distinguished_name = dn

[ dn ]
CN = gosight-agent

[ req_ext ]
subjectAltName = @alt_names

[ alt_names ]
DNS.1 = gosight-agent
EOF

# === 5. Generate client cert ===
openssl genrsa -out client.key 4096
openssl req -new -key client.key -out client.csr -config client.cnf
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out client.crt -days 365 -sha256 -extensions req_ext -extfile client.cnf

echo "All certs generated in ./certs/"
ls -l
