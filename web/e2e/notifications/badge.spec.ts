import { test, expect } from "@playwright/test"
import { registerAndLogin, SEED } from "../helpers"

test.describe("notification badge", () => {
  test("badge appears after receiving a notification", async ({ page }) => {
    // Register a fresh user.
    const { email } = await registerAndLogin(page, "nbdg")

    // Get auth token.
    const loginResp = await page.request.post("/api/v1/auth/login", {
      data: { email, password: "password123" },
    })
    expect(loginResp.status()).toBe(200)
    const { access_token: recipientToken } = await loginResp.json()

    // Create a post.
    const createResp = await page.request.post("/api/v1/posts", {
      headers: { Authorization: `Bearer ${recipientToken}` },
      data: { body: `Badge test ${Date.now()}` },
    })
    expect(createResp.status()).toBe(201)
    const post = await createResp.json()

    // Another user likes the post — triggers notification + WS push.
    const seedLogin = await page.request.post("/api/v1/auth/login", {
      data: { email: SEED.user.email, password: SEED.user.password },
    })
    expect(seedLogin.status()).toBe(200)
    const { access_token: actorToken } = await seedLogin.json()

    const likeResp = await page.request.post(`/api/v1/posts/${post.id}/like`, {
      headers: { Authorization: `Bearer ${actorToken}` },
    })
    expect(likeResp.status()).toBe(200)

    // Wait for WebSocket push and React state update.
    await page.waitForTimeout(2000)

    // Verify badge appears on the Notifications nav link.
    const nav = page.getByRole("navigation")
    const badge = nav.locator("[class*='destructive']")
    await expect(badge).toBeVisible({ timeout: 10000 })
  })
})
