<?php

namespace App\Http\Controllers;

use App\Http\Requests\StoreApiKeyRequest;
use App\Models\ApiKey;
use Illuminate\Http\RedirectResponse;
use Illuminate\Support\Str;
use Inertia\Inertia;
use Inertia\Response;

class ApiKeyController extends Controller
{
    public function index(): Response
    {
        $this->authorize('viewAny', ApiKey::class);

        $apiKeys = ApiKey::where('user_id', request()->user()->id)
            ->latest()
            ->get(['id', 'name', 'last_used_at', 'created_at']);

        return Inertia::render('api-keys/index', [
            'apiKeys' => $apiKeys,
            'newKey' => session('new_api_key'),
        ]);
    }

    public function store(StoreApiKeyRequest $request): RedirectResponse
    {
        $plainKey = Str::random(64);

        ApiKey::create([
            'user_id' => $request->user()->id,
            'name' => $request->validated('name'),
            'key_hash' => hash('sha256', $plainKey),
        ]);

        return redirect()->route('api-keys.index')
            ->with('new_api_key', $plainKey);
    }

    public function destroy(ApiKey $apiKey): RedirectResponse
    {
        $this->authorize('delete', $apiKey);

        $apiKey->delete();

        return redirect()->route('api-keys.index');
    }
}
