<?php

namespace App\Http\Middleware;

use App\Models\Agent;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class AgentTokenMiddleware
{
    public function handle(Request $request, Closure $next): Response
    {
        $token = $request->bearerToken();

        if (! $token) {
            return response()->json(['message' => 'Missing bearer token.'], Response::HTTP_UNAUTHORIZED);
        }

        $agent = Agent::where('token', hash('sha256', $token))->first();

        if (! $agent) {
            return response()->json(['message' => 'Invalid bearer token.'], Response::HTTP_UNAUTHORIZED);
        }

        $request->attributes->set('agent', $agent);

        return $next($request);
    }
}
