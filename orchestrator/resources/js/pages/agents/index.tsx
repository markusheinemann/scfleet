import { Button } from '@/components/ui/button';
import DashboardLayout from '@/layouts/dashboard-layout';
import { Link } from '@inertiajs/react';
import { Bot, Plus } from 'lucide-react';

type Agent = {
  id: number;
  name: string;
  is_online: boolean;
  last_heartbeat_at: string | null;
  registered_at: string | null;
  created_at: string;
};

type Props = {
  agents: Agent[];
};

function OnlineBadge({ online }: { online: boolean }) {
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium ${
        online
          ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
          : 'bg-muted text-muted-foreground'
      }`}
    >
      <span className={`size-1.5 rounded-full ${online ? 'bg-green-500' : 'bg-muted-foreground/50'}`} />
      {online ? 'Online' : 'Offline'}
    </span>
  );
}

export default function AgentsIndex({ agents }: Props) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold">Agents</h1>
          <p className="text-sm text-muted-foreground">Manage your scraper agents.</p>
        </div>
        <Button asChild>
          <Link href="/agents/create">
            <Plus />
            Register Agent
          </Link>
        </Button>
      </div>

      {agents.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-dashed p-12 text-center">
          <Bot className="mb-3 size-10 text-muted-foreground" />
          <p className="font-medium">No agents registered</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Register your first agent to get started.
          </p>
          <Button asChild className="mt-4">
            <Link href="/agents/create">
              <Plus />
              Register Agent
            </Link>
          </Button>
        </div>
      ) : (
        <div className="overflow-hidden rounded-xl border">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="px-4 py-3 text-left font-medium">Name</th>
                <th className="px-4 py-3 text-left font-medium">Status</th>
                <th className="px-4 py-3 text-left font-medium">Last Heartbeat</th>
                <th className="px-4 py-3 text-left font-medium">Registered</th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody>
              {agents.map(agent => (
                <tr key={agent.id} className="border-b last:border-0 hover:bg-muted/30">
                  <td className="px-4 py-3 font-medium">{agent.name}</td>
                  <td className="px-4 py-3">
                    <OnlineBadge online={agent.is_online} />
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {agent.last_heartbeat_at
                      ? new Date(agent.last_heartbeat_at).toLocaleString()
                      : 'Never'}
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {agent.registered_at
                      ? new Date(agent.registered_at).toLocaleDateString()
                      : '—'}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <Button variant="ghost" size="sm" asChild>
                      <Link href={`/agents/${agent.id}`}>View</Link>
                    </Button>
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

AgentsIndex.layout = (page: React.ReactNode) => <DashboardLayout>{page}</DashboardLayout>;
