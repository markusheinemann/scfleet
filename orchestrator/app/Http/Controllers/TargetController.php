<?php

namespace App\Http\Controllers;

use App\Http\Requests\StoreTargetRequest;
use App\Http\Requests\UpdateTargetRequest;
use App\Models\Target;
use Illuminate\Http\RedirectResponse;
use Inertia\Inertia;
use Inertia\Response;

class TargetController extends Controller
{
    public function index(): Response
    {
        $this->authorize('viewAny', Target::class);

        return Inertia::render('targets/index', [
            'targets' => Target::query()
                ->select(['id', 'title', 'url', 'created_at'])
                ->latest()
                ->get(),
        ]);
    }

    public function create(): Response
    {
        $this->authorize('create', Target::class);

        return Inertia::render('targets/create');
    }

    public function store(StoreTargetRequest $request): RedirectResponse
    {
        Target::create([
            ...$request->safe()->except('schema'),
            'schema' => json_decode($request->validated('schema'), true),
            'user_id' => $request->user()->id,
        ]);

        return redirect()->route('targets.index');
    }

    public function edit(Target $target): Response
    {
        $this->authorize('update', $target);

        return Inertia::render('targets/edit', [
            'target' => $target,
        ]);
    }

    public function update(UpdateTargetRequest $request, Target $target): RedirectResponse
    {
        $target->update([
            ...$request->safe()->except('schema'),
            'schema' => json_decode($request->validated('schema'), true),
        ]);

        return redirect()->route('targets.index');
    }
}
