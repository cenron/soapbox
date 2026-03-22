import { test, expect } from "@playwright/test"
import { registerAndLogin, gotoProfileAndWaitForAuth } from "../helpers"

test.describe("followers list", () => {
  test("follower appears in the Followers tab after following", async ({ page }) => {
    // Fresh user — avoids stale follow state from other tests.
    const { username } = await registerAndLogin(page, "fl")

    // Navigate to admin's profile and wait for auth to settle.
    await gotoProfileAndWaitForAuth(page, "admin")

    const followBtn = page.getByRole("button", { name: "Follow", exact: true })
    const followingBtn = page.getByRole("button", { name: "Following", exact: true })

    // Follow admin first.
    const followRes = page.waitForResponse(
      (r) => r.url().endsWith("/follow") && r.request().method() === "POST",
    )
    await followBtn.click()
    expect((await followRes).status()).toBe(204)
    await expect(followingBtn).toBeVisible({ timeout: 15000 })

    // Switch to Followers tab — our user should appear.
    await page.getByRole("tab", { name: /Followers/ }).click()

    const card = page.getByRole("main").locator("div").filter({ hasText: `@${username}` }).first()
    await expect(card).toBeVisible({ timeout: 10000 })
  })
})
