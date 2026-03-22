import { test, expect } from "@playwright/test"

test.describe("layout", () => {
  test("desktop shows sidebar navigation", async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 })
    await page.goto("/login")

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
