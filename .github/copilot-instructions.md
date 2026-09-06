# Roled — GitHub Copilot Custom Instructions

Read `AGENTS.md` at the repository root first. It is the canonical, detailed source of:
- project overview & structure map for /auth (Go) and /console (React)
- layered architecture rules for both subprojects
- detailed naming, file-splitting, and error-handling patterns
- end-to-end checklists for adding new CRUD domains

This file is kept short on purpose. Apply all rules in AGENTS.md. Below are short project anchors so you orient correctly even when AGENTS.md isn't injected.

## Project Identity

Roled is a centralized User & Role Management Platform (OAuth2/OIDC + RBAC). Monorepo:

- `/auth/`    — Go 1.26 backend: Fiber v3, sqlx + MariaDB, Redis cache decorator, Zap logger, Goose migrations, Testcontainers. Module: `github.com/roledio/roled/auth`.
- `/console/` — React 18 + Vite + TypeScript admin console: shadcn/ui, Tailwind v3, TanStack Query v5, React Router v6, Vitest, posthog-js, axios.

## Hard Rules (no exceptions)

### When editing Go under /auth:

1. Strict layers: handlers → services → repositories. Handlers ONLY do BindAndValidate → service call → SendSuccess/SendError. Repositories ONLY do SQL.
2. Service file-splitting: one public method per file named `<verb>_<noun>.go`. `service.go` contains the interface, struct, and constructor.
3. All writes through repositories use soft-delete (`deleted_at = NOW(4)`); all reads filter `WHERE deleted_at IS NULL`.
4. Errors = `CustomError` from `pkg/errors` + `internal/errors/<domain>.go`. `log.WithContext(ctx).Errorw(...)` then `return errors.ErrXxx.WithError(err)`.
5. After every mutation success: `shared.Invalidate*Cache(ctx, s.redis, entity)`. Cache keys from `internal/constants/rediskeys/*.go` only.
6. Multi-entity writes wrap in `s.registry.Tx(func(registry repositories.Registry) error { ... })`. Use the closure param `registry`, not `s.registry`.
7. Route registration via `h.protectedGet/Post/Put/Delete/Patch(path, constants.RouteXxx, handler)`. Add the route name constant to `internal/constants/route.go`.

### When editing React/TS under /console:

1. Imports use `@/` alias exclusively. No `../../` paths into `src/`.
2. HttpClient, TokenService, ConfigService are props on pages; hooks accept them via options objects; services accept as first+second function args `(httpClient, baseUrl, …)`. Never create `axios` instances directly.
3. List pages = URL-driven state (search/filter/sortBy/sortDir/pageNum ↔ `useSearchParams`), debounce 300 ms, persist params in `paramsStore` for back-button restore. TanStack Query keys: `['domain', idOrParamsObj]`.
4. Forms: per-field `*Touched` booleans, pure `validateXxxForm(...)` in `@/lib/validation`, show errors only when touched + error present, `aria-invalid` + `aria-describedby`.
5. Destructive actions use `<ConfirmDialog destructive>`; for Projects/Clients/Users require confirm-by-typing-name flow.
6. UI = shadcn/ui from `@/components/ui/*`. Tailwind = semantic tokens only (`bg-background`, `text-muted-foreground`, `border-border`, `bg-primary text-primary-foreground`).
7. Non-default tabs under `ProjectDetails.tsx` are `React.lazy(...)` + `<Suspense>`; default tab preloaded when project data resolves.

## Files to Always Read for Pattern Reference

### /auth pattern references
- `auth/internal/services/project/service.go` + `auth/internal/services/project/create_project.go`
- `auth/internal/handlers/api/project.go`
- `auth/internal/repositories/mariadb/project.go`
- `auth/internal/repositories/registry.go`
- `auth/internal/services/shared/invalidate_cache.go`

### /console pattern references
- `console/src/pages/projects/Projects.tsx`
- `console/src/pages/projects/NewProject.tsx`
- `console/src/pages/projects/details/ProjectDetails.tsx`
- `console/src/services/projects/projects.ts`
- `console/src/hooks/projects/index.ts`

## Quick verify commands (run & inspect output before finishing work)

```bash
# console
cd console && npm run lint && npm run test

# auth
cd auth && go build ./... && golangci-lint run
cd auth && go test ./internal/services/<changed>/... ./internal/handlers/... -count=1
```
