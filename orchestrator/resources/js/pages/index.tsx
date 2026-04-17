import DashboardLayout from '@/layouts/dashboard-layout';

export default function Index() {
  return <div className="text-muted-foreground text-sm">Welcome to Scraper Fleet.</div>;
}

Index.layout = (page: React.ReactNode) => <DashboardLayout>{page}</DashboardLayout>;
