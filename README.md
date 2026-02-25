# Golang Basic - API Documentation Platform

Full-stack platform for creating, managing, and sharing REST, GraphQL, and gRPC API documentation with attribute-based access control.

## Overview

Create projects with multiple API types (REST, GraphQL, gRPC), document with structured parameters/headers/examples, share via public links or authenticated access, and control access with ABAC.

## Architecture

```
Frontend (React) → Nginx → Go API (Chi v5) → PostgreSQL 17
                         ↓                      ↓
                      MinIO                 pgx/v5 pool
```

**Clean Architecture Layers:** Controller → Service → Repository → Model

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.24, Chi v5, pgx/v5 |
| Database | PostgreSQL 17, pgxpool |
| Storage | MinIO (S3-compatible) |
| Frontend | React 18, Vite 5, TypeScript, Zustand |
| Proxy | Nginx (rate limiting, caching) |

## Quick Start

```bash
# Backend
cd api
cp .env.example .env
go mod download
psql -U postgres -d golang_basic < sql/abac_schema.sql
psql -U postgres -d golang_basic < sql/api_docs_schema.sql
go run main.go

# Podman (recommended)
podman-compose up -d
```

See [PODMAN.md](PODMAN.md) for complete deployment guide.

## Project Structure

```
api/
├── internal/
│   ├── config/          # Database, MinIO configuration
│   ├── controller/      # HTTP handlers
│   ├── middleware/      # JWT, ABAC, logging
│   ├── model/           # Data models
│   ├── repository/      # Database access (pgx/v5)
│   ├── routes/          # Chi router setup
│   ├── service/         # Business logic
│   └── utility/         # Helper functions
├── sql/                 # Database migrations
├── nginx/               # Reverse proxy config
└── main.go              # Entry point
```

## API Endpoints

**Base URL:** `https://localhost:3003`

### Authentication (Public)
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout` (JWT)

### Users
- `GET /api/v1/users` (JWT)
- `POST /api/v1/users` (Public)
- `GET/PUT /api/v1/users/{id}` (JWT)
- `PUT /api/v1/users/{id}/password` (JWT)

### Profiles
- `GET /api/v1/profiles` (Optional Auth)
- `POST /api/v1/profiles` (JWT)
- `GET/PUT/DELETE /api/v1/profiles/{id}` (JWT)

### Projects
- `GET /api/v1/projects` (Optional Auth)
- `GET /api/v1/projects/with-stats` (Optional Auth, optimized)
- `POST /api/v1/projects` (JWT + Permission)
- `GET /api/v1/projects/{id}` (Optional Auth)
- `GET /api/v1/projects/{id}/api-stats` (Optional Auth)
- `PUT /api/v1/projects/{id}` (JWT + Permission)
- `DELETE /api/v1/projects/{id}` (JWT + Permission)

### REST APIs
- `GET /api/v1/rest-apis` (JWT + Permission)
- `GET/PUT/DELETE /api/v1/rest-apis/{id}` (JWT + Permission)
- `POST /api/v1/projects/{projectID}/rest-apis` (JWT + Permission)

### GraphQL APIs
- `GET /api/v1/graphql-apis` (JWT + Permission)
- `GET/PUT/DELETE /api/v1/graphql-apis/{id}` (JWT + Permission)
- `POST /api/v1/projects/{projectID}/graphql-apis` (JWT + Permission)

### gRPC APIs
- `GET /api/v1/grpc-apis` (JWT + Permission)
- `GET/PUT/DELETE /api/v1/grpc-apis/{id}` (JWT + Permission)
- `POST /api/v1/projects/{projectID}/grpc-apis` (JWT + Permission)

### MinIO Storage
- `GET /api/v1/minio/upload/{object}` (JWT)
- `GET /api/v1/minio/download/{object}` (JWT)

### Permissions (ABAC)
- `GET/POST /api/v1/permissions/attributes` (Admin)
- `GET/POST /api/v1/permissions/resources` (Admin)
- `GET/POST /api/v1/permissions/permissions` (Admin)

### Health & Debug
- `GET /health`, `GET /ping` (Public)
- `/debug/*` (Admin, pprof endpoints)

## Development

```bash
# Testing
go test ./...
go test -race ./...
go test -cover ./...
go test -bench=. -benchmem ./...

# Building
go build -o bin/golang-basic main.go
go build -ldflags="-s -w" -o bin/golang-basic main.go
podman build -t golang-basic:latest .

# Database
psql -U postgres -d golang_basic
pg_dump -U postgres golang_basic > backup.sql
psql -U postgres -d golang_basic < backup.sql
```

## Performance

**Budgets:** API p95 <100ms, DB <50ms p95, Cache hit >95%

**Go:** pgxpool MaxConns=CPU*4, errgroup with limits, strings.Builder for concat

**PostgreSQL:** Index FK/JOIN/WHERE columns, prepared statements, batch queries

**Frontend:** React.lazy code splitting, React.memo for expensive components, virtualize long lists

## Production Deployment

### Environment Variables

```env
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secure_password
DB_NAME=golang_basic

# JWT (RS256) - Generate with: openssl genrsa -out private.pem 2048
JWT_PRIVATE_KEY=base64_encoded_private_key
JWT_PUBLIC_KEY=base64_encoded_public_key

# MinIO
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=secure_access_key
MINIO_SECRET_KEY=secure_secret_key
MINIO_BUCKET_NAME=golang-basic
MINIO_USE_SSL=true
```

### Security Checklist

- [ ] Generate secure RSA keys for JWT
- [ ] Change all default passwords
- [ ] Enable database SSL/TLS
- [ ] Configure Nginx rate limiting
- [ ] Set up database backups
- [ ] Enable log aggregation

### Container Hardening

```yaml
cap_drop: [ALL]
cap_add: [NET_BIND_SERVICE]
read_only: true
tmpfs: ["/tmp"]
security_opt: ["no-new-privileges"]
```

### Scaling

**Horizontal:** Multiple app containers behind Nginx, external PostgreSQL (RDS/Cloud SQL), shared MinIO

**Vertical:** Increase CPU/memory limits, adjust pgxpool MaxConns

## Monitoring

```bash
# Go profiling (http://localhost:3003/debug/)
go tool pprof http://localhost:3003/debug/pprof/profile?seconds=30
go tool pprof http://localhost:3003/debug/pprof/heap

# Container stats
podman stats

# Database slow queries
psql -c "SELECT query, calls, mean_time FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"
```

## Port Reference

| Service | Port |
|---------|------|
| Go API | 3003 |
| PostgreSQL | 5432 |
| Nginx HTTP | 80 |
| Nginx HTTPS | 443 |
| MinIO API | 9000 |
| MinIO Console | 9001 |

## Resources

- [CLAUDE.md](../CLAUDE.md) - Coding guidelines
- [PODMAN.md](PODMAN.md) - Deployment guide
- [ABAC_README.md](../ABAC_README.md) - Authorization docs
- [Chi](https://go-chi.io/) - Router documentation
- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver

## License

MIT
