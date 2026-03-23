import { PostComposer } from "@/modules/posts/components/post-composer"
import { useAuth } from "@/shared/auth/auth-context"

export function HomePage() {
  const { isAuthenticated } = useAuth()

  return (
    <div>
      {isAuthenticated && <PostComposer />}

      <div className="p-6 text-center text-sm text-muted-foreground">
        Your timeline will appear here once the feed module is built.
      </div>
    </div>
  )
}
