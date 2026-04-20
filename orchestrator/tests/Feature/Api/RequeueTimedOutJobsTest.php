<?php

use App\Models\Agent;
use App\Models\ScrapeJob;

describe('scrape:requeue-timed-out command', function (): void {
    it('requeues a timed-out job when attempts are below max_attempts', function (): void {
        $agent = Agent::factory()->create(['status' => 'processing']);
        $job = ScrapeJob::factory()->timedOut($agent)->create(['attempts' => 1, 'max_attempts' => 3]);

        $this->artisan('scrape:requeue-timed-out')->assertExitCode(0);

        $fresh = $job->fresh();
        expect($fresh->status)->toBe('pending');
        expect($fresh->agent_id)->toBeNull();
        expect($fresh->claimed_at)->toBeNull();
        expect($fresh->timeout_at)->toBeNull();
    });

    it('permanently fails a timed-out job when attempts equal max_attempts', function (): void {
        $agent = Agent::factory()->create(['status' => 'processing']);
        $job = ScrapeJob::factory()->exhausted($agent)->create();

        $this->artisan('scrape:requeue-timed-out')->assertExitCode(0);

        $fresh = $job->fresh();
        expect($fresh->status)->toBe('failed');
        expect($fresh->error_type)->toBe('agent_timeout');
        expect($fresh->error_message)->not->toBeEmpty();
        expect($fresh->completed_at)->not->toBeNull();
    });

    it('resets agent status to ready after a requeue', function (): void {
        $agent = Agent::factory()->create(['status' => 'processing']);
        ScrapeJob::factory()->timedOut($agent)->create();

        $this->artisan('scrape:requeue-timed-out');

        expect($agent->fresh()->status)->toBe('ready');
    });

    it('resets agent status to ready after a permanent fail', function (): void {
        $agent = Agent::factory()->create(['status' => 'processing']);
        ScrapeJob::factory()->exhausted($agent)->create();

        $this->artisan('scrape:requeue-timed-out');

        expect($agent->fresh()->status)->toBe('ready');
    });

    it('does not touch pending jobs', function (): void {
        $job = ScrapeJob::factory()->create(['status' => 'pending']);

        $this->artisan('scrape:requeue-timed-out');

        expect($job->fresh()->status)->toBe('pending');
    });

    it('does not touch completed jobs', function (): void {
        $job = ScrapeJob::factory()->completed()->create();

        $this->artisan('scrape:requeue-timed-out');

        expect($job->fresh()->status)->toBe('completed');
    });

    it('does not touch processing jobs whose timeout has not expired', function (): void {
        $agent = Agent::factory()->create(['status' => 'processing']);
        $job = ScrapeJob::factory()->processing($agent)->create();

        $this->artisan('scrape:requeue-timed-out');

        expect($job->fresh()->status)->toBe('processing');
        expect($agent->fresh()->status)->toBe('processing');
    });

    it('handles multiple timed-out jobs in one run', function (): void {
        $agent1 = Agent::factory()->create(['status' => 'processing']);
        $agent2 = Agent::factory()->create(['status' => 'processing']);
        $job1 = ScrapeJob::factory()->timedOut($agent1)->create(['attempts' => 1]);
        $job2 = ScrapeJob::factory()->exhausted($agent2)->create();

        $this->artisan('scrape:requeue-timed-out');

        expect($job1->fresh()->status)->toBe('pending');
        expect($job2->fresh()->status)->toBe('failed');
        expect($agent1->fresh()->status)->toBe('ready');
        expect($agent2->fresh()->status)->toBe('ready');
    });
});
