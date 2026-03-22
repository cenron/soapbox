import { test, expect } from "@playwright/test"
import { SEED } from "../helpers"

test.describe("protected routes", () => {
  test("redirects unauthenticated user from /settings to /login", async ({ page }) => {
    await page.goto("/settings")
    await page.waitForURL("/login")
    await expect(page).toHaveURL("/login")
  })

  test("redirects back to /settings after login", async ({ page }) => {
    await page.goto("/settings")
    await page.waitForURL("/login")

    await page.getByLabel("Email").fill(SEED.user.email)
    await page.getByLabel("Password").fill(SEED.user.password)
    await page.getByRole("button", { name: "Log in" }).click()

    await page.waitForURL("/settings")
    await expect(page).toHaveURL("/settings")
    await expect(page.getByRole("heading", { name: "Edit profile" })).toBeVisible()
  })
})
