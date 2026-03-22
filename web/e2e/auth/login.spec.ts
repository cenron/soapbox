import { test, expect } from "@playwright/test"
import { SEED } from "../helpers"

test.describe("login", () => {
  test("logs in with valid credentials and redirects to home", async ({ page }) => {
    await page.goto("/login")

    await page.getByLabel("Email").fill(SEED.user.email)
    await page.getByLabel("Password").fill(SEED.user.password)
    await page.getByRole("button", { name: "Log in" }).click()

    await page.waitForURL("/")

    await expect(page.locator("button").filter({ hasText: `@${SEED.user.username}` })).toBeVisible()
  })

  test("shows an error for invalid credentials", async ({ page }) => {
    await page.goto("/login")

    await page.getByLabel("Email").fill(SEED.user.email)
    await page.getByLabel("Password").fill("wrongpassword")
    await page.getByRole("button", { name: "Log in" }).click()

    await expect(page.getByText("Invalid email or password.")).toBeVisible()
    await expect(page).toHaveURL("/login")
  })
})
