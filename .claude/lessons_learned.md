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

## [2026-03-22] Goose splits dollar-quoted PL/pgSQL blocks on semicolons
**What happened:** A `DO $$ ... $$;` block in a goose SQL migration failed with "unterminated dollar-quoted string" because goose splits statements on `;` by default, breaking the block mid-body.
**Takeaway:** Wrap PL/pgSQL blocks (DO, CREATE FUNCTION, etc.) with `-- +goose StatementBegin` / `-- +goose StatementEnd` so goose treats them as a single statement.

## [2026-03-22] Don't use DEFAULT gen_random_uuid() when app generates UUIDv7
**What happened:** Migrations used `DEFAULT gen_random_uuid()` (v4) but the Go app uses `types.NewID()` which produces v7. This creates two populations of IDs with different ordering properties, breaking cursor pagination assumptions. Also, `gen_random_uuid()` requires pgcrypto on older Postgres versions.
**Takeaway:** When the app always supplies IDs (e.g., UUIDv7), omit the `DEFAULT` clause entirely. Let the NOT NULL + PRIMARY KEY constraint enforce that the app provides the ID.

## [2026-03-22] Never store raw refresh tokens — hash them
**What happened:** PR review flagged that storing plaintext refresh tokens means a database compromise leaks replayable tokens.
**Takeaway:** Store SHA-256 hashes of refresh tokens. The column should be `refresh_token_hash`, not `refresh_token`. Hash on write, hash on lookup. Same pattern as API keys.

## [2026-03-22] Seed migrations must be environment-guarded
**What happened:** A seed migration that creates a dev admin account with a well-known password would run in production since goose migrations are unconditional.
**Takeaway:** Guard seed migrations with `current_setting('app.env')` check. Default to `'development'` if unset. Skip with a NOTICE log in non-dev environments. Alternatively, use a separate seed script outside the migration system.

## [2026-03-22] Skeleton stubs must fail loudly, not pass silently
**What happened:** Middleware stubs called `next.ServeHTTP(w, r)` without any auth check, meaning accidentally wiring them into routes would create open endpoints. Password stubs returned generic "not implemented" errors with no function context.
**Takeaway:** Skeleton middleware stubs should return 401/403. Skeleton function stubs should either be implemented or return errors that identify which function is unimplemented. Never let a stub silently succeed — it masks integration bugs.

## [2026-03-22] goose provider.Close() closes the shared *sql.DB connection
**What happened:** `goose.NewProvider()` receives a `*sql.DB` from the shared sqlx pool. Calling `provider.Close()` closes that underlying connection, making the entire DB pool unusable after migrations.
**Takeaway:** Never call `provider.Close()` when the `*sql.DB` is shared with the rest of the app. Add a comment explaining why.

## [2026-03-22] httpkit.Error must log non-AppError errors
**What happened:** All non-AppError errors were silently swallowed — the 500 response had no corresponding log line, making debugging impossible.
**Takeaway:** Always `slog.Error("unhandled error", "error", err)` before returning a 500. Errors should never disappear into a void.

## [2026-03-22] Public endpoints need AuthOptional middleware for viewer-relative fields
**What happened:** `GET /users/:username` returned `is_following: false` even for authenticated users because the JWT was in the header but no middleware decoded it on public routes.
**Takeaway:** Use `AuthOptional` middleware on public endpoints that have viewer-relative response fields (e.g., is_following). It decodes the JWT if present but doesn't reject unauthenticated requests.

## [2026-03-22] golangci-lint gocritic requires named results when both returns are the same type
**What happened:** `func VerifiedFrom(ctx) (bool, bool)` triggered `unnamedResult` lint error. Then `(verified bool, ok bool)` triggered `paramTypeCombine`. And named returns require `=` not `:=` in the body.
**Takeaway:** When returning two values of the same type, use combined named results: `(verified, ok bool)` and assign with `=` not `:=`.

## [2026-03-22] swaggo does not support Go generics in annotations
**What happened:** `@Success 200 {object} types.CursorPage[ProfileResponse]` caused `ParseComment error: cannot find type definition`. swaggo's parser doesn't resolve generic type instantiations.
**Takeaway:** Create a concrete swagger-only struct (e.g., `ProfileCursorPage`) that mirrors the generic type's fields and use that in annotations. The runtime code still uses the generic `types.CursorPage[T]`.

## [2026-03-22] Escape ILIKE wildcards in user-supplied search input
**What happened:** `SearchUsers` built an ILIKE pattern with `"%" + query + "%"` — if a user searched for `%` or `_`, those are ILIKE wildcards and match unintended rows.
**Takeaway:** Escape `\`, `%`, and `_` in user input before wrapping with wildcards. This is pattern injection, not SQL injection (the value is parameterized), but it produces incorrect results.

## [2026-03-22] Frontend error display must use detail, not message
**What happened:** Register page showed "conflict" to the user instead of "email already registered". The backend `AppError` has `message` (generic: "conflict", "not found") and `detail` (human-friendly: "email already registered"). The frontend was displaying `message`.
**Takeaway:** Always prefer `error.detail` over `error.message` for user-facing error display. The `message` field is a category ("conflict", "validation error"), while `detail` is the explanation. Pattern: `err.detail ?? err.message ?? "Something went wrong."`.

## [2026-03-22] @hey-api/openapi-ts bundles client-fetch since v0.73.0
**What happened:** Installing `@hey-api/client-fetch` separately triggered a deprecation warning — it's been bundled into `@hey-api/openapi-ts` since v0.73.0.
**Takeaway:** Only install `@hey-api/openapi-ts` (pinned with `-E`). The client runtime is included. Don't add `@hey-api/client-fetch` as a separate dependency.

## [2026-03-22] React Router v7 does not match literal `@` before route params
**What happened:** Route `/@:username` never matched URLs like `/@alice` — `matchPath("/@:username", "/@alice")` returns `null`. The catch-all `*` route fired instead, showing a 404 page. All E2E tests passed because they used `page.goto()` which bypasses client-side routing.
**Takeaway:** Use `/:username` instead of `/@:username` in React Router v7. The `@` becomes part of the captured param — strip it with `.replace(/^@/, "")` in the component. React Router's ranked routing ensures static paths (`/login`, `/settings`) still take priority over the dynamic segment. Always add a unit test for non-trivial route patterns using `matchPath()` or `matchRoutes()`.

## [2026-03-22] Swagger apiKey security type does not add Bearer prefix
**What happened:** `@securityDefinitions.apikey BearerAuth` in swaggo generates `{ type: "apiKey", name: "Authorization" }`. The hey-api generated client sends the raw token as the header value without `Bearer ` prefix. The Go backend middleware expects `Authorization: Bearer <token>`, so all authenticated requests returned 401.
**Takeaway:** When using swaggo's `apikey` security definition (Swagger 2.0 limitation), the frontend auth callback must prepend `Bearer ` to the token: `auth: () => token ? \`Bearer ${token}\` : ""`. Add a unit test that verifies the actual `Authorization` header value on outgoing fetch requests, not just the auth callback's return value.

## [2026-03-22] E2E tests must click through the UI, not navigate directly to URLs
**What happened:** E2E tests used `page.goto("/@admin")` to reach the profile page. This loads `index.html` from the backend's SPA handler but never exercises React Router's client-side route matching. The profile route was broken (`/@:username` doesn't match in RR v7) but all E2E tests passed because `goto()` bypasses the router entirely.
**Takeaway:** E2E tests for internal pages must navigate by clicking links and buttons — the way a real user would. Use `page.goto()` only for the initial entry point (login page, home page). Every subsequent page transition must happen through UI interaction. This tests the full stack: link `href` generation, React Router matching, component rendering, and API calls.

## [2026-03-22] Always kill existing services before starting new ones
**What happened:** Multiple `make run` / `./bin/web` instances were spawned without killing previous ones, causing port conflicts and stale builds being served during Playwright MCP testing. Tests passed against old code while new bugs went undetected.
**Takeaway:** Before starting any service, always run `lsof -ti :5173,:8080,:3000 | xargs kill 2>/dev/null; sleep 1`. This applies to all contexts: unit tests, E2E tests, Playwright MCP manual testing, rebuilds. No exceptions.

## [2026-03-22] Public endpoints with viewer-relative fields need @Security in swagger
**What happened:** `GET /users/{username}` returned `is_following: false` for authenticated users because the swagger annotation had no `@Security BearerAuth`. The generated client didn't send the auth token on this "public" endpoint, so the backend's `AuthOptional` middleware never received the JWT.
**Takeaway:** Any endpoint that uses `AuthOptional` middleware (public but returns viewer-relative fields like `is_following`) must have `@Security BearerAuth` in its swagger annotation. Without it, the generated client won't send the auth token.

## [2026-03-22] FollowButton must invalidate queries on error, not just on success
**What happened:** When the follow API returned 409 "already following", the FollowButton showed the error but didn't refetch the profile. The button stayed as "Follow" even though the user was actually following — stale `is_following: false` persisted because only `onSuccess` triggered query invalidation.
**Takeaway:** Mutation error handlers should invalidate queries to sync UI with server state, not just display the error message. A 409 "already following" means the server state differs from what the UI shows.

## [2026-03-22] Playwright E2E selectors must be scoped and verified against actual DOM
**What happened:** Multiple test failures from: (1) `getByRole("heading")` for shadcn `CardTitle` which renders as `<div>`, not `<h1>`; (2) `getByText("Sign up")` matching two elements (nav bar + login form); (3) `getByText("Followers")` matching both stat label and tab button.
**Takeaway:** Always verify element roles against actual component source — shadcn components often don't render semantic HTML elements. Scope selectors to landmarks: `page.getByRole("banner").getByRole("link", ...)`. Use `{ exact: true }` when text appears as substring of other elements.

## [2026-03-22] Set up Playwright response listeners BEFORE the triggering action
**What happened:** `page.waitForResponse()` was called after `button.click()`. The response fired before the listener was attached, causing the test to hang until timeout.
**Takeaway:** Always set up `page.waitForResponse()` (or `page.waitForRequest()`) before the action that triggers the network call. Pattern: `const promise = page.waitForResponse(...)` then `await button.click()` then `const res = await promise`.

## [2026-03-22] Webkit drops in-memory auth tokens after page.goto() under concurrency
**What happened:** Follow test POST returned 401 on webkit when 3 browsers ran in parallel. The in-memory access token was lost after `page.goto("/@admin")` because webkit's full navigation remounts the SPA. The `refreshAccessToken()` call on mount was slow under concurrent load and the follow button was clicked before auth settled.
**Takeaway:** After `page.goto()` in E2E tests, always `await page.waitForLoadState("networkidle")` before interacting with authenticated features. This ensures the token refresh completes. Also enable `retries: 1` locally in playwright.config.ts to handle transient webkit timing issues.

## [2026-03-22] Use unique test users for E2E tests that mutate shared state
**What happened:** Follow/unfollow tests used the shared seed user `testuser`. When 3 browsers ran in parallel, they raced on the same follow state — one browser's follow conflicted with another's unfollow, causing 409 errors and stale UI.
**Takeaway:** Add a `registerAndLogin(page, prefix)` helper that creates a unique user per test run (timestamp-based). Use it for any test that mutates shared state (follow, post, block). Use `{ exact: true }` on button selectors to avoid matching usernames that start with the button label (e.g., "Follow" matching "F @follow_123").

## [2026-03-22] Nav items behind auth must be hidden when logged out
**What happened:** Sidebar showed Notifications, Profile, and Settings links to unauthenticated users. Clicking them either redirected to login or showed broken pages.
**Takeaway:** Use an `authOnly` flag on nav items and filter them based on auth state. Test both states in E2E: verify auth-only items are hidden when logged out and visible when logged in.

## [2026-03-22] Sequential browser projects don't fix webkit auth flakes — keep parallel
**What happened:** Tried splitting user module tests into sequential browser projects (chromium → firefox → webkit via `dependencies`) to fix webkit 401s on follow. It didn't help — the root cause is webkit's slower token refresh after `page.goto()`, not browser concurrency. Sequential runs nearly doubled test time (14s → 23s) with no reliability gain.
**Takeaway:** Keep `fullyParallel: true`. Fix webkit flakes with: (1) unique users via `registerAndLogin()`, (2) `waitForLoadState("networkidle")` after `goto()`, (3) verify auth-only UI is visible before clicking authenticated features, (4) `retries: 1` locally. Don't add complexity that doesn't solve the problem.

## [2026-03-22] Login/register pages must redirect authenticated users
**What happened:** Navigating to `/login` while already authenticated showed the login form with the nav bar displaying the logged-in user. The page didn't check auth state.
**Takeaway:** Add `if (auth.isAuthenticated) return <Navigate to={from} replace />` after all hooks in login/register pages. Use `<Navigate>` component (not `navigate()`) to avoid hook ordering issues — hooks must be called unconditionally. Preserve the `from` location state so protected-route redirects still work (e.g., `/settings` → `/login` → login → `/settings`).

## [2026-03-22] MCP manual validation must cover every page and tab, not just the happy path
**What happened:** Profile page Posts tab showed "Posts will appear here." placeholder because the `GET /users/{username}/posts` endpoint didn't exist. The MCP walkthrough tested compose, like, reply, hashtag, delete — but never clicked the Profile link to check the Posts tab. The gap shipped unnoticed.
**Takeaway:** MCP validation is QA, not a demo. Before marking complete: (1) visit every page reachable from the nav, (2) click every tab, (3) verify data loads in each one, (4) test as a second user too. Use a checklist, not intuition. If a feature creates data, verify it shows up everywhere it should (profile, detail page, search).

## [2026-03-22] PostActions buttons need title attributes for accessibility and test selectors
**What happened:** E2E tests couldn't find like/repost/delete buttons with `getByRole('button', { name: 'Like' })` because buttons only contained SVG icons with no text content. `aria-label` or `title` was missing.
**Takeaway:** Always add `title` attributes to icon-only buttons. This provides accessibility (tooltip on hover, screen reader label) and a stable test selector via `page.getByTitle("Like")`. Don't rely on `getByRole('button', { name })` for icon-only buttons — there's no accessible name without explicit labeling.
