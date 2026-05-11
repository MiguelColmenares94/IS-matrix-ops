# Pre-Deploy Checklist

Everything required before running `docker compose up --build -d` for the first time.
Applies to both local development and EC2 deployment.

---

## 1. Prerequisites

### 1.1 Install Docker + Compose plugin

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
# Log out and back in for the group change to take effect
```

### 1.2 Install PostgreSQL 15

```sh
sudo apt-get install -y postgresql-15
sudo systemctl enable postgresql
sudo systemctl start postgresql
```

### 1.3 Install Python bcrypt (required by db/seed.sh)

```sh
sudo apt-get install -y python3-bcrypt
# Verify:
python3 -c "import bcrypt; print(bcrypt.__version__)"
```

### 1.4 Verify all prerequisites

```sh
docker --version          # expect 24+
docker compose version    # expect v2+
psql --version            # expect 15+
python3 --version         # expect 3.x
python3 -c "import bcrypt; print('bcrypt ok')"
```

---

## 2. PostgreSQL Configuration

### 2.1 Create database and user

```sh
sudo -u postgres psql -c "CREATE USER matrixops WITH PASSWORD 'your-db-password';"
sudo -u postgres psql -c "CREATE DATABASE matrixops OWNER matrixops;"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE matrixops TO matrixops;"
```

### 2.2 Allow Docker containers to connect

Docker containers connect via `host.docker.internal` (resolves to the host IP on the
Docker bridge network, typically `172.x.x.x`). PostgreSQL must accept those connections.

```sh
# Allow connections from Docker bridge network
sudo bash -c 'echo "host    matrixops       matrixops       172.0.0.0/8             md5" >> /etc/postgresql/15/main/pg_hba.conf'
sudo systemctl reload postgresql
```

### 2.3 Ensure PostgreSQL listens on all interfaces

```sh
sudo grep "listen_addresses" /etc/postgresql/15/main/postgresql.conf
# Must show: listen_addresses = '*'
# If it shows 'localhost', change it:
sudo sed -i "s/#listen_addresses = 'localhost'/listen_addresses = '*'/" /etc/postgresql/15/main/postgresql.conf
sudo systemctl restart postgresql
```

---

## 3. Generate Lockfiles

These files are required by the Dockerfiles but are not committed to the repo.
Run once after cloning, or whenever dependencies change.

### 3.1 Go — generate go.sum

```sh
docker run --rm -v "$(pwd)/api-go:/app" -w /app golang:1.22-alpine go mod tidy
```

### 3.2 Node.js — generate package-lock.json for api-node and frontend

```sh
docker run --rm -v "$(pwd)/api-node:/app" -w /app node:20-alpine npm install
docker run --rm -v "$(pwd)/frontend:/app" -w /app node:20-alpine npm install
```

---

## 4. Environment File

Create the root `.env` file (read by Docker Compose for all services):

```sh
cat > .env <<EOF
DATABASE_URL=postgres://matrixops:your-db-password@host.docker.internal:5432/matrixops
JWT_SECRET=your-strong-random-secret
JWT_EXPIRY_MINUTES=10
JWT_REFRESH_EXPIRY_DAYS=7
ALLOWED_ORIGIN=http://localhost
VITE_API_BASE_URL=http://localhost
EOF
```

> On EC2, replace `http://localhost` with `http://<EC2_PUBLIC_IP>`.

---

## 5. Seed the Evaluator User

Writes the bcrypt hash into `db/migrations/000003_seed.up.sql`.
Must be run before the first `docker compose up`.

```sh
bash db/seed.sh "your-evaluator-password"
```

> Only run once. If migrations have already been applied, run `docker compose run migrate down`
> first, then re-run seed.sh and `docker compose up` again.

---

## 6. Deploy

```sh
docker compose up --build -d
```

Expected startup order:
1. `migrate` — runs migrations, exits 0
2. `api-go`, `api-node`, `frontend` — start after migrate
3. `nginx` — starts after all three are up

### Verify

```sh
docker compose ps
# migrate should show Exited (0), all others Up

curl -s -o /dev/null -w "%{http_code}" http://localhost/
# expect: 200

curl -s -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"evaluator@example.com","password":"your-evaluator-password"}'
# expect: {"access_token":"...","refresh_token":"...","expires_at":"..."}
```

---

## Summary

| Step | Command |
|------|---------|
| Install Docker | `apt-get install docker-ce docker-compose-plugin` |
| Install PostgreSQL 15 | `apt-get install postgresql-15` |
| Install Python bcrypt | `apt-get install python3-bcrypt` |
| Create DB + user | `sudo -u postgres psql -c "CREATE USER ..."` |
| Allow Docker → PG | append to `pg_hba.conf`, reload |
| Generate go.sum | `docker run ... golang:1.22-alpine go mod tidy` |
| Generate package-lock.json | `docker run ... node:20-alpine npm install` (×2) |
| Create .env | `cat > .env <<EOF ...` |
| Seed evaluator user | `bash db/seed.sh "password"` |
| Deploy | `docker compose up --build -d` |
