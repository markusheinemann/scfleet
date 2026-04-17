<?php

use App\Models\Agent;
use Illuminate\Support\Str;

describe('POST /api/v1/register', function (): void {
    it('returns 204 on success', function (): void {
        $plain = Str::random(64);
        Agent::factory()->create(['token' => hash('sha256', $plain)]);

        $this->postJson('/api/v1/register', [], ['Authorization' => "Bearer $plain"])
            ->assertNoContent();
    });

    it('sets registered_at on first call', function (): void {
        $plain = Str::random(64);
        $agent = Agent::factory()->create([
            'token' => hash('sha256', $plain),
            'registered_at' => null,
        ]);

        $this->postJson('/api/v1/register', [], ['Authorization' => "Bearer $plain"]);

        expect($agent->fresh()->registered_at)->not->toBeNull();
    });

    it('does not overwrite registered_at on subsequent calls', function (): void {
        $plain = Str::random(64);
        $registeredAt = now()->subDay();
        $agent = Agent::factory()->create([
            'token' => hash('sha256', $plain),
            'registered_at' => $registeredAt,
        ]);

        $this->postJson('/api/v1/register', [], ['Authorization' => "Bearer $plain"]);

        expect($agent->fresh()->registered_at->toDateTimeString())
            ->toBe($registeredAt->toDateTimeString());
    });

    it('returns 401 when no token is provided', function (): void {
        $this->postJson('/api/v1/register')
            ->assertUnauthorized();
    });

    it('returns 401 for an invalid token', function (): void {
        Agent::factory()->create();

        $this->postJson('/api/v1/register', [], ['Authorization' => 'Bearer invalid-token'])
            ->assertUnauthorized();
    });
});
