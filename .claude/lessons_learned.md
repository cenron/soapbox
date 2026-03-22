# Lessons Learned

<!-- Entry format:
## [YYYY-MM-DD] Short title
**What happened:** Description
**Takeaway:** The rule or insight
-->

## [2026-03-21] TypeScript erasableSyntaxOnly blocks parameter properties
**What happened:** Vite's React TS template enables `erasableSyntaxOnly` in tsconfig.app.json. Using `constructor(public readonly x: string)` syntax fails with TS1294.
**Takeaway:** Always use explicit field declarations + constructor assignment instead of parameter properties in this project. E.g., declare `readonly x: string` as a field, then `this.x = x` in the constructor.

## [2026-03-21] Response body can only be consumed once in tests
**What happened:** `mockFetch.mockResolvedValue(new Response(...))` returns the same Response object for every call. After the first `res.json()`, the body is consumed. Subsequent calls to `res.json()` throw.
**Takeaway:** When testing multiple fetch calls that read the body, use `mockResolvedValueOnce` for each call, or avoid calling the API twice in the same test.

## [2026-03-21] shadcn buttonVariants export triggers react-refresh lint error
**What happened:** shadcn generates `export { Button, buttonVariants }` — the `cva()` call result is a `CallExpression`, not covered by `allowConstantExport`.
**Takeaway:** Add `buttonVariants` (and similar shadcn non-component exports) to `allowExportNames` in the eslint config. This will recur with each new shadcn component that exports variants.

## [2026-03-21] shadcn init places files based on components.json aliases
**What happened:** Default shadcn init puts components in `src/components/ui/` and utils in `src/lib/utils.ts`. Our project uses `src/shared/ui/` and `src/shared/lib/utils.ts`.
**Takeaway:** Update `components.json` aliases immediately after init — before adding any components — to point `ui` to `@/shared/ui`, `lib` to `@/shared/lib`, `hooks` to `@/shared/hooks`. Future `npx shadcn add` commands will then place files correctly.

## [2026-03-22] Avoid circular dependencies between auth token storage and API client
**What happened:** `token-storage.ts` imported `api` from `client.ts`, which imports `getAccessToken` from `token-storage.ts`. Works due to ESM hoisting but is fragile.
**Takeaway:** `refreshAccessToken()` should call `fetch` directly — it's a bootstrap operation that shouldn't depend on the API client it helps configure.

## [2026-03-22] SPA catch-all handler must exclude API paths
**What happened:** Using `router.NotFound(SPAHandler(...))` catches all unmatched routes including `/api/v1/nonexistent`, returning `index.html` instead of a JSON 404.
**Takeaway:** SPA handler must check path prefixes (`/api/`, `/swagger/`, `/healthz`, `/ws`) and return a proper JSON 404 for API paths. Also check `stat.IsDir()` to prevent directory listings.

## [2026-03-22] Vitest picks up Playwright test files
**What happened:** Vitest's default include pattern matches `e2e/*.spec.ts` alongside `src/**/*.test.ts`, causing Playwright's `test.describe()` to fail in the Vitest runner.
**Takeaway:** Add `exclude: ["e2e/**", "node_modules/**"]` to `vitest.config.ts` when Playwright tests live in the same package.

## [2026-03-22] npm create vite@latest initializes a nested git repo
**What happened:** `npm create vite@latest web` runs `git init` inside `web/`, creating a nested `.git` directory. The parent repo then treats `web/` as a submodule-like entry and `git add` silently does nothing.
**Takeaway:** After scaffolding with Vite, immediately `rm -rf web/.git` before staging any files.
