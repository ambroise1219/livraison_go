# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

Project overview
- Language/runtime: Go (module: github.com/ambroise1219/livraison_go)
- Web framework: gin-gonic
- ORM/DB: Prisma Client Go targeting PostgreSQL (schema in prisma/schema.prisma, migrations in prisma/migrations)
- Realtime: Redis-backed SSE/WebSocket utilities
- Docs: Swagger annotations + docs/swagger.yaml (served at /swagger)

Common commands (Windows PowerShell)
- Dependencies
  - Ensure a PostgreSQL instance and DATABASE_URL are available.
  - Generate Prisma client code (required before first build if prisma/db client code is not present):
    ```powershell path=null start=null
    # Option A: using Prisma CLI (requires Node.js + npx)
    npx prisma generate
    
    # Optionally, apply migrations for dev
    npx prisma migrate dev --name init
    ```
    ```powershell path=null start=null
    # Option B: using prisma-client-go generator via Go (generator only)
    go run github.com/steebchen/prisma-client-go generate
    ```

- Build and run
  ```powershell path=null start=null
  # Set env as needed (example values)
  $env:SERVER_HOST = "localhost"
  $env:SERVER_PORT = "8080"
  $env:DATABASE_URL = "{{DATABASE_URL}}"  # e.g., postgresql://user:pass@host:5432/db?sslmode=disable
  $env:JWT_SECRET = "{{JWT_SECRET}}"
  
  # Run in dev
  go run .
  
  # Or build a binary
  go build -o ilex-backend
  ./ilex-backend
  ```

- Tests
  ```powershell path=null start=null
  # Run all tests
  go test ./...
  
  # Verbose / with coverage
  go test -v ./...
  go test -cover ./...
  
  # Run a single package
  go test ./services/support -v
  
  # Run a single test by name (regex)
  go test ./services/support -run TestTicketStatusUpdates -v
  ```

- Formatting and basic checks
  ```powershell path=null start=null
  go fmt ./...
  go vet ./...
  ```

- Redis utility
  ```powershell path=null start=null
  # Small utility under cmd/
  go run ./cmd/redis_check
  ```

- Swagger UI (served by the app)
  - After starting the server, open: http://localhost:8080/swagger/index.html
  - BasePath: /api/v1

Important environment notes
- The application reads configuration from environment variables (with .env support via godotenv). Key variables include:
  - DATABASE_URL, SERVER_HOST, SERVER_PORT, ENVIRONMENT, JWT_SECRET, Redis (REDIS_HOST/PORT/PASSWORD/DB), Cloudinary (CLOUDINARY_CLOUD_NAME/API_KEY/API_SECRET/FOLDER), and OTP/Email settings.
- main.go defines development defaults inside init() if vars are unset; set your own env to override. Migrations should be applied to your target database before running tests that hit the DB.
- Realtime features try to connect to Redis at startup (non-fatal in development, required in production per config.GetRedisClient logic).

High-level architecture
- Entry point: main.go
  - Loads config (config.GetConfig), sets Gin mode based on ENVIRONMENT.
  - Initializes Prisma connection (database.InitPrisma) and exposes a global Prisma client for services via db.InitializePrisma.
  - Initializes all handlers (handlers.InitHandlers) which in turn set up validators and construct service singletons (auth, delivery, promo, realtime, support, vehicle, storage uploader for Cloudinary).
  - Configures routes (routes.SetupRoutes) and mounts Swagger at /swagger/*any.
  - Binds and runs the Gin server on SERVER_HOST:SERVER_PORT.

- Configuration: config/config.go
  - Loads structured Config from env (with defaults) and exposes GetConfig().
  - Provides a lazily initialized Redis client (GetRedisClient) used by realtime features. Development continues if Redis is unavailable; production requires it.

- Data layer (PostgreSQL via Prisma Client Go)
  - prisma/schema.prisma defines the DB schema with migrations in prisma/migrations.
  - database/prisma.go manages lifecycle of the Prisma client (connect, disconnect, simple stats helpers).
  - db/database.go holds a global PrismaDB pointer referencing the generated client package (github.com/ambroise1219/livraison_go/prisma/db) used by services. Includes small compatibility helpers (table counts/stats) now routed through Prisma.
  - Note: The generated prisma/db Go code is not checked in; generate it before building.

- HTTP layer
  - Router and groups: routes/routes.go
    - Global middlewares: recovery, logging, CORS, security headers, basic rate limit wrapper.
    - Health endpoints: / and /health, plus /db for simple DB stats via Prisma.
    - Versioned API under /api/v1 with groups:
      - Public: /auth (OTP send/verify, refresh), /promo (validate), /delivery/price/calculate (optional auth).
      - Protected (AuthMiddleware):
        - /auth (logout, profile, profile picture)
        - /users (profiles, deliveries, vehicles)
        - /delivery (CRUD-like operations, driver/client subgroups)
        - /promo (use, history, referral)
        - /admin (users, deliveries, drivers, promotions, vehicles, stats)
        - Realtime endpoints: /sse, /ws, /realtime
        - Support system: /support (tickets/messages/stats), /internal (staff chat), /admin/support (direct contact)
  - Middlewares: routes/middleware.go
    - AuthMiddleware currently contains a TODO for actual JWT validation; it sets a temporary user context. Role-based helpers (RequireClient/Driver/Admin/etc.) enforce access across routes.
  - Handlers: handlers/handlers.go (+ handlers/realtime.go)
    - Wire HTTP to services; perform input validation with go-playground/validator.
    - Emit realtime delivery updates via services.RealtimeService; optional Cloudinary-backed profile image upload if configured.

- Domain/services
  - services/auth: OTP, user, JWT generation/validation; JWT claims wrap user identity and role.
  - services/delivery: creation/update/query split by delivery type (simple/express/etc.).
  - services/promo: promo code operations and validation.
  - services/support: ticketing system (creation, access control, status transitions, messages). Has integration tests that talk to the real DB; ensure DATABASE_URL and migrations are set before running.
  - services/realtime (+ handlers/realtime.go): Redis-backed SSE/WebSocket for delivery tracking, driver location, notifications, ETA, and realtime stats.
  - services/storage/cloudinary: uploader abstraction used for profile images; graceful degradation when not configured.
  - services/validation: phone number normalization/validation.
  - services/vehicle: CRUD and validation around vehicles.
  - models: core enums (roles, statuses), request/response types, and helpers; used across services/handlers.

Repo docs and artifacts
- README.md contains general setup and endpoint listings. Prefer the DB info in code (Prisma + PostgreSQL) over outdated mentions of SurrealDB/SQLite.
- Swagger: docs/swagger.yaml and generated docs/docs.go. Served at /swagger when the server is running.
- Postman: postman_collections/README.md for client testing context.

Development tips specific to this repo
- Prisma must be generated before building if prisma/db is missing. If build errors reference github.com/ambroise1219/livraison_go/prisma/db, run a generate step.
- JWT middleware is stubbed; protected routes may rely on temporary context until JWT validation is fully wired. Tests in services/support hit the real DB and assume a working Prisma connection + migrated schema.
- If Redis is unavailable locally, realtime routes will degrade gracefully in development, but plan for Redis in any realistic environment.
