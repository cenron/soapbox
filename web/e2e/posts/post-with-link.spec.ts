import { test, expect } from "@playwright/test"
import { registerAndLogin } from "../helpers"

test.describe("post with link", () => {
  test("creates a post containing a URL and verifies body on detail page", async ({ page }) => {
    await registerAndLogin(page, "pwl")

    const body = `Check out https://example.com for more info ${Date.now()}`
    const textarea = page.getByPlaceholder("What's happening?")
    await expect(textarea).toBeVisible()

    await textarea.fill(body)

    // Wire up before clicking.
    const postRes = page.waitForResponse(
      (r) => r.url().includes("/api/v1/posts") && r.request().method() === "POST",
    )

    await page.getByRole("button", { name: "Post", exact: true }).click()

    const res = await postRes
    expect(res.status()).toBe(201)

    const post = await res.json()
    const postId: string = post.id

    // Navigate to post detail. No click path exists without the feed module.
    await page.goto(`/post/${postId}`)
    await page.waitForLoadState("networkidle")

    // The post body text is intact on the detail page.
    // Link preview fetch is async and may fail for the external URL in CI;
    // we only assert that the body text is present.
    await expect(page.getByText(/Check out https:\/\/example\.com/)).toBeVisible()
  })
})
