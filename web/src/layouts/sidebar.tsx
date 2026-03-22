import { Link, useLocation } from "react-router"
import { useAuth } from "@/shared/auth/auth-context"
import { useMediaQuery } from "@/shared/hooks/use-media-query"
import { Button } from "@/shared/ui/button"
import { Sheet, SheetContent, SheetTrigger } from "@/shared/ui/sheet"
import { Separator } from "@/shared/ui/separator"
import { cn } from "@/shared/lib/utils"

interface NavItem {
  label: string
  to: string
  authOnly?: boolean
}

function useNavItems(): NavItem[] {
  const { user } = useAuth()

  const items: NavItem[] = [
    { label: "Home", to: "/" },
    { label: "Search", to: "/search" },
    { label: "Notifications", to: "/notifications", authOnly: true },
    { label: "Profile", to: user ? `/@${user.username}` : "/login", authOnly: true },
    { label: "Settings", to: "/settings", authOnly: true },
  ]

  return items.filter((item) => !item.authOnly || user)
}

function NavLinks({ onClick }: { onClick?: () => void }) {
  const location = useLocation()
  const { user } = useAuth()
  const navItems = useNavItems()

  return (
    <nav className="flex flex-col gap-1">
      {navItems.map((item) => (
        <Button
          key={item.to}
          variant="ghost"
          className={cn("justify-start", location.pathname === item.to && "bg-muted")}
          asChild
        >
          <Link to={item.to} onClick={onClick}>
            {item.label}
          </Link>
        </Button>
      ))}

      {user?.role === "admin" && (
        <>
          <Separator className="my-2" />
          <Button
            variant="ghost"
            className={cn("justify-start", location.pathname === "/admin" && "bg-muted")}
            asChild
          >
            <Link to="/admin" onClick={onClick}>
              Admin
            </Link>
          </Button>
        </>
      )}
    </nav>
  )
}

export function Sidebar() {
  const isDesktop = useMediaQuery("(min-width: 768px)")

  if (isDesktop) {
    return (
      <aside className="sticky top-14 hidden h-[calc(100vh-3.5rem)] w-56 shrink-0 border-r p-4 md:block">
        <NavLinks />
      </aside>
    )
  }

  return null
}

export function MobileNav() {
  const isDesktop = useMediaQuery("(min-width: 768px)")

  if (isDesktop) return null

  return (
    <Sheet>
      <SheetTrigger asChild>
        <Button variant="ghost" size="icon" className="md:hidden">
          <span className="sr-only">Menu</span>
          <MenuIcon />
        </Button>
      </SheetTrigger>
      <SheetContent side="left" className="w-64 p-4">
        <NavLinks />
      </SheetContent>
    </Sheet>
  )
}

function MenuIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <line x1="4" x2="20" y1="12" y2="12" />
      <line x1="4" x2="20" y1="6" y2="6" />
      <line x1="4" x2="20" y1="18" y2="18" />
    </svg>
  )
}
