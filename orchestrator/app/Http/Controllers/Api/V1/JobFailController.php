<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Http\Requests\FailJobRequest;
use App\Models\Agent;
use App\Models\ScrapeJob;
use Illuminate\Http\Response;

class JobFailController extends Controller
{
    public function __invoke(FailJobRequest $request, ScrapeJob $scrapeJob): Response
    {
        /** @var Agent $agent */
        $agent = $request->attributes->get('agent');

        if ($scrapeJob->agent_id !== $agent->id) {
            abort(403, 'This job belongs to a different agent.');
        }

        if ($scrapeJob->attempts < $scrapeJob->max_attempts) {
            $scrapeJob->forceFill([
                'status' => 'pending',
                'agent_id' => null,
                'claimed_at' => null,
                'timeout_at' => null,
            ])->save();
        } else {
            $scrapeJob->forceFill([
                'status' => 'failed',
                'error_type' => $request->validated('error_type'),
                'error_message' => $request->validated('error_message'),
                'completed_at' => now(),
            ])->save();
        }

        $agent->forceFill(['status' => 'ready'])->save();

        return response()->noContent();
    }
}
