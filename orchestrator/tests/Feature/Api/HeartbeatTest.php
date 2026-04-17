<?php

use App\Models\Agent;
use Illuminate\Support\Str;

describe('POST /api/v1/heartbeat', function (): void {
    it('returns 204 on success', function (): void {
        $plain = Str::random(64);
        Agent::factory()->create(['token' => hash('sha256', $plain)]);

        $this->postJson('/api/v1/heartbeat', [], ['Authorization' => "Bearer $plain"])
            ->assertNoContent();
    });

    it('updates last_heartbeat_at', function (): void {
        $plain = Str::random(64);
        $agent = Agent::factory()->create([
            'token' => hash('sha256', $plain),
            'last_heartbeat_at' => null,
        ]);

        $this->postJson('/api/v1/heartbeat', [], ['Authorization' => "Bearer $plain"]);

        expect($agent->fresh()->last_heartbeat_at)->not->toBeNull();
    });

    it('overwrites last_heartbeat_at on each call', function (): void {
        $plain = Str::random(64);
        $previousHeartbeat = now()->subMinute();
        $agent = Agent::factory()->create([
            'token' => hash('sha256', $plain),
            'last_heartbeat_at' => $previousHeartbeat,
        ]);

        $this->travel(30)->seconds();

        $this->postJson('/api/v1/heartbeat', [], ['Authorization' => "Bearer $plain"]);

        expect($agent->fresh()->last_heartbeat_at->gt($previousHeartbeat))->toBeTrue();
    });

    it('returns 401 when no token is provided', function (): void {
        $this->postJson('/api/v1/heartbeat')
            ->assertUnauthorized();
    });

    it('returns 401 for an invalid token', function (): void {
        Agent::factory()->create();

        $this->postJson('/api/v1/heartbeat', [], ['Authorization' => 'Bearer invalid-token'])
            ->assertUnauthorized();
    });
});

describe('Agent::is_online', function (): void {
    it('is false when last_heartbeat_at is null', function (): void {
        $agent = Agent::factory()->create(['last_heartbeat_at' => null]);

        expect($agent->is_online)->toBeFalse();
    });

    it('is true when last heartbeat is within the offline threshold', function (): void {
        config(['agent.offline_after' => 120]);
        $agent = Agent::factory()->create(['last_heartbeat_at' => now()->subSeconds(60)]);

        expect($agent->is_online)->toBeTrue();
    });

    it('is false when last heartbeat exceeds the offline threshold', function (): void {
        config(['agent.offline_after' => 120]);
        $agent = Agent::factory()->create(['last_heartbeat_at' => now()->subSeconds(121)]);

        expect($agent->is_online)->toBeFalse();
    });

    it('respects a custom offline threshold from config', function (): void {
        config(['agent.offline_after' => 60]);
        $agent = Agent::factory()->create(['last_heartbeat_at' => now()->subSeconds(61)]);

        expect($agent->is_online)->toBeFalse();
    });
});
