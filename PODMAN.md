# Running Application with Podman

This guide provides instructions for running the Golang Basic application using Podman.

**Performance Budgets:** API p95 <100ms, DB queries <50ms p95, Container health <5s

## Prerequisites

- Podman installed on your system
- Podman Compose installed (or use `podman-compose`)

## Environment Variables

Create a `.env` file in the project root:
```env
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=golang_basic
DB_PORT=5432

# JWT (generate secure keys for production)
JWT_SECRET=your_jwt_secret_change_in_production
```

## Option 1: Run Database Only

If you want to run only the PostgreSQL database (for local development where you run the Go app directly):

```bash
# Start only the PostgreSQL container
podman-compose up -d postgres

# Check if the database is running
podman-compose ps

# View database logs
podman-compose logs -f postgres

# Stop the database
podman-compose stop postgres

# Remove the database container
podman-compose down
```

**Connect to the database from your local Go app:**
The database will be available at `localhost:5432`.

## Database Migrations

Run SQL migrations from `sql/` folder after starting the database:

```bash
# Start the database first
podman-compose up -d postgres

# Wait for database to be healthy
podman-compose ps

# Run ABAC schema
podman-compose exec -T postgres psql -U apiproj -d api-doc < sql/abac_schema.sql

# Run API docs schema
podman-compose exec -T postgres psql -U apiproj -d api-doc < sql/api_docs_schema.sql

# Verify tables were created
podman-compose exec postgres psql -U postgres -d golang_basic -c "\dt"

# Verify indexes
podman-compose exec postgres psql -U postgres -d golang_basic -c "\di"
```

**Single migration command:**
```bash
# Run all migrations at once
podman-compose up -d postgres && \
sleep 5 && \
podman-compose exec -T postgres psql -U postgres -d golang_basic < sql/abac_schema.sql && \
podman-compose exec -T postgres psql -U postgres -d golang_basic < sql/api_docs_schema.sql
```

## Option 2: Run Database and Application

To run both the PostgreSQL database and the Golang application in containers:

```bash
# Build and start all services
podman-compose up -d

# Or build with --build flag to rebuild images
podman-compose up -d --build

# Check running containers
podman-compose ps

# View logs for all services
podman-compose logs -f

# View logs for specific service
podman-compose logs -f app
podman-compose logs -f postgres

# Stop all services
podman-compose stop

# Stop and remove all containers
podman-compose down

# Stop and remove all containers with volumes
podman-compose down -v
```

**Access the application:**
- Application API: `https://localhost:3003` (with self-signed certificate)
- Health check: `https://localhost:3003/api/v1/health`
- PostgreSQL: `localhost:5432`

**Verify health status:**
```bash
# Check container health
podman-compose ps

# Check application health endpoint
curl -k https://localhost:3003/api/v1/health
```

## Option 3: Using Podman Pods (Production)

For production, use a pod to share namespace between containers:

```bash
# Create a pod with exposed ports
podman pod create --name golang-basic-pod -p 3003:3003 -p 5432:5432

# Run PostgreSQL in the pod
podman run --pod golang-basic-pod -d \
    --name golang-basic-postgres \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=postgres \
    -e POSTGRES_DB=golang_basic \
    -v ./pgData:/var/lib/postgresql/data:Z \
    postgres:17

# Run application in the pod
podman run --pod golang-basic-pod -d \
    --name golang-basic-app \
    --restart unless-stopped \
    -e DB_HOST=localhost \
    -e DB_PORT=5432 \
    -e DB_USER=postgres \
    -e DB_PASSWORD=postgres \
    -e DB_NAME=golang_basic \
    localhost/golang-basic-app:latest

# Stop pod and all containers
podman pod stop golang-basic-pod

# Remove pod and all containers
podman pod rm golang-basic-pod -f
```

## Useful Commands

### Rebuild the application
```bash
podman-compose build app
podman-compose up -d app
```

### Execute commands inside containers
```bash
# Access PostgreSQL CLI
podman-compose exec postgres psql -U postgres -d golang_basic

# Access application container shell
podman-compose exec app sh
```

### View container resource usage
```bash
podman stats
```

### Clean up everything
```bash
# Remove all containers, networks, and volumes
podman-compose down -v

# Remove unused images
podman image prune -a
```

## Troubleshooting

### Database connection issues
- Ensure the database container is healthy: `podman-compose ps`
- Check database logs: `podman-compose logs postgres`
- Verify environment variables are set correctly in `.env`
- Test database connection: `podman-compose exec postgres psql -U postgres -d golang_basic -c "SELECT 1;"`

### Application not starting
- Check if database is ready: `podman-compose logs postgres`
- Rebuild the application: `podman-compose up -d --build`
- Check application logs: `podman-compose logs app`
- Verify healthcheck: `curl -k https://localhost:3003/api/v1/health`

### Container permission issues (non-root user)
The application runs as non-root user (UID 1001). If you encounter permission issues:
```bash
# Fix volume permissions for database
podman-compose exec postgres chown -R 1001:1001 /var/lib/postgresql/data
```

### Healthcheck failing
- Check if health endpoint exists: `curl -k https://localhost:3003/api/v1/health`
- Verify container is running: `podman-compose ps`
- Check application logs for errors: `podman-compose logs app`
- Inspect healthcheck status: `podman inspect golang-basic-app | grep -A 10 Health`

### Port conflicts
If ports 3003 or 5432 are already in use, modify the port mappings in `docker-compose.yml`:
```yaml
ports:
  - "3004:3003"  # Change host port to 3004
```

### Performance verification
```bash
# Check container resource usage
podman stats

# Verify database query performance
podman-compose exec postgres psql -U postgres -d golang_basic -c "EXPLAIN ANALYZE SELECT * FROM projects LIMIT 10;"

# Check connection pool status
podman-compose exec postgres psql -U postgres -d golang_basic -c "SELECT count(*), state FROM pg_stat_activity GROUP BY state;"
```

## Security Considerations

**Container Security:**
- Non-root user (UID 1001) - aligned with CLAUDE.md
- Read-only root filesystem (add to docker-compose if needed): `read_only: true`
- Drop all capabilities (add to docker-compose): `cap_drop: [ALL]`
- Resource limits enforced (CPU, memory)

**Database Security:**
- SCRAM-SHA-256 authentication enforced
- Separate user per application (recommended for production)
- Network isolation via bridge network
- Volume mounted with `:Z` for SELinux contexts

## Production Checklist

- [ ] Change default passwords in `.env`
- [ ] Generate strong JWT secret
- [ ] Enable TLS/SSL for database connections
- [ ] Set up database backups
- [ ] Configure log aggregation
- [ ] Set up monitoring (pprof, pg_stat_statements)
- [ ] Review and adjust resource limits
- [ ] Enable rate limiting (Nginx)
- [ ] Configure MinIO for file storage
- [ ] Set up CI/CD pipeline
