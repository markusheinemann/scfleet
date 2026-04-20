<?php

namespace App\Http\Controllers;

use App\Models\ScrapeJob;
use Illuminate\Http\Response;
use Illuminate\Support\Facades\Storage;

class ScrapeJobArtifactController extends Controller
{
    public function screenshot(ScrapeJob $scrapeJob): Response
    {
        $path = "job-artifacts/{$scrapeJob->ulid}/screenshot.png";

        abort_unless(Storage::exists($path), 404);

        return response(Storage::get($path), 200, ['Content-Type' => 'image/png']);
    }

    public function html(ScrapeJob $scrapeJob): Response
    {
        $path = "job-artifacts/{$scrapeJob->ulid}/page.html";

        abort_unless(Storage::exists($path), 404);

        $contentType = request()->boolean('plain')
            ? 'text/plain; charset=utf-8'
            : 'text/html; charset=utf-8';

        return response(Storage::get($path), 200, ['Content-Type' => $contentType]);
    }
}
