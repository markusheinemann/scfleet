import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Field, FieldError, FieldGroup, FieldLabel, FieldSet } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Form } from '@inertiajs/react';
import { LoaderCircle } from 'lucide-react';

export default function Setup() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 p-4">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>Create admin account</CardTitle>
          <CardDescription>
            No users exist yet. Set up the first admin account to get started.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form action="/setup" method="post">
            {({ errors, processing }) => (
              <FieldGroup>
                <FieldSet>
                  <FieldGroup>
                    <Field data-invalid={!!errors.username}>
                      <FieldLabel htmlFor="username">Username</FieldLabel>
                      <Input
                        id="username"
                        name="username"
                        type="text"
                        autoComplete="username"
                        required
                      />
                      <FieldError>{errors.username}</FieldError>
                    </Field>

                    <Field data-invalid={!!errors.email}>
                      <FieldLabel htmlFor="email">Email</FieldLabel>
                      <Input id="email" name="email" type="email" autoComplete="email" required />
                      <FieldError>{errors.email}</FieldError>
                    </Field>

                    <Field data-invalid={!!errors.password}>
                      <FieldLabel htmlFor="password">Password</FieldLabel>
                      <Input
                        id="password"
                        name="password"
                        type="password"
                        autoComplete="new-password"
                        required
                      />
                      <FieldError>{errors.password}</FieldError>
                    </Field>

                    <Field>
                      <FieldLabel htmlFor="password_confirmation">Confirm password</FieldLabel>
                      <Input
                        id="password_confirmation"
                        name="password_confirmation"
                        type="password"
                        autoComplete="new-password"
                        required
                      />
                    </Field>
                  </FieldGroup>
                </FieldSet>

                <Field>
                  <Button type="submit" className="w-full" disabled={processing}>
                    {processing && <LoaderCircle className="animate-spin" />}
                    Create admin account
                  </Button>
                </Field>
              </FieldGroup>
            )}
          </Form>
        </CardContent>
      </Card>
    </div>
  );
}
