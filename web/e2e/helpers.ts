import type { Page } from "@playwright/test"

export async function loginAs(page: Page, email: string, password: string) {
  await page.goto("/login")
  await page.getByLabel("Email").fill(email)
  await page.getByLabel("Password").fill(password)
  await page.getByRole("button", { name: "Log in" }).click()
  await page.waitForURL("/")
}

export const SEED = {
  admin: { email: "admin@soapbox.dev", password: "admin123", username: "admin" },
  mod: { email: "mod@soapbox.dev", password: "mod12345", username: "moderator" },
  user: { email: "user@soapbox.dev", password: "user1234", username: "testuser" },
} as const
