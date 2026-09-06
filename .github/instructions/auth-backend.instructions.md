# Auth / Backend Scoped Instructions
applyTo:
 - auth/**/*.go
---

Canonical detailed source: `AGENTS.md` at repository root (sections 3, 4, 6 "Adding a new backend CRUD domain").

Active patterns for this scope:
1. Handler skeleton (every handler):
```go
func (h *handler) createXxx(c fiber.Ctx) error {
    ctx := c.Context()
    var req models.CreateXxxRequest
    if err := requestutil.BindAndValidate(c, &req); err != nil { return responseutil.SendError(c, err) }
    xxx, err := h.xxxService.CreateXxx(ctx, &req)
    if err != nil { return responseutil.SendError(c, err) }
    return responseutil.SendSuccess(c, xxx)
}
```
List routes: `responseutil.Paginate(req.PageRequest, len(items), total)` then `SendSuccessWithPagination`.

2. Service method file = `<verb>_<noun>.go` inside `internal/services/<domain>/`. Never accumulate multiple public methods in one file.

3. Multi-write → wrap in `s.registry.Tx(func(registry repositories.Registry) error { ... })`. Invalidation AFTER commit: `shared.Invalidate*Cache(ctx, s.redis, entity)`.

4. Repository reads: all SELECT filter `WHERE deleted_at IS NULL`. Repository deletes = soft UPDATE to `deleted_at = NOW(4)`.

5. Error flow: CustomError wrap + log + return. Never raw errors across API boundary.
