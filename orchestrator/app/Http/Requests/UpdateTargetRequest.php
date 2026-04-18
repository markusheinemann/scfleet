<?php

namespace App\Http\Requests;

use App\Rules\ValidExtractionSchema;
use Illuminate\Contracts\Validation\ValidationRule;
use Illuminate\Foundation\Http\FormRequest;

class UpdateTargetRequest extends FormRequest
{
    public function authorize(): bool
    {
        return $this->user()->can('update', $this->route('target'));
    }

    /**
     * @return array<string, ValidationRule|array<mixed>|string>
     */
    public function rules(): array
    {
        return [
            'title' => ['required', 'string', 'max:255'],
            'url' => ['required', 'string', 'url', 'max:2048'],
            'schema' => ['bail', 'required', 'json', new ValidExtractionSchema],
        ];
    }
}
