import { createFileRoute } from '@tanstack/react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Plus, Trash2, ExternalLink } from 'lucide-react'
import { useState } from 'react'

export const Route = createFileRoute('/tunnels')({
  component: Tunnels,
})

function Tunnels() {
  const queryClient = useQueryClient()
  const [showCreate, setShowCreate] = useState(false)

  const { data: tunnels, isLoading } = useQuery({
    queryKey: ['tunnels'],
    queryFn: () => fetch('/api/v1/tunnels').then(r => r.json()),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) =>
      fetch(`/api/v1/tunnels/${id}`, { method: 'DELETE' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['tunnels'] }),
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Tunnels</h1>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="mr-2 h-4 w-4" /> New Tunnel
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Tunnels</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8">Loading...</div>
          ) : tunnels?.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No tunnels found. Create your first tunnel to get started.
            </div>
          ) : (
            <div className="space-y-3">
              {tunnels?.map((tunnel: any) => (
                <div key={tunnel.id} className="border rounded-lg p-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div>
                        <p className="font-medium">{tunnel.name}</p>
                        <p className="text-sm text-muted-foreground">
                          {tunnel.subdomain && (
                            <span className="font-mono">
                              {tunnel.subdomain}.tunnel.localhost
                            </span>
                          )}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge variant={tunnel.status === 'active' ? 'default' : 'secondary'}>
                        {tunnel.status}
                      </Badge>
                      <Badge variant="outline">{tunnel.protocol.toUpperCase()}</Badge>
                      {tunnel.subdomain && (
                        <Button size="sm" variant="ghost" asChild>
                          <a href={`http://${tunnel.subdomain}.tunnel.localhost`} target="_blank">
                            <ExternalLink className="h-4 w-4" />
                          </a>
                        </Button>
                      )}
                      <Button
                        size="sm"
                        variant="destructive"
                        onClick={() => deleteMutation.mutate(tunnel.id)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                  <div className="mt-2 text-xs text-muted-foreground">
                    Created: {new Date(tunnel.created_at).toLocaleString()}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
