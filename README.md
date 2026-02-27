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
go run main.go -migrate    # Run migrations first
go run main.go              # Start server

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

## Database Migrations

The application includes a built-in migration system that tracks applied migrations and supports rollback.

### Built-in Migration (Recommended)

```bash
# Run pending migrations
go run main.go -migrate

# Rollback last migration
go run main.go -migrate-down

# Check migration status
go run main.go -migrate-status

# Run migrations then start server
go run main.go -migrate && go run main.go
```

**Benefits:**
- Tracks which migrations have been applied
- Safe transactions with automatic rollback on failure
- Version tracking prevents duplicate migrations
- Rollback support for schema changes

### Manual Migration (Alternative)

```bash
# Local psql
psql -U postgres -d golang_basic < sql/abac_schema.sql
psql -U postgres -d golang_basic < sql/api_docs_schema.sql

# Podman container
podman-compose exec -T postgres psql -U postgres -d golang_basic < sql/abac_schema.sql
podman-compose exec -T postgres psql -U postgres -d golang_basic < sql/api_docs_schema.sql

# Verify tables
podman-compose exec postgres psql -U postgres -d golang_basic -c "\dt"
podman-compose exec postgres psql -U postgres -d golang_basic -c "\di"
```

### Migration Files

Naming convention: `YYYYMMDD_XXX_description.sql`

| File | Version | Description |
|------|---------|-------------|
| `sql/abac_schema.sql` | 20260225001 | ABAC authorization schema (attributes, policies, permissions) |
| `sql/api_docs_schema.sql` | 20260225002 | API documentation schema (projects, REST/GraphQL/gRPC APIs) |

**Important:** Migrations run in version order automatically. The built-in system ensures correct execution order.

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
