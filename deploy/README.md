# Deployment Guide

## Quick Start

### Development
```bash
# Start all services in development mode
make dev
```

### Production
```bash
# 1. Copy and configure environment
cp .env.example .env
# Edit .env with your production values

# 2. Build all images
make build

# 3. Generate SSL certificates (choose one):

# Option A: Self-signed (development/testing)
make ssl-dev

# Option B: Let's Encrypt (production)
# First, update DOMAIN in .env, then:
make ssl-init

# 4. Deploy
make deploy

# 5. View logs
make logs
```

## Services

| Service   | Port  | Description                          |
|-----------|-------|--------------------------------------|
| nginx     | 80/443| Reverse proxy with SSL termination   |
| frontend  | 3000  | Next.js application                  |
| api       | 8080  | Go backend API                       |
| worker    | 8081  | Background job processor             |
| postgres  | 5432  | PostgreSQL database                  |
| redis     | 6379  | Redis cache                          |
| certbot   | -     | SSL certificate renewal              |

## Makefile Commands

### Primary Commands
- `make dev` - Run locally in development mode
- `make build` - Build all Docker images
- `make deploy` - Deploy to production (docker-compose up -d)
- `make logs` - Tail all service logs
- `make migrate` - Run database migrations

### SSL Commands
- `make ssl-init` - Initialize Let's Encrypt certificates
- `make ssl-renew` - Renew SSL certificates
- `make ssl-dev` - Generate self-signed certs for development

### Utility Commands
- `make status` - Check service status and health
- `make stop` - Stop all services
- `make restart` - Restart all services
- `make logs-api` - View API logs only
- `make logs-frontend` - View frontend logs only
- `make logs-nginx` - View nginx logs only

### Database Commands
- `make db-backup` - Create database backup
- `make db-restore file=backup.sql` - Restore from backup
- `make shell-postgres` - Connect to PostgreSQL shell
- `make shell-redis` - Connect to Redis CLI

## SSL Setup

### Let's Encrypt (Production)

1. Ensure your domain points to your server
2. Update DOMAIN in `.env`:
   ```
   DOMAIN=yourdomain.com
   ```
3. Run SSL initialization:
   ```bash
   make ssl-init
   ```
4. Certificates are stored in `certbot/conf/`
5. Auto-renewal runs every 12 hours via certbot container

### Self-Signed (Development)

```bash
make ssl-dev
```

Or manually with OpenSSL:
```bash
mkdir -p deploy/nginx/ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout deploy/nginx/ssl/privkey.pem \
  -out deploy/nginx/ssl/fullchain.pem \
  -subj "/CN=localhost"
cp deploy/nginx/ssl/fullchain.pem deploy/nginx/ssl/chain.pem
```

## Environment Variables

### Required for Production
```env
# Database
DB_PASSWORD=<strong-password>

# Redis
REDIS_PASSWORD=<strong-password>

# JWT Secrets
JWT_SECRET=<32+ char random string>
ENCRYPTION_KEY=<32 byte key>

# Domain
DOMAIN=yourdomain.com

# Platform API Keys
META_APP_ID=...
META_APP_SECRET=...
TIKTOK_APP_ID=...
TIKTOK_APP_SECRET=...
```

### Frontend
In production, the frontend uses relative URLs:
```env
NEXT_PUBLIC_API_URL=/api/v1
```
Nginx proxies `/api/` to the backend service.

## Nginx Configuration

### SSE (Server-Sent Events)
The nginx config includes special handling for SSE at `/api/v1/events/`:
- Buffering disabled
- No caching
- 24-hour connection timeout
- Chunked transfer encoding

### WebSocket (Future)
WebSocket support is pre-configured at `/api/v1/ws/`.

### Rate Limiting
- General API: 10 req/s with burst of 20
- Auth endpoints: 5 req/min
- Webhooks: 100 req/s

## Health Checks

All services include health checks:
- Nginx: `GET /nginx-health`
- API: `GET /health`
- Frontend: `GET /`

Check status:
```bash
make status
```

## Troubleshooting

### Containers not starting
```bash
# Check logs
make logs

# Check container status
docker-compose ps
```

### SSL Issues
```bash
# Test nginx config
docker-compose exec nginx nginx -t

# View certificate details
openssl x509 -in deploy/nginx/ssl/fullchain.pem -text -noout
```

### Database Connection
```bash
# Connect to database
make shell-postgres

# Check connection from API container
docker-compose exec api sh -c 'wget -qO- http://localhost:8080/health'
```

### Reset Everything
```bash
# WARNING: Removes all data!
make nuke
```
