<?php

namespace App\Http\Controllers;

use App\Models\ScrapeJob;
use App\Models\Template;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;

class TemplateJobController extends Controller
{
    public function store(Request $request, Template $template): RedirectResponse
    {
        $this->authorize('view', $template);

        $validated = $request->validate([
            'url' => ['required', 'string', 'url', 'max:2048'],
        ]);

        ScrapeJob::create([
            'template_id' => $template->id,
            'url' => $validated['url'],
            'template' => $template->template,
            'status' => 'pending',
        ]);

        return redirect()->route('templates.show', $template)
            ->with('flash', 'Scrape job queued.');
    }
}
