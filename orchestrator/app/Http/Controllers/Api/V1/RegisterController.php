<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Models\Agent;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Http\Response;

class RegisterController extends Controller
{
    public function __invoke(Request $request): JsonResponse|Response
    {
        /** @var Agent $agent */
        $agent = $request->attributes->get('agent');

        Agent::query()
            ->whereKey($agent->getKey())
            ->whereNull('registered_at')
            ->update(['registered_at' => now()]);

        return response()->noContent();
    }
}
