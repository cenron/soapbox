import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { NewPostsBanner } from "./new-posts-banner"

describe("NewPostsBanner", () => {
  it("renders nothing when count is 0", () => {
    const { container } = render(
      <NewPostsBanner count={0} onClick={vi.fn()} />,
    )
    expect(container.firstChild).toBeNull()
  })

  it("renders singular label for count 1", () => {
    render(<NewPostsBanner count={1} onClick={vi.fn()} />)
    expect(screen.getByText("Show 1 new post")).toBeInTheDocument()
  })

  it("renders plural label for count > 1", () => {
    render(<NewPostsBanner count={5} onClick={vi.fn()} />)
    expect(screen.getByText("Show 5 new posts")).toBeInTheDocument()
  })

  it("calls onClick when clicked", async () => {
    const user = userEvent.setup()
    const onClick = vi.fn()

    render(<NewPostsBanner count={3} onClick={onClick} />)

    await user.click(screen.getByRole("button"))
    expect(onClick).toHaveBeenCalledOnce()
  })
})
