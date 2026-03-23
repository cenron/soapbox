import { useMutation, useQueryClient } from "@tanstack/react-query"
import {
  putNotificationsByIdReadMutation,
  putNotificationsReadAllMutation,
  getNotificationsInfiniteQueryKey,
} from "@/shared/api/generated/@tanstack/react-query.gen"

export function useMarkRead() {
  const queryClient = useQueryClient()

  const markRead = useMutation({
    ...putNotificationsByIdReadMutation(),
    onSettled: () => {
      void queryClient.invalidateQueries({
        queryKey: getNotificationsInfiniteQueryKey(),
      })
    },
  })

  const markAllRead = useMutation({
    ...putNotificationsReadAllMutation(),
    onSettled: () => {
      void queryClient.invalidateQueries({
        queryKey: getNotificationsInfiniteQueryKey(),
      })
    },
  })

  return { markRead, markAllRead }
}
