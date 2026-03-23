import { test, expect } from "@playwright/test"
import { registerAndLogin } from "../helpers"

test.describe("profile posts tab", () => {
  test("posts created by user appear on their profile page", async ({ page }) => {
    const { email, username } = await registerAndLogin(page, "pp")

    // Retrieve auth token to create posts via API.
    const loginResp = await page.request.post("/api/v1/auth/login", {
      data: { email, password: "password123" },
    })
    expect(loginResp.status()).toBe(200)
    const { access_token } = await loginResp.json()

    // Create two posts via API.
    const post1 = await page.request.post("/api/v1/posts", {
      headers: { Authorization: `Bearer ${access_token}` },
      data: { body: `First post ${Date.now()}` },
    })
    expect(post1.status()).toBe(201)
    const post1Body = (await post1.json()).body as string

    const post2 = await page.request.post("/api/v1/posts", {
      headers: { Authorization: `Bearer ${access_token}` },
      data: { body: `Second post ${Date.now()}` },
    })
    expect(post2.status()).toBe(201)
    const post2Body = (await post2.json()).body as string

    // Navigate to own profile via nav link.
    await page.getByRole("navigation").getByRole("link", { name: "Profile" }).click()
    await expect(page).toHaveURL(new RegExp(`/@${username}`))

    // Posts tab should be selected by default and show both posts.
    const postsRes = page.waitForResponse(
      (r) => r.url().includes(`/api/v1/users/${username}/posts`) && r.request().method() === "GET",
    )
    await expect(page.getByRole("tab", { name: "Posts" })).toHaveAttribute("aria-selected", "true")
    expect((await postsRes).status()).toBe(200)

    await expect(page.getByText(post1Body)).toBeVisible()
    await expect(page.getByText(post2Body)).toBeVisible()
  })

  test("profile shows 'No posts yet.' for user with no posts", async ({ page }) => {
    await registerAndLogin(page, "noposts")

    // Navigate to own profile.
    await page.getByRole("navigation").getByRole("link", { name: "Profile" }).click()

    await expect(page.getByText("No posts yet.")).toBeVisible()
  })
})
