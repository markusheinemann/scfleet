<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Models\Agent;
use App\Models\ScrapeJob;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Http\Response;
use Illuminate\Support\Facades\DB;

class JobClaimController extends Controller
{
    public function __invoke(Request $request): JsonResponse|Response
    {
        /** @var Agent $agent */
        $agent = $request->attributes->get('agent');

        $job = DB::transaction(function () use ($agent): ?ScrapeJob {
            $job = ScrapeJob::query()
                ->where('status', 'pending')
                ->whereColumn('attempts', '<', 'max_attempts')
                ->orderBy('created_at')
                ->lock('for update skip locked')
                ->first();

            if (! $job) {
                return null;
            }

            $pageTimeoutS = (int) data_get($job->template, 'page_timeout_s', 30);

            $job->forceFill([
                'status' => 'processing',
                'agent_id' => $agent->id,
                'claimed_at' => now(),
                'timeout_at' => now()->addSeconds($pageTimeoutS + 30),
                'attempts' => $job->attempts + 1,
            ])->save();

            $agent->forceFill(['status' => 'processing'])->save();

            return $job;
        });

        if (! $job) {
            return response()->noContent();
        }

        return response()->json([
            'job_id' => $job->ulid,
            'url' => $job->url,
            'template' => $job->template,
            'timeout_s' => (int) data_get($job->template, 'page_timeout_s', 30),
        ]);
    }
}
