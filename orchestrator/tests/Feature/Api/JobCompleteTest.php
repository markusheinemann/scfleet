<?php

use App\Models\Agent;
use App\Models\ScrapeJob;
use Illuminate\Support\Str;

describe('POST /api/v1/jobs/{ulid}/complete', function (): void {
    beforeEach(function (): void {
        $this->plainToken = Str::random(64);
        $this->agent = Agent::factory()->create([
            'token' => hash('sha256', $this->plainToken),
            'status' => 'processing',
        ]);
        $this->job = ScrapeJob::factory()->processing($this->agent)->create();
    });

    it('marks the job as completed', function (): void {
        $this->postJson("/api/v1/jobs/{$this->job->ulid}/complete", [
            'result' => ['title' => 'Widget Pro'],
        ], ['Authorization' => "Bearer {$this->plainToken}"])
            ->assertNoContent();

        expect($this->job->fresh()->status)->toBe('completed');
    });

    it('stores the extracted result', function (): void {
        $this->postJson("/api/v1/jobs/{$this->job->ulid}/complete", [
            'result' => ['title' => 'Widget Pro', 'price' => 29.99],
        ], ['Authorization' => "Bearer {$this->plainToken}"]);

        $result = $this->job->fresh()->result;
        expect($result['title'])->toBe('Widget Pro');
        expect((float) $result['price'])->toBe(29.99);
    });

    it('stores field_errors when provided', function (): void {
        $this->postJson("/api/v1/jobs/{$this->job->ulid}/complete", [
            'result' => ['title' => 'Widget Pro'],
            'field_errors' => ['rating' => 'no extractor yielded a value'],
        ], ['Authorization' => "Bearer {$this->plainToken}"]);

        expect($this->job->fresh()->field_errors['rating'])->toBe('no extractor yielded a value');
    });

    it('sets completed_at', function (): void {
        $this->postJson("/api/v1/jobs/{$this->job->ulid}/complete", [
            'result' => ['title' => 'Widget Pro'],
        ], ['Authorization' => "Bearer {$this->plainToken}"]);

        expect($this->job->fresh()->completed_at)->not->toBeNull();
    });

    it('resets agent status to ready', function (): void {
        $this->postJson("/api/v1/jobs/{$this->job->ulid}/complete", [
            'result' => ['title' => 'Widget Pro'],
        ], ['Authorization' => "Bearer {$this->plainToken}"]);

        expect($this->agent->fresh()->status)->toBe('ready');
    });

    it('returns 403 when the agent does not own the job', function (): void {
        $otherToken = Str::random(64);
        Agent::factory()->create(['token' => hash('sha256', $otherToken)]);

        $this->postJson("/api/v1/jobs/{$this->job->ulid}/complete", [
            'result' => ['title' => 'Widget Pro'],
        ], ['Authorization' => "Bearer $otherToken"])
            ->assertForbidden();
    });

    it('returns 422 when result is missing', function (): void {
        $this->postJson("/api/v1/jobs/{$this->job->ulid}/complete", [], ['Authorization' => "Bearer {$this->plainToken}"])
            ->assertUnprocessable()
            ->assertJsonValidationErrors(['result']);
    });

    it('returns 401 without an agent token', function (): void {
        $this->postJson("/api/v1/jobs/{$this->job->ulid}/complete", ['result' => []])
            ->assertUnauthorized();
    });
});
