# Roled AI Agent Instructions

## Overview

Roled is a centralized User & Role Management Platform with two main sub-projects:
- **`/auth/`** — Go backend (Fiber v3 + MariaDB + Redis + GORM/sqlx) providing OAuth2/OIDC auth, RBAC, project/client/resource/role management APIs
- **`/console/`** — React + TypeScript + Vite + shadcn/ui + Tailwind + TanStack Query admin console

---

## 1. Project Structure & Ownership

```
/                                     # Monorepo root — .github/, .scripts/
├── auth/                             # Go backend
│   ├── cmd/                          # Entrypoint, app wiring, server init
│   │   ├── main.go                   # bootstrap
│   │   ├── app.go                    # App struct + Run/Shutdown/Close
│   │   ├── setup.go                  # DB, logger, services, DI wiring
│   │   ├── handlers.go               # Dependencies structs for api + web handlers
│   │   └── server.go                 # Fiber app setup, routes, middleware
│   ├── configs/default.yml           # Default YAML config
│   ├── internal/
│   │   ├── configs/                  # DefaultConfig + env parsing via Viper
│   │   ├── constants/                # Plain Go constants + rediskeys/, singleflightkeys/ sub-packages
│   │   ├── entities/                 # DB entities (structs with `db:` tags) — pure data
│   │   ├── errors/                   # Domain CustomError variables per entity (reuses pkg/errors)
│   │   ├── handlers/
│   │   │   ├── api/                  # REST API handlers (fiber.Ctx → service → response)
│   │   │   └── web/                  # Web/form handlers (authorize, password, email flows)
│   │   ├── mail/                     # SMTP email sending + templates/
│   │   ├── middlewares/              # JWT, Permission, CORS, CSRF, RequestLogger
│   │   ├── models/                   # Internal request/response DTOs with validate tags
│   │   ├── queues/                   # Redis-backed queue: Publisher / Handler / Worker / Message / Registry
│   │   ├── repositories/
│   │   │   ├── interfaces/           # Repository interfaces (one file per aggregate)
│   │   │   ├── mariadb/              # MariaDB/sqlx impl; namedExecOne/execOne helpers; testutil/ with testcontainers
│   │   │   ├── redis/                # Redis cache decorator wrapping mariadb impl
│   │   │   └── registry.go           # Registry interface + Tx(fn) + decorator composition
│   │   ├── services/
│   │   │   ├── <domain>/             # e.g. project/, user/, member/, accesstoken/, authorize/
│   │   │   │   ├── service.go        # Interface + constructor + struct + singleflight group
│   │   │   │   ├── <verb>_<noun>.go  # One file per method: create_project.go, get_projects.go
│   │   │   │   └── <verb>_<noun>_test.go
│   │   │   ├── infra/                # RedisService, EmailService, NewrelicService
│   │   │   └── shared/               # Cross-cutting helpers: invalidate_cache.go, validate_project.go
│   │   ├── utils/contextutil/        # CtxGetAccount / CtxGetUser etc.
│   │   └── views/                    # Go HTML templates (html/template) + static assets
│   ├── migrations/                   # Goose migrations: 000001_init.sql, 000002_seed.go
│   └── pkg/                          # Shared/reusable packages
│       ├── constants/                # Env, KeyPurpose, Time constants
│       ├── databases/                # DB open/migrate (gorm, sqlx, goose, zap adapters)
│       ├── errors/                   # CustomError base type (Code, Msg, HttpCode, Err, DebugMessage)
│       ├── models/                   # Shared DTOs: Request/Pagination, Response/ErrorBody, BuildInfo
│       ├── repositories/             # WithPercentAround, FixSortDir helpers
│       ├── types/                    # Custom types (e.g. EncryptedString)
│       └── utils/                    # cacheutil, copyutil, encryptionutil, flashutil, idutil,
│                                     # jsonutil, jwtutil, numberutil, passwordutil, pkceutil,
│                                     # randstrutil, requestutil (BindAndValidate), responseutil,
│                                     # singleflightutil, validationutil
└── console/                          # React + TS frontend
    ├── src/
    │   ├── components/
    │   │   ├── layout/               # DashboardLayout, AppSidebar
    │   │   ├── ui/                   # shadcn/ui components (button, card, dialog, form, etc.)
    │   │   ├── ApiPagination.tsx     # Server-driven pagination control
    │   │   ├── CheckboxList.tsx
    │   │   ├── ConfirmDialog.tsx     # Destructive confirm with optional name-type confirmation
    │   │   ├── InviteUserDialog.tsx
    │   │   ├── NavLink.tsx
    │   │   └── StatusBadge.tsx
    │   ├── hooks/
    │   │   ├── use-auth.tsx          # Auth state hook
    │   │   ├── use-current-token-info.ts  # Token info + revoke
    │   │   ├── use-mobile.tsx
    │   │   ├── use-toast.ts
    │   │   ├── projects/             # Domain hooks: useProject, useProjectClients, useProjectSettings
    │   │   ├── members/, users/, accounts/
    │   │   └── index.ts              # Barrel exports
    │   ├── lib/                      # Pure utility modules (no React)
    │   │   ├── utils.ts              # cn() — clsx + tailwind-merge
    │   │   ├── crypto.ts, jwt.ts, pkce.ts, date.ts, logger.ts,
    │   │   ├── paramsStore.ts        # Persist URL params in sessionStorage for back-nav
    │   │   ├── redirect.ts, slugify.ts, telemetry.ts, validation.ts
    │   ├── pages/
    │   │   ├── auth/                 # SignIn, SignInCallback (PKCE flow)
    │   │   ├── account/, Index.tsx, NotFound.tsx
    │   │   └── projects/
    │   │       ├── Projects.tsx, NewProject.tsx
    │   │       └── details/
    │   │           ├── ProjectDetails.tsx  # Tabs: Project, Resources, Roles, Users, Settings
    │   │           ├── tabs/         # Individual tab components (code-split via React.lazy)
    │   │           └── clients|resources|roles|users/  # Detail + New sub-pages
    │   ├── services/
    │   │   ├── core/                 # HttpClient, AuthService, ConfigService, TokenService
    │   │   ├── projects/             # Domain API functions: fetchProjects, createProject, … + types.ts
    │   │   ├── members/, users/, accounts/
    │   │   └── index.ts
    │   ├── storage/                  # StorageService (LocalStorage / SessionStorage wrappers)
    │   └── test/                     # Vitest setup + example tests
    ├── components.json               # shadcn/ui manifest
    ├── eslint.config.js              # Flat config eslint + typescript-eslint
    └── package.json                  # Scripts: dev | build | lint | test | test:coverage
```

---

## 2. Console (Frontend) — Code Standards & Patterns

### 2.1 Tech Stack

- **Framework**: Vite + React 18 + TypeScript (strictNullChecks off, noUnusedLocals off, noUnusedParameters off)
- **Router**: React Router v6 (BrowserRouter, useSearchParams, useNavigate, useParams)
- **Data fetching**: TanStack Query v5 (`useQuery`, `useMutation`, `useQueryClient`)
- **Forms**: Controlled state (useState) + `validate*()` utility in `@/lib/validation`; NOT react-hook-form for main CRUD (forms are small). React Hook Form + zod are available if needed for complex flows.
- **UI kit**: shadcn/ui components in `@/components/ui/*` — generated via components.json. **Always** extend or use these instead of writing raw Tailwind UI primitives from scratch.
- **Styling**: Tailwind CSS v3 + CSS variables via `src/index.css` (slate base, class-variance-authority for variants). Use `cn()` utility for merging.
- **Icons**: `lucide-react` — consistent size h-4 w-4 for inline icons, h-6 w-6 for standalone.
- **HTTP**: `axios` wrapped by `HttpClient` class — NEVER create raw axios instances.
- **Testing**: Vitest + `@testing-library/react` + `vitest-mock-extended`
- **Telemetry**: `posthog-js` via `@/lib/telemetry` — identify on sign-in, reset on sign-out, track events like `project_created`.
- **Storage**: `LocalStorageService` / `SessionStorageService` — always go through these wrappers; never `localStorage.*` directly.

### 2.2 File & Folder Naming

- **Files**: `kebab-case.ts` or `kebab-case.tsx` — NEVER PascalCase file names. Exception: page file names (e.g. `Projects.tsx`) to indicate they are default-exported page components.
- **Folders**: `kebab-case` flat per feature, or nested by route.
- **Barrels**: Use `index.ts` to re-export from a folder (see `src/hooks/index.ts`, `src/services/index.ts`).

### 2.3 Import Paths & Aliases

- Use path alias `@/*` → `./src/*` (defined in tsconfig + vite). **Never** use relative imports like `../../lib/utils` — always `@/lib/utils`.
- Group imports: React (1st), libraries (2nd), `@/` internal (3rd), relative same-folder (4th). Blank line between groups.
- Default-import components, named-import utilities where possible.

### 2.4 Component Patterns

- **Page components** receive typed props: `{ httpClient: HttpClient }` or `{ httpClient, tokenService }` — service deps are injected, NOT imported directly. Example: [Projects.tsx](file:///home/muhammad/Workspace/github/roledio/roled/console/src/pages/projects/Projects.tsx#L35-L37).
- **Hooks are thin wrappers** around TanStack Query. Accept options bag with `httpClient` + `baseUrl` + domain IDs, return immutable `{ data, isLoading, error }` or domain-specific shape. Pattern: [useProject](file:///home/muhammad/Workspace/github/roledio/roled/console/src/hooks/projects/index.ts#L7-L18).
- **Query keys**: `['resource', idOrObj]` — tuple first element is string domain key, second is stable identity object/string.
  - Examples: `['projects', { search, filter, pageNum, pageSize, sortBy, sortDir }]`, `['project', projectId]`, `['currentTokenInfo']`.
- **Mutation lifecycle**: Always invalidate relevant query keys in `onSuccess`. Use `toast` for user feedback on both success and error.
- **Code splitting**: Secondary tabs are `React.lazy(() => import('./tabs/XxxTab'))` wrapped in `<Suspense>`. See [ProjectDetails.tsx](file:///home/muhammad/Workspace/github/roledio/roled/console/src/pages/projects/details/ProjectDetails.tsx#L11-L15).
- **Stable IDs**: Every interactive element gets a meaningful `aria-label` and often `data-testid` in loading/error states.

### 2.5 State & URL Synchronization

- **List/search/filter/pagination/sort state lives in URL search params** + shallow mirror in useState.
  - Initialize state from `useSearchParams()` on mount via `useEffect(..., [])`.
  - Debounce search input (300 ms) into the "real" search state that drives the query.
  - Any filter change that affects server results → reset `pageNum` to 1.
  - `saveProjectsParams(location.search)` into `paramsStore` (sessionStorage) so the back-button can restore list state when returning from Details → New → List. See pattern in [Projects.tsx](file:///home/muhammad/Workspace/github/roledio/roled/console/src/pages/projects/Projects.tsx#L60-L92).
- **ProjectDetails tabs**: Each tab saves its own params (excluding `tab=`) into `saveProjectTabParams(projectId, tabName, searchStr)` and restores via `getProjectTabParams` on tab switch.

### 2.6 Services Layer (HTTP)

- Three core services: `HttpClient`, `TokenService`, `ConfigService`, `AuthService` are instantiated once at the root `App.tsx` via `useMemo` and threaded as props. See [App.tsx](file:///home/muhammad/Workspace/github/roledio/roled/console/src/App.tsx#L35-L39).
- **Domain service file** (`services/projects/projects.ts`): One async function per HTTP verb/endpoint. Signature convention:

  ```ts
  export async function fetchProjects(
    httpClient: HttpClient,
    baseUrl: string,
    params: Record<string, any> = {},
  ): Promise<{ data: Project[]; pagination?: PaginationInfo }>
  ```

- Error handling: Try/catch with layered message extraction `err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? fallback`.
- Always strip trailing slashes from baseUrl: `` `${baseUrl.replace(/\/$/, '')}/api/v1/...` ``.
- Success response shape is `{ success: boolean; data: T; error?: { message: string }; pagination?: PaginationInfo }` — check `res.data?.success` before trusting `data`.

### 2.7 Forms & Validation

- Client-side validation: dedicated pure `validateProjectForm(...)` in [lib/validation.ts](file:///home/muhammad/Workspace/github/roledio/roled/console/src/lib/validation.ts) that returns `{ isValid, errors: { name, description, … }, rowErrors? }`.
- `*Touched` state toggled `onBlur` + first `onChange`. Only show error visuals when field is touched.
- Add `aria-invalid`, `aria-describedby` for a11y.
- Add destructive "confirm by typing name" flow using `ConfirmDialog` + extra `<Input>` when deleting critical resources (projects, clients). See Projects.tsx `removeTarget` + `confirmName`.

### 2.8 shadcn/ui Component Usage

- Always import from `@/components/ui/<component>` — never redefine.
- Prefer variants: `Button variant="outline" | "secondary" | "ghost" | "destructive"` and `size="sm" | "default" | "lg" | "icon"`.
- `Card + CardHeader/CardContent/CardFooter` is the standard container for list items and sections.
- Layout composition: DashboardLayout wraps protected routes. Never bypass it for authenticated pages.

### 2.9 Console Testing Conventions

- Tests colocated with the unit: `foo.tsx` → `foo.test.tsx` (or `.test.ts`).
- Vitest runner: `npm run test`, coverage via `--coverage`.
- Render with `render` from `@testing-library/react`, query by role/label/`data-testid`.
- Mock `HttpClient` and TanStack Query via pattern in existing tests (see `src/hooks/members/index.test.tsx`, `src/services/core/authService.test.ts`).

---

## 3. Auth (Backend) — Code Standards & Patterns

### 3.1 Tech Stack

- **Go 1.26** with module `github.com/roledio/roled/auth`
- **Web**: Fiber v3 (zero-allocation HTTP). Templates: `gofiber/template/html/v3`
- **Router**: Fiber group-based, with helpers `protectedGet/Post/Put/Delete/Patch` that chain `middlewares.JWT` + `middlewares.Permission`.
- **DB**: MariaDB/MySQL via `sqlx` + raw SQL (**not** ORM-generated queries). Query builder: `github.com/Masterminds/squirrel` for dynamic SELECTs. `gorm` available only for migrations, Goose dialect.
- **Cache**: Redis v9 (`redis/go-redis/v9`) with decorator-pattern cached repositories.
- **Logging**: Zap via `fiber/contrib/zap` + lumberjack rotation. All handlers use `log.WithContext(ctx)`.
- **Validation**: `go-playground/validator/v10` via `requestutil.BindAndValidate(c, &req)`.
- **Config**: Viper reads YAML + env overrides (env-var prefix via Viper setup). Config struct in `internal/configs`.
- **Migration**: Goose `goose` in `/migrations/` — SQL (`.sql`) + Go (`.go`) files prefixed with zero-padded numbers.
- **Testing**: testify (`suite`, `assert`, `require`, `mock`). Integration tests use testcontainers (`modules/mariadb`) via `mariadb/testutil`.
- **Observability**: New Relic APM + NR log context; `matoous/go-nanoid/v2` for IDs, `lithammer/shortuuid/v4` in some places.
- **DI strategy**: Manual constructor injection, no DI framework. App struct holds concrete services; they receive `repositories.Registry` + infrastructure.

### 3.2 Layered Architecture — Strict Flow

```
Handler (fiber.Ctx)
  ├► requestutil.BindAndValidate → *models.XxxRequest
  ├► service.Xxx(ctx, req) → (resp, err)
  ├► responseutil.SendError | SendSuccess[WithPagination]
  └► NEVER call repositories directly from handlers

Service struct  (1 interface + 1 private impl per domain)
  ├► read/validate: registry.XxxRepository().Find*(ctx, …)
  ├► business rules, authorization, orchestration
  ├► write paths wrap in registry.Tx(fn func(Registry) error) for multi-entity
  ├► after mutation: shared.Invalidate*Cache(ctx, redis, entity)
  ├► singleflightutil / sfGroup.Do for hot read paths (GetProjectDetails, etc.)
  └► return: response model DTO OR nil + custom error

Registry (decorator composition)
  ├► mariadb/*Repository  → raw SQL/sqlx
  └► if Redis available: redis/*Repository wrapping mariadb impl (CACHE aside pattern)
```

### 3.3 Package / File Conventions

- **PascalCase for public symbols** (Go standard), **camelCase for private fields/structs**.
- One-file-per-use-case in services: `create_project.go` holds only `func (s *projectService) CreateProject(...)`. Constructor, interface, and shared helpers live in `service.go`. This is non-negotiable — do NOT collapse methods into a single large file.
- Repository interfaces: one file per aggregate in `internal/repositories/interfaces/<domain>.go`.
- Redis cache keys: dedicated `func <Name>(<args>) string` in `internal/constants/rediskeys/<domain>.go`. Never build cache key strings inline.
- Singleflight keys: similarly in `internal/constants/singleflightkeys/`.
- Domain constants: resource codes (accounts, projects, clients, …) + action codes (read, create, update, delete) in `internal/constants/resource.go` and `internal/constants/action.go`.
- Route name constants: `internal/constants/route.go` (used in handler registration AND permission middleware).

### 3.4 Models & DTOs

- **Request models** in `internal/models/<domain>.go` structs use **three tag types**:
  - `uri:"field"` for path params
  - `query:"field"` for querystring
  - `json:"field"` for JSON body
  - `validate:"required,notblank,max=50,uri,omitempty,dive"` for validator v10
  - Example: [models/project.go](file:///home/muhammad/Workspace/github/roledio/roled/auth/internal/models/project.go#L9-L66)
- Response DTOs — the same file. Use pointer types for optional fields, slice + struct for nested.
- Pagination: every list request embeds `models.PageRequest` (which has `SetDefaults()`, `Offset()`, `Limit()`).

### 3.5 Entities & DB

- **Entities** (pure DB reflection structs) are in `internal/entities/<domain>.go` with only `db:"col"` tags. Timestamps, soft-delete `DeletedAt *time.Time` always present. No business methods.
- **Soft delete**: Use `UPDATE SET deleted_at = NOW(4)` never `DELETE FROM`. Every SELECT has `WHERE deleted_at IS NULL`.
- **Naming conventions**: snake_case columns mirror struct fields via `db:` tags.
- **Write helpers in mariadb package**:
  - `namedExecOne(ctx, qx, query, struct)` — `sqlx.NamedExecContext` + asserts exactly 1 row affected (panics otherwise for update/delete single).
  - `execOne(ctx, qx, query, args…)` — positional variant.
- **Squirrel for reads**: Use `sq.SelectBuilder`, `sq.Eq/Like/GtOrEq`, `ToSql()`. For fixed/simple queries prefer raw string + GetContext/SelectContext.
- **Sort whitelist**: Map sortBy string to allowed DB column; default fallback to `created_at DESC` to avoid SQLi.

### 3.6 Registry, Transactions & Cache Decorators

- `Registry.Tx(fn)` wraps work in a `*sqlx.Tx` and re-creates the registry with the tx as `QueryExecutor`. Rollback happens if fn returns non-nil. **All multi-entity writes must go through Tx**.
- Cache-aside in redis repos:
  - Read pattern: try `redis.Get` by key → on miss call underlying mariadb → `redis.Set` with TTL → return.
  - Write pattern: always write to DB in service, then **always invalidate** via `shared.Invalidate*Cache` (NOT write-through).
  - TTL comes from `config.CacheDefaultTTLDuration`.
- Repository lookup in registry: constructor returns `mariadb` instance, then if redis is enabled, wraps in `redis.<Name>CacheRepository`. This is done once per accessor — see [registry.go](file:///home/muhammad/Workspace/github/roledio/roled/auth/internal/repositories/registry.go#L98-L248).
- **After every service mutation**, explicit `shared.InvalidateXxxCache(ctx, s.redis, entity)`. If you forget, stale data will persist for TTL.

### 3.7 Error Handling

- **Always** use `pkg/errors.CustomError`. The generic sentinels in `pkg/errors/` are `ErrSystemError`, `ErrInvalidParams`, `ErrInvalidAuthorizationToken`, `ErrInsufficientPermission`, etc. — wrap with `.WithError(err)` to attach cause.
- Domain-specific errors live in `internal/errors/<domain>.go`. Examples: [errors/project.go](file:///home/muhammad/Workspace/github/roledio/roled/auth/internal/errors/project.go).
- In handlers, never `fmt.Errorf` — use the correct CustomError. `responseutil.SendError(c, err)` matches via `errors.As` and sends the correct HTTP code + JSON shape `{ success: false, error: { code, message, debug? } }`.
- Logging pattern on error: `log.WithContext(ctx).Errorw("human description of failure", "error", err, "extra_field", value)` — then return the error. This preserves request correlation (request-id, user, etc.) set by RequestLogger middleware keys.

### 3.8 Request Context (contextutil)

- JWT middleware injects access token + account into both `fiber.Ctx.UserContext()` AND `c.Locals(constants.CtxXxx)`.
- Access them in services via `contextutil.GetAccount(ctx)`, `contextutil.GetUser(ctx)` etc. — they return nil if missing; services guard with domain errors (e.g. `errors.ErrCtxAccountNotFound`).
- **Always pass the original ctx through** — do not create `context.Background()` inside service business logic (only allowed in framework setup code).

### 3.9 Handler Pattern

1. `ctx := c.Context()`
2. Declare typed req struct: `var req models.XxxRequest`
3. `if err := requestutil.BindAndValidate(c, &req); err != nil { return responseutil.SendError(c, err) }`
4. Call service, store triple: `res, total, err` or `res, err`
5. On error: `return responseutil.SendError(c, err)`
6. On success, either:
   - list: `pagination := responseutil.Paginate(req.PageRequest, len(items), total)` → `SendSuccessWithPagination`
   - single: `SendSuccess(c, res)`
- Route registration helpers: `h.protectedGet(path, constants.RouteName, handler)` — these wire JWT + Permission middleware and name the route for the permission check.

### 3.10 Service Pattern

- **service.go** defines the `XxxService` interface, private `xxxService` struct with fields `(defaultConfig, registry, redis, sfGroup singleflight.Group, uploadService …)`.
- Constructor `NewXxxService(...) XxxService` returns interface. Injected deps: always keep the list minimal.
- Singleflight is a zero-value `sfGroup singleflight.Group` in struct. Use for:
  - Heavy aggregate lookups: GetProjectDetails, GetCurrentAccessToken.
  - Pattern: `sfGroup.Do(key, func() (any, error) { … })`. Don't forget to use keys from `constants/singleflightkeys`.
- ID generation: `idutil.NewID()` for standard primary keys, `idutil.NanoID(n)` for token/secret strings.

### 3.11 Middleware

- **JWT**: validates signature, reads claims → looks up access token in repo (Redis→DB) → verifies token status is `AccessTokenStatusIssued` → verifies project_id/client_id/user_id match → injects access token + account into context. Skips `/system/*` and `/api/v1/tokens`.
- **Permission**: looks up route name → maps to resource+action → resolves member role via user → checks role permissions or client permissions (for service principals). 403 if missing.
- **CORS/CSRF**: standard; CSRF only for form-backed web handlers.
- **RequestLogger**: injects extra keys into the log context (request-id, route, user, latency).

### 3.12 Queues (Email)

- Redis stream-based queues. Publisher → stream key → Worker pulls → Handler dispatches.
- Queue names + DLQ names in `constants/queue.go`.
- Publisher interface: `Publish(ctx, queue, payload)`. Payload structs in `queues/payloads/` (e.g. `EmailPayload`).
- Queue message processing: retry, DLQ, dead-letter after max attempts.

### 3.13 Testing (Auth)

- **Unit tests** sit next to unit: `create_project_test.go`. Use `mockery` (`.mockery.yml`) for interface mocks + testify suite.
- **Repository integration tests** use `mariadb/testutil/suite.go` — testcontainers spins up a MariaDB, applies migrations, loads fixtures. Example `mariadb/account_test.go`.
- Always assert both error code (via `errors.Is` for CustomError) AND returned fields where applicable.

---

## 4. Cross-Project / Full-Stack Patterns

### 4.1 API Contract

- All routes are `/api/v1/<domain>` and guarded by route-name constants that also define the permission check. Console services file always mirrors the backend route exactly.
- Route name (e.g. `constants.RouteGetProjects`) is the string used by `Permission` middleware to resolve a required permission. When adding a route, also register: (1) route constant, (2) permission code seeding, (3) console service function, (4) console hook if used across components.

### 4.2 Naming Mirroring

| Auth Concept | Console Mirror |
|---|---|
| `services/project/create_project.go:CreateProject` → returns `*models.ProjectDetails` | `services/projects/projects.ts:createProject()` → returns `Project` (`types.ts`) |
| `entities.Project` fields + snake_case JSON | `Project` type with snake_case keys (matches backend JSON tags) |
| `constants.ResourceCodeProjects` + `constants.ActionRead` | Permission string array in `CurrentTokenInfo.permissions` |

### 4.3 Error Propagation Front→Back

- Backend returns `{ success: false, error: { code, message } }` with appropriate HTTP status.
- Console catches axios error and prioritizes `err.response.data.error.message`, falling back through several layers. **Preserve exact backend error strings** — do not invent new error messages in UI unless the backend truly threw something raw.

---

## 5. Lint & Build Commands

### Console
```bash
cd console
npm run dev            # Start Vite dev server
npm run build          # Production build
npm run lint           # ESLint (typescript-eslint flat config)
npm run test           # Vitest run
npm run test:coverage  # Vitest with coverage
```

### Auth
```bash
cd auth
make build             # Compile binary
make test              # Run unit tests
make lint              # golangci-lint run (see .golangci.yml)
go test ./...          # All tests including integration (requires DB/Redis or uses testcontainers)
```

---

## 6. Common Refactors / Anti-Patterns to Avoid

- **Don't** inline SQL into services. Use repositories.
- **Don't** forget `shared.InvalidateXxxCache` after writes.
- **Don't** write list filters in-memory when the DB query can do it (except client-side search augmentation when explicitly needed).
- **Don't** `var err error` + naked `return`; use named returns only when necessary; nakedret linter forbids long ones.
- **Don't** create new axios instances in console; use `HttpClient.instanceRef`.
- **Don't** add new global CSS unless it's a CSS variable in `index.css`; prefer Tailwind utility classes on components.
- **Don't** skip request validation — every `*Request` struct needs `validate:` tags and calls `BindAndValidate`.
- **Do** keep service methods focused and split across files `verb_noun.go`.
- **Do** prefer `log.WithContext(ctx)...` over `fmt.Println`/`log.Print`.
- **Do** add `aria-label` / `data-testid` to new interactive UI elements.

---

## 7. Quick Start Templates

### Adding a new backend CRUD domain (example: Widget)

1. `internal/entities/widget.go`
2. `internal/models/widget.go` — GetWidgetsRequest, WidgetDetails, CreateWidgetRequest, UpdateWidgetRequest, DeleteWidgetRequest
3. `internal/errors/widget.go` — ErrWidgetNotFound, ErrWidgetNameDuplicate, …
4. `internal/repositories/interfaces/widget.go` — WidgetRepository interface
5. `internal/repositories/mariadb/widget.go` + test if it's critical
6. `internal/repositories/redis/widget.go` (cache decorator)
7. Add accessors to Registry interface + impl (registry.go)
8. `internal/constants/rediskeys/widget.go`, `singleflightkeys/widget.go` if relevant
9. `internal/services/widget/service.go` + `get_widgets.go`, `get_widget_details.go`, `create_widget.go`, `update_widget.go`, `delete_widget.go`
10. Invalidate helpers in `internal/services/shared/invalidate_cache.go`
11. Wire into `Services` struct + `setupServices()` in setup.go
12. `internal/handlers/api` handler methods + protectedXxx registrations in SetupRoutes
13. Add route constants to `constants/route.go` + permission seeding (migration or seed file)
14. Update `newApiHandlerDeps` (handlers.go)

### Adding a new console page under project details

1. `src/pages/projects/details/newwidget/NewWidget.tsx` with props `{ httpClient: HttpClient }`
2. `src/pages/projects/details/newwidget/NewWidget.test.tsx`
3. Domain service: `src/services/projects/widgets.ts` (fetch/create/update/delete)
4. Hook: `src/hooks/widgets/index.ts` with `useWidget`, `useWidgets`
5. Route: Register in `App.tsx` Routes: `/projects/:project_id/widgets/new`, `/projects/:project_id/widgets/:widget_id/details`
6. If inside ProjectDetails tabs: add a new `tabs/WidgetsTab.tsx` (React.lazy), import + register in ProjectDetails TabsContent
