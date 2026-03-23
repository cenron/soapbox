import { test, expect } from "@playwright/test"
import { registerAndLogin } from "../helpers"

test.describe("create post", () => {
  test("types in composer, updates character count, posts, and clears textarea", async ({
    page,
  }) => {
    await registerAndLogin(page, "cp")

    const textarea = page.getByPlaceholder("What's happening?")
    await expect(textarea).toBeVisible()

    // Default character count at 280.
    await expect(page.getByText("280", { exact: true })).toBeVisible()

    const postBody = "Hello from E2E test!"
    await textarea.fill(postBody)

    // Character count decrements by the number of characters typed.
    const expectedCount = (280 - postBody.length).toString()
    await expect(page.getByText(expectedCount)).toBeVisible()

    // Wire up response listener BEFORE clicking.
    const postRes = page.waitForResponse(
      (r) => r.url().includes("/api/v1/posts") && r.request().method() === "POST",
    )

    await page.getByRole("button", { name: "Post", exact: true }).click()

    expect((await postRes).status()).toBe(201)

    // Composer clears on success.
    await expect(textarea).toHaveValue("")
  })

  test("Post button is disabled when textarea is empty", async ({ page }) => {
    await registerAndLogin(page, "cp_empty")

    const postBtn = page.getByRole("button", { name: "Post", exact: true })
    await expect(postBtn).toBeDisabled()
  })

  test("Post button is disabled when over 280 character limit", async ({ page }) => {
    await registerAndLogin(page, "cp_over")

    const textarea = page.getByPlaceholder("What's happening?")
    // Fill with 281 characters.
    await textarea.fill("a".repeat(281))

    await expect(page.getByRole("button", { name: "Post", exact: true })).toBeDisabled()
  })
})
