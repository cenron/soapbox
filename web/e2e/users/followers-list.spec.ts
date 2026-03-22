import { test, expect } from "@playwright/test"
import { loginAs, SEED } from "../helpers"

test.describe("followers list", () => {
  test("shows testuser in @admin's followers list after following", async ({ page }) => {
    await loginAs(page, SEED.user.email, SEED.user.password)

    await page.goto("/admin")

    // Ensure testuser is following admin before checking the list.
    const followButton = page.getByRole("button", { name: "Follow" })
    const followingButton = page.getByRole("button", { name: "Following" })

    if (await followButton.isVisible()) {
      await followButton.click()
      await expect(followingButton).toBeVisible()
    }

    // Switch to the Followers tab.
    await page.getByRole("tab", { name: /Followers/ }).click()

    // testuser should appear as a user card in the list.
    const card = page.locator("div").filter({ hasText: `@${SEED.user.username}` }).first()
    await expect(card).toBeVisible()
  })
})
