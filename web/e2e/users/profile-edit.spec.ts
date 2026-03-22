import { test, expect } from "@playwright/test"
import { loginAs, SEED } from "../helpers"

test.describe("profile edit", () => {
  test("saves display name and bio, then verifies on profile page", async ({ page }) => {
    await loginAs(page, SEED.user.email, SEED.user.password)

    // Navigate to settings via sidebar click.
    await page.getByRole("link", { name: "Settings" }).click()
    await expect(page).toHaveURL("/settings")
    await expect(page.getByText("Edit profile")).toBeVisible()

    const ts = Date.now()
    const newDisplayName = `Updated Name ${ts}`
    const newBio = `Bio updated at ${ts}`

    await page.getByLabel("Display name").clear()
    await page.getByLabel("Display name").fill(newDisplayName)
    await page.getByLabel("Bio").clear()
    await page.getByLabel("Bio").fill(newBio)

    // Listen for save response BEFORE clicking.
    const savePromise = page.waitForResponse(
      (res) => res.url().includes("/api/v1/users/me") && res.request().method() === "PUT",
    )
    await page.getByRole("button", { name: "Save changes" }).click()

    expect((await savePromise).status()).toBe(200)
    await expect(page.getByText("Profile updated.")).toBeVisible()

    // Verify changes persisted — navigate to profile by clicking.
    await page.getByRole("link", { name: "Profile" }).click()
    await expect(page).toHaveURL(`/@${SEED.user.username}`)
    await expect(page.getByRole("heading", { name: newDisplayName })).toBeVisible()
    await expect(page.getByText(newBio)).toBeVisible()
  })

  test("settings reachable via dropdown menu", async ({ page }) => {
    await loginAs(page, SEED.user.email, SEED.user.password)

    await page.locator("button").filter({ hasText: `@${SEED.user.username}` }).click()
    await page.getByRole("menuitem", { name: "Settings" }).click()

    await expect(page).toHaveURL("/settings")
    await expect(page.getByText("Edit profile")).toBeVisible()
  })
})
