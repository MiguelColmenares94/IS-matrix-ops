#!/usr/bin/env bash
set -euo pipefail

# ── Colours ──────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
info()  { echo -e "${GREEN}[+]${NC} $*"; }
warn()  { echo -e "${YELLOW}[!]${NC} $*"; }
die()   { echo -e "${RED}[✗]${NC} $*" >&2; exit 1; }

# ── Must run from project root ────────────────────────────────────────────────
cd "$(dirname "$0")"
[[ -f docker-compose.yml ]] || die "Run this script from the IS-matrix-ops project root."

# ── Collect inputs ────────────────────────────────────────────────────────────
echo ""
echo "=== IS-matrix-ops Pre-Deploy Setup ==="
echo ""
read -rp "DB password for 'matrixops' user:       " DB_PASSWORD
read -rp "Evaluator login password:                " SEED_PASSWORD
read -rp "JWT secret (leave blank to auto-generate): " JWT_SECRET
[[ -z "$JWT_SECRET" ]] && JWT_SECRET=$(openssl rand -hex 32)
read -rp "App origin URL [http://localhost]:       " APP_ORIGIN
APP_ORIGIN="${APP_ORIGIN:-http://localhost}"
echo ""

# ── 1. Install prerequisites ──────────────────────────────────────────────────
info "Step 1/6 — Installing prerequisites"

install_if_missing() {
  dpkg -s "$1" &>/dev/null || sudo apt-get install -y "$1"
}

sudo apt-get update -qq

# Docker
if ! command -v docker &>/dev/null; then
  info "Installing Docker..."
  sudo apt-get install -y ca-certificates curl
  sudo install -m 0755 -d /etc/apt/keyrings
  sudo curl -fsSL https://download.docker.com/linux/$(. /etc/os-release && echo "$ID")/gpg \
    -o /etc/apt/keyrings/docker.asc
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] \
    https://download.docker.com/linux/$(. /etc/os-release && echo "$ID") \
    $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
    sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
  sudo apt-get update -qq
  sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
  sudo usermod -aG docker "$USER"
  warn "Docker installed. You may need to log out and back in for group changes."
else
  info "Docker already installed: $(docker --version)"
fi

# PostgreSQL 15
if ! command -v psql &>/dev/null; then
  info "Installing PostgreSQL 15..."
  sudo apt-get install -y curl ca-certificates gnupg
  sudo install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | \
    sudo gpg --dearmor -o /etc/apt/keyrings/postgresql.gpg
  echo "deb [signed-by=/etc/apt/keyrings/postgresql.gpg] \
    https://apt.postgresql.org/pub/repos/apt $(. /etc/os-release && echo "$VERSION_CODENAME")-pgdg main" | \
    sudo tee /etc/apt/sources.list.d/pgdg.list > /dev/null
  sudo apt-get update -qq
  sudo apt-get install -y postgresql-15
  sudo systemctl enable postgresql
  sudo systemctl start postgresql
else
  info "PostgreSQL already installed: $(psql --version)"
fi

# Python bcrypt
if ! python3 -c "import bcrypt" &>/dev/null; then
  info "Installing python3-bcrypt..."
  install_if_missing python3-bcrypt
else
  info "python3-bcrypt already installed"
fi

# ── 2. PostgreSQL configuration ───────────────────────────────────────────────
info "Step 2/6 — Configuring PostgreSQL"

# Create user and DB (ignore errors if they already exist)
sudo -u postgres psql -c "CREATE USER matrixops WITH PASSWORD '${DB_PASSWORD}';" 2>/dev/null || warn "User 'matrixops' already exists"
sudo -u postgres psql -c "CREATE DATABASE matrixops OWNER matrixops;" 2>/dev/null || warn "Database 'matrixops' already exists"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE matrixops TO matrixops;" 2>/dev/null || true

# Allow Docker bridge network (172.x.x.x) in pg_hba.conf
PG_HBA=$(sudo find /etc/postgresql -name pg_hba.conf | head -1)
if ! sudo grep -q "172.0.0.0/8" "$PG_HBA"; then
  sudo bash -c "echo 'host    matrixops       matrixops       172.0.0.0/8             md5' >> $PG_HBA"
  info "Added Docker network rule to pg_hba.conf"
fi

# Ensure listen_addresses = '*'
PG_CONF=$(sudo find /etc/postgresql -name postgresql.conf | head -1)
if ! sudo grep -qE "^listen_addresses\s*=\s*'\*'" "$PG_CONF"; then
  sudo sed -i "s/#\?listen_addresses\s*=\s*'[^']*'/listen_addresses = '*'/" "$PG_CONF"
  info "Set listen_addresses = '*' in postgresql.conf"
fi

sudo systemctl reload postgresql

# ── 3. Generate lockfiles ─────────────────────────────────────────────────────
info "Step 3/6 — Generating lockfiles"

if [[ ! -f api-go/go.sum ]]; then
  info "Generating go.sum..."
  docker run --rm -v "$(pwd)/api-go:/app" -w /app golang:1.22-alpine go mod tidy
else
  info "go.sum already exists"
fi

if [[ ! -f api-node/package-lock.json ]]; then
  info "Generating api-node/package-lock.json..."
  docker run --rm -v "$(pwd)/api-node:/app" -w /app node:20-alpine npm install --silent
else
  info "api-node/package-lock.json already exists"
fi

if [[ ! -f frontend/package-lock.json ]]; then
  info "Generating frontend/package-lock.json..."
  docker run --rm -v "$(pwd)/frontend:/app" -w /app node:20-alpine npm install --silent
else
  info "frontend/package-lock.json already exists"
fi

# ── 4. Create .env ────────────────────────────────────────────────────────────
info "Step 4/6 — Creating .env"

if [[ -f .env ]]; then
  warn ".env already exists — skipping (delete it manually to regenerate)"
else
  cat > .env <<EOF
DATABASE_URL=postgres://matrixops:${DB_PASSWORD}@host.docker.internal:5432/matrixops
JWT_SECRET=${JWT_SECRET}
JWT_EXPIRY_MINUTES=10
JWT_REFRESH_EXPIRY_DAYS=7
ALLOWED_ORIGIN=${APP_ORIGIN}
VITE_API_BASE_URL=${APP_ORIGIN}
EOF
  info ".env created"
fi

# ── 5. Seed evaluator user ────────────────────────────────────────────────────
info "Step 5/6 — Seeding evaluator user"

if grep -q "PLACEHOLDER_HASH" db/migrations/000003_seed.up.sql; then
  bash db/seed.sh "${SEED_PASSWORD}"
else
  warn "Seed migration already has a hash — skipping seed.sh"
fi

# ── 6. Build and start ────────────────────────────────────────────────────────
info "Step 6/6 — Building and starting containers"
docker compose up --build -d

# ── Verify ────────────────────────────────────────────────────────────────────
echo ""
info "Waiting for services to be ready..."
sleep 5

HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${APP_ORIGIN}/")
if [[ "$HTTP_CODE" == "200" ]]; then
  info "Frontend reachable at ${APP_ORIGIN}/ ✓"
else
  warn "Frontend returned HTTP ${HTTP_CODE} — check: docker compose logs frontend"
fi

LOGIN_RESP=$(curl -s -X POST "${APP_ORIGIN}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"evaluator@example.com\",\"password\":\"${SEED_PASSWORD}\"}")

if echo "$LOGIN_RESP" | grep -q "access_token"; then
  info "Login endpoint working ✓"
else
  warn "Login failed — response: ${LOGIN_RESP}"
fi

echo ""
echo "=== Setup complete ==="
echo ""
echo "  App:      ${APP_ORIGIN}/"
echo "  Email:    evaluator@example.com"
echo "  Password: ${SEED_PASSWORD}"
echo ""
