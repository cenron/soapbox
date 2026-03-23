import { test, expect } from "@playwright/test"
import { registerAndLogin } from "../helpers"

test.describe("delete post", () => {
  test("creates a post then deletes it from the detail page", async ({ page }) => {
    await registerAndLogin(page, "dp")

    // Create a post via the UI composer.
    const body = `Delete me ${Date.now()}`
    const textarea = page.getByPlaceholder("What's happening?")
    await expect(textarea).toBeVisible()
    await textarea.fill(body)

    const createRes = page.waitForResponse(
      (r) => r.url().includes("/api/v1/posts") && r.request().method() === "POST",
    )
    await page.getByRole("button", { name: "Post", exact: true }).click()

    const created = await createRes
    expect(created.status()).toBe(201)
    const post = await created.json()
    const postId: string = post.id

    // Navigate to the post detail page. No click path without the feed module.
    await page.goto(`/post/${postId}`)
    await page.waitForLoadState("networkidle")

    // Post body is visible before deletion.
    await expect(page.getByText(body)).toBeVisible()

    // Handle the confirm() dialog that fires before deletion.
    page.once("dialog", (dialog) => dialog.accept())

    // Wire up delete response BEFORE clicking the button.
    const deleteRes = page.waitForResponse(
      (r) =>
        r.url().includes(`/api/v1/posts/${postId}`) && r.request().method() === "DELETE",
    )

    await page.getByTitle("Delete").click()

    expect((await deleteRes).status()).toBe(204)

    // After deletion the component calls navigate(-1), landing back on home.
    await expect(page).toHaveURL("/")
  })
})
