import { test, expect } from "@playwright/test"

test.describe("registration", () => {
  test("registers a new user and redirects to home", async ({ page }) => {
    const ts = Date.now()
    const email = `testuser_${ts}@example.com`
    const username = `testuser_${ts}`
    const displayName = `Test User ${ts}`

    await page.goto("/register")

    await page.getByLabel("Email").fill(email)
    await page.getByLabel("Username").fill(username)
    await page.getByLabel("Display name").fill(displayName)
    await page.getByLabel("Password").fill("password123")

    await page.getByRole("button", { name: "Create account" }).click()

    await page.waitForURL("/")

    await expect(page.locator("button").filter({ hasText: `@${username}` })).toBeVisible()
  })
})
