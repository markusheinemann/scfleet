<?php

use App\Models\Agent;
use App\Models\User;
use Illuminate\Support\Str;

describe('agents index', function (): void {
    it('shows the agents page', function (): void {
        $user = User::factory()->create();

        $this->actingAs($user)
            ->get(route('agents.index'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page->component('agents/index'));
    });

    it('requires authentication', function (): void {
        $this->get(route('agents.index'))
            ->assertRedirect(route('login'));
    });

    it('lists all registered agents', function (): void {
        $user = User::factory()->create();
        Agent::factory(3)->create();

        $this->actingAs($user)
            ->get(route('agents.index'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page
                ->component('agents/index')
                ->has('agents', 3)
            );
    });
});

describe('agents create', function (): void {
    it('shows the create agent page', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->get(route('agents.create'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page->component('agents/create'));
    });

    it('requires authentication', function (): void {
        $this->get(route('agents.create'))
            ->assertRedirect(route('login'));
    });
});

describe('agents store', function (): void {
    it('creates a new agent', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)->post(route('agents.store'), [
            'name' => 'My Production Server',
        ]);

        expect(Agent::count())->toBe(1)
            ->and(Agent::first()->name)->toBe('My Production Server')
            ->and(Agent::first()->user_id)->toBe($user->id);
    });

    it('stores the token as a sha256 hash', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)->post(route('agents.store'), ['name' => 'My Agent']);

        $agent = Agent::withoutGlobalScopes()->first();
        expect(strlen($agent->getRawOriginal('token')))->toBe(64);
    });

    it('redirects to the agent show page after creation', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->post(route('agents.store'), ['name' => 'My Agent'])
            ->assertRedirect(route('agents.show', Agent::first()));
    });

    it('flashes the plain token to the session scoped to the agent', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->post(route('agents.store'), ['name' => 'My Agent'])
            ->assertSessionHas('token_'.Agent::first()->id);
    });

    it('requires a name', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->post(route('agents.store'), [])
            ->assertSessionHasErrors('name');
    });

    it('requires name to be a string', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->post(route('agents.store'), ['name' => ['array']])
            ->assertSessionHasErrors('name');
    });

    it('requires name within 255 characters', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->post(route('agents.store'), ['name' => Str::repeat('a', 256)])
            ->assertSessionHasErrors('name');
    });

    it('requires authentication', function (): void {
        $this->post(route('agents.store'), ['name' => 'My Agent'])
            ->assertRedirect(route('login'));
    });
});

describe('agents show', function (): void {
    it('shows the agent page with setup instructions', function (): void {
        $user = User::factory()->create();
        $agent = Agent::factory()->create();

        $this->actingAs($user)
            ->get(route('agents.show', $agent))
            ->assertOk()
            ->assertInertia(fn ($page) => $page
                ->component('agents/show')
                ->where('agent.id', $agent->id)
                ->where('agent.name', $agent->name)
            );
    });

    it('passes the plain token from session flash on first view', function (): void {
        $user = User::factory()->create();
        $agent = Agent::factory()->create();

        $this->actingAs($user)
            ->withSession(["token_{$agent->id}" => 'plain-token-value'])
            ->get(route('agents.show', $agent))
            ->assertOk()
            ->assertInertia(fn ($page) => $page
                ->where('token', 'plain-token-value')
            );
    });

    it('does not show a token flashed for a different agent', function (): void {
        $user = User::factory()->create();
        $agent = Agent::factory()->create();
        $otherId = $agent->id + 1;

        $this->actingAs($user)
            ->withSession(["token_{$otherId}" => 'other-token-value'])
            ->get(route('agents.show', $agent))
            ->assertOk()
            ->assertInertia(fn ($page) => $page
                ->where('token', null)
            );
    });

    it('passes null token when session has no token', function (): void {
        $user = User::factory()->create();
        $agent = Agent::factory()->create();

        $this->actingAs($user)
            ->get(route('agents.show', $agent))
            ->assertOk()
            ->assertInertia(fn ($page) => $page
                ->where('token', null)
            );
    });

    it('does not expose the hashed token', function (): void {
        $user = User::factory()->create();
        $agent = Agent::factory()->create();

        $this->actingAs($user)
            ->get(route('agents.show', $agent))
            ->assertOk()
            ->assertInertia(fn ($page) => $page
                ->missing('agent.token')
            );
    });

    it('passes canRegenerate true for admin', function (): void {
        $agent = Agent::factory()->create();

        $this->actingAs(User::factory()->admin()->create())
            ->get(route('agents.show', $agent))
            ->assertOk()
            ->assertInertia(fn ($page) => $page->where('canRegenerate', true));
    });

    it('passes canRegenerate false for editor', function (): void {
        $agent = Agent::factory()->create();

        $this->actingAs(User::factory()->editor()->create())
            ->get(route('agents.show', $agent))
            ->assertOk()
            ->assertInertia(fn ($page) => $page->where('canRegenerate', false));
    });

    it('passes canRegenerate false for viewer', function (): void {
        $agent = Agent::factory()->create();

        $this->actingAs(User::factory()->create())
            ->get(route('agents.show', $agent))
            ->assertOk()
            ->assertInertia(fn ($page) => $page->where('canRegenerate', false));
    });

    it('requires authentication', function (): void {
        $agent = Agent::factory()->create();

        $this->get(route('agents.show', $agent))
            ->assertRedirect(route('login'));
    });
});

describe('agents regenerate-token', function (): void {
    it('regenerates the token', function (): void {
        $user = User::factory()->admin()->create();
        $agent = Agent::factory()->create();
        $oldToken = $agent->getRawOriginal('token');

        $this->actingAs($user)
            ->post(route('agents.regenerate-token', $agent));

        expect($agent->fresh()->getRawOriginal('token'))
            ->not->toBe($oldToken)
            ->toHaveLength(64);
    });

    it('stores the new token as a sha256 hash', function (): void {
        $user = User::factory()->admin()->create();
        $agent = Agent::factory()->create();

        $this->actingAs($user)
            ->post(route('agents.regenerate-token', $agent));

        expect(strlen($agent->fresh()->getRawOriginal('token')))->toBe(64);
    });

    it('redirects to the agent show page after regeneration', function (): void {
        $user = User::factory()->admin()->create();
        $agent = Agent::factory()->create();

        $this->actingAs($user)
            ->post(route('agents.regenerate-token', $agent))
            ->assertRedirect(route('agents.show', $agent));
    });

    it('flashes the new plain token to the session scoped to the agent', function (): void {
        $user = User::factory()->admin()->create();
        $agent = Agent::factory()->create();

        $this->actingAs($user)
            ->post(route('agents.regenerate-token', $agent))
            ->assertSessionHas("token_{$agent->id}");
    });

    it('forbids editor from regenerating token', function (): void {
        $agent = Agent::factory()->create();

        $this->actingAs(User::factory()->editor()->create())
            ->post(route('agents.regenerate-token', $agent))
            ->assertForbidden();
    });

    it('forbids viewer from regenerating token', function (): void {
        $agent = Agent::factory()->create();

        $this->actingAs(User::factory()->create())
            ->post(route('agents.regenerate-token', $agent))
            ->assertForbidden();
    });

    it('requires authentication', function (): void {
        $agent = Agent::factory()->create();

        $this->post(route('agents.regenerate-token', $agent))
            ->assertRedirect(route('login'));
    });
});

describe('agent permissions', function (): void {
    it('allows admin to view agents', function (): void {
        $this->actingAs(User::factory()->admin()->create())
            ->get(route('agents.index'))
            ->assertOk();
    });

    it('allows editor to view agents', function (): void {
        $this->actingAs(User::factory()->editor()->create())
            ->get(route('agents.index'))
            ->assertOk();
    });

    it('allows viewer to view agents', function (): void {
        $this->actingAs(User::factory()->create())
            ->get(route('agents.index'))
            ->assertOk();
    });

    it('allows admin to create agents', function (): void {
        $this->actingAs(User::factory()->admin()->create())
            ->get(route('agents.create'))
            ->assertOk();
    });

    it('allows editor to create agents', function (): void {
        $this->actingAs(User::factory()->editor()->create())
            ->get(route('agents.create'))
            ->assertOk();
    });

    it('forbids viewer from creating agents', function (): void {
        $this->actingAs(User::factory()->create())
            ->get(route('agents.create'))
            ->assertForbidden();
    });

    it('forbids viewer from storing agents', function (): void {
        $this->actingAs(User::factory()->create())
            ->post(route('agents.store'), ['name' => 'My Agent'])
            ->assertForbidden();
    });

    it('allows admin to regenerate agent tokens', function (): void {
        $agent = Agent::factory()->create();

        $this->actingAs(User::factory()->admin()->create())
            ->post(route('agents.regenerate-token', $agent))
            ->assertRedirect();
    });

    it('forbids editor from regenerating agent tokens', function (): void {
        $agent = Agent::factory()->create();

        $this->actingAs(User::factory()->editor()->create())
            ->post(route('agents.regenerate-token', $agent))
            ->assertForbidden();
    });
});
