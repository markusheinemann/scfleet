<?php

use App\Models\Agent;
use App\Models\ApiKey;
use App\Models\ScrapeJob;
use App\Models\Template;
use Illuminate\Support\Str;

describe('POST /api/v1/scrape', function (): void {
    beforeEach(function (): void {
        $this->plainKey = Str::random(64);
        ApiKey::factory()->create(['key_hash' => hash('sha256', $this->plainKey)]);
        $this->template = Template::factory()->create();
    });

    it('submits a scrape job and returns 202 with job_id', function (): void {
        $this->postJson('/api/v1/scrape', [
            'url' => 'https://example.com',
            'template_id' => $this->template->id,
        ], ['Authorization' => "Bearer {$this->plainKey}"])
            ->assertStatus(202)
            ->assertJsonStructure(['job_id', 'status', 'created_at'])
            ->assertJsonPath('status', 'pending');

        expect(ScrapeJob::count())->toBe(1);
    });

    it('stores the template json and template_id from the referenced template', function (): void {
        $this->postJson('/api/v1/scrape', [
            'url' => 'https://example.com',
            'template_id' => $this->template->id,
        ], ['Authorization' => "Bearer {$this->plainKey}"]);

        $job = ScrapeJob::first();
        expect($job->template_id)->toBe($this->template->id)
            ->and($job->template)->toEqual($this->template->template);
    });

    it('computes and stores a template_hash', function (): void {
        $this->postJson('/api/v1/scrape', [
            'url' => 'https://example.com',
            'template_id' => $this->template->id,
        ], ['Authorization' => "Bearer {$this->plainKey}"]);

        expect(ScrapeJob::first()->template_hash)->not->toBeNull()->toHaveLength(64);
    });

    it('returns 401 without an api key', function (): void {
        $this->postJson('/api/v1/scrape', [
            'url' => 'https://example.com',
            'template_id' => $this->template->id,
        ])->assertUnauthorized();
    });

    it('returns 401 for an invalid api key', function (): void {
        $this->postJson('/api/v1/scrape', [
            'url' => 'https://example.com',
            'template_id' => $this->template->id,
        ], ['Authorization' => 'Bearer invalid'])->assertUnauthorized();
    });

    it('returns 422 for a missing url', function (): void {
        $this->postJson('/api/v1/scrape', [
            'template_id' => $this->template->id,
        ], ['Authorization' => "Bearer {$this->plainKey}"])
            ->assertUnprocessable()
            ->assertJsonValidationErrors(['url']);
    });

    it('returns 422 for an invalid url', function (): void {
        $this->postJson('/api/v1/scrape', [
            'url' => 'not-a-url',
            'template_id' => $this->template->id,
        ], ['Authorization' => "Bearer {$this->plainKey}"])
            ->assertUnprocessable()
            ->assertJsonValidationErrors(['url']);
    });

    it('returns 422 for a missing template_id', function (): void {
        $this->postJson('/api/v1/scrape', [
            'url' => 'https://example.com',
        ], ['Authorization' => "Bearer {$this->plainKey}"])
            ->assertUnprocessable()
            ->assertJsonValidationErrors(['template_id']);
    });

    it('returns 422 for a non-existent template_id', function (): void {
        $this->postJson('/api/v1/scrape', [
            'url' => 'https://example.com',
            'template_id' => 99999,
        ], ['Authorization' => "Bearer {$this->plainKey}"])
            ->assertUnprocessable()
            ->assertJsonValidationErrors(['template_id']);
    });
});

describe('GET /api/v1/scrape/{ulid}', function (): void {
    beforeEach(function (): void {
        $this->plainKey = Str::random(64);
        ApiKey::factory()->create(['key_hash' => hash('sha256', $this->plainKey)]);
    });

    it('returns job status for a pending job', function (): void {
        $job = ScrapeJob::factory()->create();

        $this->getJson("/api/v1/scrape/{$job->ulid}", ['Authorization' => "Bearer {$this->plainKey}"])
            ->assertOk()
            ->assertJsonPath('job_id', $job->ulid)
            ->assertJsonPath('status', 'pending')
            ->assertJsonPath('result', null);
    });

    it('returns result for a completed job', function (): void {
        $job = ScrapeJob::factory()->completed()->create();

        $this->getJson("/api/v1/scrape/{$job->ulid}", ['Authorization' => "Bearer {$this->plainKey}"])
            ->assertOk()
            ->assertJsonPath('status', 'completed')
            ->assertJsonPath('result.title', 'Test Title');
    });

    it('returns 404 for an unknown ulid', function (): void {
        $this->getJson('/api/v1/scrape/01JUNK00000000000000000000', ['Authorization' => "Bearer {$this->plainKey}"])
            ->assertNotFound();
    });

    it('returns 401 without an api key', function (): void {
        $job = ScrapeJob::factory()->create();

        $this->getJson("/api/v1/scrape/{$job->ulid}")->assertUnauthorized();
    });
});

describe('Full scrape lifecycle', function (): void {
    it('submit → claim → complete → show', function (): void {
        $plainKey = Str::random(64);
        ApiKey::factory()->create(['key_hash' => hash('sha256', $plainKey)]);

        $plainToken = Str::random(64);
        Agent::factory()->create(['token' => hash('sha256', $plainToken)]);

        $template = Template::factory()->create();

        $jobId = $this->postJson('/api/v1/scrape', [
            'url' => 'https://example.com/product/1',
            'template_id' => $template->id,
        ], ['Authorization' => "Bearer $plainKey"])
            ->assertStatus(202)
            ->json('job_id');

        $this->postJson('/api/v1/jobs/claim', [], ['Authorization' => "Bearer $plainToken"])
            ->assertOk()
            ->assertJsonPath('job_id', $jobId)
            ->assertJsonPath('url', 'https://example.com/product/1')
            ->assertJsonStructure(['template', 'timeout_s']);

        $this->postJson("/api/v1/jobs/{$jobId}/complete", [
            'result' => ['title' => 'Awesome Widget'],
            'field_errors' => [],
        ], ['Authorization' => "Bearer $plainToken"])
            ->assertNoContent();

        $this->getJson("/api/v1/scrape/{$jobId}", ['Authorization' => "Bearer $plainKey"])
            ->assertOk()
            ->assertJsonPath('status', 'completed')
            ->assertJsonPath('result.title', 'Awesome Widget')
            ->assertJsonPath('field_errors', []);
    });

    it('submit → claim → fail → requeue → claim again', function (): void {
        $plainKey = Str::random(64);
        ApiKey::factory()->create(['key_hash' => hash('sha256', $plainKey)]);

        $plainToken = Str::random(64);
        Agent::factory()->create(['token' => hash('sha256', $plainToken)]);

        $template = Template::factory()->create();

        $jobId = $this->postJson('/api/v1/scrape', [
            'url' => 'https://example.com',
            'template_id' => $template->id,
        ], ['Authorization' => "Bearer $plainKey"])->json('job_id');

        $this->postJson('/api/v1/jobs/claim', [], ['Authorization' => "Bearer $plainToken"])
            ->assertOk()->assertJsonPath('job_id', $jobId);

        $this->postJson("/api/v1/jobs/{$jobId}/fail", [
            'error_type' => 'navigation_error',
            'error_message' => 'DNS lookup failed',
        ], ['Authorization' => "Bearer $plainToken"])->assertNoContent();

        expect(ScrapeJob::find($jobId)->status)->toBe('pending');

        $this->postJson('/api/v1/jobs/claim', [], ['Authorization' => "Bearer $plainToken"])
            ->assertOk()->assertJsonPath('job_id', $jobId);

        expect(ScrapeJob::find($jobId)->attempts)->toBe(2);
    });
});
