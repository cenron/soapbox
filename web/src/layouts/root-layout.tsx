import { Outlet } from "react-router"
import { NavBar } from "./nav-bar"
import { MobileNav } from "./sidebar"
import { Sidebar } from "./sidebar"

export function RootLayout() {
  return (
    <div className="min-h-screen bg-background">
      <div className="flex items-center">
        <MobileNav />
        <div className="flex-1">
          <NavBar />
        </div>
      </div>

      <div className="flex">
        <Sidebar />
        <main className="flex-1">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
