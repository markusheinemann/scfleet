<?php

namespace App\Http\Controllers;

use App\Http\Requests\StoreTemplateRequest;
use App\Http\Requests\UpdateTemplateRequest;
use App\Models\Template;
use Illuminate\Http\RedirectResponse;
use Inertia\Inertia;
use Inertia\Response;

class TemplateController extends Controller
{
    public function index(): Response
    {
        $this->authorize('viewAny', Template::class);

        return Inertia::render('templates/index', [
            'templates' => Template::query()
                ->select(['id', 'title', 'created_at'])
                ->latest()
                ->get(),
        ]);
    }

    public function create(): Response
    {
        $this->authorize('create', Template::class);

        return Inertia::render('templates/create');
    }

    public function store(StoreTemplateRequest $request): RedirectResponse
    {
        Template::create([
            'title' => $request->validated('title'),
            'template' => json_decode($request->validated('template'), true),
            'user_id' => $request->user()->id,
        ]);

        return redirect()->route('templates.index');
    }

    public function show(Template $template): Response
    {
        $this->authorize('view', $template);

        $jobs = $template->scrapeJobs()
            ->select(['ulid', 'template_id', 'status', 'error_type', 'error_message', 'result', 'field_errors', 'attempts', 'claimed_at', 'completed_at', 'created_at', 'has_artifacts'])
            ->latest()
            ->limit(50)
            ->get();

        $stats = $template->scrapeJobs()
            ->selectRaw('
                COUNT(*) as total,
                SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as completed,
                SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as failed,
                SUM(CASE WHEN status IN (?, ?) THEN 1 ELSE 0 END) as active,
                MAX(created_at) as last_run_at
            ', ['completed', 'failed', 'pending', 'processing'])
            ->first();

        return Inertia::render('templates/show', [
            'template' => $template->only(['id', 'title', 'template', 'created_at']),
            'jobs' => $jobs,
            'stats' => $stats,
        ]);
    }

    public function edit(Template $template): Response
    {
        $this->authorize('update', $template);

        return Inertia::render('templates/edit', [
            'template' => $template,
        ]);
    }

    public function update(UpdateTemplateRequest $request, Template $template): RedirectResponse
    {
        $template->update([
            'title' => $request->validated('title'),
            'template' => json_decode($request->validated('template'), true),
        ]);

        return redirect()->route('templates.index');
    }
}
