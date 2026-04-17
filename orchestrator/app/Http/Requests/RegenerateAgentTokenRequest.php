<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;

class RegenerateAgentTokenRequest extends FormRequest
{
    public function authorize(): bool
    {
        return $this->user()->can('regenerateToken', $this->route('agent'));
    }

    public function rules(): array
    {
        return [];
    }
}
