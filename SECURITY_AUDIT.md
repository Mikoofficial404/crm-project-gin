# 🔐 CRM Project — Security Audit Report

## Overview
- **Stack:** Go 1.26, Gin, GORM + PostgreSQL, Redis, Docker Compose
- **Auth:** JWT (HS256), Argon2id, TOTP 2FA, Google OAuth
- **Audit Date:** 2026-07-21

---

## 🔴 CRITICAL (Sudah Diperbaiki)

### 1. `.env` EXPOSED SECRETS — Semua credential bocor di git history
✅ **Fixed** — `.env` dihapus dari git history via `git filter-branch`

### 2. `USER` MODEL LEAKING SENSITIVE FIELDS
✅ **Fixed** — `json:"-"` untuk `Password`, `TwoFactorSecret`, `RefreshToken`, `RefreshTokenExpiry`

### 3. PUBLIC FILE UPLOADS
✅ **Fixed** — Static serve dihapus, upload validasi ekstensi + MIME, download via `GET /api/v1/files/:file_id`

### 4. NO FILE TYPE RESTRICTION
✅ **Fixed** — CSV import: 10MB limit + ekstensi `.csv` validation

---

## 🟠 HIGH (Sudah Diperbaiki)

5. CORS — `AllowOrigins` spesifik dari env `CORS_ORIGIN`
6. RateLimitUser — terpasang di semua protected routes
7. DB SSL — `sslmode=require` di production
8. Redis — password dari `REDIS_PASSWORD`
9. Error messages — `ServerError()` helper + logrus
10. WebSocket — CheckOrigin validasi `CORS_ORIGIN`
11. Webhook — wajibkan `WEBHOOK_SECRET` di production
12. OAuth — state random per-request, Redis-backed
13. Ownership checks — GET lead/deal/contact/task by ID

---

## 🟡 MEDIUM (Sudah Diperbaiki)

14. Metrics + Swagger — only in dev mode
15. Account lockout — 5x failed = 15 menit
16. Password complexity — 8 char, upper, lower, num, symbol
17. 2FA recovery codes — 8 codes + `/login/recovery`
18. Refresh token — `POST /auth/refresh`
19. Security headers — HSTS, X-Frame, nosniff, dll
20. Docker Compose — env vars bukan hardcoded
21. Email enumeration — uniform message

---

## ⚠️ TINDAKAN SETELAH AUDIT

- **Rotate semua secret** (Google OAuth, Gemini, SMTP, WhatsApp)
- **Generate JWT_SECRET baru**
- **Set env vars di VPS**, jangan copy `.env`
- **Force push** ke remote setelah history di-rewrite
