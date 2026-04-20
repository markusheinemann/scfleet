<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Http\Requests\CompleteJobRequest;
use App\Models\Agent;
use App\Models\ScrapeJob;
use Illuminate\Http\Response;

class JobCompleteController extends Controller
{
    public function __invoke(CompleteJobRequest $request, ScrapeJob $scrapeJob): Response
    {
        /** @var Agent $agent */
        $agent = $request->attributes->get('agent');

        if ($scrapeJob->agent_id !== $agent->id) {
            abort(403, 'This job belongs to a different agent.');
        }

        $scrapeJob->forceFill([
            'status' => 'completed',
            'result' => $request->validated('result'),
            'field_errors' => $request->validated('field_errors', []),
            'completed_at' => now(),
        ])->save();

        $agent->forceFill(['status' => 'ready'])->save();

        return response()->noContent();
    }
}
