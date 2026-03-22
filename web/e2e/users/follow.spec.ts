import { test, expect } from "@playwright/test"
import { registerAndLogin, gotoProfileAndWaitForAuth } from "../helpers"

test.describe("follow / unfollow", () => {
  test("follows and unfollows @admin", async ({ page }) => {
    // Fresh user per run — no shared state between parallel browsers.
    await registerAndLogin(page, "fw")

    // Navigate to admin's profile and wait for auth to settle.
    await gotoProfileAndWaitForAuth(page, "admin")

    const followBtn = page.getByRole("button", { name: "Follow", exact: true })
    const followingBtn = page.getByRole("button", { name: "Following", exact: true })

    // Fresh user — must start unfollowed.
    await expect(followBtn).toBeVisible()

    // --- Follow ---
    const followRes = page.waitForResponse(
      (r) => r.url().endsWith("/follow") && r.request().method() === "POST",
    )
    await followBtn.click()
    expect((await followRes).status()).toBe(204)
    await expect(followingBtn).toBeVisible({ timeout: 15000 })

    // --- Unfollow ---
    const unfollowRes = page.waitForResponse(
      (r) => r.url().endsWith("/follow") && r.request().method() === "DELETE",
    )
    await followingBtn.click()
    expect((await unfollowRes).status()).toBe(204)
    await expect(followBtn).toBeVisible({ timeout: 15000 })

    // Verify unfollow persisted across navigation.
    await page.reload()
    await page.waitForLoadState("networkidle")
    await expect(followBtn).toBeVisible({ timeout: 10000 })
  })
})
