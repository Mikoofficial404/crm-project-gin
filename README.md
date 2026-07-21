# Miko Tech CRM

Sistem CRM (Customer Relationship Management) dengan integrasi WhatsApp, AI chatbot, campaign management, deal pipeline, dan team collaboration.

## Tech Stack

- **Go 1.26** — Backend
- **Gin** — HTTP framework
- **PostgreSQL 15** — Database (GORM)
- **Redis 7** — Cache / Session / Queue
- **Asynq** — Background task queue
- **WebSocket** — Real-time notifications
- **Docker Compose** — Local infrastructure

## Fitur

- Lead & Contact Management (CSV import/export)
- Deal Pipeline (Kanban stages, SLA tracking, PDF/Excel export)
- WhatsApp Integration (inbound webhook → auto-lead → AI reply)
- AI Chatbot (Google Gemini)
- Campaign (Email + WhatsApp scheduling)
- Task Management (due dates, reminders)
- Dashboard & Analytics
- Authentication (JWT + Argon2id + 2FA TOTP + Google OAuth)
- RBAC (Admin / Sales roles)
- Real-time Notifications (WebSocket)
- Audit Log & Login History

## Quick Start

### Prasyarat

- Go 1.26+
- Docker & Docker Compose
- PostgreSQL 15, Redis 7

### Setup

```bash
# Clone repo
git clone https://github.com/your-username/crm-project.git
cd crm-project

# Copy environment config
cp .env.example .env
# Edit .env dengan credential kamu

# Start database & Redis
docker-compose up -d crm-postgres crm-redis

# Run aplikasi
go run ./cmd/api/
```

Akses API di `http://localhost:8080`

Swagger docs tersedia di `http://localhost:8080/swagger/index.html` (development mode only)

## Environment Variables

| Variable | Deskripsi |
|----------|-----------|
| `APP_ENV` | `development` atau `production` |
| `APP_PORT` | Port aplikasi (default: 8080) |
| `JWT_SECRET` | Secret key untuk JWT signing |
| `DB_HOST/PORT/USER/PASSWORD/NAME` | PostgreSQL config |
| `DB_SSLMODE` | `disable` (dev) atau `require` (production) |
| `REDIS_HOST/PORT/PASSWORD` | Redis config |
| `CORS_ORIGIN` | Frontend origin untuk CORS |
| `GOOGLE_CLIENT_ID/SECRET` | Google OAuth credentials |
| `GEMINI_API_KEY` | Google Gemini API key |
| `SMTP_*` | Email (Mailtrap untuk dev, SMTP production) |
| `WA_GOWA_URL/DEVICE_ID/BASIC_AUTH` | WhatsApp gateway |
| `WEBHOOK_SECRET` | Secret untuk webhook auth (production) |
| `SENTRY_DSN` | Sentry error tracking |

## Deployment (VPS)

```bash
# Set environment variables di systemd/docker
export APP_ENV=production
export CORS_ORIGIN=https://domain-kamu.com
export DB_SSLMODE=require
export REDIS_PASSWORD=your-password
export WEBHOOK_SECRET=random-string

# Build binary
go build -o crm-api ./cmd/api/

# Run
./crm-api
```

**Penting:** Jangan copy `.env` ke server. Gunakan environment variables langsung dari VPS/systemd.

## License

MIT
