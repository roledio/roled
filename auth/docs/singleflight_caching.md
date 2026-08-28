# Singleflight & Cache-Aside Architecture in `auth`

This document details the design, purpose, and implementation guidelines for the **Singleflight Cache Mechanism** combined with the **Cache-Aside Decorator Pattern** in the `auth` service.

---

## 1. Executive Summary & Purpose

In high-concurrency environments, popular cache keys (e.g. project settings, client credentials, account metadata) can expire or experience cache misses. When this occurs, hundreds or thousands of simultaneous requests hit the service at the same instant.

Without protection, all concurrent requests miss the cache and fall back to querying the database simultaneously—a problem known as **Cache Stampede (or Thundering Herd)**. This causes CPU spikes, database pool exhaustion, and potential cascading failures.

To address this:
1. **Cache-Aside Pattern (Repository Layer)** provides temporal caching via Redis to reduce latency and database read pressure across consecutive requests over time.
2. **Singleflight Mechanism (Service Layer)** provides instantaneous concurrency deduplication in memory, guaranteeing that for any given key, **only one request executes the underlying query/computation**, while all other concurrent callers wait and receive the exact same result.

---

## 2. Architecture & Patterns

```
                          [ Concurrent Incoming Requests ]
                                         │
                                         ▼
                      ┌────────────────────────────────────┐
                      │    Service Layer (Singleflight)    │
                      │  Keys: sf:get_project_settings:<id>│
                      └──────────────────┬─────────────────┘
                                         │  (Only 1 request proceeds)
                                         ▼
                      ┌────────────────────────────────────┐
                      │    Repository (Cache-Aside)        │
                      │    Keys: project_settings:<id>     │
                      └──────────┬──────────────────┬──────┘
                                 │                  │
                         Cache Hit│          Cache Miss
                                 ▼                  ▼
                           ┌──────────┐       ┌───────────┐
                           │  Redis   │       │ Database  │
                           └──────────┘       └───────────┘
```

### Why Both Patterns Are Used Together (2-Tier Resiliency)

| Layer & Pattern | Primary Responsibility | Time Horizon | Protection Target |
| :--- | :--- | :--- | :--- |
| **Cache-Aside** (`project_setting_cache.go`) | Store & retrieve entities from Redis | Seconds to Hours | Database read load across sequential requests |
| **Singleflight** (`get_project_settings.go`) | Deduplicate in-flight concurrent execution | Milliseconds (In-flight) | Thundering Herd / Cache Stampede on cache miss |

---

## 3. Multi-Process & Microservice Pod Considerations

### Singleflight Behavior in Distributed Deployments
In a microservice deployment with multiple running instances/pods (e.g., 10 pods behind a Load Balancer):
- Singleflight operates **in-memory per application process**.
- When **10,000 concurrent requests** hit `GetProjectSettings` across 10 pods during a cold start / cache miss:
  1. The Load Balancer distributes requests across the 10 pods (~1,000 requests per pod).
  2. Each pod deduplicates its local share of incoming requests down to **1 DB query per pod**.
  3. Total database queries across the entire cluster: **Max 10 DB queries** (1 per pod).
  4. The first pod that completes its query populates Redis (e.g., `project_settings:<id>`).
  5. Subsequent requests across all pods immediately hit Redis directly.
- **Database Read Load Reduction**: **99.9%** (from 10,000 queries down to ~10 queries).

### In-Memory Singleflight vs. Distributed Redis Locks

| Feature | In-Memory Singleflight (Chosen Strategy) | Distributed Redis Lock (`SET NX`) |
| :--- | :--- | :--- |
| **Primary Use Case** | **Read APIs & Cache Stampede Protection** | **Critical Write / Non-Idempotent Operations** |
| **Latency Overhead** | **~10–50 nanoseconds** (In-memory `sync.Mutex`) | **+10ms–100ms** (Redis network round-trips & polling) |
| **Cluster Load Reduction** | Reduces 10,000 queries to $N$ queries ($N$ = pod count) | Reduces 10,000 queries to 1 query cluster-wide |
| **Failure Modes** | **Zero** (No lock deadlocks, no stale lock cleanup) | Risk of lock timeouts, polling CPU spikes, stale locks |

> **Conclusion**: In-memory singleflight per process combined with Redis Cache-Aside is the industry-standard choice for read operations. Executing $N$ queries across $N$ pods during a cold start is completely negligible for databases while eliminating distributed lock complexity and network latency.

---

## 4. Key Design & Best Practices

### A. Scoped Cache Key Conventions
To prevent key collisions, maintain clear separation of concerns, and aid in debugging/telemetry, keys are explicitly namespaced into separate files:
- **Singleflight Keys (In-Memory)** in `internal/constants/singleflight_key.go`: `sf:get_project_settings:<project_id>` (defined via `constants.SingleflightKeyGetProjectSettings`)
- **Redis Cache Keys (Distributed)** in `internal/constants/redis_key.go`: `project_settings:<project_id>` (defined via `constants.RedisKeyProjectSettingsByProjectID`)

### B. Generic Reusable Utility (`singleflightutil`)
The singleflight functionality is encapsulated in `auth/pkg/utils/singleflightutil`:
- Uses Go 1.18+ generics (`Do[T any]`) for type safety.
- Handles `nil` group references defensively for safe execution in tests and fallback mode.

```go
res, err, shared := singleflightutil.Do(&s.sfGroup, key, func() (*models.ProjectSettings, error) {
    return s.getProjectSettings(ctx, req)
})
```

### C. Invalidation Lifecycle
When an entity is updated (`UpdateProjectSettings`) or deleted (`DeleteProject`):
1. **Singleflight Invalidation**: `s.sfGroup.Forget(sfKey)` removes any stale in-flight execution handle.
2. **Redis Invalidation**: `s.redis.DeleteWithContext(ctx, redisKey)` purges the cached entity from Redis.

---

## 5. How to Apply Singleflight to Other APIs

When adding Singleflight cache stampede protection to new APIs in `auth`:

1. **Define Singleflight Key in `internal/constants/singleflight_key.go`**:
   ```go
   func SingleflightKeyGetFoo(id string) string {
       return "sf:get_foo:" + id
   }
   ```

2. **Wrap Service Method with `singleflightutil.Do`**:
   ```go
   func (s *fooService) GetFoo(ctx context.Context, req *models.GetFooRequest) (*models.FooResponse, error) {
       key := constants.SingleflightKeyGetFoo(req.ID)
       res, err, _ := singleflightutil.Do(&s.sfGroup, key, func() (*models.FooResponse, error) {
           return s.getFoo(ctx, req)
       })
       return res, err
   }
   ```

3. **Purge Cache on Mutating Operations**:
   ```go
   func (s *fooService) invalidateFooCache(ctx context.Context, id string) {
       s.sfGroup.Forget(constants.SingleflightKeyGetFoo(id))
       if s.redis != nil {
           _ = s.redis.DeleteWithContext(ctx, constants.RedisKeyFooByID(id))
       }
   }
   ```

---

## 6. Verification & Test Coverage

- Unit tests in `auth/pkg/utils/singleflightutil/singleflightutil_test.go` verify generics, deduplication, error propagation, and nil fallback.
- Integration/Service unit tests in `auth/internal/services/project/get_project_settings_test.go` (`TestProjectService_GetProjectSettings_SingleflightDeduplication`) verify that 10 concurrent requests result in exactly **one** underlying DB/repository execution.
