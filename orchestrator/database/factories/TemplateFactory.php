<?php

namespace Database\Factories;

use App\Models\Template;
use App\Models\User;
use Illuminate\Database\Eloquent\Factories\Factory;

/**
 * @extends Factory<Template>
 */
class TemplateFactory extends Factory
{
    public function definition(): array
    {
        return [
            'user_id' => User::factory(),
            'title' => fake()->words(3, true),
            'template' => [
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
