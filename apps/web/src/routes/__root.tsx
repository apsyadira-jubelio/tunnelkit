import {
  createRootRoute,
  Link,
  Outlet,
} from '@tanstack/react-router'

export const Route = createRootRoute({
  component: RootLayout,
})

function RootLayout() {
  return (
    <div className="min-h-screen bg-background">
      <nav className="border-b">
        <div className="container mx-auto px-4">
          <div className="flex h-16 items-center justify-between">
            <div className="flex items-center gap-6">
              <Link to="/" className="text-xl font-bold">
                TunnelKit
              </Link>
              <div className="flex gap-4">
                <Link
                  to="/"
                  className="text-sm font-medium text-muted-foreground hover:text-foreground"
                  activeProps={{ className: 'text-foreground' }}
                >
                  Dashboard
                </Link>
                <Link
                  to="/tunnels"
                  className="text-sm font-medium text-muted-foreground hover:text-foreground"
                  activeProps={{ className: 'text-foreground' }}
                >
                  Tunnels
                </Link>
                <Link
                  to="/api-keys"
                  className="text-sm font-medium text-muted-foreground hover:text-foreground"
                  activeProps={{ className: 'text-foreground' }}
                >
                  API Keys
                </Link>
              </div>
            </div>
            <div className="text-sm text-muted-foreground">
              v0.1.0
            </div>
          </div>
        </div>
      </nav>
      <main className="container mx-auto px-4 py-6">
        <Outlet />
      </main>
    </div>
  )
}
