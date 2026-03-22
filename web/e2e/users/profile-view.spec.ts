import { test, expect } from "@playwright/test"
import { loginAs, gotoProfileAndWaitForAuth, SEED } from "../helpers"

test.describe("profile view", () => {
  test("own profile via nav link shows profile data", async ({ page }) => {
    await loginAs(page, SEED.user.email, SEED.user.password)

    await page.getByRole("link", { name: "Profile" }).click()

    await expect(page).toHaveURL(`/@${SEED.user.username}`)

    const main = page.getByRole("main")
    await expect(main.getByRole("heading", { level: 1 })).toBeVisible()
    await expect(main.getByText(`@${SEED.user.username}`)).toBeVisible()
    await expect(main.getByText("Followers", { exact: true })).toBeVisible()
    await expect(main.getByText("Following", { exact: true })).toBeVisible()
  })

  test("own profile via dropdown menu", async ({ page }) => {
    await loginAs(page, SEED.user.email, SEED.user.password)

    await page.locator("button").filter({ hasText: `@${SEED.user.username}` }).click()
    await page.getByRole("menuitem", { name: "Profile" }).click()

    await expect(page).toHaveURL(`/@${SEED.user.username}`)
    await expect(page.getByRole("main").getByText(`@${SEED.user.username}`)).toBeVisible()
  })

  test("another user's profile shows name, bio, verified badge, and counts", async ({ page }) => {
    await loginAs(page, SEED.user.email, SEED.user.password)

    await gotoProfileAndWaitForAuth(page, "admin")

    const main = page.getByRole("main")
    await expect(main.getByRole("heading", { name: "Admin" })).toBeVisible()
    await expect(main.getByText("Soapbox administrator")).toBeVisible()
    await expect(main.getByText("Verified")).toBeVisible()
    await expect(main.locator("span.text-muted-foreground").filter({ hasText: /^Followers$/ })).toBeVisible()
    await expect(main.locator("span.text-muted-foreground").filter({ hasText: /^Following$/ })).toBeVisible()
  })
})
