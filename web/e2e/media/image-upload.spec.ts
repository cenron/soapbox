import { test, expect } from "@playwright/test"
import { loginAs, SEED } from "../helpers"

test.describe("Media: image upload", () => {
  test("upload avatar via settings page", async ({ page }) => {
    await loginAs(page, SEED.admin.email, SEED.admin.password)

    // Navigate to settings by clicking the nav link
    await page.getByRole("navigation").getByRole("link", { name: "Settings" }).click()
    await expect(page.getByText("Edit profile")).toBeVisible()

    // Verify the upload drop zone is present
    await expect(page.getByText(/drop an image here/i)).toBeVisible()

    // Create a test image file and attach it via the file input
    const buffer = Buffer.from(
      "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
      "base64",
    )

    const fileInput = page.locator('[data-testid="file-input"]')
    await fileInput.setInputFiles({
      name: "test-avatar.png",
      mimeType: "image/png",
      buffer,
    })

    // Verify preview appears
    await expect(page.getByAltText("Upload preview")).toBeVisible()

    // Verify upload and clear buttons appear
    await expect(page.getByRole("button", { name: /^upload$/i })).toBeVisible()
    await expect(page.getByRole("button", { name: /clear/i })).toBeVisible()

    // Click upload and wait for the API response
    const uploadUrlPromise = page.waitForResponse(
      (res) => res.url().includes("/media/upload-url") && res.status() === 201,
    )
    await page.getByRole("button", { name: /^upload$/i }).click()
    const uploadUrlRes = await uploadUrlPromise

    // Verify the API returned a presigned URL
    const urlBody = await uploadUrlRes.json()
    expect(urlBody.upload_url).toBeTruthy()
    expect(urlBody.id).toBeTruthy()
    expect(urlBody.file_key).toContain("uploads/")

    // Wait for confirm request
    const confirmPromise = page.waitForResponse(
      (res) => res.url().includes("/confirm") && res.status() === 200,
    )
    const confirmRes = await confirmPromise
    const confirmBody = await confirmRes.json()
    expect(confirmBody.status).toBe("ready")
    expect(confirmBody.url).toBeTruthy()

    // Verify upload complete message
    await expect(page.getByText("Upload complete")).toBeVisible()
  })

  test("rejects invalid file type", async ({ page }) => {
    await loginAs(page, SEED.admin.email, SEED.admin.password)

    await page.getByRole("navigation").getByRole("link", { name: "Settings" }).click()
    await expect(page.getByText("Edit profile")).toBeVisible()

    const fileInput = page.locator('[data-testid="file-input"]')
    await fileInput.setInputFiles({
      name: "document.pdf",
      mimeType: "application/pdf",
      buffer: Buffer.from("fake pdf"),
    })

    await expect(page.getByText(/not supported/i)).toBeVisible()
  })
})
