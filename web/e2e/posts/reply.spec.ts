import { test, expect } from "@playwright/test"
import { registerAndLogin } from "../helpers"

test.describe("reply to post", () => {
  test("submits a reply via the detail page composer and verifies it appears in thread", async ({
    page,
  }) => {
    const { email } = await registerAndLogin(page, "rply")

    // Get auth token.
    const loginResp = await page.request.post("/api/v1/auth/login", {
      data: { email, password: "password123" },
    })
    expect(loginResp.status()).toBe(200)
    const { access_token } = await loginResp.json()

    // Create root post via API.
    const createResp = await page.request.post("/api/v1/posts", {
      headers: { Authorization: `Bearer ${access_token}` },
      data: { body: `Root for reply test ${Date.now()}` },
    })
    expect(createResp.status()).toBe(201)
    const post = await createResp.json()
    const postId: string = post.id

    // Navigate to post detail.
    await page.goto(`/post/${postId}`)
    await page.waitForLoadState("networkidle")

    // The reply composer is below the root post. There may be multiple
    // "What's happening?" textareas (root PostCard area is absent here),
    // but only the ReplyComposer is on this page alongside the "Replying to" label.
    await expect(page.getByText(/Replying to/)).toBeVisible()

    const replyText = `Reply via UI ${Date.now()}`
    const textarea = page.getByPlaceholder("What's happening?")
    await textarea.fill(replyText)

    // Wire up before clicking Post.
    const replyRes = page.waitForResponse(
      (r) => r.url().includes("/api/v1/posts") && r.request().method() === "POST",
    )

    await page.getByRole("button", { name: "Post", exact: true }).click()

    expect((await replyRes).status()).toBe(201)

    // Reply appears in the thread.
    await expect(page.getByText(replyText)).toBeVisible({ timeout: 10000 })
  })
})
