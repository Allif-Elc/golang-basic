# Running Application with Podman

This guide provides instructions for running the Golang Basic application using Podman.

## Prerequisites

- Podman installed on your system
- Podman Compose installed (or use `podman-compose`)

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

**Environment Variables for Database:**
Create a `.env` file in the project root:
```env
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=golang_basic
DB_PORT=5432
```

**Connect to the database from your local Go app:**
The database will be available at `localhost:5432`.

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
- Application API: `http://localhost:3003`
- PostgreSQL: `localhost:5432`

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
- Verify environment variables are set correctly

### Application not starting
- Check if database is ready: `podman-compose logs postgres`
- Rebuild the application: `podman-compose up -d --build`
- Check application logs: `podman-compose logs app`

### Port conflicts
If ports 3003 or 5432 are already in use, modify the port mappings in `docker-compose.yml`:
```yaml
ports:
  - "3004:3003"  # Change host port to 3004
```
