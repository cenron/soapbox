import { describe, it, expect, vi } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { ImageUpload } from "./image-upload"

vi.mock("@/shared/api/generated/@tanstack/react-query.gen", () => ({
  postMediaUploadUrlMutation: () => ({ mutationKey: ["upload-url"] }),
  postMediaByIdConfirmMutation: () => ({ mutationKey: ["confirm"] }),
}))

vi.mock("@tanstack/react-query", async () => {
  const actual = await vi.importActual("@tanstack/react-query")
  return {
    ...actual,
    useMutation: () => ({
      mutate: vi.fn(),
      mutateAsync: vi.fn(),
      isPending: false,
    }),
  }
})

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>,
  )
}

describe("ImageUpload", () => {
  it("renders the drop zone", () => {
    renderWithProviders(<ImageUpload onUploadComplete={vi.fn()} />)

    expect(screen.getByText(/drop an image here/i)).toBeDefined()
  })

  it("shows accepted file types", () => {
    renderWithProviders(<ImageUpload onUploadComplete={vi.fn()} />)

    expect(screen.getByText(/jpeg, png, gif, webp/i)).toBeDefined()
  })

  it("shows max size info", () => {
    renderWithProviders(
      <ImageUpload onUploadComplete={vi.fn()} maxSizeMB={5} />,
    )

    expect(screen.getByText(/max 5 MB/i)).toBeDefined()
  })

  it("shows error for invalid file type", () => {
    renderWithProviders(<ImageUpload onUploadComplete={vi.fn()} />)

    const input = document.querySelector(
      '[data-testid="file-input"]',
    ) as HTMLInputElement

    const invalidFile = new File(["content"], "doc.pdf", {
      type: "application/pdf",
    })

    fireEvent.change(input, { target: { files: [invalidFile] } })

    expect(screen.getByText(/not supported/i)).toBeDefined()
  })

  it("shows error for oversized file", () => {
    renderWithProviders(
      <ImageUpload onUploadComplete={vi.fn()} maxSizeMB={0.001} />,
    )

    const input = document.querySelector(
      '[data-testid="file-input"]',
    ) as HTMLInputElement

    const bigFile = new File(["x".repeat(2000)], "big.jpg", {
      type: "image/jpeg",
    })

    fireEvent.change(input, { target: { files: [bigFile] } })

    expect(screen.getByText(/too large/i)).toBeDefined()
  })

  it("shows preview and upload button after valid file selection", () => {
    const objectUrl = "blob:http://localhost/test"
    vi.spyOn(URL, "createObjectURL").mockReturnValue(objectUrl)

    renderWithProviders(<ImageUpload onUploadComplete={vi.fn()} />)

    const input = document.querySelector(
      '[data-testid="file-input"]',
    ) as HTMLInputElement

    const file = new File(["img"], "photo.jpg", { type: "image/jpeg" })
    fireEvent.change(input, { target: { files: [file] } })

    expect(screen.getByAltText("Upload preview")).toBeDefined()
    expect(screen.getByRole("button", { name: /^upload$/i })).toBeDefined()
  })

  it("shows clear button after file selection", () => {
    vi.spyOn(URL, "createObjectURL").mockReturnValue("blob:test")

    renderWithProviders(<ImageUpload onUploadComplete={vi.fn()} />)

    const input = document.querySelector(
      '[data-testid="file-input"]',
    ) as HTMLInputElement

    const file = new File(["img"], "photo.jpg", { type: "image/jpeg" })
    fireEvent.change(input, { target: { files: [file] } })

    expect(screen.getByRole("button", { name: /clear/i })).toBeDefined()
  })
})
