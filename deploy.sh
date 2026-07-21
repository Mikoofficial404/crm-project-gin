#!/bin/bash
set -e

echo "========================================="
echo "  Miko Tech CRM - Deploy Script"
echo "========================================="

# === Config ===
APP_DIR="/opt/crm-project"
SERVICE_NAME="crm-api"
REPO_URL="https://github.com/Mikoofficial404/crm-project-gin.git"

# === 1. Clone / Pull ===
if [ -d "$APP_DIR/.git" ]; then
    echo "[1/5] Pulling latest code..."
    cd "$APP_DIR"
    git pull origin master
else
    echo "[1/5] Cloning repo..."
    git clone "$REPO_URL" "$APP_DIR"
    cd "$APP_DIR"
fi

# === 2. Check .env ===
if [ ! -f "$APP_DIR/.env" ]; then
    echo "[WARNING] .env not found! Create one or ensure env vars are set in service file."
fi

# === 3. Build ===
echo "[2/5] Building binary..."
go build -o "$APP_DIR/crm-api" ./cmd/api/

# === 4. DB & Redis via Docker ===
echo "[3/5] Starting PostgreSQL & Redis..."
docker-compose up -d crm-postgres crm-redis

# === 5. Install Service ===
echo "[4/5] Installing systemd service..."
cp "$APP_DIR/crm-api.service" /etc/systemd/system/
systemctl daemon-reload
systemctl enable "$SERVICE_NAME"

# === 6. Restart ===
echo "[5/5] Restarting service..."
systemctl restart "$SERVICE_NAME"

echo ""
echo "========================================="
echo "  Deploy Selesai!"
echo "  Cek status: systemctl status $SERVICE_NAME"
echo "  Cek log:    journalctl -u $SERVICE_NAME -f"
echo "========================================="
