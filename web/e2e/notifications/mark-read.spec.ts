import { test, expect } from "@playwright/test"
import { registerAndLogin, SEED, loginAs } from "../helpers"

test.describe("mark notifications as read", () => {
  test("mark all read clears unread indicators", async ({ page }) => {
    // Register a fresh user.
    const { email } = await registerAndLogin(page, "nmr")

    // Get auth tokens.
    const loginResp = await page.request.post("/api/v1/auth/login", {
      data: { email, password: "password123" },
    })
    expect(loginResp.status()).toBe(200)
    const { access_token: recipientToken } = await loginResp.json()

    // Create a post.
    const createResp = await page.request.post("/api/v1/posts", {
      headers: { Authorization: `Bearer ${recipientToken}` },
      data: { body: `Mark read test ${Date.now()}` },
    })
    expect(createResp.status()).toBe(201)
    const post = await createResp.json()

    // Seed user likes the post.
    const seedLogin = await page.request.post("/api/v1/auth/login", {
      data: { email: SEED.user.email, password: SEED.user.password },
    })
    expect(seedLogin.status()).toBe(200)
    const { access_token: actorToken } = await seedLogin.json()

    const likeResp = await page.request.post(`/api/v1/posts/${post.id}/like`, {
      headers: { Authorization: `Bearer ${actorToken}` },
    })
    expect(likeResp.status()).toBe(200)

    // Wait for async processing.
    await page.waitForTimeout(1500)

    // Log back in as recipient.
    await loginAs(page, email, "password123")

    // Navigate to notifications page.
    await page.getByRole("navigation").getByRole("link", { name: "Notifications" }).click()
    await page.waitForURL("/notifications")
    await page.waitForLoadState("networkidle")

    // Verify "Mark all as read" button is visible.
    const markAllBtn = page.getByRole("button", { name: "Mark all as read" })
    await expect(markAllBtn).toBeVisible({ timeout: 10000 })

    // Click mark all as read and verify API call.
    const markAllResp = page.waitForResponse(
      (r) => r.url().includes("/api/v1/notifications/read-all") && r.request().method() === "PUT",
    )
    await markAllBtn.click()
    expect((await markAllResp).status()).toBe(204)

    // After marking all read, the button should disappear.
    await expect(markAllBtn).not.toBeVisible({ timeout: 10000 })
  })
})
