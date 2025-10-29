# 🔍 TLS Client Initialization Flow - Từ main.go đến Connection

## 📋 Flow Overview

```
main.go
    ↓ calls
config.Load()
    ↓ returns Config with TLS settings
client.NewClients(cfg)
    ↓ creates
grpcpool.Manager
    ↓ for each service
grpcpool.Pool (with TLS)
    ↓ creates
5 gRPC Connections per service
    ↓ each connection
TLS Handshake with server
    ↓ ready
Client can call service methods
```

---

## 📊 Complete Flow Diagram

```
main.go
  │
  ├─> config.Load()
  │     │
  │     └─> LoadServerConfig()
  │           │
  │           └─> LoadTLSConfig()  ← Read from .env
  │                 │
  │                 └─> Return TLSConfig{Enabled, CertFile, KeyFile, CAFile}
  │
  ├─> clients.NewClients(cfg)
  │     │
  │     ├─> For each service:
  │     │     │
  │     │     └─> sharedTLS.ClientTLSConfig(caFile, serverName)
  │     │           │
  │     │           ├─> os.ReadFile(caFile)  ← Read ca-cert.pem
  │     │           ├─> x509.NewCertPool()
  │     │           ├─> certPool.AppendCertsFromPEM(caCert)
  │     │           └─> credentials.NewTLS(&tls.Config{
  │     │                   RootCAs: certPool,
  │     │                   ServerName: serviceName  ← "order-service"
  │     │               })
  │     │
  │     ├─> poolManager.CreateCommonPools(configs)
  │     │     │
  │     │     └─> For each config:
  │     │           │
  │     │           └─> NewPool(target, size, PoolConfig{TLSCreds})
  │     │                 │
  │     │                 └─> For i := 0; i < 5; i++:
  │     │                       │
  │     │                       └─> grpc.Dial(target, opts...)
  │     │                             │
  │     │                             └─> TLS HANDSHAKE
  │     │                                   │
  │     │                                   ├─> TCP Connect
  │     │                                   ├─> ClientHello
  │     │                                   ├─> Receive ServerCertificate
  │     │                                   ├─> Verify Certificate:
  │     │                                   │     - Signed by CA?
  │     │                                   │     - ServerName match?
  │     │                                   │     - Not expired?
  │     │                                   ├─> Key Exchange
  │     │                                   └─> Encrypted Connection 
  │     │
  │     └─> Return Clients{User, Product, Order, ...}
  │
  └─> API Gateway ready to handle requests
        │
        └─> Client can call: orderClient.CreateOrder()
              │
              └─> Get connection from pool (already TLS encrypted)
                    │
                    └─> Send encrypted request → Receive encrypted response
```

---

## 🔑 Key Points

1. **TLS Config loaded từ .env** → `LoadTLSConfig()`
2. **MỖI service có TLS credentials riêng** → `ClientTLSConfig(caFile, serverName)`
3. **TLS Handshake xảy ra khi `grpc.Dial()`** → Tự động bởi Go
4. **Pool có 5 connections, tất cả đã TLS** → Ready to use
5. **Client call = Get connection from pool** → Already encrypted

---



