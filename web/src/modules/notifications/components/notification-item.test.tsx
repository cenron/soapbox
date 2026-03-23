import { describe, it, expect, vi } from "vitest"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { MemoryRouter } from "react-router"
import { NotificationItem } from "./notification-item"
import type { NotificationsNotificationResponse } from "@/shared/api/generated/types.gen"

function renderWithRouter(ui: React.ReactElement) {
  return render(<MemoryRouter>{ui}</MemoryRouter>)
}

function makeNotification(
  overrides: Partial<NotificationsNotificationResponse> = {},
): NotificationsNotificationResponse {
  return {
    id: "notif-1",
    type: "like",
    actor_id: "actor-1",
    actor_username: "alice",
    actor_display_name: "Alice",
    actor_avatar_url: "",
    post_id: "post-1",
    read: false,
    created_at: new Date().toISOString(),
    ...overrides,
  }
}

describe("NotificationItem", () => {
  it("renders like notification text", () => {
    renderWithRouter(
      <NotificationItem notification={makeNotification()} onMarkRead={vi.fn()} />,
    )
    expect(screen.getByText("Alice")).toBeInTheDocument()
    expect(screen.getByText("liked your post")).toBeInTheDocument()
  })

  it("renders repost notification text", () => {
    renderWithRouter(
      <NotificationItem
        notification={makeNotification({ type: "repost" })}
        onMarkRead={vi.fn()}
      />,
    )
    expect(screen.getByText("reposted your post")).toBeInTheDocument()
  })

  it("renders reply notification text", () => {
    renderWithRouter(
      <NotificationItem
        notification={makeNotification({ type: "reply" })}
        onMarkRead={vi.fn()}
      />,
    )
    expect(screen.getByText("replied to your post")).toBeInTheDocument()
  })

  it("renders follow notification text", () => {
    renderWithRouter(
      <NotificationItem
        notification={makeNotification({ type: "follow", post_id: undefined })}
        onMarkRead={vi.fn()}
      />,
    )
    expect(screen.getByText("followed you")).toBeInTheDocument()
  })

  it("shows unread indicator for unread notifications", () => {
    const { container } = renderWithRouter(
      <NotificationItem notification={makeNotification({ read: false })} onMarkRead={vi.fn()} />,
    )
    expect(container.querySelector(".bg-blue-500")).toBeInTheDocument()
  })

  it("hides unread indicator for read notifications", () => {
    const { container } = renderWithRouter(
      <NotificationItem notification={makeNotification({ read: true })} onMarkRead={vi.fn()} />,
    )
    expect(container.querySelector(".bg-blue-500")).not.toBeInTheDocument()
  })

  it("calls onMarkRead when clicking an unread notification", async () => {
    const user = userEvent.setup()
    const onMarkRead = vi.fn()

    renderWithRouter(
      <NotificationItem
        notification={makeNotification({ read: false })}
        onMarkRead={onMarkRead}
      />,
    )

    await user.click(screen.getByRole("link"))
    expect(onMarkRead).toHaveBeenCalledWith("notif-1")
  })

  it("does not call onMarkRead when clicking a read notification", async () => {
    const user = userEvent.setup()
    const onMarkRead = vi.fn()

    renderWithRouter(
      <NotificationItem
        notification={makeNotification({ read: true })}
        onMarkRead={onMarkRead}
      />,
    )

    await user.click(screen.getByRole("link"))
    expect(onMarkRead).not.toHaveBeenCalled()
  })
})
