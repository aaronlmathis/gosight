Write-Host "🔐 GoSight TLS Certificate Generator (Windows)"

$hostname = Read-Host "Enter server hostname (e.g. localhost or gosight.internal)"
New-Item -ItemType Directory -Force -Path "certs" | Out-Null
Set-Location -Path "certs"

# Generate CA
openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.crt -subj "/CN=GoSightCA"

# Server key + cert
openssl genrsa -out server.key 4096
openssl req -new -key server.key -out server.csr -subj "/CN=$hostname"
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365 -sha256

# Client key + cert
openssl genrsa -out client.key 4096
openssl req -new -key client.key -out client.csr -subj "/CN=agent"
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365 -sha256

Write-Host "✅ Certificates generated in ./certs/"
