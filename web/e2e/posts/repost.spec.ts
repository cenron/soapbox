import { test, expect } from "@playwright/test"
import { registerAndLogin } from "../helpers"

test.describe("repost / undo repost", () => {
  test("reposts then undoes a repost", async ({ page }) => {
    const { email } = await registerAndLogin(page, "rp")

    // Retrieve auth token to create the post via API.
    const loginResp = await page.request.post("/api/v1/auth/login", {
      data: { email, password: "password123" },
    })
    expect(loginResp.status()).toBe(200)
    const { access_token } = await loginResp.json()

    // Create post via API.
    const createResp = await page.request.post("/api/v1/posts", {
      headers: { Authorization: `Bearer ${access_token}` },
      data: { body: `Repost test ${Date.now()}` },
    })
    expect(createResp.status()).toBe(201)
    const post = await createResp.json()
    const postId: string = post.id

    // Navigate to post detail.
    await page.goto(`/post/${postId}`)
    await page.waitForLoadState("networkidle")

    const repostBtn = page.getByTitle("Repost")
    const undoRepostBtn = page.getByTitle("Undo repost")

    await expect(repostBtn).toBeVisible()

    // --- Repost ---
    const repostRes = page.waitForResponse(
      (r) =>
        r.url().includes(`/api/v1/posts/${postId}/repost`) && r.request().method() === "POST",
    )
    await repostBtn.click()
    expect((await repostRes).status()).toBe(200)
    await expect(undoRepostBtn).toBeVisible({ timeout: 10000 })

    // --- Undo repost ---
    const undoRes = page.waitForResponse(
      (r) =>
        r.url().includes(`/api/v1/posts/${postId}/repost`) && r.request().method() === "DELETE",
    )
    await undoRepostBtn.click()
    expect((await undoRes).status()).toBe(200)
    await expect(repostBtn).toBeVisible({ timeout: 10000 })
  })
})
