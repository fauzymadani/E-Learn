# Docker/Podman Deployment Guide

This guide explains how to build and run the E-Learning Platform backend using Docker or Podman.

## Prerequisites

- Podman installed on your system
- PostgreSQL running on localhost (host machine)
- Go 1.23+ (for local development)

## Important Notes

### Network Mode: Host

The docker-compose.yml uses `network_mode: "host"` to allow the container to access your localhost PostgreSQL database. This means:
- The container shares the host's network stack
- The application can connect to `localhost:5432` (your PostgreSQL)
- The application will be accessible on `localhost:8080`

### Database Connection

Make sure your PostgreSQL is configured to accept connections from localhost. Update your `pg_hba.conf` if needed:

```
# IPv4 local connections:
host    all             all             127.0.0.1/32            md5
```

## Quick Start

### Using Podman Compose

1. **Build and start the container:**

```bash
podman-compose up -d --build
```

2. **View logs:**

```bash
podman-compose logs -f backend
```

3. **Stop the container:**

```bash
podman-compose down
```

### Using Podman Directly

1. **Build the image:**

```bash
podman build -t elearning-backend .
```

2. **Run the container:**

```bash
podman run -d \
  --name elearning-backend \
  --network host \
  -v ./uploads:/app/uploads \
  -v ./.env:/app/.env:ro \
  -p 8080:8080 \
  elearning-backend
```

3. **View logs:**

```bash
podman logs -f elearning-backend
```

4. **Stop and remove:**

```bash
podman stop elearning-backend
podman rm elearning-backend
```

## Environment Variables

The application uses the following environment variables (configured in `.env` or docker-compose.yml):

### Database
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name (default: elearning)
- `DB_SSLMODE` - SSL mode (default: disable)

### Server
- `SERVER_PORT` - Server port (default: 8080)
- `GIN_MODE` - Gin mode (release/debug)

### JWT
- `JWT_SECRET` - JWT secret key (change in production!)
- `JWT_EXPIRATION` - Token expiration duration (default: 24h)

### Notification Service (gRPC)
- `NOTIFICATION_GRPC_ADDRESS` - Notification service address (default: localhost:50051)

### Google Cloud Storage
- `GCS_ENABLED` - Enable GCS (default: false)
- `GCS_BUCKET_NAME` - GCS bucket name

## Configuration Options

### Using .env File

Create a `.env` file in the project root:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=elearning
DB_SSLMODE=disable

# Server
SERVER_PORT=8080
GIN_MODE=release

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-this
JWT_EXPIRATION=24h

# Notification gRPC
NOTIFICATION_GRPC_ADDRESS=localhost:50051

# GCS
GCS_ENABLED=false
```

### Override in docker-compose.yml

You can also override environment variables in the `docker-compose.yml` file.

## Troubleshooting

### Container can't connect to PostgreSQL

1. **Check if PostgreSQL is running:**
   ```bash
   systemctl status postgresql
   # or
   sudo systemctl status postgresql
   ```

2. **Check PostgreSQL is listening on localhost:**
   ```bash
   netstat -an | grep 5432
   # or
   ss -an | grep 5432
   ```

3. **Test connection from host:**
   ```bash
   psql -h localhost -U postgres -d elearning
   ```

4. **Check pg_hba.conf allows local connections:**
   ```bash
   sudo cat /var/lib/postgres/data/pg_hba.conf | grep 127.0.0.1
   ```

### Permission Issues with Uploads

If you encounter permission issues with the uploads directory:

```bash
# Fix permissions
chmod -R 755 uploads/
chown -R 1000:1000 uploads/
```

### Container Health Check Failing

The health check uses `wget` to check `/health` endpoint. If it fails:

1. Check container logs:
   ```bash
   podman logs elearning-backend
   ```

2. Check if the application started:
   ```bash
   podman exec -it elearning-backend ps aux
   ```

3. Test the health endpoint manually:
   ```bash
   curl http://localhost:8080/health
   ```

### Rebuilding After Code Changes

```bash
# Stop and remove old container
podman-compose down

# Rebuild and start
podman-compose up -d --build

# Or with Podman directly
podman stop elearning-backend
podman rm elearning-backend
podman build -t elearning-backend .
podman run -d --name elearning-backend --network host -v ./uploads:/app/uploads elearning-backend
```

## Production Deployment

For production deployment, consider:

1. **Use a reverse proxy** (nginx/traefik) in front of the application
2. **Enable HTTPS** with valid SSL certificates
3. **Change default secrets** in environment variables
4. **Set proper file permissions** for uploads directory
5. **Configure log rotation** for container logs
6. **Set resource limits** in docker-compose.yml:

```yaml
services:
  backend:
    # ... other config
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 1G
        reservations:
          cpus: '1'
          memory: 512M
```

## Development vs Production

### Development

```bash
# Use debug mode
GIN_MODE=debug podman-compose up
```

### Production

```bash
# Use release mode
GIN_MODE=release podman-compose up -d
```

## Useful Commands

```bash
# View running containers
podman ps

# View all containers (including stopped)
podman ps -a

# View container logs
podman logs -f elearning-backend

# Execute command in container
podman exec -it elearning-backend sh

# View container stats
podman stats elearning-backend

# Inspect container
podman inspect elearning-backend

# Remove all stopped containers
podman container prune

# Remove all unused images
podman image prune
```

## Podman vs Docker

This setup works with both Podman and Docker. To use Docker instead:

```bash
# Replace 'podman' with 'docker'
docker-compose up -d --build

# Or
docker build -t elearning-backend .
docker run -d --name elearning-backend --network host elearning-backend
```

The main difference is that Podman runs rootless by default, which is more secure.

