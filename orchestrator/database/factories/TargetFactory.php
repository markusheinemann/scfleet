<?php

namespace Database\Factories;

use App\Models\Target;
use App\Models\User;
use Illuminate\Database\Eloquent\Factories\Factory;

/**
 * @extends Factory<Target>
 */
class TargetFactory extends Factory
{
    public function definition(): array
    {
        return [
            'user_id' => User::factory(),
            'title' => fake()->words(3, true),
            'url' => fake()->url(),
            'schema' => [
                'version' => '1',
                'fields' => [
                    [
                        'name' => 'title',
                        'type' => 'string',
                        'extractors' => [
                            ['strategy' => 'css', 'selector' => 'h1'],
                        ],
                    ],
                ],
            ],
        ];
    }
}
