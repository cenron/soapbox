import { test, expect } from "@playwright/test"
import { loginAs, SEED } from "../helpers"

test.describe("logout", () => {
  test("logs out and redirects to login page", async ({ page }) => {
    await loginAs(page, SEED.user.email, SEED.user.password)

    const trigger = page.locator("button").filter({ hasText: `@${SEED.user.username}` })
    await trigger.click()

    await page.getByRole("menuitem", { name: "Log out" }).click()

    await page.waitForURL("/login")

    await expect(page.getByRole("link", { name: "Log in" })).toBeVisible()
    await expect(page.getByRole("link", { name: "Sign up" })).toBeVisible()
  })
})
