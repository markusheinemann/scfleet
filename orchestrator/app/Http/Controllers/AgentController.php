<?php

namespace App\Http\Controllers;

use App\Http\Requests\RegenerateAgentTokenRequest;
use App\Http\Requests\StoreAgentRequest;
use App\Models\Agent;
use Illuminate\Http\RedirectResponse;
use Illuminate\Support\Str;
use Inertia\Inertia;
use Inertia\Response;

class AgentController extends Controller
{
    public function index(): Response
    {
        $this->authorize('viewAny', Agent::class);

        return Inertia::render('agents/index', [
            'agents' => Agent::latest()->get(),
        ]);
    }

    public function create(): Response
    {
        $this->authorize('create', Agent::class);

        return Inertia::render('agents/create');
    }

    public function store(StoreAgentRequest $request): RedirectResponse
    {
        $plainToken = Str::random(64);

        $agent = Agent::create([
            'user_id' => $request->user()->id,
            'name' => $request->validated('name'),
            'token' => hash('sha256', $plainToken),
        ]);

        return redirect()->route('agents.show', $agent)
            ->with("token_{$agent->id}", $plainToken);
    }

    public function show(Agent $agent): Response
    {
        $this->authorize('view', $agent);

        return Inertia::render('agents/show', [
            'agent' => $agent,
            'token' => session("token_{$agent->id}"),
            'canRegenerate' => request()->user()->can('regenerateToken', $agent),
        ]);
    }

    public function regenerateToken(RegenerateAgentTokenRequest $request, Agent $agent): RedirectResponse
    {
        $plainToken = Str::random(64);

        $agent->update(['token' => hash('sha256', $plainToken)]);

        return redirect()->route('agents.show', $agent)
            ->with("token_{$agent->id}", $plainToken);
    }
}
