# RabbitMQ connection URL (queue)
QUEUE_URL=amqp://guest:guest@localhost:5672/

# Redis connection URL (quota store)
REDIS_URL=redis://localhost:6379/0

# Provider mapping: JSON string mapping tenantId to provider key
# Example: tenant1 uses SMTP, tenant2 uses Google, tenant3 uses Outlook
PROVIDER_MAP='{"tenant1":"smtp","tenant2":"google","tenant3":"outlook"}'

# Worker and retry configuration
WORKER_COUNT=5
RETRY_POLICY_MAX_RETRIES=3
RETRY_POLICY_INITIAL_DELAY=1s

QUOTA_SCORE_THRESHOLD=0.8
QUOTA_SCALE_FACTOR=1.5

# Validator: comma-separated list of disposable email domains
VALIDATOR_DISPOSABLE_DOMAINS=mailinator.com,trashmail.com,dispostable.com

# SMTP provider credentials
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USER=
SMTP_PASS=
SMTP_FROM=no-reply@example.com

GOOGLE_CREDENTIALS_JSON=
GOOGLE_ACCESS_TOKEN=
GOOGLE_REFRESH_TOKEN=
GOOGLE_EMAIL_SENDER=

ZERO_BOUNCE_API_KEY=

# Outlook provider credentials
OUTLOOK_TOKEN=
OUTLOOK_FROM=
