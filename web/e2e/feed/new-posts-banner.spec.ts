import { test, expect } from "@playwright/test"
import { loginAs, registerAndLogin, gotoProfileAndWaitForAuth } from "../helpers"

test.describe("feed new posts banner", () => {
  test("new post from followed user appears after re-navigating home", async ({
    page,
  }) => {
    // Register user A (will post later).
    const userA = await registerAndLogin(page, "banner_a")

    // Log out user A.
    await page.locator("button").filter({ hasText: `@${userA.username}` }).click()
    const logoutRes = page.waitForResponse(
      (r) => r.url().includes("/api/v1/auth/logout") && r.request().method() === "POST",
    )
    await page.getByRole("menuitem", { name: "Log out" }).click()
    await logoutRes
    await page.waitForURL("/login")

    // Register user B (the follower).
    const userB = await registerAndLogin(page, "banner_b")

    // User B follows user A.
    await gotoProfileAndWaitForAuth(page, userA.username)
    const followRes = page.waitForResponse(
      (r) => r.url().includes("/follow") && r.request().method() === "POST",
    )
    await page.getByRole("button", { name: "Follow", exact: true }).click()
    expect((await followRes).ok()).toBeTruthy()

    // User B goes home.
    await page.getByRole("navigation").getByRole("link", { name: "Home" }).click()
    await page.waitForURL("/")

    // Log out user B.
    await page.locator("button").filter({ hasText: `@${userB.username}` }).click()
    const logoutRes2 = page.waitForResponse(
      (r) => r.url().includes("/api/v1/auth/logout") && r.request().method() === "POST",
    )
    await page.getByRole("menuitem", { name: "Log out" }).click()
    await logoutRes2
    await page.waitForURL("/login")

    // Log in as user A and create a post.
    await loginAs(page, userA.email, "password123")

    const postBody = `New post from ${userA.username} ${Date.now()}`
    await page.getByPlaceholder("What's happening?").fill(postBody)
    const postRes = page.waitForResponse(
      (r) => r.url().includes("/api/v1/posts") && r.request().method() === "POST",
    )
    await page.getByRole("button", { name: "Post", exact: true }).click()
    expect((await postRes).status()).toBe(201)

    // Wait for async fan-out.
    await page.waitForTimeout(1000)

    // Log out user A.
    await page.locator("button").filter({ hasText: `@${userA.username}` }).click()
    const logoutRes3 = page.waitForResponse(
      (r) => r.url().includes("/api/v1/auth/logout") && r.request().method() === "POST",
    )
    await page.getByRole("menuitem", { name: "Log out" }).click()
    await logoutRes3
    await page.waitForURL("/login")

    // Log in as user B and check timeline.
    await loginAs(page, userB.email, "password123")

    // The new post should appear in user B's timeline.
    await expect(page.getByText(postBody)).toBeVisible({ timeout: 15000 })
  })
})
