<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Models\ScrapeJob;
use Illuminate\Http\Request;
use Illuminate\Http\Response;
use Illuminate\Support\Facades\Storage;

class JobArtifactsController extends Controller
{
    public function __invoke(Request $request, ScrapeJob $scrapeJob): Response
    {
        $validated = $request->validate([
            'screenshot' => ['sometimes', 'nullable', 'string'],
            'html' => ['sometimes', 'nullable', 'string'],
        ]);

        $base = "job-artifacts/{$scrapeJob->ulid}";

        if (! empty($validated['screenshot'])) {
            Storage::put("{$base}/screenshot.png", base64_decode($validated['screenshot']));
        }

        if (! empty($validated['html'])) {
            Storage::put("{$base}/page.html", $validated['html']);
        }

        $scrapeJob->forceFill(['has_artifacts' => true])->save();

        return response()->noContent();
    }
}
