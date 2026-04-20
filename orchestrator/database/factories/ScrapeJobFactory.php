<?php

namespace Database\Factories;

use App\Models\Agent;
use App\Models\ScrapeJob;
use Illuminate\Database\Eloquent\Factories\Factory;

/**
 * @extends Factory<ScrapeJob>
 */
class ScrapeJobFactory extends Factory
{
    /** Minimal valid extraction template for tests. */
    private static function minimalTemplate(): array
    {
        return [
            'version' => '1',
            'fields' => [
                [
                    'name' => 'title',
                    'type' => 'string',
                    'required' => true,
                    'extractors' => [
                        ['strategy' => 'css', 'selector' => 'h1'],
                    ],
                ],
            ],
        ];
    }

    public function definition(): array
    {
        return [
            'url' => fake()->url(),
            'template' => self::minimalTemplate(),
            'status' => 'pending',
        ];
    }

    public function processing(?Agent $agent = null): static
    {
        return $this->state(fn () => [
            'status' => 'processing',
            'agent_id' => $agent?->id ?? Agent::factory(),
            'claimed_at' => now(),
            'timeout_at' => now()->addMinutes(2),
            'attempts' => 1,
        ]);
    }

    public function timedOut(?Agent $agent = null): static
    {
        return $this->state(fn () => [
            'status' => 'processing',
            'agent_id' => $agent?->id ?? Agent::factory(),
            'claimed_at' => now()->subMinutes(5),
            'timeout_at' => now()->subMinutes(1),
            'attempts' => 1,
        ]);
    }

    public function exhausted(?Agent $agent = null): static
    {
        return $this->state(fn () => [
            'status' => 'processing',
            'agent_id' => $agent?->id ?? Agent::factory(),
            'claimed_at' => now()->subMinutes(5),
            'timeout_at' => now()->subMinutes(1),
            'attempts' => 3,
            'max_attempts' => 3,
        ]);
    }

    public function completed(): static
    {
        return $this->state([
            'status' => 'completed',
            'result' => ['title' => 'Test Title'],
            'completed_at' => now(),
            'attempts' => 1,
        ]);
    }

    public function failed(): static
    {
        return $this->state([
            'status' => 'failed',
            'error_type' => 'navigation_error',
            'error_message' => 'Could not reach server',
            'completed_at' => now(),
            'attempts' => 3,
        ]);
    }
}
