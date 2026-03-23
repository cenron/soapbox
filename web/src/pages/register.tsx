import { useState } from "react"
import { Link, Navigate, useNavigate } from "react-router"
import { useMutation } from "@tanstack/react-query"
import { postAuthRegisterMutation } from "@/shared/api/generated/@tanstack/react-query.gen"
import { useAuth } from "@/shared/auth/auth-context"
import { Button } from "@/shared/ui/button"
import { Input } from "@/shared/ui/input"
import { Label } from "@/shared/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card"

export function RegisterPage() {
  const auth = useAuth()
  const navigate = useNavigate()

  const [email, setEmail] = useState("")
  const [username, setUsername] = useState("")
  const [displayName, setDisplayName] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState<string | null>(null)

  const { mutate, isPending } = useMutation({
    ...postAuthRegisterMutation(),
    onSuccess(data) {
      if (!data?.access_token || !data?.user) {
        setError("Unexpected response from server.")
        return
      }
      auth.login(data.access_token, data.user)
      void navigate("/", { replace: true })
    },
    onError(err: { detail?: string; message?: string }) {
      setError(err?.detail ?? err?.message ?? "Registration failed. Please try again.")
    },
  })

  if (auth.isAuthenticated) {
    return <Navigate to="/" replace />
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    mutate({ body: { email, username, display_name: displayName, password } })
  }

  return (
    <div className="flex min-h-[60vh] items-center justify-center p-6">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle className="text-xl">Create your account</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                autoComplete="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="username">Username</Label>
              <Input
                id="username"
                type="text"
                autoComplete="username"
                required
                value={username}
                onChange={(e) => setUsername(e.target.value)}
              />
              <p className="text-xs text-muted-foreground">
                Letters, numbers, and underscores only
              </p>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="displayName">Display name</Label>
              <Input
                id="displayName"
                type="text"
                autoComplete="name"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                autoComplete="new-password"
                required
                minLength={8}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>

            {error && <p className="text-sm text-red-500">{error}</p>}

            <Button type="submit" className="w-full" disabled={isPending}>
              {isPending ? "Creating account..." : "Create account"}
            </Button>

            <p className="text-center text-sm text-muted-foreground">
              Already have an account?{" "}
              <Link to="/login" className="text-foreground underline underline-offset-4">
                Log in
              </Link>
            </p>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
