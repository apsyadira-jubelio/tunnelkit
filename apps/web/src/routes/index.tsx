import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

export const Route = createFileRoute('/')({
  component: Dashboard,
})

function Dashboard() {
  const { data: tunnels, isLoading: tunnelsLoading } = useQuery({
    queryKey: ['tunnels'],
    queryFn: () => fetch('/api/v1/tunnels').then(r => r.json()),
  })

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Dashboard</h1>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Tunnels</CardTitle>
          </CardHeader>
          <CardContent>
            {tunnelsLoading ? (
              <Skeleton className="h-8 w-20" />
            ) : (
              <div className="text-2xl font-bold">
                {tunnels?.filter((t: any) => t.status === 'active').length || 0}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">0</div>
            <p className="text-xs text-muted-foreground">Metrics coming soon</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Error Rate</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">0%</div>
            <p className="text-xs text-muted-foreground">No errors</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Recent Tunnels</CardTitle>
        </CardHeader>
        <CardContent>
          {tunnelsLoading ? (
            <div className="space-y-2">
              <Skeleton className="h-12 w-full" />
              <Skeleton className="h-12 w-full" />
            </div>
          ) : tunnels?.length === 0 ? (
            <p className="text-muted-foreground">No tunnels yet. Create one to get started.</p>
          ) : (
            <div className="space-y-2">
              {tunnels?.slice(0, 5).map((tunnel: any) => (
                <div key={tunnel.id} className="flex items-center justify-between border rounded-lg p-3">
                  <div>
                    <p className="font-medium">{tunnel.name}</p>
                    <p className="text-sm text-muted-foreground">
                      {tunnel.subdomain && `${tunnel.subdomain}.tunnel.localhost`}
                    </p>
                  </div>
                  <span className={cn(
                    "px-2 py-1 rounded-full text-xs",
                    tunnel.status === 'active' ? "bg-green-100 text-green-800" : "bg-gray-100 text-gray-800"
                  )}>
                    {tunnel.status}
                  </span>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function cn(...classes: (string | undefined | false)[]) {
  return classes.filter(Boolean).join(' ')
}
