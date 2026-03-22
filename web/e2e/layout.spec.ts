import { test, expect } from "@playwright/test"
import { loginAs, SEED } from "./helpers"

test.describe("layout", () => {
  test("desktop shows public sidebar links when unauthenticated", async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 })
    await page.goto("/login")

    await expect(page.getByRole("link", { name: "Home" })).toBeVisible()
    await expect(page.getByRole("link", { name: "Search" })).toBeVisible()

    // Auth-only links should be hidden
    await expect(page.getByRole("link", { name: "Notifications" })).not.toBeVisible()
    await expect(page.getByRole("link", { name: "Profile" })).not.toBeVisible()
    await expect(page.getByRole("link", { name: "Settings" })).not.toBeVisible()
  })

  test("desktop shows all sidebar links when authenticated", async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 })
    await loginAs(page, SEED.user.email, SEED.user.password)

    await expect(page.getByRole("link", { name: "Home" })).toBeVisible()
    await expect(page.getByRole("link", { name: "Search" })).toBeVisible()
    await expect(page.getByRole("link", { name: "Notifications" })).toBeVisible()
    await expect(page.getByRole("link", { name: "Profile" })).toBeVisible()
    await expect(page.getByRole("link", { name: "Settings" })).toBeVisible()
  })

  test("mobile hides sidebar", async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto("/login")

    // Sidebar links should not be visible on mobile (they're in the drawer)
    const sidebar = page.locator("aside")
    await expect(sidebar).not.toBeVisible()
  })

  test("search input is visible on desktop", async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 })
    await page.goto("/login")

    await expect(page.getByPlaceholder("Search...")).toBeVisible()
  })
})
