<?php

namespace App\Http\Requests;

use Illuminate\Contracts\Validation\ValidationRule;
use Illuminate\Foundation\Http\FormRequest;
use Illuminate\Validation\Rule;

class FailJobRequest extends FormRequest
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
            'error_type' => [
                'required',
                'string',
                Rule::in(['missing_required_field', 'page_timeout', 'navigation_error', 'extraction_error']),
            ],
            'error_message' => ['required', 'string', 'max:1000'],
        ];
    }
}
