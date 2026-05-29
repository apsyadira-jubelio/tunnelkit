import { createFileRoute } from '@tanstack/react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Plus, Copy, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export const Route = createFileRoute('/api-keys')({
  component: APIKeys,
})

function APIKeys() {
  const queryClient = useQueryClient()
  const [showCreate, setShowCreate] = useState(false)
  const [newKey, setNewKey] = useState<any>(null)

  const { data: keys, isLoading } = useQuery({
    queryKey: ['api-keys'],
    queryFn: () => fetch('/api/v1/api-keys').then(r => r.json()),
  })

  const createMutation = useMutation({
    mutationFn: (data: { name: string; scopes: string[] }) =>
      fetch('/api/v1/api-keys', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      }).then(r => r.json()),
    onSuccess: (data) => {
      setNewKey(data)
      queryClient.invalidateQueries({ queryKey: ['api-keys'] })
    },
  })

  const revokeMutation = useMutation({
    mutationFn: (id: string) =>
      fetch(`/api/v1/api-keys/${id}`, { method: 'DELETE' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['api-keys'] }),
  })

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">API Keys</h1>
        <Button onClick={() => setShowCreate(true)}>
          <Plus className="mr-2 h-4 w-4" /> Create Key
        </Button>
      </div>

      {newKey && (
        <Card className="border-green-200 bg-green-50">
          <CardHeader>
            <CardTitle className="text-green-800">✓ API Key Created</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-green-700 mb-2">
              Copy this key now. It will not be shown again.
            </p>
            <div className="flex items-center gap-2">
              <code className="p-2 bg-white rounded flex-1">{newKey.plain_key}</code>
              <Button size="sm" onClick={() => copyToClipboard(newKey.plain_key)}>
                <Copy className="h-4 w-4" />
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Your API Keys</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-8">Loading...</div>
          ) : keys?.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No API keys yet. Create one for CLI access.
            </div>
          ) : (
            <div className="space-y-3">
              {keys?.map((key: any) => (
                <div key={key.id} className="border rounded-lg p-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="font-medium">{key.name}</p>
                      <p className="text-sm text-muted-foreground font-mono">
                        {key.key_prefix}...
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      {key.scopes.map((scope: string) => (
                        <Badge key={scope} variant="outline">{scope}</Badge>
                      ))}
                      <Button
                        size="sm"
                        variant="destructive"
                        onClick={() => revokeMutation.mutate(key.id)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                  <div className="mt-2 text-xs text-muted-foreground">
                    Created: {new Date(key.created_at).toLocaleString()}
                    {key.last_used && ` • Last used: ${new Date(key.last_used).toLocaleString()}`}
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
