<?php

namespace App\Http\Requests;

use App\Rules\ValidExtractionSchema;
use Illuminate\Contracts\Validation\ValidationRule;
use Illuminate\Foundation\Http\FormRequest;

class StoreScrapeJobRequest extends FormRequest
{
    public function authorize(): bool
    {
        return true;
    }

    /**
     * @return array<string, ValidationRule|array<mixed>|string>
     */
    public function rules(): array
    {
        return [
            'url' => ['required', 'string', 'url', 'max:2048'],
            'template' => ['required', 'array', ValidExtractionSchema::forDecodedInput()],
        ];
    }
}
