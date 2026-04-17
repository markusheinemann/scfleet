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
import DashboardLayout from '@/layouts/dashboard-layout';
import { Link, router } from '@inertiajs/react';
import { ArrowLeft, Copy, RefreshCw, Terminal, TriangleAlert } from 'lucide-react';
import { useState } from 'react';

type Agent = {
  id: number;
  name: string;
  last_heartbeat_at: string | null;
  created_at: string;
};

type Props = {
  agent: Agent;
  token: string | null;
  canRegenerate: boolean;
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

export default function AgentsShow({ agent, token, canRegenerate }: Props) {
  return (
    <div className="max-w-2xl space-y-6">
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon-sm" asChild>
          <Link href="/agents">
            <ArrowLeft />
            <span className="sr-only">Back to agents</span>
          </Link>
        </Button>
        <div>
          <h1 className="text-xl font-semibold">{agent.name}</h1>
          <p className="text-sm text-muted-foreground">
            Registered {new Date(agent.created_at).toLocaleDateString()}
          </p>
        </div>
      </div>

      {token && (
        <Card className="border-amber-500/50 bg-amber-50/50 dark:bg-amber-950/20">
          <CardHeader>
            <div className="flex items-center gap-2">
              <TriangleAlert className="size-4 text-amber-600 dark:text-amber-400" />
              <CardTitle className="text-amber-700 dark:text-amber-400">Save your token</CardTitle>
            </div>
            <CardDescription>
              This token is shown only once. Copy it now — you won't be able to see it again.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <CodeBlock value={token} />
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Terminal className="size-4 text-muted-foreground" />
            <CardTitle>Setup instructions</CardTitle>
          </div>
          <CardDescription>
            Run the following commands on the server where you want to deploy this agent.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-5">
          <div className="space-y-2">
            <p className="text-sm font-medium">1. Download the agent binary</p>
            <CodeBlock value="curl -fsSL https://releases.scraperfleet.io/agent/install.sh | sh" />
          </div>

          <div className="space-y-2">
            <p className="text-sm font-medium">2. Start the agent with your token</p>
            <CodeBlock value={`scraperfleet-agent --token=${token ?? '<YOUR_TOKEN>'}`} />
          </div>

          <div className="space-y-2">
            <p className="text-sm font-medium">3. (Optional) Install as a system service</p>
            <CodeBlock value={`scraperfleet-agent install --token=${token ?? '<YOUR_TOKEN>'}`} />
          </div>
        </CardContent>
      </Card>

      {canRegenerate && (
        <Card className="border-destructive/40">
          <CardHeader>
            <CardTitle className="text-base">Regenerate token</CardTitle>
            <CardDescription>
              Generate a new token for this agent. The current token will stop working immediately.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button variant="destructive">
                  <RefreshCw className="size-4" />
                  Regenerate token
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Regenerate token?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This will immediately invalidate the current token for{' '}
                    <strong>{agent.name}</strong>. Any running agent using the old token will
                    disconnect and need to be restarted with the new token.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction
                    className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                    onClick={() => router.post(`/agents/${agent.id}/regenerate-token`)}
                  >
                    Yes, regenerate
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

AgentsShow.layout = (page: React.ReactNode) => <DashboardLayout>{page}</DashboardLayout>;
