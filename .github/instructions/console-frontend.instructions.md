# Console / Frontend Scoped Instructions
applyTo:
 - console/src/**/*.ts
 - console/src/**/*.tsx
---

Canonical detailed source: `AGENTS.md` at repository root (sections 2, 4, 6 "Adding a new console page under project details").

Active patterns for this scope:
1. Imports: `@/` alias only. No `../../` for src-relative paths.
2. HttpClient injection as prop on pages/hooks/services. No raw `axios` instance creation.
3. List page state machine: useState X N → hydrate from useSearchParams on mount → effect serializes back → debounce search → TanStack Query with tuple query key `['domain', { ...params }]` + `keepPreviousData: true`. Persist in `paramsStore`.
4. Forms: touched booleans per field, pure `validateXxxForm` returns `{ isValid, errors, rowErrors? }`, error visuals only when touched.
5. Destructive actions = `<ConfirmDialog destructive>` + confirm-by-typing-name for critical resources.
6. UI components: import only from `@/components/ui/*`. Tailwind uses semantic tokens exclusively.
7. Tabs under ProjectDetails: `React.lazy` + `<Suspense>` for non-default tabs; default preloaded when project data resolves.
