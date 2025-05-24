## 0.2.0 (2025-05-24)

### Feat

- **mtls**: server side mtls implementation

## 0.1.0 (2025-05-24)

### Feat

- **sharding**: basic sharding initialization and scaffold
- **server**: grpc server bootstrap
- **server**: implement grpc store server and interceptors
- **proto**: add baseline proto definitions and generation using Makefile
- **configuration**: add tls mode configuration and certificate/key location
- **database**: initial database struct and methods, a wrapper for boltdb
- **configuration**: add configuration validation
- **server**: basic project scaffold, configuration and logging initialization

### Fix

- **server**: fix incorrect usage of UnaryServerInterceptor chaining
- **database**: create default bucket
- **configuration**: fix default configuration values bug

### Refactor

- **config**: move config to seperate package
