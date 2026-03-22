import { Link } from "react-router"
import { Button } from "@/shared/ui/button"

export function NotFoundPage() {
  return (
    <div className="flex min-h-[60vh] flex-col items-center justify-center gap-4 p-6">
      <h1 className="text-4xl font-bold">404</h1>
      <p className="text-muted-foreground">This page doesn't exist.</p>
      <Button asChild variant="outline">
        <Link to="/">Go home</Link>
      </Button>
    </div>
  )
}
