import { test, expect } from "@playwright/test"

test.describe("navigation", () => {
  test("login link navigates to /login", async ({ page }) => {
    await page.goto("/")
    await page.getByRole("link", { name: "Log in" }).click()
    await expect(page).toHaveURL("/login")
    await expect(page.getByRole("heading", { name: "Log in" })).toBeVisible()
  })

  test("signup link navigates to /register", async ({ page }) => {
    await page.goto("/")
    await page.getByRole("link", { name: "Sign up" }).click()
    await expect(page).toHaveURL("/register")
    await expect(page.getByRole("heading", { name: "Create account" })).toBeVisible()
  })

  test("unknown routes show 404 page", async ({ page }) => {
    await page.goto("/not/a/real/page")
    await expect(page.getByRole("heading", { name: "404" })).toBeVisible()
    await expect(page.getByText("This page doesn't exist")).toBeVisible()
  })

  test("404 page has a link back to home", async ({ page }) => {
    await page.goto("/not/a/real/page")
    await page.getByRole("link", { name: "Go home" }).click()
    // Home is a protected route — unauthenticated users redirect to /login
    await expect(page).toHaveURL("/login")
  })

  test("search page is accessible", async ({ page }) => {
    await page.goto("/search")
    await expect(page.getByRole("heading", { name: "Search" })).toBeVisible()
  })
})
