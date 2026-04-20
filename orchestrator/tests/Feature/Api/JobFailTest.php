<?php

use App\Models\Agent;
use App\Models\ScrapeJob;
use Illuminate\Support\Str;

describe('POST /api/v1/jobs/{ulid}/fail', function (): void {
    beforeEach(function (): void {
        $this->plainToken = Str::random(64);
        $this->agent = Agent::factory()->create([
            'token' => hash('sha256', $this->plainToken),
            'status' => 'processing',
        ]);
    });

    it('requeues the job when attempts are below max_attempts', function (): void {
        $job = ScrapeJob::factory()->processing($this->agent)->create(['attempts' => 1, 'max_attempts' => 3]);

        $this->postJson("/api/v1/jobs/{$job->ulid}/fail", [
            'error_type' => 'navigation_error',
            'error_message' => 'DNS lookup failed',
        ], ['Authorization' => "Bearer {$this->plainToken}"])
            ->assertNoContent();

        $fresh = $job->fresh();
        expect($fresh->status)->toBe('pending');
        expect($fresh->agent_id)->toBeNull();
        expect($fresh->claimed_at)->toBeNull();
        expect($fresh->timeout_at)->toBeNull();
    });

    it('permanently fails the job when attempts equal max_attempts', function (): void {
        $job = ScrapeJob::factory()->exhausted($this->agent)->create();

        $this->postJson("/api/v1/jobs/{$job->ulid}/fail", [
            'error_type' => 'page_timeout',
            'error_message' => 'Timed out after 30s',
        ], ['Authorization' => "Bearer {$this->plainToken}"])
            ->assertNoContent();

        $fresh = $job->fresh();
        expect($fresh->status)->toBe('failed');
        expect($fresh->error_type)->toBe('page_timeout');
        expect($fresh->error_message)->toBe('Timed out after 30s');
        expect($fresh->completed_at)->not->toBeNull();
    });

    it('resets agent status to ready after a requeue', function (): void {
        $job = ScrapeJob::factory()->processing($this->agent)->create(['attempts' => 1]);

        $this->postJson("/api/v1/jobs/{$job->ulid}/fail", [
            'error_type' => 'navigation_error',
            'error_message' => 'error',
        ], ['Authorization' => "Bearer {$this->plainToken}"]);

        expect($this->agent->fresh()->status)->toBe('ready');
    });

    it('resets agent status to ready after a permanent fail', function (): void {
        $job = ScrapeJob::factory()->exhausted($this->agent)->create();

        $this->postJson("/api/v1/jobs/{$job->ulid}/fail", [
            'error_type' => 'page_timeout',
            'error_message' => 'error',
        ], ['Authorization' => "Bearer {$this->plainToken}"]);

        expect($this->agent->fresh()->status)->toBe('ready');
    });

    it('returns 403 when the agent does not own the job', function (): void {
        $job = ScrapeJob::factory()->processing($this->agent)->create();

        $otherToken = Str::random(64);
        Agent::factory()->create(['token' => hash('sha256', $otherToken)]);

        $this->postJson("/api/v1/jobs/{$job->ulid}/fail", [
            'error_type' => 'navigation_error',
            'error_message' => 'error',
        ], ['Authorization' => "Bearer $otherToken"])
            ->assertForbidden();
    });

    it('returns 422 for an invalid error_type', function (): void {
        $job = ScrapeJob::factory()->processing($this->agent)->create();

        $this->postJson("/api/v1/jobs/{$job->ulid}/fail", [
            'error_type' => 'unknown_type',
            'error_message' => 'error',
        ], ['Authorization' => "Bearer {$this->plainToken}"])
            ->assertUnprocessable()
            ->assertJsonValidationErrors(['error_type']);
    });

    it('accepts all valid error_type values', function (string $errorType): void {
        $job = ScrapeJob::factory()->processing($this->agent)->create(['attempts' => 1]);

        $this->postJson("/api/v1/jobs/{$job->ulid}/fail", [
            'error_type' => $errorType,
            'error_message' => 'test error',
        ], ['Authorization' => "Bearer {$this->plainToken}"])
            ->assertNoContent();
    })->with(['missing_required_field', 'page_timeout', 'navigation_error', 'extraction_error']);

    it('returns 401 without an agent token', function (): void {
        $job = ScrapeJob::factory()->processing($this->agent)->create();

        $this->postJson("/api/v1/jobs/{$job->ulid}/fail", [
            'error_type' => 'navigation_error',
            'error_message' => 'error',
        ])->assertUnauthorized();
    });
});
