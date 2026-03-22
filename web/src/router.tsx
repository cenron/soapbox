import { createBrowserRouter } from "react-router"
import { RootLayout } from "@/layouts/root-layout"
import { ProtectedRoute } from "@/shared/auth/protected-route"
import { HomePage } from "@/pages/home"
import { LoginPage } from "@/pages/login"
import { RegisterPage } from "@/pages/register"
import { ProfilePage } from "@/pages/profile"
import { PostDetailPage } from "@/pages/post-detail"
import { SearchPage } from "@/pages/search"
import { NotificationsPage } from "@/pages/notifications"
import { SettingsPage } from "@/pages/settings"
import { AdminPage } from "@/pages/admin"
import { NotFoundPage } from "@/pages/not-found"

export const router = createBrowserRouter([
  {
    element: <RootLayout />,
    children: [
      // Public routes
      { path: "/login", element: <LoginPage /> },
      { path: "/register", element: <RegisterPage /> },
      { path: "/search", element: <SearchPage /> },
      { path: "/@:username", element: <ProfilePage /> },
      { path: "/post/:id", element: <PostDetailPage /> },

      // Protected routes
      {
        element: <ProtectedRoute />,
        children: [
          { path: "/", element: <HomePage /> },
          { path: "/notifications", element: <NotificationsPage /> },
          { path: "/settings", element: <SettingsPage /> },
          { path: "/admin", element: <AdminPage /> },
        ],
      },

      // Catch-all
      { path: "*", element: <NotFoundPage /> },
    ],
  },
])
