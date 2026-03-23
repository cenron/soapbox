import { test, expect } from "@playwright/test"
import { registerAndLogin } from "../helpers"

test.describe("thread view", () => {
  test("reply created via API appears in thread below the root post", async ({ page }) => {
    const { email } = await registerAndLogin(page, "th")

    // Get auth token.
    const loginResp = await page.request.post("/api/v1/auth/login", {
      data: { email, password: "password123" },
    })
    expect(loginResp.status()).toBe(200)
    const { access_token } = await loginResp.json()

    const headers = { Authorization: `Bearer ${access_token}` }

    // Create root post via API.
    const rootResp = await page.request.post("/api/v1/posts", {
      headers,
      data: { body: `Root post ${Date.now()}` },
    })
    expect(rootResp.status()).toBe(201)
    const root = await rootResp.json()
    const rootId: string = root.id
    const rootBody: string = root.body

    // Create reply via API.
    const replyBody = `Reply to root ${Date.now()}`
    const replyResp = await page.request.post("/api/v1/posts", {
      headers,
      data: { body: replyBody, parent_id: rootId },
    })
    expect(replyResp.status()).toBe(201)

    // Navigate to root post detail page.
    await page.goto(`/post/${rootId}`)
    await page.waitForLoadState("networkidle")

    // Root post body is displayed.
    await expect(page.getByText(rootBody)).toBeVisible()

    // Reply appears in the thread below.
    await expect(page.getByText(replyBody)).toBeVisible({ timeout: 10000 })
  })
})
