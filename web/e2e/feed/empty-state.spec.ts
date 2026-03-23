import { test, expect } from "@playwright/test"
import { registerAndLogin } from "../helpers"

test.describe("feed empty state", () => {
  test("new user with no follows sees empty state message", async ({
    page,
  }) => {
    await registerAndLogin(page, "feed_empty")

    // The timeline should show the empty state message.
    await expect(
      page.getByText(/timeline is empty/i),
    ).toBeVisible({ timeout: 10000 })
  })
})
