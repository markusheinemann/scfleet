<?php

namespace App\Http\Middleware;

use App\Models\ApiKey;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class ApiKeyMiddleware
{
    public function handle(Request $request, Closure $next): Response
    {
        $key = $request->bearerToken();

        if (! $key) {
            return response()->json(['message' => 'Missing bearer token.'], Response::HTTP_UNAUTHORIZED);
        }

        $apiKey = ApiKey::where('key_hash', hash('sha256', $key))->first();

        if (! $apiKey) {
            return response()->json(['message' => 'Invalid bearer token.'], Response::HTTP_UNAUTHORIZED);
        }

        $apiKey->forceFill(['last_used_at' => now()])->save();
        $request->attributes->set('api_key', $apiKey);

        return $next($request);
    }
}
