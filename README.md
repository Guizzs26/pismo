# Pismo Challenge

REST API for customer account and transaction management.

## Running

```bash
make run
```

The API will be available at `http://localhost:8080`.

## Stopping

```bash
make down
```

## Tests

```bash
make test
```

Just unit tests


```bash
make test-unit
```

Just integration tests

```bash
make test-integration
```

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/accounts` | Create a new account |
| GET | `/accounts/{accountId}` | Retrieve an account by ID |
| POST | `/transactions` | Create a new transaction |

## Headers

| Header | Required | Description |
|--------|----------|-------------|
| `Idempotency-Key` | Yes (POST /transactions) | Unique key to prevent duplicate transactions |

## Environment Variables

All variables have defaults and are configured automatically via `docker compose`.
See `.env.example` for reference.