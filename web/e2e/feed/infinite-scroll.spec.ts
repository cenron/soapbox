import { test, expect } from "@playwright/test"
import { registerAndLogin } from "../helpers"

test.describe("feed infinite scroll", () => {
  // This test creates 25 posts and verifies pagination loads more.
  // Use a longer timeout since creating many posts is slow.
  test.setTimeout(120000)

  test("loads more posts when scrolling to bottom", async ({ page }) => {
    await registerAndLogin(page, "feed_scroll")

    // Create enough posts to fill more than one page (default page size is 20).
    for (let i = 1; i <= 25; i++) {
      const textarea = page.getByPlaceholder("What's happening?")
      await textarea.fill(`Scroll test post ${i}`)

      const postRes = page.waitForResponse(
        (r) =>
          r.url().includes("/api/v1/posts") &&
          r.request().method() === "POST",
      )
      await page.getByRole("button", { name: "Post", exact: true }).click()
      await postRes
    }

    // Wait for async fan-out to complete for all posts.
    await page.waitForTimeout(2000)

    // Reload to get a fresh timeline fetch.
    await page.getByRole("navigation").getByRole("link", { name: "Home" }).click()

    // Wait for timeline to load with the first page (newest posts).
    await expect(page.getByText("Scroll test post 25")).toBeVisible({
      timeout: 10000,
    })

    // Set up listener for the paginated feed request BEFORE scrolling.
    const feedRes = page.waitForResponse(
      (r) => r.url().includes("/api/v1/feed") && r.url().includes("cursor="),
    )

    // Scroll to the bottom to trigger infinite scroll.
    await page.evaluate("window.scrollTo(0, document.body.scrollHeight)")

    // Wait for the paginated response.
    const res = await feedRes
    expect(res.status()).toBe(200)

    // Older posts should now be visible.
    await expect(page.getByText("Scroll test post 1", { exact: true })).toBeVisible({
      timeout: 10000,
    })
  })
})
