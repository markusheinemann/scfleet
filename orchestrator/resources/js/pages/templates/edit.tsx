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

type Template = {
  id: number;
  title: string;
  template: object;
};

type Props = {
  template: Template;
};

export default function TemplatesEdit({ template }: Props) {
  const [schema, setSchema] = useState(JSON.stringify(template.template, null, 2));

  return (
    <div className="max-w-3xl space-y-6">
      <div>
        <h1 className="text-xl font-semibold">Edit Template</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Update the extraction template configuration.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Template details</CardTitle>
          <CardDescription>Configure what data to extract.</CardDescription>
        </CardHeader>
        <CardContent>
          <Form action={`/templates/${template.id}`} method="put">
            {({ errors, processing }) => (
              <FieldGroup>
                <Field data-invalid={!!errors.title}>
                  <FieldLabel htmlFor="title">Title</FieldLabel>
                  <Input
                    id="title"
                    name="title"
                    type="text"
                    placeholder="e.g. Product prices"
                    defaultValue={template.title}
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
                    fallback={
                      <div className="bg-muted animate-pulse h-[320px] w-full rounded-md" />
                    }
                  >
                    <SchemaEditor
                      value={schema}
                      onChange={setSchema}
                      invalid={!!errors.template}
                    />
                  </Suspense>
                  <input type="hidden" name="template" value={schema} />
                  <FieldError>{errors.template}</FieldError>
                </Field>

                <Button type="submit" className="w-full" disabled={processing}>
                  {processing && <LoaderCircle className="animate-spin" />}
                  Save Changes
                </Button>
              </FieldGroup>
            )}
          </Form>
        </CardContent>
      </Card>
    </div>
  );
}

TemplatesEdit.layout = (page: React.ReactNode) => <DashboardLayout>{page}</DashboardLayout>;
