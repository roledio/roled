# Trae AI — Project Rules

Always read `AGENTS.md` at the repository root first. It is the authoritative source for:
- project architecture /auth (Go) and /console (React)
- layered code standards, naming conventions, file-splitting patterns
- end-to-end checklists for adding new CRUD domains or console pages
- lint, build, test commands
- anti-patterns

This file only gives Trae-specific orientation anchors so you can correctly identify project context on session start, before AGENTS.md is fully loaded into context.

## Project: Roled

Two subprojects in the monorepo:

| Path | Stack | What it does |
|---|---|---|
| `/auth/`    | Go 1.26, Fiber v3, sqlx/MariaDB, Redis v9, Zap, Goose, Testcontainers | OAuth2/OIDC + RBAC auth server. REST APIs for Projects, Clients, Resources, Roles, Users, Members, Access Tokens. |
| `/console/` | Vite, React 18, TypeScript, shadcn/ui, Tailwind, TanStack Query v5, React Router v6, Vitest, PostHog, Axios | Admin console UI for project owners to manage the above. |

## Trae session startup checklist

1. Open `AGENTS.md`.
2. Scope the work to /auth or /console. Load pattern reference files listed in AGENTS.md §2/§3 "Live examples of patterns".
3. Baseline sanity run before edits:
   - `cd auth && go build ./...` (backend)
   - `cd console && npm run lint` (frontend)
4. After edits, run and report:
   - Frontend: `cd console && npm run lint && npm run test`
   - Backend:  `cd auth && go build ./... && golangci-lint run && go test ./internal/services/<changed>/... ./internal/handlers/... -count=1`

## Hard rules (mirrored from AGENTS.md; always enforce)

### In /auth/** (Go)

1. 3-tier layering. Handlers don't call repos. Services don't write raw SQL.
2. `internal/services/<domain>/<verb>_<noun>.go` — one file per public method. Never accumulate public methods in `service.go` beyond constructor.
3. Soft delete (UPDATE `deleted_at = NOW(4)`) everywhere; all reads filter `WHERE deleted_at IS NULL`.
4. CustomError + log with context before return.
5. Every mutation → cache invalidation via `shared.Invalidate*Cache(ctx, s.redis, entity)`.
6. Multi-entity writes → `s.registry.Tx(fn(registry repositories.Registry) error { ... })`.
7. Routes → `h.protectedGet/Post/Put/Delete/Patch(path, constants.RouteXxx, handler)`.

### In /console/** (React/TS)

1. `@/` alias; no ../../ paths into `src/`.
2. HttpClient is a prop (pages) / options-bag param (hooks) / first arg (service fns). Never raw axios.
3. List state synced with URL via useSearchParams. Search debounced 300 ms. Params persisted for back-nav.
4. TanStack Query key = tuple `['domain', id | paramsObj]`.
5. Forms: touched state + pure validate. Aria invalid + describedby.
6. Destructive actions → `<ConfirmDialog destructive>` + name confirm for critical resources.
7. UI components from `@/components/ui/*`.
8. Semantic Tailwind tokens only; no raw slate/zinc palette colors.
9. Non-default ProjectDetails tabs = `React.lazy + Suspense`.

## Pattern files to open for reference

When creating something new, first open the closest analogue:

| Goal | File to read |
|---|---|
| Go service with Tx + cache invalidation | `auth/internal/services/project/create_project.go` |
| Go handler skeleton | `auth/internal/handlers/api/project.go` |
| Go MariaDB repository | `auth/internal/repositories/mariadb/project.go` |
| Go redis cache-aside decorator | `auth/internal/repositories/redis/project.go` |
| Go domain CustomError sentinels | `auth/internal/errors/project.go` |
| React list page (URL state, pagination, confirm delete) | `console/src/pages/projects/Projects.tsx` |
| React form (touched, validation, upload) | `console/src/pages/projects/NewProject.tsx` |
| React tabbed detail page (lazy tabs, paramsStore restore) | `console/src/pages/projects/details/ProjectDetails.tsx` |
| HTTP service function shape | `console/src/services/projects/projects.ts` |
| TanStack Query hook wrapper pattern | `console/src/hooks/projects/index.ts` |
