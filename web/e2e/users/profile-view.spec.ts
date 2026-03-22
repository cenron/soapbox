import { test, expect } from "@playwright/test"
import { loginAs, SEED } from "../helpers"

test.describe("profile view", () => {
  test("shows display name, bio, verified badge, and follower counts for @admin", async ({ page }) => {
    await loginAs(page, SEED.user.email, SEED.user.password)

    await page.goto("/@admin")

    await expect(page.getByRole("heading", { name: "Admin" })).toBeVisible()
    await expect(page.getByText("Soapbox administrator")).toBeVisible()
    await expect(page.getByText("Verified")).toBeVisible()

    await expect(page.getByText("Followers")).toBeVisible()
    await expect(page.getByText("Following")).toBeVisible()
  })
})
