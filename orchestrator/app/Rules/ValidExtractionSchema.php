<?php

namespace App\Rules;

use Closure;
use Illuminate\Contracts\Validation\ValidationRule;
use Illuminate\Translation\PotentiallyTranslatedString;
use Opis\JsonSchema\Errors\ErrorFormatter;
use Opis\JsonSchema\Validator;

class ValidExtractionSchema implements ValidationRule
{
    /**
     * @param  Closure(string, ?string=): PotentiallyTranslatedString  $fail
     */
    private static ?object $schema = null;

    private static ?Validator $validator = null;

    public function validate(string $attribute, mixed $value, Closure $fail): void
    {
        $data = json_decode($value);

        if (! is_object($data)) {
            $fail('The :attribute must be a valid JSON object.');

            return;
        }

        if (self::$schema === null) {
            // Load the canonical schema from the monorepo root (single source of truth).
            $base = self::readJson(base_path('../api/schemas/template.v1.json'), $fail);
            if ($base === null) {
                return;
            }

            // Merge opis $error annotations (PHP-only, not part of the public schema).
            $overlay = self::readJson(resource_path('schemas/template.v1.errors.json'), $fail);
            if ($overlay === null) {
                return;
            }

            self::$schema = self::mergeErrorOverlay($base, $overlay);
        }

        self::$validator ??= new Validator;

        $result = self::$validator->validate($data, self::$schema);

        if ($result->isValid()) {
            return;
        }

        // format() respects $error annotations in the schema and yields one message
        // per JSON pointer path. Pick the message at the deepest path — it is the
        // most specific and avoids intermediate "must match $ref" noise.
        $errors = (new ErrorFormatter)->format($result->error(), false);

        $message = null;
        $deepest = -1;

        foreach ($errors as $pointer => $msg) {
            $depth = count(array_filter(explode('/', $pointer)));

            if ($depth > $deepest) {
                $deepest = $depth;
                $message = $msg;
            }
        }

        $fail('The :attribute is not a valid extraction schema: '.($message ?? 'unknown error').'.');
    }

    private static function readJson(string $path, Closure $fail): ?object
    {
        $json = file_get_contents($path);

        if ($json === false) {
            $fail("Could not read schema file: {$path}.");

            return null;
        }

        $decoded = json_decode($json);

        if (json_last_error() !== JSON_ERROR_NONE || ! is_object($decoded)) {
            $fail("Schema file contains invalid JSON: {$path}.");

            return null;
        }

        return $decoded;
    }

    private static function mergeErrorOverlay(object $base, object $overlay): object
    {
        $merged = clone $base;

        foreach ($overlay as $key => $value) {
            if ($key === '$error') {
                $merged->{'$error'} = $value;
            } elseif (is_object($value) && isset($merged->{$key}) && is_object($merged->{$key})) {
                $merged->{$key} = self::mergeErrorOverlay($merged->{$key}, $value);
            }
        }

        return $merged;
    }
}
