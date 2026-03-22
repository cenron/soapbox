import { test, expect } from "@playwright/test"
import { loginAs, SEED } from "../helpers"

test.describe("profile edit", () => {
  test("updates display name and bio, then reflects changes on the profile page", async ({ page }) => {
    await loginAs(page, SEED.user.email, SEED.user.password)

    await page.goto("/settings")

    await expect(page.getByRole("heading", { name: "Edit profile" })).toBeVisible()

    const ts = Date.now()
    const newDisplayName = `Updated Name ${ts}`
    const newBio = `Bio updated at ${ts}`

    const displayNameInput = page.getByLabel("Display name")
    await displayNameInput.clear()
    await displayNameInput.fill(newDisplayName)

    const bioInput = page.getByLabel("Bio")
    await bioInput.clear()
    await bioInput.fill(newBio)

    await page.getByRole("button", { name: "Save changes" }).click()

    await expect(page.getByText("Profile updated.")).toBeVisible()

    await page.goto(`/${SEED.user.username}`)

    await expect(page.getByRole("heading", { name: newDisplayName })).toBeVisible()
    await expect(page.getByText(newBio)).toBeVisible()
  })
})
