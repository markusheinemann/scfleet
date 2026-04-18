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

const DEFAULT_SCHEMA = JSON.stringify(
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

export default function TargetsCreate() {
  const [schema, setSchema] = useState(DEFAULT_SCHEMA);

  return (
    <div className="max-w-3xl space-y-6">
      <div>
        <h1 className="text-xl font-semibold">New Target</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Define a website to scrape and the data to extract.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Target details</CardTitle>
          <CardDescription>Configure what to scrape.</CardDescription>
        </CardHeader>
        <CardContent>
          <Form action="/targets" method="post">
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

                <Field data-invalid={!!errors.url}>
                  <FieldLabel htmlFor="url">URL</FieldLabel>
                  <Input
                    id="url"
                    name="url"
                    type="url"
                    placeholder="https://example.com/page"
                    required
                  />
                  <FieldError>{errors.url}</FieldError>
                </Field>

                <Field data-invalid={!!errors.schema}>
                  <div className="flex items-center justify-between">
                    <FieldLabel>Extraction Schema</FieldLabel>
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
                    <SchemaEditor value={schema} onChange={setSchema} invalid={!!errors.schema} />
                  </Suspense>
                  <FieldError>{errors.schema}</FieldError>
                </Field>

                <Button type="submit" className="w-full" disabled={processing}>
                  {processing && <LoaderCircle className="animate-spin" />}
                  Create Target
                </Button>
              </FieldGroup>
            )}
          </Form>
        </CardContent>
      </Card>
    </div>
  );
}

TargetsCreate.layout = (page: React.ReactNode) => <DashboardLayout>{page}</DashboardLayout>;
