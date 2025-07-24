# Email Warmup Service

Automated, scalable, and extensible email warmup platform for modern email infrastructure.  
Supports multiple providers, quota management, event-driven processing, and is designed for easy extension and integration.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Usage](#usage)
- [Extending the Service](#extending-the-service)
  - [Adding a New Email Provider](#adding-a-new-email-provider)
  - [Adding a New Queue Backend](#adding-a-new-queue-backend)
  - [Adding a New Quota Store](#adding-a-new-quota-store)
- [Environment Variables](#environment-variables)
- [Dependencies](#dependencies)
- [License](#license)

---

## Overview

**Email Warmup Service** automates the process of warming up email accounts by sending, tracking, and scoring emails, managing quotas, and scaling sending limits based on performance.  
It is built with Go, leverages Redis for quota management, RabbitMQ for event queueing, and supports multiple email providers (SMTP, Google, Outlook).

---

## Features

- **Multi-provider support:** SMTP, Google, Outlook, and easily extensible to more.
- **Quota management:** Redis-backed daily quotas, automatic scaling based on email scores.
- **Event-driven architecture:** RabbitMQ queue for email send events.
- **Disposable domain & ZeroBounce validation:** Prevents sending to disposable email addresses and uses [ZeroBounce](https://zerobounce.net/) for advanced email validation.
- **Configurable retry policy:** Control retries and delays for email sending.
- **Daily scheduler:** Automatically checks scores and scales quotas.
- **Extensible abstractions:** Add new providers, queue backends, or quota stores with minimal changes.

---

## Architecture

The service is modular and highly extensible.  
Key abstractions are implemented as interfaces, making it easy to add new providers, queue backends, or quota stores.

### Main Components

- **Entrypoint:** [`main.go`](main.go) — Loads config, initializes components, starts workers and scheduler.
- **Configuration:** [`internal/config/config.go`](internal/config/config.go) — Loads configuration from `.env` and environment variables.
- **Queue:** [`internal/queue/client.go`](internal/queue/client.go) — Abstraction for event queue (RabbitMQ).
- **Quota:** [`internal/quota/redis-store.go`](internal/quota/redis-store.go) — Redis-backed quota store and scoring.
- **Processor:** [`internal/processor/processor.go`](internal/processor/processor.go) — Handles email send events, scoring, quota deduction.
- **Scheduler:** [`internal/scheduler/scheduler.go`](internal/scheduler/scheduler.go) — Daily job for scaling quotas.
- **Providers:** [`internal/providers/factory.go`](internal/providers/factory.go), [`smtp.go`](internal/providers/smtp.go) — Provider factory and SMTP implementation.
- **Validator:** [`internal/validator/validator.go`](internal/validator/validator.go), [`internal/validator/zerobounce.go`](internal/validator/zerobounce.go) — Disposable domain validator and ZeroBounce integration.

---

## Getting Started

### Prerequisites

- Go 1.23+
- Redis server
- RabbitMQ server

### Installation

```sh
git clone https://github.com/ilivestrong/email_warmup_service.git
cd email_warmup_service
go build -o email-warmup-service
```

---

## Configuration

All configuration is managed via the `.env` file.  
See the provided `.env` for example values.

### Example `.env`

```env
QUEUE_URL=amqp://guest:guest@localhost:5672/
REDIS_URL=redis://localhost:6379/0
PROVIDER_MAP='{"tenant1":"smtp","tenant2":"google","tenant3":"outlook"}'
WORKER_COUNT=5
RETRY_POLICY_MAX_RETRIES=3
RETRY_POLICY_INITIAL_DELAY=1s
VALIDATOR_DISPOSABLE_DOMAINS=mailinator.com,trashmail.com,dispostable.com
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USER=
SMTP_PASS=
SMTP_FROM=no-reply@example.com
ZERO_BOUNCE_API_KEY=your-zerobounce-api-key
```

---

## Usage

### Running the Service

```sh
./email-warmup-service
```

### How It Works

1. **Startup:** Loads config, connects to Redis and RabbitMQ, starts worker goroutines.
2. **Event Queue:** Listens for `SendEmailEvent` messages from RabbitMQ on a queue named `"send_email"`. _Please ensure that a queue with this name is created before running the service._
3. **Processing:** Each event is validated, quota checked, and sent via the appropriate provider.
4. **Scoring:** Delivery, open, bounce, and spam status are scored and saved.
5. **Quota Scaling:** Daily scheduler checks scores and increases quotas for high-performing tenants.

---

## ZeroBounce Integration

This service uses [ZeroBounce](https://zerobounce.net/) to validate recipient email addresses before sending.  
ZeroBounce helps detect invalid, risky, or disposable emails, improving deliverability and sender reputation.

**How it works:**

- Before sending, the validator checks if the email is from a disposable domain.
- Then, it calls the ZeroBounce API to validate the address.
- If the address is invalid or risky, the email is not sent.

**Configuration:**  
Set your ZeroBounce API key in the `.env` file as `ZERO_BOUNCE_API_KEY`.

See [`internal/validator/validator.go`](internal/validator/validator.go) and [`internal/validator/zerobounce.go`](internal/validator/zerobounce.go) for implementation details.

---

## Scheduler

The scheduler runs a daily check (every 20 seconds for demo/testing) to:

- Retrieve the previous day's scores for each tenant.
- Calculate the average score.
- If the average score is high (≥ 0.8), increase the tenant's quota.

See [`internal/scheduler/scheduler.go`](internal/scheduler/scheduler.go) for details.

---

## Extending the Service

### Adding a New Email Provider

Providers are managed via a factory and must implement the `Provider` interface.

**Steps:**

1. Create a new file, e.g., `internal/providers/myprovider.go`.
2. Implement the `Provider` interface:

   ```go
   // filepath: internal/providers/myprovider.go
   package providers

   type MyProvider struct {
       // provider-specific fields
   }

   func (p *MyProvider) Send(email Email) error {
       // Implement sending logic
   }
   ```

3. Register your provider in the factory:

   ```go
   // filepath: internal/providers/factory.go
   // ...existing code...
   func NewFactory(cfg Config) ProviderFactory {
       return ProviderFactory{
           "smtp": NewSMTPProvider(cfg),
           "google": NewGoogleProvider(cfg),
           "outlook": NewOutlookProvider(cfg),
           "myprovider": NewMyProvider(cfg), // Add your provider here
       }
   }
   // ...existing code...
   ```

### Adding a New Queue Backend

Queue backends must implement the `QueueClient` interface.

**Steps:**

1. Create a new file, e.g., `internal/queue/myqueue.go`.
2. Implement the `QueueClient` interface:

   ```go
   // filepath: internal/queue/myqueue.go
   package queue

   type MyQueueClient struct {
       // queue-specific fields
   }

   func (q *MyQueueClient) Publish(event SendEmailEvent) error {
       // Implement publish logic
   }

   func (q *MyQueueClient) Consume(handler func(SendEmailEvent)) error {
       // Implement consume logic
   }
   ```

3. Register your queue client in the application wiring.

### Adding a New Quota Store

Quota stores must implement the `QuotaStore` interface.

**Steps:**

1. Create a new file, e.g., `internal/quota/my_store.go`.
2. Implement the `QuotaStore` interface:

   ```go
   // filepath: internal/quota/my_store.go
   package quota

   type MyQuotaStore struct {
       // store-specific fields
   }

   func (s *MyQuotaStore) GetQuota(tenant string) (int, error) {
       // Implement quota retrieval
   }

   func (s *MyQuotaStore) SetQuota(tenant string, quota int) error {
       // Implement quota setting
   }
   ```

3. Register your quota store in the application wiring.

---

## Environment Variables

| Variable                                              | Description                                 |
| ----------------------------------------------------- | ------------------------------------------- |
| QUEUE_URL                                             | RabbitMQ connection string                  |
| REDIS_URL                                             | Redis connection string                     |
| PROVIDER_MAP                                          | JSON mapping of tenant IDs to provider keys |
| WORKER_COUNT                                          | Number of concurrent email workers          |
| RETRY_POLICY_MAX_RETRIES                              | Max retries for sending emails              |
| RETRY_POLICY_INITIAL_DELAY                            | Initial delay between retries               |
| VALIDATOR_DISPOSABLE_DOMAINS                          | Comma-separated list of disposable domains  |
| SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, SMTP_FROM | SMTP credentials                            |
| ZERO_BOUNCE_API_KEY                                   | API key for ZeroBounce email validation     |

---

## Dependencies

Key dependencies from [`go.mod`](go.mod):

- [github.com/go-redis/redis/v8](https://github.com/go-redis/redis) — Redis client
- [github.com/joho/godotenv](https://github.com/joho/godotenv) — .env loader
- [github.com/spf13/viper](https://github.com/spf13/viper) — Configuration management
- [github.com/streadway/amqp](https://github.com/streadway/amqp) — RabbitMQ client
- [golang.org/x/oauth2](https://pkg.go.dev/golang.org/x/oauth2) — OAuth2 support
- [google.golang.org/api](https://pkg.go.dev/google.golang.org/api) — Google API support

See [`go.mod`](go.mod) for a full list.

---

## License

MIT

---

**Note:** This project is intended for educational and development purposes. Use responsibly and comply with email sending
