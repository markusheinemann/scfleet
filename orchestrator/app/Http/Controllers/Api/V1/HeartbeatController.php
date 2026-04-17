<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Models\Agent;
use Illuminate\Http\Request;
use Illuminate\Http\Response;

class HeartbeatController extends Controller
{
    public function __invoke(Request $request): Response
    {
        /** @var Agent $agent */
        $agent = $request->attributes->get('agent');

        $agent->forceFill(['last_heartbeat_at' => now()])->save();

        return response()->noContent();
    }
}
