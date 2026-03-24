import { test, expect } from "@playwright/test"
import { registerAndLogin, SEED, loginAs } from "../helpers"

test.describe("notifications list", () => {
  test("shows notifications after receiving likes and follows", async ({ page }) => {
    // Register a fresh user who will be the notification recipient.
    const { email, username } = await registerAndLogin(page, "nlist")

    // Get auth token to create content via API.
    const loginResp = await page.request.post("/api/v1/auth/login", {
      data: { email, password: "password123" },
    })
    expect(loginResp.status()).toBe(200)
    const { access_token: recipientToken } = await loginResp.json()

    // Create a post by the recipient.
    const createResp = await page.request.post("/api/v1/posts", {
      headers: { Authorization: `Bearer ${recipientToken}` },
      data: { body: `Notification test ${Date.now()}` },
    })
    expect(createResp.status()).toBe(201)
    const post = await createResp.json()

    // Log in as seed user and like the post + follow the recipient.
    const seedLogin = await page.request.post("/api/v1/auth/login", {
      data: { email: SEED.user.email, password: SEED.user.password },
    })
    expect(seedLogin.status()).toBe(200)
    const { access_token: actorToken } = await seedLogin.json()

    const likeResp = await page.request.post(`/api/v1/posts/${post.id}/like`, {
      headers: { Authorization: `Bearer ${actorToken}` },
    })
    expect(likeResp.status()).toBe(200)

    const followResp = await page.request.post(`/api/v1/users/${username}/follow`, {
      headers: { Authorization: `Bearer ${actorToken}` },
    })
    // 200 or 409 (already following) — both are acceptable.
    expect([200, 409]).toContain(followResp.status())

    // Wait for async event handlers to process.
    await page.waitForTimeout(1500)

    // Log back in as the recipient and navigate to notifications.
    await loginAs(page, email, "password123")

    // Click Notifications in the sidebar.
    await page.getByRole("navigation").getByRole("link", { name: "Notifications" }).click()
    await page.waitForURL("/notifications")

    // Verify notification list loads with API success.
    const notifResp = page.waitForResponse(
      (r) => r.url().includes("/api/v1/notifications") && r.request().method() === "GET",
    )
    // The query may already be in flight, so reload to ensure we catch it.
    await page.reload()
    expect((await notifResp).status()).toBe(200)

    // Verify notification items appear.
    await expect(page.getByText("liked your post")).toBeVisible({ timeout: 10000 })
  })
})
