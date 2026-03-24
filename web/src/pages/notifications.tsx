import { useEffect } from "react"
import { NotificationList } from "@/modules/notifications/components/notification-list"
import { useNotificationBadge } from "@/modules/notifications/hooks/use-notification-badge"

export function NotificationsPage() {
  const { reset } = useNotificationBadge()

  useEffect(() => {
    reset()
  }, [reset])

  return (
    <div className="mx-auto max-w-2xl">
      <div className="border-b px-4 py-3">
        <h1 className="text-xl font-bold">Notifications</h1>
      </div>
      <NotificationList />
    </div>
  )
}
