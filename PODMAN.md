# Podman Deployment Guide

Complete guide for running the Golang Basic application using Podman containers.

**Performance Budgets:** API p95 <100ms, DB queries <50ms p95, Container health <5s

## Quick Start

One-command setup for local development:
```bash
# Clone and setup
podman-compose up -d && \
sleep 5 && \
podman-compose exec -T postgres psql -U postgres -d golang_basic < sql/abac_schema.sql && \
podman-compose exec -T postgres psql -U postgres -d golang_basic < sql/api_docs_schema.sql

# Verify health
curl -k https://localhost:3003/health
```

## Prerequisites

- Podman installed on your system
- Podman Compose installed (or use `podman-compose`)

## Environment Variables

Create a `.env` file in the project root:

```env
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=golang_basic

# JWT RSA Keys (RS256 algorithm)
# Generate new keys with:
#   openssl genrsa -out private.pem 2048
#   openssl rsa -in private.pem -pubout -out public.pem
# Then convert to base64 single line:
#   cat private.pem | tr -d '\n' > private_key.txt
#   cat public.pem | tr -d '\n' > public_key.txt
JWT_PRIVATE_KEY=your_base64_encoded_private_key_here
JWT_PUBLIC_KEY=your_base64_encoded_public_key_here

# MinIO Configuration
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET_NAME=golang-basic
MINIO_USE_SSL=false
```

## Deployment Options

### Option 1: Database Only (Local Development)

Run only PostgreSQL for local Go development:

```bash
# Start PostgreSQL
podman-compose up -d postgres

# Check status
podman-compose ps

# View logs
podman-compose logs -f postgres

# Stop
podman-compose stop postgres
```

**Database connection:** `localhost:5432`

### Option 2: Full Stack (Containers)

Run all services including the application:

```bash
# Build and start all services
podman-compose up -d

# Rebuild with latest changes
podman-compose up -d --build

# Check status
podman-compose ps

# View logs
podman-compose logs -f
podman-compose logs -f app

# Stop all services
podman-compose stop

# Remove containers and volumes
podman-compose down -v
```

**Access points:**
- API: `https://localhost:3003` (self-signed cert)
- Health: `https://localhost:3003/health`
- PostgreSQL: `localhost:5432`
- MinIO Console: `http://localhost:9001`

### Option 3: Production Pods

For production, use a pod for shared namespace:

```bash
# Create pod with exposed ports
podman pod create --name golang-basic-pod -p 3003:3003 -p 5432:5432 -p 9000:9000

# Run PostgreSQL in pod
podman run --pod golang-basic-pod -d \
    --name golang-basic-postgres \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=postgres \
    -e POSTGRES_DB=golang_basic \
    -v ./pgData:/var/lib/postgresql/data:Z \
    postgres:17

# Run application in pod
podman run --pod golang-basic-pod -d \
    --name golang-basic-app \
    --restart unless-stopped \
    -e DB_HOST=localhost \
    -e DB_PORT=5432 \
    -e DB_USER=postgres \
    -e DB_PASSWORD=postgres \
    -e DB_NAME=golang_basic \
    localhost/golang-basic-app:latest

# Stop and remove pod
podman pod stop golang-basic-pod
podman pod rm golang-basic-pod -f
```

## Database Migrations

Run SQL migrations after starting the database:

```bash
# Single command for all migrations
podman-compose up -d postgres && \
sleep 5 && \
podman-compose exec -T postgres psql -U postgres -d golang_basic < sql/abac_schema.sql && \
podman-compose exec -T postgres psql -U postgres -d golang_basic < sql/api_docs_schema.sql

# Verify tables
podman-compose exec postgres psql -U postgres -d golang_basic -c "\dt"

# Verify indexes
podman-compose exec postgres psql -U postgres -d golang_basic -c "\di"
```

## Performance Monitoring

### Go Profiling

The application exposes pprof endpoints at `/debug`:

```bash
# CPU profile (30 seconds)
curl http://localhost:3003/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof -http=:8080 cpu.prof

# Memory profile
curl http://localhost:3003/debug/pprof/heap > heap.prof
go tool pprof -http=:8080 heap.prof

# Goroutine profile
curl http://localhost:3003/debug/pprof/goroutine > goroutine.prof

# Interactive profiling
go tool pprof http://localhost:3003/debug/pprof/profile
```

### Database Monitoring

```sql
-- Enable pg_stat_statements (run once)
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Slow queries (top 10 by mean time)
SELECT query, calls, mean_time, total_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- Index usage (find unused indexes)
SELECT indexname, idx_scan
FROM pg_stat_user_indexes
ORDER BY idx_scan ASC;

-- Connection pool status
SELECT count(*), state
FROM pg_stat_activity
GROUP BY state;

-- Table bloat analysis
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

### Container Resource Monitoring

```bash
# Real-time resource usage
podman stats

# Container-specific stats
podman stats golang-basic-app golang-basic-postgres

# Health check status
podman inspect golang-basic-app | grep -A 10 Health
```

## Nginx Reverse Proxy

The Nginx container provides rate limiting, caching, and SSL termination.

**Configuration:** `nginx/nginx.conf`

### Rate Limits

| Zone | Rate | Burst | Applied To |
|------|------|-------|------------|
| `api_limit` | 10 req/s | 20 | `/api/*` |
| `login_limit` | 5 req/s | - | `/api/v1/auth/login` |

### Caching

```nginx
# API cache: 1 minute for 200 responses
proxy_cache_path /var/cache/nginx/api keys_zone=api_cache:10m;

# Static cache: 24 hours
proxy_cache_path /var/cache/nginx/static keys_zone=static_cache:10m;
```

### SSL/TLS

- Protocols: TLSv1.2, TLSv1.3
- HTTP/2 enabled
- Self-signed certs (replace in production)
- HSTS header enabled

### Nginx Commands

```bash
# Test configuration
podman-compose exec nginx nginx -t

# Reload config (no downtime)
podman-compose exec nginx nginx -s reload

# View access logs
podman-compose logs -f nginx

# Cache statistics
podman-compose exec nginx ls -lh /var/cache/nginx/
```

## Useful Commands

### Application

```bash
# Rebuild application
podman-compose build app
podman-compose up -d app

# Access application shell
podman-compose exec app sh

# View application logs
podman-compose logs -f app
```

### Database

```bash
# Access PostgreSQL CLI
podman-compose exec postgres psql -U postgres -d golang_basic

# Run single query
podman-compose exec postgres psql -U postgres -d golang_basic -c "SELECT version();"

# Database backup
podman-compose exec postgres pg_dump -U postgres golang_basic > backup.sql

# Restore from backup
podman-compose exec -T postgres psql -U postgres golang_basic < backup.sql
```

### Maintenance

```bash
# View container resource usage
podman stats

# Remove unused images
podman image prune -a

# Clean up everything
podman-compose down -v
podman volume prune
```

## Troubleshooting

### Database Connection Issues

```bash
# Check database health
podman-compose ps
podman-compose logs postgres

# Test connection
podman-compose exec postgres psql -U postgres -d golang_basic -c "SELECT 1;"

# Check network
podman network inspect golang-basic_golang-basic-network
```

### Application Not Starting

```bash
# Check if database is ready
podman-compose logs postgres

# Rebuild application
podman-compose up -d --build

# Check application logs
podman-compose logs app

# Verify health endpoint
curl -k https://localhost:3003/health
```

### Performance Issues

```bash
# Check container resources
podman stats

# Profile CPU (30 seconds)
curl http://localhost:3003/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof -http=:8080 cpu.prof

# Check slow queries
podman-compose exec postgres psql -U postgres -d golang_basic -c "
SELECT query, calls, mean_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;"
```

### Permission Issues

The application runs as non-root user (UID 1001). Fix volume permissions:

```bash
# Fix database permissions
podman-compose exec postgres chown -R 1001:1001 /var/lib/postgresql/data
```

### Port Conflicts

If ports are in use, modify mappings in `docker-compose.yml`:

```yaml
ports:
  - "3004:3003"  # Change host port to 3004
```

## Production Security

### Container Hardening (Per CLAUDE.md)

```yaml
# Add to docker-compose.yml for production
cap_drop:
  - ALL
cap_add:
  - NET_BIND_SERVICE
read_only: true
tmpfs:
  - /tmp
security_opt:
  - no-new-privileges
```

### Database Security

- SCRAM-SHA-256 authentication enforced
- Separate user per application
- Network isolation via bridge network
- Volume mounted with `:Z` for SELinux

### Network Security

- All services on internal bridge network
- Only Nginx exposes ports (80/443)
- Database not accessible from outside
- Rate limiting on all public endpoints

## Production Checklist

- [ ] Change default passwords in `.env`
- [ ] Generate strong RSA keys for JWT
- [ ] Replace self-signed SSL certificates
- [ ] Enable TLS/SSL for database connections
- [ ] Set up automated database backups
- [ ] Configure log aggregation
- [ ] Enable pprof monitoring (limit access in production)
- [ ] Enable pg_stat_statements extension
- [ ] Review and adjust resource limits
- [ ] Set up MinIO backups
- [ ] Configure CI/CD pipeline
- [ ] Set up health check alerts
- [ ] Configure firewall rules
- [ ] Enable security scanning in CI/CD

## Port Reference

| Service | Internal Port | External Port |
|---------|---------------|---------------|
| Go API | 3003 | 3003 |
| PostgreSQL | 5432 | 5432 |
| Nginx HTTP | 80 | 80 |
| Nginx HTTPS | 443 | 443 |
| MinIO API | 9000 | 9000 |
| MinIO Console | 9001 | 9001 |
