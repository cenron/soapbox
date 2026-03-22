import { test, expect } from "@playwright/test"
import { loginAs, SEED } from "../helpers"

test.describe("follow / unfollow", () => {
  test("follows and unfollows @admin", async ({ page }) => {
    await loginAs(page, SEED.user.email, SEED.user.password)

    await page.goto("/admin")

    const followButton = page.getByRole("button", { name: "Follow" })
    const followingButton = page.getByRole("button", { name: "Following" })

    // If already following, unfollow first so the test starts from a clean state.
    if (await followingButton.isVisible()) {
      await followingButton.hover()
      await page.getByRole("button", { name: "Unfollow" }).click()
      await expect(followButton).toBeVisible()
    }

    // Read the follower count before following.
    const followersTabPattern = /Followers \((\d+)\)/
    const tabText = await page.getByRole("tab", { name: followersTabPattern }).textContent()
    const beforeCount = parseInt(tabText?.match(/\((\d+)\)/)?.[1] ?? "0", 10)

    // Follow.
    await followButton.click()
    await expect(page.getByRole("button", { name: "Following" })).toBeVisible()

    // Follower count should increment.
    await expect(page.getByRole("tab", { name: `Followers (${beforeCount + 1})` })).toBeVisible()

    // Unfollow by hovering to reveal the Unfollow label.
    await page.getByRole("button", { name: "Following" }).hover()
    await page.getByRole("button", { name: "Unfollow" }).click()

    await expect(page.getByRole("button", { name: "Follow" })).toBeVisible()
  })
})
