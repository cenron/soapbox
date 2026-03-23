import { describe, it, expect } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"
import { ImageDisplay } from "./image-display"

describe("ImageDisplay", () => {
  it("renders an image with the correct src", () => {
    render(<ImageDisplay src="http://localhost:9000/soapbox/uploads/test.jpg" alt="test" />)

    const img = screen.getByAltText("test")
    expect(img).toBeDefined()
    expect(img.getAttribute("src")).toBe("http://localhost:9000/soapbox/uploads/test.jpg")
  })

  it("has lazy loading attribute", () => {
    render(<ImageDisplay src="http://example.com/img.jpg" alt="photo" />)

    const img = screen.getByAltText("photo")
    expect(img.getAttribute("loading")).toBe("lazy")
  })

  it("shows loading skeleton initially", () => {
    const { container } = render(<ImageDisplay src="http://example.com/img.jpg" alt="photo" />)

    const skeleton = container.querySelector(".animate-pulse")
    expect(skeleton).toBeTruthy()
  })

  it("hides skeleton after image loads", () => {
    const { container } = render(<ImageDisplay src="http://example.com/img.jpg" alt="photo" />)

    const img = screen.getByAltText("photo")
    fireEvent.load(img)

    const skeleton = container.querySelector(".animate-pulse")
    expect(skeleton).toBeFalsy()
  })

  it("shows error message on image error", () => {
    render(<ImageDisplay src="http://example.com/broken.jpg" alt="broken" />)

    const img = screen.getByAltText("broken")
    fireEvent.error(img)

    expect(screen.getByText(/failed to load image/i)).toBeDefined()
  })
})
