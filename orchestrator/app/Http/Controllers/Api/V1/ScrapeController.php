<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Http\Requests\StoreScrapeJobRequest;
use App\Models\ScrapeJob;
use App\Models\Template;
use Illuminate\Http\JsonResponse;

class ScrapeController extends Controller
{
    public function store(StoreScrapeJobRequest $request): JsonResponse
    {
        $template = Template::findOrFail($request->validated('template_id'));

        $job = ScrapeJob::create([
            'url' => $request->validated('url'),
            'template_id' => $template->id,
            'template' => $template->template,
            'status' => 'pending',
        ]);

        return response()->json([
            'job_id' => $job->ulid,
            'status' => $job->status,
            'created_at' => $job->created_at,
        ], 202);
    }

    public function show(ScrapeJob $scrapeJob): JsonResponse
    {
        return response()->json([
            'job_id' => $scrapeJob->ulid,
            'status' => $scrapeJob->status,
            'url' => $scrapeJob->url,
            'result' => $scrapeJob->result,
            'field_errors' => $scrapeJob->field_errors,
            'error_type' => $scrapeJob->error_type,
            'error_message' => $scrapeJob->error_message,
            'created_at' => $scrapeJob->created_at,
            'completed_at' => $scrapeJob->completed_at,
        ]);
    }
}
