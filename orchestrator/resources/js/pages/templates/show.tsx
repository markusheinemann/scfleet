import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Field, FieldError } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import DashboardLayout from '@/layouts/dashboard-layout';
import { Form, Link, usePage, usePoll } from '@inertiajs/react';
import {
  ArrowLeft,
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  Clock,
  Code2,
  Copy,
  LoaderCircle,
  Pencil,
  Play,
  XCircle,
} from 'lucide-react';
import { useState } from 'react';

type ScrapeJob = {
  ulid: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  error_type: string | null;
  error_message: string | null;
  result: Record<string, unknown> | null;
  field_errors: Record<string, string> | null;
  attempts: number;
  claimed_at: string | null;
  completed_at: string | null;
  created_at: string;
  has_artifacts: boolean;
};

type Stats = {
  total: number;
  completed: number;
  failed: number;
  active: number;
  last_run_at: string | null;
};

type Template = {
  id: number;
  title: string;
  template: Record<string, unknown>;
  created_at: string;
};

type Props = {
  template: Template;
  jobs: ScrapeJob[];
  stats: Stats;
  appUrl: string;
};

const STATUS_STYLES = {
  pending: 'bg-muted text-muted-foreground',
  processing: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
  completed: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
  failed: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
};

function StatusBadge({ status }: { status: ScrapeJob['status'] }) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_STYLES[status]}`}
    >
      {status}
    </span>
  );
}

function duration(job: ScrapeJob): string {
  if (!job.claimed_at || !job.completed_at) return '—';
  const ms = new Date(job.completed_at).getTime() - new Date(job.claimed_at).getTime();
  return ms < 1000 ? `${ms}ms` : `${(ms / 1000).toFixed(1)}s`;
}

function JobRow({ job }: { job: ScrapeJob }) {
  const [expanded, setExpanded] = useState(false);
  const hasDetail = job.result || job.error_message || job.field_errors || job.has_artifacts;

  return (
    <>
      <tr
        className={`border-b last:border-0 ${hasDetail ? 'cursor-pointer hover:bg-muted/30' : ''}`}
        onClick={() => hasDetail && setExpanded(e => !e)}
      >
        <td className="px-4 py-3">
          <div className="flex items-center gap-1.5">
            {hasDetail ? (
              expanded ? (
                <ChevronDown className="size-3.5 shrink-0 text-muted-foreground" />
              ) : (
                <ChevronRight className="size-3.5 shrink-0 text-muted-foreground" />
              )
            ) : (
              <span className="size-3.5 shrink-0" />
            )}
            <StatusBadge status={job.status} />
          </div>
        </td>
        <td className="px-4 py-3 text-sm text-muted-foreground">
          {new Date(job.created_at).toLocaleString()}
        </td>
        <td className="px-4 py-3 text-sm text-muted-foreground">{duration(job)}</td>
        <td className="px-4 py-3 text-sm">
          {job.status === 'failed' && (
            <span className="text-red-600 dark:text-red-400">{job.error_type ?? 'error'}</span>
          )}
          {job.status === 'completed' && job.result && (
            <span className="truncate text-muted-foreground">
              {Object.entries(job.result)
                .slice(0, 2)
                .map(([k, v]) => `${k}: ${String(v).slice(0, 40)}`)
                .join(' · ')}
            </span>
          )}
          {job.status === 'processing' && (
            <span className="text-blue-600 dark:text-blue-400">attempt {job.attempts}</span>
          )}
        </td>
      </tr>
      {expanded && hasDetail && (
        <tr className="border-b bg-muted/20 last:border-0">
          <td colSpan={4} className="px-6 py-4">
            <div className="space-y-3">
              {job.has_artifacts && (
                <div>
                  <p className="mb-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                    Page capture
                  </p>
                  <div className="flex gap-2">
                    <a
                      href={`/scrape-jobs/${job.ulid}/html`}
                      target="_blank"
                      rel="noreferrer"
                      className="inline-flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-xs font-medium hover:bg-muted"
                    >
                      Rendered HTML
                    </a>
                    <a
                      href={`/scrape-jobs/${job.ulid}/html?plain=1`}
                      target="_blank"
                      rel="noreferrer"
                      className="inline-flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-xs font-medium hover:bg-muted"
                    >
                      Page source
                    </a>
                  </div>
                  <a
                    href={`/scrape-jobs/${job.ulid}/screenshot`}
                    target="_blank"
                    rel="noreferrer"
                    className="mt-3 block"
                  >
                    <img
                      src={`/scrape-jobs/${job.ulid}/screenshot`}
                      alt="Page screenshot"
                      className="max-h-96 rounded-md border object-top shadow-sm"
                    />
                  </a>
                </div>
              )}
              {job.result && (
                <div>
                  <p className="mb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                    Result
                  </p>
                  <pre className="overflow-x-auto rounded-md bg-muted p-3 text-xs">
                    {JSON.stringify(job.result, null, 2)}
                  </pre>
                </div>
              )}
              {job.error_message && (
                <div>
                  <p className="mb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                    Error
                  </p>
                  <p className="rounded-md bg-red-50 p-3 text-xs text-red-700 dark:bg-red-950/30 dark:text-red-400">
                    <span className="font-medium">{job.error_type}: </span>
                    {job.error_message}
                  </p>
                </div>
              )}
              {job.field_errors && Object.keys(job.field_errors).length > 0 && (
                <div>
                  <p className="mb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                    Field errors
                  </p>
                  <ul className="space-y-1">
                    {Object.entries(job.field_errors).map(([field, msg]) => (
                      <li key={field} className="text-xs text-muted-foreground">
                        <span className="font-mono font-medium text-foreground">{field}</span>:{' '}
                        {msg}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          </td>
        </tr>
      )}
    </>
  );
}

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);

  const copy = () => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <button
      onClick={copy}
      title={copied ? 'Copied!' : 'Copy to clipboard'}
      className="shrink-0 text-muted-foreground transition-colors hover:text-foreground"
    >
      <Copy className="size-3.5" />
      <span className="sr-only">{copied ? 'Copied' : 'Copy'}</span>
    </button>
  );
}

function CodeSnippet({ code }: { code: string }) {
  return (
    <div className="relative rounded-md bg-muted">
      <div className="absolute right-3 top-3">
        <CopyButton text={code} />
      </div>
      <pre className="overflow-x-auto p-3 pr-10 font-mono text-xs leading-relaxed">{code}</pre>
    </div>
  );
}

function ApiUsage({ template, appUrl }: { template: Template; appUrl: string }) {
  const [open, setOpen] = useState(false);

  const submitSnippet = `curl -X POST ${appUrl}/api/v1/scrape \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{"url": "https://example.com", "template_id": ${template.id}}'`;

  const pollSnippet = `curl ${appUrl}/api/v1/scrape/{job_id} \\
  -H "Authorization: Bearer YOUR_API_KEY"`;

  return (
    <Card>
      <CardHeader className="cursor-pointer select-none" onClick={() => setOpen(v => !v)}>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Code2 className="size-4 text-muted-foreground" />
            <CardTitle>API usage</CardTitle>
          </div>
          {open ? (
            <ChevronDown className="size-4 text-muted-foreground" />
          ) : (
            <ChevronRight className="size-4 text-muted-foreground" />
          )}
        </div>
      </CardHeader>
      {open && (
        <CardContent className="space-y-5">
          <p className="text-sm text-muted-foreground">
            Use the REST API to submit scrape jobs programmatically. You'll need an{' '}
            <Link href="/api-keys" className="underline underline-offset-4 hover:text-foreground">
              API key
            </Link>{' '}
            to authenticate.
          </p>

          <div className="space-y-2">
            <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              1. Submit a scrape job
            </p>
            <CodeSnippet code={submitSnippet} />
            <p className="text-xs text-muted-foreground">
              Returns <code className="font-mono">202</code> with a{' '}
              <code className="font-mono">job_id</code> you can use to poll for results.
            </p>
          </div>

          <div className="space-y-2">
            <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              2. Poll for results
            </p>
            <CodeSnippet code={pollSnippet} />
            <p className="text-xs text-muted-foreground">
              Check <code className="font-mono">status</code> until it's{' '}
              <code className="font-mono">completed</code> or{' '}
              <code className="font-mono">failed</code>. Results are in the{' '}
              <code className="font-mono">result</code> field.
            </p>
          </div>
        </CardContent>
      )}
    </Card>
  );
}

export default function TemplatesShow({ template, jobs, stats, appUrl }: Props) {
  const { props } = usePage<{ flash?: string }>();
  const [showRunForm, setShowRunForm] = useState(false);

  usePoll(1000, { only: ['jobs', 'stats'] });

  const successRate =
    stats.total > 0 ? Math.round(((stats.completed ?? 0) / stats.total) * 100) : null;
  const failureRate =
    stats.total > 0 ? Math.round(((stats.failed ?? 0) / stats.total) * 100) : null;

  return (
    <div className="space-y-6">
      <div className="flex items-start gap-3">
        <Button variant="ghost" size="icon-sm" asChild>
          <Link href="/templates">
            <ArrowLeft />
            <span className="sr-only">Back to templates</span>
          </Link>
        </Button>
        <div className="flex-1">
          <h1 className="text-xl font-semibold">{template.title}</h1>
          <p className="text-sm text-muted-foreground">Extraction template</p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" asChild>
            <Link href={`/templates/${template.id}/edit`}>
              <Pencil />
              Edit
            </Link>
          </Button>
          <Button size="sm" onClick={() => setShowRunForm(v => !v)}>
            <Play />
            Run Scrape
          </Button>
        </div>
      </div>

      {showRunForm && (
        <Card className="border-primary/30">
          <CardContent className="pt-5">
            <Form
              action={`/templates/${template.id}/jobs`}
              method="post"
              onSuccess={() => setShowRunForm(false)}
            >
              {({ errors, processing }) => (
                <div className="flex gap-3">
                  <Field className="flex-1" data-invalid={!!errors.url}>
                    <Input
                      name="url"
                      type="url"
                      placeholder="https://example.com/page-to-scrape"
                      autoFocus
                      required
                    />
                    <FieldError>{errors.url}</FieldError>
                  </Field>
                  <Button type="submit" disabled={processing}>
                    {processing ? <LoaderCircle className="animate-spin" /> : <Play />}
                    Queue job
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    onClick={() => setShowRunForm(false)}
                    disabled={processing}
                  >
                    Cancel
                  </Button>
                </div>
              )}
            </Form>
          </CardContent>
        </Card>
      )}

      {props.flash && (
        <div className="flex items-center gap-2 rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-700 dark:border-green-800 dark:bg-green-950/30 dark:text-green-400">
          <CheckCircle2 className="size-4 shrink-0" />
          {props.flash}
        </div>
      )}

      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        <Card>
          <CardContent className="pt-5">
            <p className="text-sm text-muted-foreground">Total runs</p>
            <p className="mt-1 text-2xl font-semibold">{stats.total ?? 0}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5">
            <p className="text-sm text-muted-foreground">Success rate</p>
            <p
              className={`mt-1 text-2xl font-semibold ${
                successRate === null
                  ? ''
                  : successRate >= 90
                    ? 'text-green-600 dark:text-green-400'
                    : successRate >= 70
                      ? 'text-yellow-600 dark:text-yellow-400'
                      : 'text-red-600 dark:text-red-400'
              }`}
            >
              {successRate !== null ? `${successRate}%` : '—'}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5">
            <p className="text-sm text-muted-foreground">Failure rate</p>
            <p
              className={`mt-1 text-2xl font-semibold ${
                failureRate === null
                  ? ''
                  : failureRate === 0
                    ? 'text-green-600 dark:text-green-400'
                    : failureRate <= 10
                      ? 'text-yellow-600 dark:text-yellow-400'
                      : 'text-red-600 dark:text-red-400'
              }`}
            >
              {failureRate !== null ? `${failureRate}%` : '—'}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5">
            <p className="text-sm text-muted-foreground">Last run</p>
            <p className="mt-1 text-sm font-medium">
              {stats.last_run_at ? new Date(stats.last_run_at).toLocaleString() : '—'}
            </p>
          </CardContent>
        </Card>
      </div>

      <ApiUsage template={template} appUrl={appUrl} />

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Job log</CardTitle>
            {stats.active > 0 && (
              <span className="flex items-center gap-1.5 text-sm text-blue-600 dark:text-blue-400">
                <Clock className="size-3.5" />
                {stats.active} active
              </span>
            )}
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {jobs.length === 0 ? (
            <div className="flex flex-col items-center justify-center px-4 py-12 text-center">
              <XCircle className="mb-3 size-8 text-muted-foreground" />
              <p className="font-medium">No jobs yet</p>
              <p className="mt-1 text-sm text-muted-foreground">
                Click <strong>Run Scrape</strong> to submit the first job.
              </p>
            </div>
          ) : (
            <div className="overflow-hidden rounded-b-xl">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-muted/50">
                    <th className="px-4 py-3 text-left font-medium">Status</th>
                    <th className="px-4 py-3 text-left font-medium">Submitted</th>
                    <th className="px-4 py-3 text-left font-medium">Duration</th>
                    <th className="px-4 py-3 text-left font-medium">Summary</th>
                  </tr>
                </thead>
                <tbody>
                  {jobs.map(job => (
                    <JobRow key={job.ulid} job={job} />
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

TemplatesShow.layout = (page: React.ReactNode) => <DashboardLayout>{page}</DashboardLayout>;
