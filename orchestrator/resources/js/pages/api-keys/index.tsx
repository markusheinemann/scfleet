import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import DashboardLayout from '@/layouts/dashboard-layout';
import { Form, router } from '@inertiajs/react';
import { Copy, KeyRound, Plus, Trash2, TriangleAlert } from 'lucide-react';
import { useState } from 'react';

type ApiKey = {
  id: number;
  name: string;
  last_used_at: string | null;
  created_at: string;
};

type Props = {
  apiKeys: ApiKey[];
  newKey: string | null;
};

function CodeBlock({ value }: { value: string }) {
  const [copied, setCopied] = useState(false);

  const copy = () => {
    navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="flex items-center gap-2 rounded-lg bg-muted px-3 py-2 font-mono text-sm">
      <span className="min-w-0 flex-1 overflow-x-auto">{value}</span>
      <button
        onClick={copy}
        title={copied ? 'Copied!' : 'Copy to clipboard'}
        className="shrink-0 text-muted-foreground transition-colors hover:text-foreground"
      >
        <Copy className="size-4" />
        <span className="sr-only">{copied ? 'Copied' : 'Copy'}</span>
      </button>
    </div>
  );
}

export default function ApiKeysIndex({ apiKeys, newKey }: Props) {
  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-xl font-semibold">API Keys</h1>
        <p className="text-sm text-muted-foreground">
          Manage keys used to authenticate requests to the scraping API.
        </p>
      </div>

      {newKey && (
        <Card className="border-amber-500/50 bg-amber-50/50 dark:bg-amber-950/20">
          <CardHeader>
            <div className="flex items-center gap-2">
              <TriangleAlert className="size-4 text-amber-600 dark:text-amber-400" />
              <CardTitle className="text-amber-700 dark:text-amber-400">
                Save your API key
              </CardTitle>
            </div>
            <CardDescription>
              This key is shown only once. Copy it now — you won't be able to see it again.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <CodeBlock value={newKey} />
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Plus className="size-4 text-muted-foreground" />
            <CardTitle>Create API key</CardTitle>
          </div>
          <CardDescription>Give your key a name to identify where it's used.</CardDescription>
        </CardHeader>
        <CardContent>
          <Form action="/api-keys" method="post" resetOnSuccess>
            {({ errors, processing }) => (
              <div className="flex gap-3">
                <div className="flex-1 space-y-1">
                  <Label htmlFor="name" className="sr-only">
                    Key name
                  </Label>
                  <Input
                    id="name"
                    name="name"
                    placeholder="e.g. Production, CI pipeline"
                    autoComplete="off"
                  />
                  {errors.name && <p className="text-sm text-destructive">{errors.name}</p>}
                </div>
                <Button type="submit" disabled={processing}>
                  <KeyRound className="size-4" />
                  Create key
                </Button>
              </div>
            )}
          </Form>
        </CardContent>
      </Card>

      {apiKeys.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-dashed p-12 text-center">
          <KeyRound className="mb-3 size-10 text-muted-foreground" />
          <p className="font-medium">No API keys yet</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Create a key above to start making API requests.
          </p>
        </div>
      ) : (
        <div className="overflow-hidden rounded-xl border">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="px-4 py-3 text-left font-medium">Name</th>
                <th className="px-4 py-3 text-left font-medium">Last used</th>
                <th className="px-4 py-3 text-left font-medium">Created</th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody>
              {apiKeys.map(apiKey => (
                <tr key={apiKey.id} className="border-b last:border-0 hover:bg-muted/30">
                  <td className="px-4 py-3 font-medium">{apiKey.name}</td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {apiKey.last_used_at ? new Date(apiKey.last_used_at).toLocaleString() : 'Never'}
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {new Date(apiKey.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <AlertDialog>
                      <AlertDialogTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          className="text-destructive hover:text-destructive"
                        >
                          <Trash2 className="size-4" />
                          <span className="sr-only">Delete key</span>
                        </Button>
                      </AlertDialogTrigger>
                      <AlertDialogContent>
                        <AlertDialogHeader>
                          <AlertDialogTitle>Delete API key?</AlertDialogTitle>
                          <AlertDialogDescription>
                            This will immediately revoke <strong>{apiKey.name}</strong>. Any
                            integrations using this key will stop working.
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel>Cancel</AlertDialogCancel>
                          <AlertDialogAction
                            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                            onClick={() => router.delete(`/api-keys/${apiKey.id}`)}
                          >
                            Delete
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

ApiKeysIndex.layout = (page: React.ReactNode) => <DashboardLayout>{page}</DashboardLayout>;
