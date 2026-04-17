import Link from 'next/link';

export default function HomePage() {
  return (
    <main className="flex flex-1 flex-col items-center justify-center gap-6 text-center p-12">
      <h1 className="text-4xl font-bold tracking-tight">Scraper Fleet</h1>
      <p className="text-muted-foreground max-w-md text-lg">
        A distributed web scraping platform. The orchestrator manages your agents; each agent
        connects, registers, and sends periodic heartbeats.
      </p>
      <Link
        href="/docs"
        className="inline-flex items-center justify-center rounded-md bg-primary px-6 py-2 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90 transition-colors"
      >
        Read the Docs
      </Link>
    </main>
  );
}
