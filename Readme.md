
# Getting started
1. Make your own copy
The cleanest path is "Use this template" → Create a new repository on GitHub. You get your own repo with no inherited commit history.
If you'd rather clone:

```bash
git clone https://github.com/hossain-asif/golang_project_template.git my-app
cd my-app
rm -rf .git
git init
git add .
git commit -m "initialize project"
git branch -M main
```

2. Rename the module
Don't skip this. The module is named go_project_structure and every internal import points at that name. Until you change it, your project is still carrying this one's identity.

```bash
go mod edit -module github.com/yourname/my-app
```

Then update the imports across every Go file:

```bash
# macOS
grep -rl 'go_project_structure' --include='*.go' . \
  | xargs sed -i '' 's|go_project_structure|github.com/yourname/my-app|g'

# Linux — same command, drop the empty quotes after -i
grep -rl 'go_project_structure' --include='*.go' . \
  | xargs sed -i 's|go_project_structure|github.com/yourname/my-app|g'
```

Finish with:

```bash
go mod tidy
```

If it compiles, the rename took.

3. Configure the environment

```bash
cp .env.example .env
```

Open .env and fill it in. Every value falls back to a sane default except JWT_SECRET — generate a real one:

```bash
openssl rand -base64 48
```

1. Start Postgres and Redis
Whatever you normally use is fine. If you want them up fast:

```bash
docker run -d --name pg -e POSTGRES_USER=user -e POSTGRES_PASSWORD=12345 \
  -e POSTGRES_DB=mydb -p 5432:5432 postgres:16

docker run -d --name redis -p 6379:6379 redis:7
```

Match the credentials to whatever you put in .env.

5. Run the app

```bash
go mod download
go run main.go
```

It listens on the PORT from your .env (default :8080).

---

# Run the migrations
Install goose once:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Then open the Makefile and update `DB_URL` at the top to match your database. It's hardcoded there — it does not read from .env .


```bash
CREATE
────────────────────────────────────
create_<table>_table

ALTER COLUMN
────────────────────────────────────
add_<column>_column_to_<table>_table
drop_<column>_column_from_<table>_table
rename_<old_column>_column_to_<new_column>_column_in_<table>_table
change_<table>_table_<column>_column_type

FOREIGN KEY
────────────────────────────────────
add_<source_table>_table_<column>_column_foreign_key_to_<target_table>_table
drop_<source_table>_table_<column>_column_foreign_key

UNIQUE
────────────────────────────────────
add_unique_<table>_table_<column>_column
drop_<table>_table_<column>_column_unique

CHECK
────────────────────────────────────
add_check_<table>_table_<business_rule>
drop_check_<table>_table_<business_rule>

NOT NULL
────────────────────────────────────
add_not_null_<table>_table_<column>_column
drop_not_null_<table>_table_<column>_column

DEFAULT
────────────────────────────────────
add_default_<table>_table_<column>_column
drop_default_<table>_table_<column>_column

INDEX
────────────────────────────────────
add_idx_<table>_table_<column>_column
drop_idx_<table>_table_<column>_column

TABLE
────────────────────────────────────
drop_<table>_table
rename_<old_table>_table_to_<new_table>_table
```

Example: 
```bash
20260212174735_create_<table_name>_table.sql # table name will be plural

20260212174735_add_<column_name>_to_<table_name>.sql 

20260212174735_add_<source_table>_table_<column_name>_column_foreign_key_to_<target_table>_table.sql

20260212175000_add_index_to_<table_name>_<column_name>.sql

```


```bash
gmake migrate-up
gmake migrate-down

```

Other migration commands:

```bash
make migrate-create name="create_orders_table"   # new migration
make migrate-status                              # what's applied
make migrate-down                                # roll back one
make migrate-reset                               # roll back everything
```

---

# Adding a new domain module

1. Add the repository at `internal/db/repositories/order/`
2. Add the model at `internal/db/models/order.go`
3. Add the DTOs at `internal/dto/order_dto.go`
4. Create internal/order/ with:

```
order_router.go      # route definitions
order_middleware.go  # middleware implementation
order_handler.go     # HTTP layer — decode, validate, respond
order_service.go     # business logic
module.go            # wires the above together
```

5. Implement Initialize in `internal/order/module.go`:

```go
type OrderModule struct {
    repository repositories.OrderRepository
    service    OrderService
}

func (om *OrderModule) Initialize(dep module.Dependency, r chi.Router) ([]scheduler.Task, error) {
    om.repository = repositories.NewOrderRepository(dep.DB)
    om.service = NewOrderService(om.repository)
    handler := NewOrderHandler(om.service)
    NewOrderRouter(handler).Register(r)

    return nil, nil   // or return background tasks — see below
}
```

6. Background tasks
A module can return recurring jobs from Initialize. The scheduler starts them with the app and shuts them down with it:

```go
return []scheduler.Task{
    {
        Name: "order.expire-stale",
        Interval: 30 * time.Minute,
        Fn: func(ctx context.Context) error {
            return om.service.ExpireStale(ctx)
        },
    },
}, nil
```

6. Register it in `internal/router/router.go`:

```go
var Modules = []module.Module{
    &user.UserModule{},
    &order.OrderModule{},   // ← the one line that turns it on
}
```

---

# Adding Infrastructure Resources
Worked example: adding RabbitMQ
Say you want a message broker available to every module.

1. declare the variables

Add to `.env.example`:

```bash
# rabbitmq 
RABBITMQ_URL=amqp://guest:guest@127.0.0.1:5672/
RABBITMQ_EXCHANGE=app.events
RABBITMQ_TIMEOUT_SECONDS=10
```

Then put real values in your own `.env`. Keep the two files in sync — `.env` is gitignored, so `.env.example` is the only thing the next developer sees.

2. write the connector

New file: `config/resources/rabbitmq.go`

The folder is called resources, but it holds anything connection-shaped.

```go
package resources

import (
	"fmt"
	"time"

	"go_project_structure/common_pkg/logger"
	env "go_project_structure/config/env"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Package-scoped logger — defined once, reused across all methods.
var rabbitLog = logger.Log.Scope("config", "resources", "rabbitmq")

// rabbitConfig holds everything needed to reach the broker.
type rabbitConfig struct {
	URL      string
	Exchange string
	Timeout  time.Duration
}

// loadRabbitConfig reads all RabbitMQ values from the environment.
func loadRabbitConfig() rabbitConfig {
	return rabbitConfig{
		URL:      env.GetString("RABBITMQ_URL", "amqp://guest:guest@127.0.0.1:5672/"),
		Exchange: env.GetString("RABBITMQ_EXCHANGE", "app.events"),
		Timeout:  time.Duration(env.GetInt("RABBITMQ_TIMEOUT_SECONDS", 10)) * time.Second,
	}
}

// RabbitMQ wraps the connection and channel so the caller can close both.
type RabbitMQ struct {
	conn     *amqp.Connection
	Channel  *amqp.Channel
	Exchange string
}

// SetupRabbitMQ dials the broker, opens a channel, declares the exchange,
// and returns a ready-to-use client.
//
// The caller is responsible for calling Close on shutdown.
func SetupRabbitMQ() (*RabbitMQ, error) {
	log := rabbitLog.Method("SetupRabbitMQ")

	cfg := loadRabbitConfig()

	conn, err := amqp.DialConfig(cfg.URL, amqp.Config{Dial: amqp.DefaultDial(cfg.Timeout)})
	if err != nil {
		log.WithError(err).Error("Failed to dial RabbitMQ.")
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		log.WithError(err).Error("Failed to open RabbitMQ channel.")
		return nil, fmt.Errorf("open channel: %w", err)
	}

	// Declaring the exchange here doubles as the connectivity check —
	// it round-trips to the broker and fails loudly if anything is wrong.
	if err := ch.ExchangeDeclare(cfg.Exchange, "topic", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		log.WithError(err).Error("Failed to declare exchange.")
		return nil, fmt.Errorf("declare exchange %q: %w", cfg.Exchange, err)
	}

	log.Infof("Successfully connected to RabbitMQ. exchange: %s", cfg.Exchange)

	return &RabbitMQ{conn: conn, Channel: ch, Exchange: cfg.Exchange}, nil
}

// Close cleanly shuts the channel and connection down.
func (r *RabbitMQ) Close() {
	if r == nil {
		return
	}
	if r.Channel != nil {
		_ = r.Channel.Close()
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			rabbitLog.Method("Close").WithError(err).Warn("RabbitMQ close encountered an error.")
		}
	}
}
```

Install the driver:

```bash
go get github.com/rabbitmq/amqp091-go
```

3. add the app-level wrapper

In `app/resources.go`, alongside `SetupDB` and `SetupRedis`:

```go
// message broker
func SetupRabbitMQ() (*dbConfig.RabbitMQ, error) {
	broker, err := dbConfig.SetupRabbitMQ()
	if err != nil {
		appLog.Method("SetupRabbitMQ").WithError(err).Error("Error setting up rabbitmq.")
		return nil, fmt.Errorf("rabbitmq setup: %w", err)
	}
	return broker, nil
}
```

This layer looks redundant — it calls one function and returns. It isn't. It's where the app-level log scope (appLog) and the app-level error context get attached, so failures read as "rabbitmq setup: dial rabbitmq: connection refused" instead of a bare driver error.

4. add the field to Dependency

In `internal/pkg/module/module.go`:

```go
type Dependency struct {
	DB          *gorm.DB
	RedisClient *redis.Client
	Broker      *resources.RabbitMQ

	// add new infra here only
}
```

That comment in the struct is the rule: this is the only place a shared resource gets registered for module consumption.

5. build it during startup

In `app/application.go`, inside `initializeResources()`:

```go
func (app *Application) initializeResources() (module.Dependency, error) {

	// db setup
	db, err := SetupDB()
	if err != nil {
		return module.Dependency{}, err
	}

	// redis setup
	redisClient, err := SetupRedis()
	if err != nil {
		return module.Dependency{}, err
	}

	// broker setup
	broker, err := SetupRabbitMQ()
	if err != nil {
		return module.Dependency{}, err
	}

	dep := module.Dependency{
		DB:          db,
		RedisClient: redisClient,
		Broker:      broker,
	}
	return dep, nil
}
```

Order matters if one resource depends on another. Otherwise put it wherever it reads best.

6. guard against nil

In `app/modules.go`, inside `dependencyInit()`:

```go
if dependency.Broker == nil {
	return nil, nil, fmt.Errorf("dependencyInit: nil broker")
}
```

This catches the case where someone adds the field to Dependency but forgets to populate it in step 5 — a nil pointer that would otherwise blow up at the first publish, in production, hours later.

7. use it

Now any module can pull it out of the dependency bundle. In `internal/order/module.go`:

```go
func (om *OrderModule) Initialize(dep module.Dependency, r chi.Router) ([]scheduler.Task, error) {
	om.repository = repositories.NewOrderRepository(dep.DB)
	om.publisher  = NewOrderPublisher(dep.Broker)          // ← here
	om.service    = NewOrderService(om.repository, om.publisher)

	NewOrderRouter(NewOrderHandler(om.service)).Register(r)
	return nil, nil
}
```

Pass the client down as a constructor argument. Don't reach for dep deeper than Initialize — that's the boundary where injection ends and business logic begins.


