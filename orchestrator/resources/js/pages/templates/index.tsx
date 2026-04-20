import { Button } from '@/components/ui/button';
import DashboardLayout from '@/layouts/dashboard-layout';
import { Link } from '@inertiajs/react';
import { Pencil, Plus, Target } from 'lucide-react';

type TemplateItem = {
  id: number;
  title: string;
  created_at: string;
};

type Props = {
  templates: TemplateItem[];
};

export default function TemplatesIndex({ templates }: Props) {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold">Templates</h1>
          <p className="text-sm text-muted-foreground">Manage your scraping templates.</p>
        </div>
        <Button asChild>
          <Link href="/templates/create">
            <Plus />
            New Template
          </Link>
        </Button>
      </div>

      {templates.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-dashed p-12 text-center">
          <Target className="mb-3 size-10 text-muted-foreground" />
          <p className="font-medium">No templates defined</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Create your first scraping template to get started.
          </p>
          <Button asChild className="mt-4">
            <Link href="/templates/create">
              <Plus />
              New Template
            </Link>
          </Button>
        </div>
      ) : (
        <div className="overflow-hidden rounded-xl border">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="px-4 py-3 text-left font-medium">Title</th>
                <th className="px-4 py-3 text-left font-medium">Created</th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody>
              {templates.map(template => (
                <tr
                  key={template.id}
                  className="border-b last:border-0 hover:bg-muted/30 cursor-pointer"
                  onClick={() => (window.location.href = `/templates/${template.id}`)}
                >
                  <td className="px-4 py-3 font-medium">{template.title}</td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {new Date(template.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3 text-right">
                    <Button variant="ghost" size="sm" asChild onClick={e => e.stopPropagation()}>
                      <Link href={`/templates/${template.id}/edit`}>
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

TemplatesIndex.layout = (page: React.ReactNode) => <DashboardLayout>{page}</DashboardLayout>;
