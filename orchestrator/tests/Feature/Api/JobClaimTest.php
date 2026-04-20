<?php

use App\Models\Agent;
use App\Models\ScrapeJob;
use Illuminate\Support\Str;

describe('POST /api/v1/jobs/claim', function (): void {
    beforeEach(function (): void {
        $this->plainToken = Str::random(64);
        $this->agent = Agent::factory()->create(['token' => hash('sha256', $this->plainToken)]);
    });

    it('returns 200 with job payload when a pending job exists', function (): void {
        $job = ScrapeJob::factory()->create(['url' => 'https://example.com']);

        $this->postJson('/api/v1/jobs/claim', [], ['Authorization' => "Bearer {$this->plainToken}"])
            ->assertOk()
            ->assertJsonPath('job_id', $job->ulid)
            ->assertJsonPath('url', 'https://example.com')
            ->assertJsonStructure(['template', 'timeout_s']);
    });

    it('returns 204 when no pending jobs exist', function (): void {
        $this->postJson('/api/v1/jobs/claim', [], ['Authorization' => "Bearer {$this->plainToken}"])
            ->assertNoContent();
    });

    it('sets job status to processing', function (): void {
        $job = ScrapeJob::factory()->create();

        $this->postJson('/api/v1/jobs/claim', [], ['Authorization' => "Bearer {$this->plainToken}"]);

        expect($job->fresh()->status)->toBe('processing');
    });

    it('assigns the claiming agent to the job', function (): void {
        $job = ScrapeJob::factory()->create();

        $this->postJson('/api/v1/jobs/claim', [], ['Authorization' => "Bearer {$this->plainToken}"]);

        expect($job->fresh()->agent_id)->toBe($this->agent->id);
    });

    it('sets agent status to processing', function (): void {
        ScrapeJob::factory()->create();

        $this->postJson('/api/v1/jobs/claim', [], ['Authorization' => "Bearer {$this->plainToken}"]);

        expect($this->agent->fresh()->status)->toBe('processing');
    });

    it('increments attempts on claim', function (): void {
        $job = ScrapeJob::factory()->create();

        $this->postJson('/api/v1/jobs/claim', [], ['Authorization' => "Bearer {$this->plainToken}"]);

        expect($job->fresh()->attempts)->toBe(1);
    });

    it('sets claimed_at and timeout_at', function (): void {
        ScrapeJob::factory()->create();

        $this->postJson('/api/v1/jobs/claim', [], ['Authorization' => "Bearer {$this->plainToken}"]);

        $job = ScrapeJob::first();
        expect($job->claimed_at)->not->toBeNull();
        expect($job->timeout_at)->not->toBeNull();
        expect($job->timeout_at->gt($job->claimed_at))->toBeTrue();
    });

    it('claims jobs in FIFO order', function (): void {
        $first = ScrapeJob::factory()->create(['created_at' => now()->subMinute()]);
        ScrapeJob::factory()->create(['created_at' => now()]);

        $this->postJson('/api/v1/jobs/claim', [], ['Authorization' => "Bearer {$this->plainToken}"])
            ->assertJsonPath('job_id', $first->ulid);
    });

    it('skips jobs where attempts equals max_attempts', function (): void {
        ScrapeJob::factory()->create(['attempts' => 3, 'max_attempts' => 3]);

        $this->postJson('/api/v1/jobs/claim', [], ['Authorization' => "Bearer {$this->plainToken}"])
            ->assertNoContent();
    });

    it('does not claim a job already being processed', function (): void {
        $other = Agent::factory()->create();
        ScrapeJob::factory()->processing($other)->create();

        $this->postJson('/api/v1/jobs/claim', [], ['Authorization' => "Bearer {$this->plainToken}"])
            ->assertNoContent();
    });

    it('returns 401 without an agent token', function (): void {
        $this->postJson('/api/v1/jobs/claim')->assertUnauthorized();
    });

    it('returns 401 for an invalid agent token', function (): void {
        $this->postJson('/api/v1/jobs/claim', [], ['Authorization' => 'Bearer invalid'])
            ->assertUnauthorized();
    });
});
