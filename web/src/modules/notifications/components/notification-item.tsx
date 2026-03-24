import { Link } from "react-router"
import { Heart, Repeat2, MessageCircle, UserPlus } from "lucide-react"
import type { NotificationsNotificationResponse } from "@/shared/api/generated/types.gen"
import { cn } from "@/shared/lib/utils"

interface NotificationItemProps {
  notification: NotificationsNotificationResponse
  onMarkRead: (id: string) => void
}

const iconMap: Record<string, typeof Heart> = {
  like: Heart,
  repost: Repeat2,
  reply: MessageCircle,
  follow: UserPlus,
}

const colorMap: Record<string, string> = {
  like: "text-red-500",
  repost: "text-green-500",
  reply: "text-blue-500",
  follow: "text-purple-500",
}

function getActionText(type: string): string {
  switch (type) {
    case "like":
      return "liked your post"
    case "repost":
      return "reposted your post"
    case "reply":
      return "replied to your post"
    case "follow":
      return "followed you"
    default:
      return "interacted with you"
  }
}

function getHref(notification: NotificationsNotificationResponse): string {
  if (notification.type === "follow") {
    return `/@${notification.actor_username}`
  }
  if (notification.post_id) {
    return `/post/${notification.post_id}`
  }
  return "#"
}

function formatTime(dateStr: string | undefined): string {
  if (!dateStr) return ""

  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMin = Math.floor(diffMs / 60000)
  const diffHr = Math.floor(diffMs / 3600000)
  const diffDay = Math.floor(diffMs / 86400000)

  if (diffMin < 1) return "now"
  if (diffMin < 60) return `${diffMin}m`
  if (diffHr < 24) return `${diffHr}h`
  if (diffDay < 7) return `${diffDay}d`

  return date.toLocaleDateString()
}

export function NotificationItem({ notification, onMarkRead }: NotificationItemProps) {
  const type = notification.type ?? "like"
  const Icon = iconMap[type] ?? Heart
  const actorName = notification.actor_display_name || notification.actor_username || "Someone"

  function handleClick() {
    if (!notification.read && notification.id) {
      onMarkRead(notification.id)
    }
  }

  return (
    <Link
      to={getHref(notification)}
      onClick={handleClick}
      className={cn(
        "flex items-start gap-3 border-b px-4 py-3 transition-colors hover:bg-muted/50",
        !notification.read && "bg-muted/30",
      )}
    >
      <div className={cn("mt-0.5 shrink-0", colorMap[type])}>
        <Icon className="h-5 w-5" />
      </div>

      <div className="min-w-0 flex-1">
        <p className="text-sm">
          <span className="font-semibold">{actorName}</span>
          {" "}
          {getActionText(type)}
        </p>
        <p className="mt-0.5 text-xs text-muted-foreground">
          {formatTime(notification.created_at)}
        </p>
      </div>

      {!notification.read && (
        <div className="mt-2 h-2 w-2 shrink-0 rounded-full bg-blue-500" />
      )}
    </Link>
  )
}
