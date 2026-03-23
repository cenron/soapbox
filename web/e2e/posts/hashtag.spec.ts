import { test, expect } from "@playwright/test"
import { registerAndLogin } from "../helpers"

test.describe("hashtag rendering and navigation", () => {
  test("hashtag in post body renders as a link and navigates to search page", async ({
    page,
  }) => {
    const { email } = await registerAndLogin(page, "ht")

    // Get auth token.
    const loginResp = await page.request.post("/api/v1/auth/login", {
      data: { email, password: "password123" },
    })
    expect(loginResp.status()).toBe(200)
    const { access_token } = await loginResp.json()

    // Create post with a hashtag via API.
    const createResp = await page.request.post("/api/v1/posts", {
      headers: { Authorization: `Bearer ${access_token}` },
      data: { body: `Loving #soapbox today ${Date.now()}` },
    })
    expect(createResp.status()).toBe(201)
    const post = await createResp.json()
    const postId: string = post.id

    // Navigate to post detail.
    await page.goto(`/post/${postId}`)
    await page.waitForLoadState("networkidle")

    // The hashtag is rendered as a clickable link.
    const hashtagLink = page.getByRole("link", { name: "#soapbox" })
    await expect(hashtagLink).toBeVisible()

    // Click the hashtag — it should navigate to the search page with the correct query.
    await hashtagLink.click()

    await expect(page).toHaveURL("/search?q=%23soapbox")
  })
})
