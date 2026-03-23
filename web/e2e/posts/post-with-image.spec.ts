import { test, expect } from "@playwright/test"
import { loginAs, SEED } from "../helpers"

// TODO: Full image-attachment testing (selecting a file, uploading via presigned URL,
// and verifying the media appears on the post) is covered by the media module's
// image-upload.spec.ts. That flow requires the ImageUpload component to complete its
// presigned-URL + confirm round-trip before the post is submitted.
//
// This test verifies the simpler path: a text-only post can be created by the admin
// user (who already has an avatar) and is reachable via its detail page.

test.describe("post with image (text-only baseline)", () => {
  test("admin creates a text post and navigates to its detail page", async ({ page }) => {
    await loginAs(page, SEED.admin.email, SEED.admin.password)

    const body = `Image post baseline ${Date.now()}`
    const textarea = page.getByPlaceholder("What's happening?")
    await expect(textarea).toBeVisible()

    await textarea.fill(body)

    const postRes = page.waitForResponse(
      (r) => r.url().includes("/api/v1/posts") && r.request().method() === "POST",
    )

    await page.getByRole("button", { name: "Post", exact: true }).click()

    const res = await postRes
    expect(res.status()).toBe(201)

    const post = await res.json()
    const postId: string = post.id

    // Navigate to the post detail page. No click path exists without the feed module.
    await page.goto(`/post/${postId}`)
    await page.waitForLoadState("networkidle")

    // Post body appears on the detail page.
    await expect(page.getByText(body)).toBeVisible()
  })
})
