import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Field, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import DashboardLayout from '@/layouts/dashboard-layout';
import { Form } from '@inertiajs/react';
import { LoaderCircle } from 'lucide-react';

export default function AgentsCreate() {
  return (
    <div className="max-w-lg space-y-6">
      <div>
        <h1 className="text-xl font-semibold">Register Agent</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Register a new agent to run on a remote server and collect data.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Agent details</CardTitle>
          <CardDescription>Give your agent a name to identify it later.</CardDescription>
        </CardHeader>
        <CardContent>
          <Form action="/agents" method="post">
            {({ errors, processing }) => (
              <FieldGroup>
                <Field data-invalid={!!errors.name}>
                  <FieldLabel htmlFor="name">Name</FieldLabel>
                  <Input
                    id="name"
                    name="name"
                    type="text"
                    placeholder="e.g. Production Server 1"
                    required
                  />
                  <FieldError>{errors.name}</FieldError>
                </Field>

                <Button type="submit" className="w-full" disabled={processing}>
                  {processing && <LoaderCircle className="animate-spin" />}
                  Register Agent
                </Button>
              </FieldGroup>
            )}
          </Form>
        </CardContent>
      </Card>
    </div>
  );
}

AgentsCreate.layout = (page: React.ReactNode) => <DashboardLayout>{page}</DashboardLayout>;
