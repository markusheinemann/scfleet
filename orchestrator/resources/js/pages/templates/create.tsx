import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Field, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import DashboardLayout from '@/layouts/dashboard-layout';
import { Form } from '@inertiajs/react';
import { LoaderCircle } from 'lucide-react';
import { lazy, Suspense, useState } from 'react';

const SchemaEditor = lazy(() =>
  typeof window !== 'undefined'
    ? import('@/components/schema-editor')
    : Promise.resolve({ default: () => null } as never)
);

const DEFAULT_TEMPLATE = JSON.stringify(
  {
    version: '1',
    fields: [
      {
        name: 'title',
        type: 'string',
        extractors: [{ strategy: 'css', selector: 'h1' }],
      },
    ],
  },
  null,
  2
);

export default function TemplatesCreate() {
  const [template, setTemplate] = useState(DEFAULT_TEMPLATE);

  return (
    <div className="max-w-3xl space-y-6">
      <div>
        <h1 className="text-xl font-semibold">New Template</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Define a reusable extraction template. The URL is provided when running a scrape.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Template details</CardTitle>
          <CardDescription>Configure what data to extract.</CardDescription>
        </CardHeader>
        <CardContent>
          <Form action="/templates" method="post">
            {({ errors, processing }) => (
              <FieldGroup>
                <Field data-invalid={!!errors.title}>
                  <FieldLabel htmlFor="title">Title</FieldLabel>
                  <Input
                    id="title"
                    name="title"
                    type="text"
                    placeholder="e.g. Product prices"
                    required
                  />
                  <FieldError>{errors.title}</FieldError>
                </Field>

                <Field data-invalid={!!errors.template}>
                  <div className="flex items-center justify-between">
                    <FieldLabel>Extraction Template</FieldLabel>
                    <a
                      href="https://markusheinemann.github.io/scfleet/docs/extraction-schema"
                      target="_blank"
                      rel="noreferrer"
                      className="text-xs text-muted-foreground underline-offset-4 hover:underline"
                    >
                      Schema reference ↗
                    </a>
                  </div>
                  <Suspense
                    fallback={<div className="bg-muted animate-pulse h-80 w-full rounded-md" />}
                  >
                    <SchemaEditor
                      value={template}
                      onChange={setTemplate}
                      invalid={!!errors.template}
                    />
                  </Suspense>
                  <input type="hidden" name="template" value={template} />
                  <FieldError>{errors.template}</FieldError>
                </Field>

                <Button type="submit" className="w-full" disabled={processing}>
                  {processing && <LoaderCircle className="animate-spin" />}
                  Create Template
                </Button>
              </FieldGroup>
            )}
          </Form>
        </CardContent>
      </Card>
    </div>
  );
}

TemplatesCreate.layout = (page: React.ReactNode) => <DashboardLayout>{page}</DashboardLayout>;
