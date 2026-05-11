# IS-matrix-ops

## Overview

Full-stack matrix operations application. Users authenticate, submit a matrix for QR
factorization (API 1 — Go), then compute statistics on the result (API 2 — Node.js).
All services run as Docker containers behind an Nginx reverse proxy. PostgreSQL runs
as a host process on the EC2 instance.

**Architecture:**
- `api-go` — Go/Fiber: auth (JWT issuance) + QR factorization
- `api-node` — Node.js/Express: statistics computation
- `frontend` — Vue 3/Vite: single-page app
- `nginx` — reverse proxy (only public port: 80)
- `db/` — PostgreSQL migrations via golang-migrate

---

## Prerequisites

| Tool         | Version     |
|--------------|-------------|
| Go           | 1.22+       |
| Node.js      | 20 LTS      |
| Docker       | 24+         |
| Docker Compose plugin | 2.x |
| PostgreSQL   | 15+         |
| Python 3     | 3.x (for seed.sh) |

---

## Local Setup

### 1. Clone

```sh
git clone <repo-url>
cd IS-matrix-ops
```

### 2. Environment variables

```sh
cp api-go/.env.example api-go/.env
cp api-node/.env.example api-node/.env
cp frontend/.env.example frontend/.env
```

Edit each `.env` file and fill in `JWT_SECRET`, `DATABASE_URL`, `ALLOWED_ORIGIN`,
and `VITE_API_BASE_URL`.

### 3. Database setup

Create the database and user in PostgreSQL:

```sh
psql -U postgres -c "CREATE USER matrixops WITH PASSWORD 'yourpassword';"
psql -U postgres -c "CREATE DATABASE matrixops OWNER matrixops;"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE matrixops TO matrixops;"
```

### 4. Seed the evaluator user

```sh
pip install bcrypt
bash db/seed.sh "your-evaluator-password"
```

### 5. Run migrations

```sh
docker compose up migrate
```

### 6. Run each service locally (without Docker)

**API 1 (Go):**
```sh
cd api-go
go mod download
go run ./cmd/main.go
```

**API 2 (Node.js):**
```sh
cd api-node
npm install
node server.js
```

**Frontend:**
```sh
cd frontend
npm install
npm run dev
```

---

## Running with Docker Compose

```sh
# Copy and fill in the root .env (Docker Compose reads it)
cp api-go/.env.example .env   # then edit with all variables

docker compose up -d
```

All services start after migrations complete. Access the app at `http://localhost/`.

---

## API Reference

### API 1 — Go (port 8080, via Nginx at `/api/v1/`)

#### `POST /api/v1/auth/login`
No auth required.
```json
// Request
{ "email": "user@example.com", "password": "plaintext" }

// Response 200
{ "access_token": "<jwt>", "refresh_token": "<uuid>", "expires_at": "2026-05-10T10:10:00Z" }

// Errors: 400 (bad input), 401 (wrong credentials)
```

#### `POST /api/v1/auth/refresh`
No auth required.
```json
// Request
{ "refresh_token": "<uuid>" }

// Response 200 — same shape as login
// Errors: 400, 401
```

#### `POST /api/v1/auth/logout`
Requires `Authorization: Bearer <access_token>`.
```json
// Request
{ "refresh_token": "<uuid>" }

// Response 204 (no body)
// Errors: 400, 401
```

#### `POST /api/v1/matrix/qr`
Requires `Authorization: Bearer <access_token>`.
```json
// Request
{ "matrix": [[1, 2, 3], [4, 5, 6], [7, 8, 9]] }

// Response 200
{ "q": [[...], ...], "r": [[...], ...] }

// Error 400
{ "error": "Invalid matrix: ...", "example": { "matrix": [[1, 2], [3, 4], [5, 6]] } }
```

---

### API 2 — Node.js (port 3000, via Nginx at `/api/v2/`)

#### `POST /api/v2/stats`
Requires `Authorization: Bearer <access_token>`.
```json
// Request
{ "q": [[...], ...], "r": [[...], ...] }

// Response 200
{ "max": 9.123, "min": -3.456, "avg": 1.234, "sum": 44.567, "q_diagonal": false, "r_diagonal": false }

// Errors: 400 (missing q/r), 401
```

---

## Running Tests

### Integration tests (requires running PostgreSQL + migrations)

```sh
docker compose --profile test up test-api-go test-api-node
```

Both test containers exit 0 on success.

### Unit tests only

**API 1 (Go):**
```sh
cd api-go
go test ./internal/...
```

**API 2 (Node.js):**
```sh
cd api-node
npm test
```

---

## Deployment Guide (EC2)

### 1. Provision EC2

- Ubuntu 22.04 LTS, t2.small or larger
- Security group: port 22 (your IP only), port 80 (0.0.0.0/0)

### 2. Install Docker

```sh
sudo apt-get update
sudo apt-get install -y ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] \
  https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo usermod -aG docker $USER
```

Log out and back in for the group change to take effect.

### 3. Install PostgreSQL 15

```sh
sudo apt-get install -y postgresql-15
sudo systemctl enable postgresql
sudo systemctl start postgresql
```

Configure to listen on localhost only (default on Ubuntu — verify in
`/etc/postgresql/15/main/postgresql.conf` that `listen_addresses = 'localhost'`).

### 4. Create database and user

```sh
sudo -u postgres psql -c "CREATE USER matrixops WITH PASSWORD 'yourpassword';"
sudo -u postgres psql -c "CREATE DATABASE matrixops OWNER matrixops;"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE matrixops TO matrixops;"
```

### 5. Install Python bcrypt

```sh
sudo apt-get install -y python3-pip
pip3 install bcrypt
```

### 6. Clone and configure

```sh
git clone <repo-url>
cd IS-matrix-ops
```

Create `.env` at the project root:
```
DATABASE_URL=postgres://matrixops:yourpassword@host.docker.internal:5432/matrixops
JWT_SECRET=<strong-random-string>
PORT=8080
JWT_EXPIRY_MINUTES=10
JWT_REFRESH_EXPIRY_DAYS=7
ALLOWED_ORIGIN=http://<EC2_PUBLIC_IP>
VITE_API_BASE_URL=http://<EC2_PUBLIC_IP>
```

### 7. Seed and deploy

```sh
bash db/seed.sh "your-evaluator-password"
docker compose up -d
```

### 8. Verify

- `http://<EC2_PUBLIC_IP>/` — frontend loads
- `POST http://<EC2_PUBLIC_IP>/api/v1/auth/login` — returns tokens
- `POST http://<EC2_PUBLIC_IP>/api/v2/stats` — returns stats

**Note:** HTTP only. A raw EC2 IP cannot obtain a TLS certificate from a public CA,
so HTTPS is not implemented. This is a documented constraint for this demo deployment.

Internal ports (8080, 3000, 5432) are not reachable from outside the instance.
