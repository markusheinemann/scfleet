import { Button } from '@/components/ui/button';
import DashboardLayout from '@/layouts/dashboard-layout';
import { Link } from '@inertiajs/react';
import { Pencil, Plus, Target } from 'lucide-react';

type TargetItem = {
  id: number;
  title: string;
  url: string;
  created_at: string;
};

type Props = {
  targets: TargetItem[];
};

export default function TargetsIndex({ targets }: Props) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold">Targets</h1>
          <p className="text-sm text-muted-foreground">Manage your scraping targets.</p>
        </div>
        <Button asChild>
          <Link href="/targets/create">
            <Plus />
            New Target
          </Link>
        </Button>
      </div>

      {targets.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-dashed p-12 text-center">
          <Target className="mb-3 size-10 text-muted-foreground" />
          <p className="font-medium">No targets defined</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Create your first scraping target to get started.
          </p>
          <Button asChild className="mt-4">
            <Link href="/targets/create">
              <Plus />
              New Target
            </Link>
          </Button>
        </div>
      ) : (
        <div className="overflow-hidden rounded-xl border">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="px-4 py-3 text-left font-medium">Title</th>
                <th className="px-4 py-3 text-left font-medium">URL</th>
                <th className="px-4 py-3 text-left font-medium">Created</th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody>
              {targets.map(target => (
                <tr key={target.id} className="border-b last:border-0 hover:bg-muted/30">
                  <td className="px-4 py-3 font-medium">{target.title}</td>
                  <td className="max-w-xs truncate px-4 py-3 text-muted-foreground">
                    <a
                      href={target.url}
                      target="_blank"
                      rel="noreferrer"
                      className="hover:underline"
                    >
                      {target.url}
                    </a>
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {new Date(target.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <Button variant="ghost" size="sm" asChild>
                      <Link href={`/targets/${target.id}/edit`}>
                        <Pencil />
                        Edit
                      </Link>
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

TargetsIndex.layout = (page: React.ReactNode) => <DashboardLayout>{page}</DashboardLayout>;
