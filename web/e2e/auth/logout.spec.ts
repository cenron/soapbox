import { test, expect } from "@playwright/test"
import { loginAs, SEED } from "../helpers"

test.describe("logout", () => {
  test("logs out and redirects to login page", async ({ page }) => {
    await loginAs(page, SEED.user.email, SEED.user.password)

    const trigger = page.locator("button").filter({ hasText: `@${SEED.user.username}` })
    await trigger.click()

    // Assert the logout API call succeeds.
    const logoutPromise = page.waitForResponse(
      (res) => res.url().includes("/api/v1/auth/logout") && res.request().method() === "POST",
    )

    await page.getByRole("menuitem", { name: "Log out" }).click()

    const logoutResponse = await logoutPromise
    expect(logoutResponse.status()).toBe(204)

    await page.waitForURL("/login")

    await expect(page.getByRole("banner").getByRole("link", { name: "Log in" })).toBeVisible()
    await expect(page.getByRole("banner").getByRole("link", { name: "Sign up" })).toBeVisible()
  })
})
