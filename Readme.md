```
go_project_structure_domain_driven_design/
├── main.go                    # Entry point
├── app/
│   └── application.go         # Server bootstrap (chi router, graceful shutdown)
├── config/
│   ├── env/env.go             # .env loader + typed getters
│   └── db/db.go               # GORM connection setup
├── internal/                  # All domain-specific code
│   ├── module/module.go       # Module interface contract
│   ├── router/router.go       # Global module registry (register domains here)
│   ├── dto/user_dto.go        # Request/Response data transfer objects
│   ├── middlewares/
│   │   └── user_middleware.go # Domain-specific request validators
│   ├── db/
│   │   ├── models/user_model.go        # GORM model
│   │   └── repositories/user_repository.go  # Data access layer
│   └── user/                  # The "user" bounded context (domain)
│       ├── module.go          # Wires repo+svc+controller+router; registers tasks
│       ├── user_router.go     # Route declarations for user domain
│       ├── user_handler.go    # HTTP handlers (UserController)
│       ├── user_service.go    # Business logic interface + impl
│       └── user_csv_service.go # CSV-related service methods (split for clarity)
├── common_pkg/                # Shared, domain-agnostic utilities
│   ├── csv/csv.go             # Generic CSV export + streaming upload with worker pool
│   ├── json/json.go           # JSON read/write helpers
│   ├── middlewares/
│   │   ├── request_logger.go  # Request ID injection + logging
│   │   ├── jwt_auth_middleware.go  # JWT Bearer token verification
│   │   └── rate_limit.go      # Token-bucket rate limiter (5 req/sec)
│   ├── proxy/proxy.go         # Reverse proxy helper (strips path prefix)
│   ├── scheduler/
│   │   ├── task_assignment.go # Public entry: TaskAssignment(ctx, tasks)
│   │   └── ticker.go          # Task runner with skip-if-busy + panic recovery
│   └── worker_pool/
│       └── worker_pool.go     # Generic goroutine pool (Thread Pool pattern)
├── utils/
│   └── authentication/
│       └── authentication.go  # bcrypt hash + verify helpers
├── db/
│   └── migrations/            # (placeholder for SQL migrations)
└── exports/                   # CSV files written here at runtime
```



|Pattern|Where|
|---|---|
|**Interface-based DI**|Repository + Service are interfaces; concrete structs injected at module init|
|**Module self-registration**|Each domain exposes a `Module` interface; app iterates `router.Modules`|
|**Context propagation**|Middlewares inject validated payloads + auth claims via `context.WithValue`|
|**Raw SQL over ORM magic**|Repository uses explicit SQL with `db.Exec`/`db.Raw` for clarity|
|**Soft delete**|`gorm.Model` provides `DeletedAt`; queries always filter `deleted_at IS NULL`|
|**Partial updates with pointer types**|`*string` fields in `UpdateUserRequest` — nil means "don't update that field"|
|**Generic utilities**|Worker pool and CSV export use Go generics for type safety|
|**Graceful shutdown**|`signal.NotifyContext` cancels ctx on SIGINT/SIGTERM; server shuts down with 10s timeout|

Worker Pool (common_pkg/worker_pool/worker_pool.go)
A generic Thread Pool using Go generics ([I any, O any]):

Producer → JobChan → [Worker 1] → ResultChan → Consumer
                  → [Worker 2] →
                  → [Worker N] →
NewWorkerPool(workerCount, bufferSize, processFunc) — full version with typed results
NewPool(workerCount, processFunc) — simplified version for fire-and-forget (no output)
Used by the CSV upload pipeline to process batches in parallel


Scheduler (common_pkg/scheduler/)
Runs periodic background tasks. Each Task has a Name, Interval, and Fn func(ctx context.Context) error.

The Ticker.run() method is robust:

Uses time.NewTicker per task
Skip-if-busy: if the previous run hasn't finished, the next tick is skipped (no pile-up)
Panic recovery: a panicking task is logged and reset, not crashing the server
Context-aware: stops cleanly when ctx is cancelled (on SIGINT/SIGTERM)
StartAll() launches each task in its own goroutine and waits with a sync.WaitGroup.


CSV Package (common_pkg/csv/csv.go)
Export (ExportToCSV):

Accepts any slice of structs
Uses reflect to read csv:"fieldname" struct tags as column headers
Writes timestamped file to exports/ directory
Upload + Stream (UploadAndStreamCSV):

Parses multipart form (max 10MB)
Reads CSV rows into batches of configurable size
Submits each batch to a WorkerPool for parallel processing
Collects and returns the first error


Proxy (common_pkg/proxy/proxy.go)
ProxyToService(targetURL, pathPrefix) creates an httputil.ReverseProxy that:

Strips the local pathPrefix from the request path before forwarding
Sets Host header to the target's host
Forwards any userId from context as X-User-Id header (inter-service auth pattern)
Example: GET /fake-store/products → GET https://fakestoreapi.com/products



Full Request Lifecycle Example: POST /signup
Client → POST /signup (JSON body)
  │
  ├─ RequestLoggerMiddleware  → logs, injects requestId into ctx
  ├─ UserRegisterRequestValidator → decodes JSON into RegisterUserRequest, injects into ctx
  │
  └─ UserController.RegisterUser
       ├─ reads registration_payload from ctx
       ├─ builds models.User{Name, Email, Password}
       ├─ UserServiceImpl.CreateUser
       │    ├─ HashPassword(plaintext) → bcrypt hash
       │    └─ UserRepositoryImpl.Create
       │         ├─ INSERT INTO users (name, email, password) VALUES (?, ?, ?)
       │         ├─ pg error code handling (23505 = duplicate email)
       │         └─ returns rows affected
       └─ WriteJsonSuccessResponse(200, message, {name, email})