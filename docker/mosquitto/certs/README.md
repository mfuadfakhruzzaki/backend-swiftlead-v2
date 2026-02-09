# Mosquitto TLS Certificates

Place your TLS certificates in this directory to enable encrypted MQTT connections.

## Required files

| File         | Description                            |
| ------------ | -------------------------------------- |
| `ca.crt`     | Certificate Authority (CA) certificate |
| `server.crt` | Server certificate signed by the CA    |
| `server.key` | Server private key                     |

## Generate self-signed certificates (development)

```bash
# Generate CA key and certificate
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt \
  -subj "/CN=Swiftlet MQTT CA"

# Generate server key and CSR
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr \
  -subj "/CN=mosquitto"

# Sign server certificate with CA
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out server.crt -days 3650

# (Optional) Generate client key and certificate for mutual TLS
openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr \
  -subj "/CN=swiftlet-backend"
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out client.crt -days 3650

# Clean up CSR files
rm -f *.csr
```

## Enabling TLS

1. Generate or obtain certificates and place them here
2. Uncomment the TLS listener block in `docker/mosquitto/mosquitto.conf`
3. Set these environment variables:
   ```
   MQTT_BROKER=ssl://mosquitto:8883
   MQTT_USE_TLS=true
   MQTT_CA_CERT=/app/certs/mqtt/ca.crt
   ```
4. Mount certs into the backend container (see docker-compose)
