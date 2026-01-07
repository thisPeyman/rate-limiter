# Distributed Sliding Window Rate Limiter

A high-performance, distributed rate limiter written in Go using Redis and Lua scripts.

## Architecture

### 1. Algorithm: Sliding Window Counter (Approximation)
We use a **Weighted Sliding Window** algorithm.
- **Why?** Standard fixed windows cause "bursts" at window edges (e.g., 100 requests at 10:00:00.999 and 100 at 10:00:01.001).
- **How?** We store the counter for the `Current Window` and the `Previous Window`.
- **Formula:** `EstimatedRate = CurrentWindowCount + (PreviousWindowCount * (1 - PercentOfWindowElapsed))`
- **Storage:** Redis keys `rate:{userID}:{window_timestamp}`.

### 2. Distributed State
- **Redis:** Acts as the centralized counter store.
- **Lua Scripts:** Used to ensure **atomicity**. The calculation, checking, and incrementing happen inside Redis in a single operation. This prevents race conditions where two concurrent requests both read "9", think they are safe, and both increment to "10", violating the limit.

### 3. Fail-Open Strategy
If Redis is down, the system logs the error but **allows the request**. This prioritizes availability over strict rate limiting during infrastructure outages.

## Project Structure
- `cmd/api`: Entry point.
- `internal/limiter`: Core rate limiting logic and Lua scripts.
- `internal/server`: HTTP Middleware and handlers.
- `internal/platform`: Infrastructure setup (Redis).

## Running the Project

### Prerequisites
- Docker & Docker Compose
- Or: Go 1.25.1+ and a local Redis instance

### Run with Docker Compose (Recommended)
```bash
docker-compose up --build
```

### Run Locally
1. Start Redis: `docker run -p 6379:6379 redis`
2. Run App: `go run cmd/api/main.go`

## Testing
Run unit and integration tests (requires local Redis):
```bash
go test ./... -v
```

## Scaling
To scale this further:
1. **Redis Cluster:** Use Redis Cluster to shard keys by UserID, distributing memory load.
2. **Local Caching:** Implement a tiny in-memory cache (e.g., 50ms TTL) in the Go app to reduce Redis round-trips for extremely high-traffic users ("Hot Keys").
