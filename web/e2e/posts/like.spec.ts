import { test, expect } from "@playwright/test"
import { registerAndLogin } from "../helpers"

test.describe("like / unlike post", () => {
  test("likes then unlikes a post", async ({ page }) => {
    const { email } = await registerAndLogin(page, "lk")

    // Retrieve the auth token so we can create the post via API directly.
    const loginResp = await page.request.post("/api/v1/auth/login", {
      data: { email, password: "password123" },
    })
    expect(loginResp.status()).toBe(200)
    const { access_token } = await loginResp.json()

    // Create post via API to avoid UI complexity.
    const createResp = await page.request.post("/api/v1/posts", {
      headers: { Authorization: `Bearer ${access_token}` },
      data: { body: `Like test ${Date.now()}` },
    })
    expect(createResp.status()).toBe(201)
    const post = await createResp.json()
    const postId: string = post.id

    // Navigate to post detail page.
    await page.goto(`/post/${postId}`)
    await page.waitForLoadState("networkidle")

    const likeBtn = page.getByTitle("Like")
    const unlikeBtn = page.getByTitle("Unlike")

    await expect(likeBtn).toBeVisible()

    // --- Like ---
    const likeRes = page.waitForResponse(
      (r) =>
        r.url().includes(`/api/v1/posts/${postId}/like`) && r.request().method() === "POST",
    )
    await likeBtn.click()
    expect((await likeRes).status()).toBe(200)
    await expect(unlikeBtn).toBeVisible({ timeout: 10000 })

    // --- Unlike ---
    const unlikeRes = page.waitForResponse(
      (r) =>
        r.url().includes(`/api/v1/posts/${postId}/like`) && r.request().method() === "DELETE",
    )
    await unlikeBtn.click()
    expect((await unlikeRes).status()).toBe(200)
    await expect(likeBtn).toBeVisible({ timeout: 10000 })
  })
})
