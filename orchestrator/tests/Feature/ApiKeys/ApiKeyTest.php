<?php

use App\Models\ApiKey;
use App\Models\User;
use Illuminate\Support\Str;

describe('api-keys index', function (): void {
    it('shows the api keys page', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->get(route('api-keys.index'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page->component('api-keys/index'));
    });

    it('requires authentication', function (): void {
        $this->get(route('api-keys.index'))
            ->assertRedirect(route('login'));
    });

    it('lists only the authenticated user\'s api keys', function (): void {
        $user = User::factory()->admin()->create();
        ApiKey::factory(2)->create(['user_id' => $user->id]);
        ApiKey::factory()->create(); // another user's key

        $this->actingAs($user)
            ->get(route('api-keys.index'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page
                ->component('api-keys/index')
                ->has('apiKeys', 2)
            );
    });

    it('passes the new key from session flash', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->withSession(['new_api_key' => 'plain-key-value'])
            ->get(route('api-keys.index'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page->where('newKey', 'plain-key-value'));
    });

    it('passes null new key when session has no key', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->get(route('api-keys.index'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page->where('newKey', null));
    });

    it('does not expose the key hash in the listing', function (): void {
        $user = User::factory()->admin()->create();
        ApiKey::factory()->create(['user_id' => $user->id]);

        $this->actingAs($user)
            ->get(route('api-keys.index'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page
                ->has('apiKeys', 1)
                ->missing('apiKeys.0.key_hash')
            );
    });

    it('forbids viewers from accessing api keys', function (): void {
        $this->actingAs(User::factory()->create())
            ->get(route('api-keys.index'))
            ->assertForbidden();
    });
});

describe('api-keys store', function (): void {
    it('creates a new api key', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)->post(route('api-keys.store'), ['name' => 'Production']);

        expect(ApiKey::count())->toBe(1)
            ->and(ApiKey::first()->name)->toBe('Production')
            ->and(ApiKey::first()->user_id)->toBe($user->id);
    });

    it('stores the key as a sha256 hash', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)->post(route('api-keys.store'), ['name' => 'CI']);

        expect(ApiKey::first()->key_hash)->toHaveLength(64);
    });

    it('flashes the plain key to the session', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->post(route('api-keys.store'), ['name' => 'CI'])
            ->assertSessionHas('new_api_key');
    });

    it('redirects to the api keys index after creation', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->post(route('api-keys.store'), ['name' => 'CI'])
            ->assertRedirect(route('api-keys.index'));
    });

    it('requires a name', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->post(route('api-keys.store'), [])
            ->assertSessionHasErrors('name');
    });

    it('requires name within 255 characters', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->post(route('api-keys.store'), ['name' => Str::repeat('a', 256)])
            ->assertSessionHasErrors('name');
    });

    it('requires authentication', function (): void {
        $this->post(route('api-keys.store'), ['name' => 'CI'])
            ->assertRedirect(route('login'));
    });

    it('forbids viewers from creating api keys', function (): void {
        $this->actingAs(User::factory()->create())
            ->post(route('api-keys.store'), ['name' => 'CI'])
            ->assertForbidden();
    });
});

describe('api-keys destroy', function (): void {
    it('deletes the api key', function (): void {
        $user = User::factory()->admin()->create();
        $apiKey = ApiKey::factory()->create(['user_id' => $user->id]);

        $this->actingAs($user)->delete(route('api-keys.destroy', $apiKey));

        expect(ApiKey::count())->toBe(0);
    });

    it('redirects to the api keys index after deletion', function (): void {
        $user = User::factory()->admin()->create();
        $apiKey = ApiKey::factory()->create(['user_id' => $user->id]);

        $this->actingAs($user)
            ->delete(route('api-keys.destroy', $apiKey))
            ->assertRedirect(route('api-keys.index'));
    });

    it('prevents a user from deleting another user\'s api key', function (): void {
        $user = User::factory()->admin()->create();
        $otherKey = ApiKey::factory()->create(); // different user

        $this->actingAs($user)
            ->delete(route('api-keys.destroy', $otherKey))
            ->assertForbidden();

        expect(ApiKey::count())->toBe(1);
    });

    it('forbids viewers from deleting api keys', function (): void {
        $viewer = User::factory()->create();
        $apiKey = ApiKey::factory()->create(['user_id' => $viewer->id]);

        $this->actingAs($viewer)
            ->delete(route('api-keys.destroy', $apiKey))
            ->assertForbidden();
    });

    it('requires authentication', function (): void {
        $apiKey = ApiKey::factory()->create();

        $this->delete(route('api-keys.destroy', $apiKey))
            ->assertRedirect(route('login'));
    });
});

describe('api-keys permissions', function (): void {
    it('allows admin to manage api keys', function (): void {
        $this->actingAs(User::factory()->admin()->create())
            ->get(route('api-keys.index'))
            ->assertOk();
    });

    it('allows editor to manage api keys', function (): void {
        $this->actingAs(User::factory()->editor()->create())
            ->get(route('api-keys.index'))
            ->assertOk();
    });

    it('forbids viewer from managing api keys', function (): void {
        $this->actingAs(User::factory()->create())
            ->get(route('api-keys.index'))
            ->assertForbidden();
    });
});
