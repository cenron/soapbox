import type { Page } from "@playwright/test"
import { expect } from "@playwright/test"

/**
 * Log in as an existing user. Navigates to /login, fills the form, and waits
 * for redirect to home.
 */
export async function loginAs(page: Page, email: string, password: string) {
  await page.goto("/login")
  await page.getByLabel("Email").fill(email)
  await page.getByLabel("Password").fill(password)
  await page.getByRole("button", { name: "Log in" }).click()
  await page.waitForURL("/")
}

/**
 * Register a fresh user and land on the home page. Uses a timestamp suffix
 * so every test run gets its own user — no shared mutable state between
 * parallel browsers.
 */
export async function registerAndLogin(page: Page, prefix: string) {
  const ts = Date.now()
  const username = `${prefix}_${ts}`
  const email = `${username}@test.dev`

  await page.goto("/register")
  await page.getByLabel("Email").fill(email)
  await page.getByLabel("Username").fill(username)
  await page.getByLabel("Display name").fill(username)
  await page.getByLabel("Password").fill("password123")
  await page.getByRole("button", { name: "Create account" }).click()
  await page.waitForURL("/")

  return { email, username }
}

/**
 * Navigate to another user's profile via page.goto() and wait for auth to
 * fully settle. This is the only acceptable use of goto() for internal pages —
 * there is no click path to another user's profile without search/feed.
 *
 * Webkit's token refresh is slower after full navigation, so we:
 * 1. Wait for networkidle (all fetches including token refresh complete)
 * 2. Verify the Profile nav link is visible (proves auth context resolved)
 */
export async function gotoProfileAndWaitForAuth(page: Page, username: string) {
  await page.goto(`/@${username}`)
  await page.waitForLoadState("networkidle")
  await expect(
    page.getByRole("navigation").getByRole("link", { name: "Profile" }),
  ).toBeVisible({ timeout: 10000 })
}

export const SEED = {
  admin: { email: "admin@soapbox.dev", password: "admin123", username: "admin" },
  mod: { email: "mod@soapbox.dev", password: "mod12345", username: "moderator" },
  user: { email: "user@soapbox.dev", password: "user1234", username: "testuser" },
} as const
