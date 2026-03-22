import { Link } from "react-router"
import { useAuth } from "@/shared/auth/auth-context"
import { Button } from "@/shared/ui/button"
import { Input } from "@/shared/ui/input"

export function NavBar() {
  const { isAuthenticated } = useAuth()

  return (
    <header className="sticky top-0 z-50 border-b bg-background">
      <div className="flex h-14 items-center gap-4 px-4">
        <Link to="/" className="text-lg font-bold tracking-tight">
          Soapbox
        </Link>

        <div className="hidden flex-1 sm:block">
          <Input placeholder="Search..." className="max-w-xs" />
        </div>

        <div className="ml-auto flex items-center gap-2">
          {isAuthenticated ? (
            <Button variant="ghost" size="sm" asChild>
              <Link to="/settings">Settings</Link>
            </Button>
          ) : (
            <>
              <Button variant="ghost" size="sm" asChild>
                <Link to="/login">Log in</Link>
              </Button>
              <Button size="sm" asChild>
                <Link to="/register">Sign up</Link>
              </Button>
            </>
          )}
        </div>
      </div>
    </header>
  )
}
