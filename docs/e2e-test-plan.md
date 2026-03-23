# E2E Test Plan — User Journey Workflows

This document maps end-to-end Playwright tests to project phases. Each phase adds new user journeys as frontend features are built. Tests accumulate — earlier phases remain and must keep passing.

All e2e tests live in `web/e2e/`. Run with `make web-test-e2e` or `cd web && npm run test:e2e`.

---

## Phase 0B: Frontend Foundation

**Status:** `complete`

| Test file | Journey | What it verifies |
|-----------|---------|-----------------|
| `smoke.spec.ts` | App loads | Title renders, nav bar visible, auth buttons present |
| `navigation.spec.ts` | Page navigation | Login/register links work, 404 page, back-to-home link, public pages accessible |
| `layout.spec.ts` | Responsive layout | Desktop sidebar visible, mobile sidebar hidden, search input on desktop |

---

## Phase 1: Auth Module

**Status:** `complete` (OAuth and session-refresh tests deferred to post-MVP)

| Test file | Journey | What it verifies |
|-----------|---------|-----------------|
| `auth/register.spec.ts` | New user registration | Fill form → submit → redirect to home, user appears in nav |
| `auth/login.spec.ts` | Existing user login | Fill form → submit → redirect to home, nav shows username |
| `auth/logout.spec.ts` | User logout | Click logout → redirect to login, nav shows login/signup |
| `auth/protected-routes.spec.ts` | Auth guard | Unauthenticated visit to /settings → redirect to /login, after login → redirect back |
| `auth/oauth.spec.ts` | OAuth flow | OAuth buttons visible, click triggers redirect (mock provider) |
| `auth/session-refresh.spec.ts` | Silent token refresh | Token expires → app refreshes silently → user stays logged in |

---

## Phase 1: Users Module

**Status:** `complete`

| Test file | Journey | What it verifies |
|-----------|---------|-----------------|
| `users/profile-view.spec.ts` | View a profile | Navigate to /@username → see display name, bio, avatar, post/follower counts |
| `users/profile-edit.spec.ts` | Edit own profile | Settings → update display name/bio → save → changes reflect on profile page |
| `users/follow.spec.ts` | Follow/unfollow | Visit profile → click follow → count updates, click again → unfollows |
| `users/followers-list.spec.ts` | Browse followers | Profile → followers tab → list of user cards with follow buttons |

---

## Phase 2: Media Module

**Status:** `complete`

| Test file | Journey | What it verifies |
|-----------|---------|-----------------|
| `media/image-upload.spec.ts` | Upload an image | Drag-and-drop or click → preview appears → progress indicator → upload completes |

---

## Phase 2: Posts Module

**Status:** `complete`

| Test file | Journey | What it verifies |
|-----------|---------|-----------------|
| `posts/create-post.spec.ts` | Compose a post | Type text → character count updates → submit → post appears in feed |
| `posts/post-with-image.spec.ts` | Post with image | Attach image → preview visible → submit → post card shows image |
| `posts/post-with-link.spec.ts` | Post with link | Type URL → link preview auto-generates → submit → preview in post card |
| `posts/delete-post.spec.ts` | Delete own post | Click delete → confirm → post removed from feed |
| `posts/like.spec.ts` | Like/unlike a post | Click heart → count increments (optimistic), click again → decrements |
| `posts/repost.spec.ts` | Repost a post | Click repost → count increments, click again → undoes |
| `posts/thread.spec.ts` | View a thread | Click post → post detail page → replies visible in order |
| `posts/reply.spec.ts` | Reply to a post | Open post detail → compose reply → submit → appears in thread |
| `posts/hashtag.spec.ts` | Hashtag in post | Post with #tag → tag is clickable → navigates to search results |

---

## Phase 3: Feed Module

**Status:** `pending`

| Test file | Journey | What it verifies |
|-----------|---------|-----------------|
| `feed/timeline.spec.ts` | Home timeline | Login → home page shows posts from followed users, chronological order |
| `feed/infinite-scroll.spec.ts` | Pagination | Scroll down → older posts load automatically |
| `feed/new-posts-banner.spec.ts` | Real-time new posts | New post arrives via WebSocket → "N new posts" banner → click → posts load at top |
| `feed/empty-state.spec.ts` | New user feed | New user with no follows → empty state with suggestion to follow people |

---

## Phase 3: Notifications Module

**Status:** `pending`

| Test file | Journey | What it verifies |
|-----------|---------|-----------------|
| `notifications/badge.spec.ts` | Nav badge | Receive notification → badge appears on nav icon with count |
| `notifications/list.spec.ts` | Notifications page | Navigate to /notifications → list of activities grouped by type |
| `notifications/mark-read.spec.ts` | Mark as read | Click notification → marked read, "mark all read" clears badge |
| `notifications/realtime.spec.ts` | Real-time push | Another user likes post → notification appears without page refresh |

---

## Phase 4: Moderation Module

**Status:** `pending`

| Test file | Journey | What it verifies |
|-----------|---------|-----------------|
| `moderation/block.spec.ts` | Block a user | Visit profile → block → their posts disappear from feed |
| `moderation/mute.spec.ts` | Mute a user | Visit profile → mute → their posts hidden, can unmute from settings |
| `moderation/report.spec.ts` | Report content | Click report on post/user → modal → select reason → submit → confirmation |
| `moderation/admin-panel.spec.ts` | Admin review queue | Login as admin → /admin → report queue → resolve report, ban user |
| `moderation/admin-access.spec.ts` | Admin route protection | Non-admin visits /admin → forbidden or hidden, admin sees admin link in sidebar |

---

## Phase 4: Search Module

**Status:** `pending`

| Test file | Journey | What it verifies |
|-----------|---------|-----------------|
| `search/search-users.spec.ts` | Search for users | Type in search bar → results tab shows matching users |
| `search/search-posts.spec.ts` | Search for posts | Type query → posts tab shows matching posts |
| `search/search-hashtags.spec.ts` | Search by hashtag | Type #tag → results show posts with that hashtag |
| `search/debounce.spec.ts` | Search UX | Typing triggers debounced search, not on every keystroke |

---

## Phase 5: Integration & Polish

**Status:** `pending`

| Test file | Journey | What it verifies |
|-----------|---------|-----------------|
| `integration/full-journey.spec.ts` | Complete user journey | Register → edit profile → follow user → compose post → like post → receive notification → search → logout |
| `integration/block-filtering.spec.ts` | Block filters across modules | Block user → their posts gone from feed, search, notifications |
| `integration/mobile-responsive.spec.ts` | Mobile full flow | Complete user journey on mobile viewport — drawer nav, responsive cards |

---

## MANDATORY: E2E testing policy

These rules exist because we shipped bugs that E2E tests should have caught. They are not optional.

### 1. Click, don't navigate

E2E tests must reach pages by **clicking links and buttons** — the way a real user does. `page.goto()` is allowed only for the initial entry point (e.g., login page). Every subsequent page transition must happen through UI interaction.

**Why:** `page.goto()` bypasses React Router's client-side route matching. A broken route pattern (`/@:username` not matching in RR v7) was invisible to every E2E test because they all used `goto()` to reach the profile page. The 404 only appeared when a real user clicked a link.

**Bad:**
```ts
await page.goto("/@admin")
await expect(page.getByText("admin")).toBeVisible()
```

**Good:**
```ts
// Start at login, authenticate, then click through
await page.goto("/login")
await login(page, "admin@test.com", "password")
await page.getByRole("link", { name: "Profile" }).click()
await expect(page.getByText("admin")).toBeVisible()
```

### 2. Assert API responses, not just UI rendering

Every E2E test that triggers an authenticated API call must verify the call succeeded. A page can render stale cached data or optimistic UI even when the API returned 401.

**How:** Either:
- Assert a success message appears (e.g., "Profile updated.")
- Use `page.waitForResponse()` to verify the HTTP status
- Assert the persisted result (navigate away and back to confirm data survived)

**Bad:**
```ts
await page.getByRole("button", { name: "Save" }).click()
// Test ends here — doesn't check if save actually worked
```

**Good:**
```ts
await page.getByRole("button", { name: "Save" }).click()
await expect(page.getByText("Profile updated.")).toBeVisible()
// Navigate to profile and verify the change persisted
await page.getByRole("link", { name: "Profile" }).click()
await expect(page.getByText("New bio text")).toBeVisible()
```

### 3. Every module phase must include route matching unit tests

For any non-trivial route pattern (dynamic segments, special characters), add a unit test using `matchRoutes()` that proves the pattern matches expected URLs and doesn't match static routes.

### 4. Every module phase must include an auth integration unit test

At least one unit test must verify that authenticated API calls send the correct `Authorization: Bearer <token>` header. Mock `fetch` and inspect the headers — don't just test the config object.

### 5. Typecheck must pass before E2E tests run

`npx tsc --noEmit` must pass with zero errors before running `make web-test-e2e`. A type error in a test file (like a non-existent property on a response type) indicates a contract mismatch that could hide real bugs.

### 6. Use unique test users for state-mutating tests

Tests that mutate shared server state (follow/unfollow, create/delete posts, block/mute) must register a unique user per test run using `registerAndLogin(page, prefix)`. This prevents race conditions when multiple browsers run in parallel.

### 7. Wait for auth to settle after page.goto()

After any `page.goto()`, wait for `networkidle` AND verify auth-only UI elements are visible (e.g., Profile nav link) before interacting with authenticated features. Webkit's token refresh is slower — clicking before auth settles causes 401 errors.

---

## How to maintain this plan

1. **When starting a phase with frontend work**, check the test file column — create those test files as part of the module implementation.
2. **Mark the phase status as `complete`** once all listed tests pass.
3. **Add new rows** if you discover user journeys not covered here.
4. **Never remove a passing test** — tests accumulate across phases.
5. **All e2e tests must pass before opening a PR** — add `make web-test-e2e` to the pre-PR checklist.
6. **Read the testing policy above** before writing any E2E test. Violations will ship bugs.
