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

---

## Setup & Deployment

For both local development and EC2, use the automated setup script:

```sh
git clone <repo-url>
cd IS-matrix-ops
bash pre-deploy.sh
```

The script will prompt for passwords and configuration, then handle everything:
installing Docker and PostgreSQL, configuring the database, generating lockfiles,
creating the root `.env`, seeding the evaluator user, and running
`docker compose up --build -d`.

See `pre-deploy.md` for a step-by-step breakdown of what the script does, or if
you need to run any step manually.

> On EC2, enter `http://<EC2_PUBLIC_IP>` when prompted for the app origin URL.

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

## EC2 Notes

- Ubuntu 22.04 LTS, t2.small or larger
- Security group: port 22 (your IP only), port 80 (0.0.0.0/0)
- HTTP only — a raw EC2 IP cannot obtain a TLS certificate from a public CA.
  This is a documented constraint for this demo deployment.
- Internal ports (8080, 3000, 5432) are not reachable from outside the instance.
