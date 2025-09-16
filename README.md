# Go OTP Auth Backend

This is a Golang backend service implementing OTP-based login and registration with basic user management features. The project uses PostgreSQL for persistent storage and Redis for OTP and rate limiting. Additionally, two monitoring utilities are included: Adminer for Postgres and RedisInsight for Redis.

---

## Features

- OTP-based login & registration
- Rate limiting (max 3 OTP requests per phone per 10 minutes)
- JWT-based authentication (token will expire after 1 hour)
- User management endpoints with pagination and search
- Swagger/OpenAPI documentation
- Dockerized with PostgreSQL, Redis, and monitoring tools (Adminer & RedisInsight)

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Run Locally](#run-locally)
- [Run with Docker](#run-with-docker)
- [API Endpoints](#api-endpoints)
- [Database Choice](#database-choice)
- [Swagger Documentation](#swagger-documentation)
- [Monitoring Utilities](#monitoring-utilities)

---

## Prerequisites

- Go >= 1.21
- Docker & Docker Compose (optional for containerized setup)

---

## Run Locally

1. Clone the repository:

```bash
git clone https://github.com/Alizadeh275/go-otp-auth.git
cd go-otp-auth
```

2. Install dependencies:

```bash
go mod tidy
```

3. Create a `.env` file (or set environment variables):

```
PORT=8080
DATABASE_URL=postgres://otpuser:otppass@db:5432/otpdb?sslmode=disable
REDIS_ADDR=redis:6379
JWT_SECRET=replace-me-with-strong-secret
OTP_TTL_SECONDS=120
RATE_LIMIT_MAX=3
RATE_LIMIT_WINDOW_SECONDS=600
```

4. Start PostgreSQL and Redis manually (if not using Docker).

5. Run the service:

```bash
go run ./cmd/server/main.go
```

6. API available at `http://localhost:8080`

---

## Run with Docker

1. Build and start services:

```bash
docker-compose up --build
```

2. Check logs:

```bash
docker-compose logs -f app
```

3. Swagger UI available at `http://localhost:8080/docs/index.html`

4. Monitoring utilities:
   - **Adminer**: `http://localhost:8088` (Postgres UI)
   - **RedisInsight**: `http://localhost:5540` (Redis monitoring UI)

---

## API Endpoints

### Request OTP

```
POST /otp/request
Body:
{
  "phone": "+1234567890"
}
```

Response:

```json
{
  "status": "otp_generated"
}
```

### Verify OTP

```
POST /otp/verify
Body:
{
  "phone": "+1234567890",
  "otp": "123456"
}
```

Response:

```json
{
  "token": "jwt_token_here",
  "user": {
    "id": 1,
    "phone": "+1234567890",
    "registered_at": "2025-09-16T12:00:00Z"
  }
}
```

### Get Current User

```
GET /users/me
Header: Authorization: Bearer <token>
```

Response:

```json
{
  "id": 1,
  "phone": "+1234567890",
  "registered_at": "2025-09-16T12:00:00Z"
}
```

### List Users

```
GET /users?page=1&size=10&search=123
```

Response:

```json
{
  "total": 1,
  "page": 1,
  "size": 10,
  "data": [
    {
      "id": 1,
      "phone": "+1234567890",
      "registered_at": "2025-09-16T12:00:00Z"
    }
  ]
}
```

---

## Database Choice

- **PostgreSQL**: reliable, ACID-compliant, and well-supported in Go via `sqlx`.
- **Redis**: used for temporary OTP storage and rate limiting due to fast in-memory operations and TTL support.
- **Docker Compose** allows quick setup of both services locally.

---

## Swagger Documentation

- Run `swag init -g ./cmd/server/main.go` to generate `docs/`
- Swagger UI available at: `http://localhost:8080/docs`/
- All endpoints documented with request/response types, parameters, and security.

---

## Monitoring Utilities

- **Adminer**: Lightweight web UI to inspect PostgreSQL database.
- **RedisInsight**: Web UI for monitoring Redis keys, TTLs, and rate limiting/OTP operations.

---

## Notes

- OTP is printed in the console; no SMS integration.
- OTP expires in 2 minutes.
- Max 3 OTP requests per phone number per 10 minutes.
- JWT authentication is required for `/users/me` endpoint.
- Pagination and search available for `/users`.

---

**Author:** Sajjad Alizadeh Fard  
**Date:** 2025-09-16

