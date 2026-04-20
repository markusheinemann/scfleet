<?php

namespace App\Http\Requests;

use App\Models\Template;
use App\Rules\ValidExtractionSchema;
use Illuminate\Contracts\Validation\ValidationRule;
use Illuminate\Foundation\Http\FormRequest;

class StoreTemplateRequest extends FormRequest
{
    public function authorize(): bool
    {
        return $this->user()->can('create', Template::class);
    }

    /**
     * @return array<string, ValidationRule|array<mixed>|string>
     */
    public function rules(): array
    {
        return [
            'title' => ['required', 'string', 'max:255'],
            'template' => ['bail', 'required', 'json', new ValidExtractionSchema],
        ];
    }
}
