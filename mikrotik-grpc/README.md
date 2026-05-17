# Production-Grade ISP User Management & RADIUS System

This project provides a complete, production-grade solution for Internet Service Providers (ISPs) to manage their user base and authenticate them using MikroTik routers. It consists of two separate Go applications that work together:

1.  **API Server (HTTP):** A backend API for administrators to create, update, and manage users. This is the single point of entry for all user management tasks.
2.  **RADIUS Server (UDP):** A high-performance AAA server that handles live authentication and accounting requests from MikroTik routers.

**Core Architecture Principle:** The MikroTik routers **do not** store any user credentials. All users are managed centrally in a PostgreSQL database, and authentication is handled exclusively via the RADIUS protocol. This creates a scalable, secure, and manageable system.

---

## System Architecture

The architecture is designed for separation of concerns and high availability.

1.  **Management Plane (API Server):**
    - An administrator uses a frontend or a tool like Postman to send an HTTP request (e.g., `POST /api/clients`) to the **Go API Server**.
    - The API server validates the request, hashes the user's password with `bcrypt`, and inserts the new user record into the **PostgreSQL** database.

2.  **Authentication Plane (RADIUS Server):**
    - An end-user tries to connect via PPPoE or Hotspot.
    - The **MikroTik Router** sends a RADIUS `Access-Request` to the **Go RADIUS Server**.
    - The RADIUS server queries **Redis** (for cache) or **PostgreSQL** (for the source of truth) to validate the user's credentials, status, and package expiry.
    - If valid, it returns an `Access-Accept` packet containing authorization attributes (like `Mikrotik-Rate-Limit`).
    - The MikroTik router establishes the user's session with the specified speed limit.
    - Throughout the session, the router sends `Accounting-Request` packets to the RADIUS server to log data usage and session duration.

---

## Getting Started

### Prerequisites

- Go 1.21+
- Docker and Docker Compose (recommended for easy setup of PostgreSQL and Redis)
- A MikroTik router or CHR instance

### 1. Setup Database & Cache

You can use Docker Compose to easily spin up PostgreSQL and Redis. Create a `docker-compose.yml` file if you don't have one.

### 2. Database Migration

Once your PostgreSQL container is running, connect to it and run the schema script to create the necessary tables and indexes.

```bash
# Example using psql client
psql "postgres://youruser:yourpassword@localhost:5432/radiusdb" -f migrations/001_initial_schema.sql
```

### 3. Configuration

The system is configured via a `.env` file in the project root.

```dotenv
# .env file
# -------------------------------------

# PostgreSQL Connection URL
POSTGRES_URL=postgres://youruser:yourpassword@localhost:5432/radiusdb

# Redis Connection
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# RADIUS Server Configuration
# This MUST match the secret configured on your MikroTik router
RADIUS_SECRET=my_very_secret_key
AUTH_PORT=1812
ACCT_PORT=1813
WORKER_POOL_SIZE=20

# API Server Configuration
API_PORT=:8080
```

### 4. Running the Applications

You must run both applications simultaneously in separate terminal sessions.

**Terminal 1: Run the RADIUS Server**

```bash
go run ./cmd/radius-server/main.go
```

**Terminal 2: Run the API Server**

```bash
go run ./cmd/api-server/main.go
```

---

## API Usage: Creating a User

To create a new user, send a `POST` request to the API server.

**Endpoint:** `POST /api/clients`

**Headers:**
- `Content-Type: application/json`

**Body:**

```json
{
    "username": "new_subscriber",
    "password": "a_strong_password_123",
    "packageName": "Standard 10M",
    "expiryDays": 30
}
```

**Success Response (`201 Created`):**

```json
{
    "id": "c3a4b5d6-e7f8-1234-5678-90abcdef1234",
    "username": "new_subscriber",
    "status": "active",
    "expiryDate": "2024-10-28T10:00:00Z",
    "message": "User created successfully"
}
```

---

## MikroTik Router Configuration

Configure your MikroTik router to use your new external RADIUS server.

1.  **Add the RADIUS Server:**
    Replace `X.X.X.X` with the IP address of the machine running your Go RADIUS server. The `secret` must exactly match the `RADIUS_SECRET` in your `.env` file.

    ```routeros
    /radius add service=ppp,hotspot address=X.X.X.X secret=my_very_secret_key authentication-port=1812 accounting-port=1813 timeout=1s
    ```

2.  **Enable RADIUS for Authentication & Accounting:**
    This command forces the PPP (PPPoE) service to use RADIUS.

    ```routeros
    /ppp aaa set use-radius=yes accounting=yes
    ```

3.  **Configure PPP Profile (Optional but Recommended):**
    Ensure your PPP profiles do not have a local address pool configured if you want the RADIUS server to assign IPs or pools.

4.  **Verify:**
    Check the RADIUS status and logs on the MikroTik router.
    ```routeros
    /radius print detail
    /log print where topics~"radius"
    ```
    You should see packets being sent and responses received.