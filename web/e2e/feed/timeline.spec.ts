import { test, expect } from "@playwright/test"
import { registerAndLogin, gotoProfileAndWaitForAuth } from "../helpers"

test.describe("feed timeline", () => {
  test("shows posts from followed users on the home timeline", async ({
    page,
  }) => {
    // Register user A (the poster).
    const userA = await registerAndLogin(page, "feed_a")

    // User A creates a post.
    const postBody = `Hello from ${userA.username}!`
    await page.getByPlaceholder("What's happening?").fill(postBody)

    const postRes = page.waitForResponse(
      (r) => r.url().includes("/api/v1/posts") && r.request().method() === "POST",
    )
    await page.getByRole("button", { name: "Post", exact: true }).click()
    expect((await postRes).status()).toBe(201)

    // Log out user A via dropdown menu.
    await page.locator("button").filter({ hasText: `@${userA.username}` }).click()
    const logoutRes = page.waitForResponse(
      (r) => r.url().includes("/api/v1/auth/logout") && r.request().method() === "POST",
    )
    await page.getByRole("menuitem", { name: "Log out" }).click()
    await logoutRes
    await page.waitForURL("/login")

    // Register user B (the follower).
    await registerAndLogin(page, "feed_b")

    // User B follows user A.
    await gotoProfileAndWaitForAuth(page, userA.username)
    await expect(page.getByRole("button", { name: "Follow", exact: true })).toBeVisible({ timeout: 10000 })

    const followRes = page.waitForResponse(
      (r) => r.url().includes("/follow") && r.request().method() === "POST",
    )
    await page.getByRole("button", { name: "Follow", exact: true }).click()
    expect((await followRes).ok()).toBeTruthy()

    // Wait for the async backfill to complete before navigating home.
    await page.waitForTimeout(1000)

    // Navigate to home by clicking the Home nav link.
    await page.getByRole("navigation").getByRole("link", { name: "Home" }).click()
    await page.waitForURL("/")

    // User A's post should appear in user B's timeline.
    await expect(page.getByText(postBody)).toBeVisible({ timeout: 10000 })
  })

  test("shows own posts in the timeline", async ({ page }) => {
    await registerAndLogin(page, "feed_own")

    const postBody = "My own post in the feed"
    await page.getByPlaceholder("What's happening?").fill(postBody)

    const postRes = page.waitForResponse(
      (r) => r.url().includes("/api/v1/posts") && r.request().method() === "POST",
    )
    await page.getByRole("button", { name: "Post", exact: true }).click()
    expect((await postRes).status()).toBe(201)

    // Wait for the async fan-out to insert into the timeline.
    await page.waitForTimeout(1000)

    // Reload to get a fresh timeline fetch (not cached empty result).
    await page.getByRole("navigation").getByRole("link", { name: "Home" }).click()

    // Post should appear in the timeline.
    await expect(page.getByText(postBody)).toBeVisible({ timeout: 10000 })
  })
})
