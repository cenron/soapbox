import { test, expect } from "@playwright/test"
import { registerAndLogin, SEED } from "../helpers"

test.describe("real-time notifications", () => {
  test("notification appears without page refresh when another user likes a post", async ({
    page,
  }) => {
    // Register a fresh user who will receive the notification.
    const { email } = await registerAndLogin(page, "nrt")

    // Get auth token.
    const loginResp = await page.request.post("/api/v1/auth/login", {
      data: { email, password: "password123" },
    })
    expect(loginResp.status()).toBe(200)
    const { access_token: recipientToken } = await loginResp.json()

    // Create a post by the recipient.
    const createResp = await page.request.post("/api/v1/posts", {
      headers: { Authorization: `Bearer ${recipientToken}` },
      data: { body: `Realtime test ${Date.now()}` },
    })
    expect(createResp.status()).toBe(201)
    const post = await createResp.json()

    // Navigate to notifications page (should be empty).
    await page.getByRole("navigation").getByRole("link", { name: "Notifications" }).click()
    await page.waitForURL("/notifications")
    await page.waitForLoadState("networkidle")

    // Now, while on the notifications page, another user likes the post via API.
    const seedLogin = await page.request.post("/api/v1/auth/login", {
      data: { email: SEED.user.email, password: SEED.user.password },
    })
    expect(seedLogin.status()).toBe(200)
    const { access_token: actorToken } = await seedLogin.json()

    const likeResp = await page.request.post(`/api/v1/posts/${post.id}/like`, {
      headers: { Authorization: `Bearer ${actorToken}` },
    })
    expect(likeResp.status()).toBe(200)

    // The WebSocket push should trigger a badge update — verify the badge
    // appears in the nav without any page refresh or navigation.
    await page.waitForTimeout(2000)

    const nav = page.getByRole("navigation")
    const badge = nav.locator("[class*='destructive']")
    await expect(badge).toBeVisible({ timeout: 10000 })
  })
})
