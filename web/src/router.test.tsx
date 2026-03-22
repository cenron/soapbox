import { describe, it, expect } from "vitest"
import { matchRoutes } from "react-router"

/**
 * Route matching tests verify that React Router resolves URLs to the correct
 * page components. These exist because React Router v7 does NOT match literal
 * special characters (like @) before dynamic segments — a bug that shipped
 * undetected when E2E tests bypassed client-side routing via page.goto().
 */

const routes = [
  {
    children: [
      { path: "/login", id: "login" },
      { path: "/register", id: "register" },
      { path: "/search", id: "search" },
      { path: "/:username", id: "profile" },
      { path: "/post/:id", id: "post" },
      { path: "/", id: "home" },
      { path: "/notifications", id: "notifications" },
      { path: "/settings", id: "settings" },
      { path: "/admin", id: "admin" },
      { path: "*", id: "not-found" },
    ],
  },
]

function matchLeaf(url: string): string | undefined {
  const matches = matchRoutes(routes, url)
  const route = matches?.[matches.length - 1]?.route as { id?: string } | undefined
  return route?.id
}

function matchParams(url: string): Record<string, string | undefined> | undefined {
  const matches = matchRoutes(routes, url)
  return matches?.[matches.length - 1]?.params
}

describe("route matching", () => {
  it("resolves /@username to the profile route", () => {
    expect(matchLeaf("/@alice")).toBe("profile")
    expect(matchLeaf("/@journeytest")).toBe("profile")
    expect(matchLeaf("/@admin")).toBe("profile")
  })

  it("captures the full @username as the param", () => {
    expect(matchParams("/@alice")).toEqual({ username: "@alice" })
    expect(matchParams("/@journeytest")).toEqual({ username: "@journeytest" })
  })

  it("static routes take priority over /:username", () => {
    expect(matchLeaf("/login")).toBe("login")
    expect(matchLeaf("/register")).toBe("register")
    expect(matchLeaf("/search")).toBe("search")
    expect(matchLeaf("/notifications")).toBe("notifications")
    expect(matchLeaf("/settings")).toBe("settings")
    expect(matchLeaf("/admin")).toBe("admin")
  })

  it("resolves /post/:id correctly", () => {
    expect(matchLeaf("/post/123")).toBe("post")
    expect(matchParams("/post/abc-def")).toEqual({ id: "abc-def" })
  })

  it("resolves / to home", () => {
    expect(matchLeaf("/")).toBe("home")
  })

  it("resolves unknown nested paths to not-found", () => {
    expect(matchLeaf("/not/a/real/page")).toBe("not-found")
  })

  it("@ prefix in param must be stripped to get the API username", () => {
    // The profile page does rawUsername.replace(/^@/, "") to extract the username.
    // This test documents that contract: the param includes @ and must be stripped.
    const params = matchParams("/@alice")
    const rawUsername = params?.username ?? ""
    const username = rawUsername.replace(/^@/, "")
    expect(username).toBe("alice")
  })
})
