Write-Host "GoSight TLS Cert Generator (with SANs for Windows)" -ForegroundColor Cyan

# Create working dir
$certsPath = Join-Path -Path $PSScriptRoot -ChildPath "certs"
New-Item -ItemType Directory -Force -Path $certsPath | Out-Null
Set-Location $certsPath

# === 1. Generate CA ===
openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.crt -subj "/CN=GoSightCA"

# === 2. Create server.cnf with SANs ===
@"
[ req ]
default_bits       = 2048
prompt             = no
default_md         = sha256
req_extensions     = req_ext
distinguished_name = dn

[ dn ]
CN =localhost

[ req_ext ]
subjectAltName = @alt_names

[ alt_names ]
DNS.1 = localhost
IP.1  = 127.0.0.1
"@ | Set-Content -Encoding ascii server.cnf

# === 3. Generate server cert ===
openssl genrsa -out server.key 4096
openssl req -new -key server.key -out server.csr -config server.cnf
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial `
  -out server.crt -days 365 -sha256 -extensions req_ext -extfile server.cnf
Get-Content server.crt, ca.crt | Set-Content server-cert.pem
Copy-Item server.key server-key.pem


# === 4. Create client.cnf with SANs ===
@"
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
"@ | Set-Content -Encoding ascii client.cnf

# === 5. Generate client cert ===
openssl genrsa -out client.key 4096
openssl req -new -key client.key -out client.csr -config client.cnf
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial `
  -out client.crt -days 365 -sha256 -extensions req_ext -extfile client.cnf

Write-Host "`n All certs generated in $certsPath" -ForegroundColor Green
Get-ChildItem -Name
