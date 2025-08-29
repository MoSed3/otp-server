# OTP Server

A REST API server for OTP-based authentication, built with Go.

## Table of Contents

- [Features](#features)
- [Project Design and Architecture](#project-design-and-architecture)
- [Why Redis?](#why-redis)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Environment Variables](#environment-variables)
  - [Running the Application](#running-the-application)
- [API Documentation](#api-documentation)
- [Database Migrations](#database-migrations)

## Features

- User authentication via OTP (One-Time Password)
- RESTful API endpoints
- Rate limiting for OTP requests
- JWT-based authentication for user sessions
- Profile management for authenticated users
- Database auto-migration option

## Project Design and Architecture

This project follows a layered architecture to ensure separation of concerns, maintainability, and scalability.

-   **`main.go`**: The entry point of the application, responsible for initializing configurations, database, Redis, and starting the HTTP server.
-   **`config/`**: Handles application configuration, loading environment variables and providing access to them.
-   **`db/`**: Contains database models (GORM structs) and functions for interacting with the PostgreSQL database. This layer abstracts database operations from the business logic.
-   **`redis/`**: Manages interactions with the Redis server, primarily for OTP session management and rate limiting.
-   **`controller/`**: Implements the business logic for handling API requests. It orchestrates interactions between the `db` and `redis` layers.
-   **`router/`**: Defines the API routes and their corresponding handlers. It uses `go-chi/chi` for routing and integrates middleware.
-   **`middleware/`**: Contains custom middleware for transaction management, rate limiting, and JWT token validation.
-   **`settings/`**: Manages application-wide settings stored in the database.

**Design Patterns Used:**

-   **Repository Pattern**: The `db` package functions act as a repository, abstracting data access logic.
-   **Dependency Injection**: Components like `db` and `redis` clients are initialized once and passed to other layers, promoting loose coupling.
-   **Middleware Pattern**: Used extensively in the `router` and `middleware` packages for cross-cutting concerns like logging, authentication, and transaction management.

## Why Redis?

Redis is utilized in this project for several critical reasons, primarily to enhance performance and reduce the load on the PostgreSQL database:

1.  **OTP Session Management**: OTP codes and associated user session data are stored in Redis with a short Time-To-Live (TTL). This allows for quick retrieval and validation of OTPs without hitting the database for every request, which is crucial for a high-traffic authentication flow.
2.  **Rate Limiting**: Redis is an excellent choice for implementing efficient rate limiting. It allows the application to track the number of OTP requests per user or IP address within a given time frame, preventing abuse and protecting the backend services from being overwhelmed.
3.  **Reduced Database Load**: By offloading frequently accessed, transient data (like OTPs and rate limit counters) to Redis, the PostgreSQL database is spared from numerous read/write operations that would otherwise occur. This keeps the database free to handle more persistent and complex data operations, improving overall system responsiveness and scalability.

## Getting Started

### Prerequisites

-   Go (version 1.25 or higher)
-   PostgreSQL database
-   Redis server
-   Docker and Docker Compose (optional, for local development setup)

### Environment Variables

The application uses environment variables for configuration. A `.env.example` file is provided as a template. Copy it to `.env` and fill in your specific values.

```ini
# PostgreSQL Configuration
POSTGRES_DB=otp
POSTGRES_USER=user
POSTGRES_PASSWORD=1234

# Database Connection URL
DATABASE_URL=postgresql://user:1234@127.0.0.1:5432/otp?sslmode=disable
DB_POOL_SIZE=10
DB_MAX_OVERFLOW=30
ECHO_SQL_QUERIES=false
AUTOGENERATE_DB=false

# Redis Configuration
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_DB=1
REDIS_PASSWORD=

# Server Configuration
WEBAPP_HOST=127.0.0.1
WEBAPP_PORT=8080

# SSL/TLS Configuration (optional)
CERT_FILE=
KEY_FILE=

# PgAdmin Configuration (optional)
PGADMIN_DEFAULT_EMAIL=pgadmin4@pgadmin.org
PGADMIN_DEFAULT_PASSWORD=admin
```

-   **`DATABASE_URL`**: Connection string for your PostgreSQL database.
-   **`DB_POOL_SIZE`**: Maximum number of open connections to the database.
-   **`DB_MAX_OVERFLOW`**: Maximum number of connections that can exceed `DB_POOL_SIZE`.
-   **`ECHO_SQL_QUERIES`**: Set to `true` to log SQL queries to the console (useful for debugging).
-   **`AUTOGENERATE_DB`**: Set to `true` to automatically run GORM migrations and seed default settings on application startup. **Use with caution in production environments.**
-   **`REDIS_HOST`**: Hostname or IP address of your Redis server.
-   **`REDIS_PORT`**: Port of your Redis server.
-   **`REDIS_DB`**: Redis database number to use.
-   **`REDIS_PASSWORD`**: Password for your Redis server (leave empty if no password).
-   **`WEBAPP_HOST`**: Host IP address for the web application.
-   **`WEBAPP_PORT`**: Port for the web application.
-   **`CERT_FILE`**, **`KEY_FILE`**: Paths to SSL/TLS certificate and key files for HTTPS (optional).

### Running the Application with Docker

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/MoSed3/otp-server.git
    cd otp-server
    ```

2.  **Set up environment variables:**
    Copy `.env.example` to `.env` and configure it.
    ```bash
    cp .env.example .env
    # Edit .env with your database and Redis credentials
    ```

3.  **Run the application:**
    ```bash
    docker compose up
    ```

### Running the Application without Docker (Manual Setup)

To run the application directly on your system without Docker, follow these steps:

1.  **Prerequisites:**
    *   Go (version 1.25 or higher)
    *   PostgreSQL database server running locally or accessible
    *   Redis server running locally or accessible
    *   `migrate` CLI tool for database migrations: `go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

2.  **Clone the repository:**
    ```bash
    git clone https://github.com/MoSed3/otp-server.git
    cd otp-server
    ```

3.  **Set up environment variables:**
    Copy `.env.example` to `.env` and configure it with your PostgreSQL and Redis connection details. Ensure `DATABASE_URL`, `REDIS_HOST`, `REDIS_PORT`, etc., are correctly set to point to your local services.
    ```bash
    cp .env.example .env
    # Edit .env with your database and Redis credentials
    ```

4.  **Install Go dependencies:**
    ```bash
    go mod download
    go mod tidy
    ```

5.  **Build the server and CLI applications:**
    ```bash
    make build # Builds the main server application
    make build-cli # Builds the admin CLI application
    ```
    This will create `server-<os>-<arch>` and `admin-cli-<os>-<arch>` executables in your project root.

6.  **Run database migrations:**
    Before starting the server, apply the database migrations. Replace the `DATABASE_URL` with the actual connection string from your `.env` file.
    ```bash
    migrate -path migrations -database "postgresql://user:1234@127.0.0.1:5432/otp?sslmode=disable" up
    ```
    (Adjust the `DATABASE_URL` as per your `.env` configuration.)

7.  **Run the server:**
    ```bash
    ./server-<os>-<arch> # Replace with your actual OS and architecture, e.g., ./server-linux-amd64
    # Alternatively, you can use:
    # go run ./cmd/server
    # make run
    ```

### Accessing the CLI

The project includes an `admin` CLI tool for administrative tasks.

#### With Manual Setup

If you are running the application without Docker, you can execute the CLI directly:

```bash
./admin-cli-<os>-<arch> # Replace with your actual OS and architecture, e.g., ./admin-cli-linux-amd64
# Alternatively, you can use:
# go run ./cmd/admin <command> [args...]
# make run-cli <command> [args...]
```
Example:
```bash
./admin-cli-linux-amd64 list
```

#### With Docker Setup

If you are running the application using `docker compose`, you can access the CLI by executing commands inside the `otp_server` container:

1.  **Find the container ID or name:**
    ```bash
    docker ps
    ```
    Look for the container running the `otp_server` service (e.g., `otp-server-otp_server-1`).

2.  **Execute CLI commands:**
    ```bash
    docker exec -it <container_id_or_name> ./admin <command> [args...]
    ```
    Example:
    ```bash
    docker exec -it otp-server-otp_server-1 ./admin list list
    ```

## API Documentation

The API documentation is generated using `swaggo` and is available via Swagger UI.
Once the server is running, you can access the Swagger UI at:
`http://localhost:8080/swagger/index.html` (adjust host and port as per your configuration).

To regenerate the Swagger documentation after making changes to API annotations:
```bash
swag init
```

## Database Migrations

This project uses SQL-based migrations. The migration files are located in the `migrations/` directory.

-   `000001_initial_schema.up.sql`: Contains SQL statements to apply the initial database schema.
-   `000001_initial_schema.down.sql`: Contains SQL statements to revert the initial database schema.

You can use a tool like `migrate` (https://github.com/golang-migrate/migrate) to manage these migrations.
Example commands (assuming `migrate` CLI is installed):

To apply migrations:
```bash
migrate -path migrations -database "postgresql://user:1234@127.0.0.1:5432/otp?sslmode=disable" up
```

To revert migrations:
```bash
migrate -path migrations -database "postgresql://user:1234@127.0.0.1:5432/otp?sslmode=disable" down
