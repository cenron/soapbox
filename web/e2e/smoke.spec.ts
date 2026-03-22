import { test, expect } from "@playwright/test"

test.describe("smoke tests", () => {
  test("app loads and shows the Soapbox title", async ({ page }) => {
    await page.goto("/")
    await expect(page).toHaveTitle("Soapbox")
  })

  test("nav bar is visible with logo", async ({ page }) => {
    await page.goto("/")
    await expect(page.getByRole("link", { name: "Soapbox" })).toBeVisible()
  })

  test("shows login and signup buttons when unauthenticated", async ({ page }) => {
    await page.goto("/")
    await expect(page.getByRole("banner").getByRole("link", { name: "Log in" })).toBeVisible()
    await expect(page.getByRole("banner").getByRole("link", { name: "Sign up" })).toBeVisible()
  })
})
