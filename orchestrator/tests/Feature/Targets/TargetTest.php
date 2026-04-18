<?php

use App\Models\Target;
use App\Models\User;

$validSchema = json_encode([
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
]);

describe('targets index', function (): void {
    it('shows the targets page', function (): void {
        $this->actingAs(User::factory()->create())
            ->get(route('targets.index'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page->component('targets/index'));
    });

    it('requires authentication', function (): void {
        $this->get(route('targets.index'))
            ->assertRedirect(route('login'));
    });

    it('lists all targets', function (): void {
        Target::factory(3)->create();

        $this->actingAs(User::factory()->create())
            ->get(route('targets.index'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page
                ->component('targets/index')
                ->has('targets', 3)
            );
    });
});

describe('targets create', function (): void {
    it('shows the create target page', function (): void {
        $this->actingAs(User::factory()->admin()->create())
            ->get(route('targets.create'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page->component('targets/create'));
    });

    it('requires authentication', function (): void {
        $this->get(route('targets.create'))
            ->assertRedirect(route('login'));
    });

    it('forbids viewer from accessing create page', function (): void {
        $this->actingAs(User::factory()->create())
            ->get(route('targets.create'))
            ->assertForbidden();
    });
});

describe('targets store', function () use ($validSchema): void {
    it('creates a new target', function () use ($validSchema): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)->post(route('targets.store'), [
            'title' => 'My Target',
            'url' => 'https://example.com',
            'schema' => $validSchema,
        ]);

        expect(Target::count())->toBe(1)
            ->and(Target::first()->title)->toBe('My Target')
            ->and(Target::first()->url)->toBe('https://example.com')
            ->and(Target::first()->user_id)->toBe($user->id);
    });

    it('stores the schema as structured data', function () use ($validSchema): void {
        $this->actingAs(User::factory()->admin()->create())->post(route('targets.store'), [
            'title' => 'My Target',
            'url' => 'https://example.com',
            'schema' => $validSchema,
        ]);

        expect(Target::first()->schema)->toBeArray()
            ->and(Target::first()->schema['version'])->toBe('1');
    });

    it('redirects to the targets index after creation', function () use ($validSchema): void {
        $this->actingAs(User::factory()->admin()->create())
            ->post(route('targets.store'), [
                'title' => 'My Target',
                'url' => 'https://example.com',
                'schema' => $validSchema,
            ])
            ->assertRedirect(route('targets.index'));
    });

    it('requires authentication', function () use ($validSchema): void {
        $this->post(route('targets.store'), [
            'title' => 'My Target',
            'url' => 'https://example.com',
            'schema' => $validSchema,
        ])->assertRedirect(route('login'));
    });

    it('requires a title', function () use ($validSchema): void {
        $this->actingAs(User::factory()->admin()->create())
            ->post(route('targets.store'), [
                'url' => 'https://example.com',
                'schema' => $validSchema,
            ])
            ->assertSessionHasErrors('title');
    });

    it('requires a valid url', function () use ($validSchema): void {
        $this->actingAs(User::factory()->admin()->create())
            ->post(route('targets.store'), [
                'title' => 'My Target',
                'url' => 'not-a-url',
                'schema' => $validSchema,
            ])
            ->assertSessionHasErrors('url');
    });

    it('requires a schema', function (): void {
        $this->actingAs(User::factory()->admin()->create())
            ->post(route('targets.store'), [
                'title' => 'My Target',
                'url' => 'https://example.com',
            ])
            ->assertSessionHasErrors('schema');
    });

    it('requires schema to be valid json', function (): void {
        $this->actingAs(User::factory()->admin()->create())
            ->post(route('targets.store'), [
                'title' => 'My Target',
                'url' => 'https://example.com',
                'schema' => 'not-json',
            ])
            ->assertSessionHasErrors('schema');
    });
});

describe('targets schema validation', function (): void {
    it('rejects schema with wrong version', function (): void {
        $schema = json_encode(['version' => '2', 'fields' => [['name' => 'title', 'type' => 'string', 'extractors' => [['strategy' => 'css', 'selector' => 'h1']]]]]);

        $this->actingAs(User::factory()->admin()->create())
            ->post(route('targets.store'), [
                'title' => 'My Target',
                'url' => 'https://example.com',
                'schema' => $schema,
            ])
            ->assertSessionHasErrors('schema');
    });

    it('rejects schema with empty fields array', function (): void {
        $schema = json_encode(['version' => '1', 'fields' => []]);

        $this->actingAs(User::factory()->admin()->create())
            ->post(route('targets.store'), [
                'title' => 'My Target',
                'url' => 'https://example.com',
                'schema' => $schema,
            ])
            ->assertSessionHasErrors('schema');
    });

    it('rejects schema with invalid field name', function (): void {
        $schema = json_encode([
            'version' => '1',
            'fields' => [
                ['name' => 'Invalid Name!', 'type' => 'string', 'extractors' => [['strategy' => 'css', 'selector' => 'h1']]],
            ],
        ]);

        $this->actingAs(User::factory()->admin()->create())
            ->post(route('targets.store'), [
                'title' => 'My Target',
                'url' => 'https://example.com',
                'schema' => $schema,
            ])
            ->assertSessionHasErrors('schema');
    });

    it('rejects schema with unknown top-level key', function (): void {
        $schema = json_encode([
            'version' => '1',
            'fields' => [['name' => 'title', 'type' => 'string', 'extractors' => [['strategy' => 'css', 'selector' => 'h1']]]],
            'unknown_key' => 'value',
        ]);

        $this->actingAs(User::factory()->admin()->create())
            ->post(route('targets.store'), [
                'title' => 'My Target',
                'url' => 'https://example.com',
                'schema' => $schema,
            ])
            ->assertSessionHasErrors('schema');
    });

    it('rejects schema with empty extractors', function (): void {
        $schema = json_encode([
            'version' => '1',
            'fields' => [
                ['name' => 'title', 'type' => 'string', 'extractors' => []],
            ],
        ]);

        $this->actingAs(User::factory()->admin()->create())
            ->post(route('targets.store'), [
                'title' => 'My Target',
                'url' => 'https://example.com',
                'schema' => $schema,
            ])
            ->assertSessionHasErrors('schema');
    });
});

describe('target permissions', function () use ($validSchema): void {
    it('allows admin to view targets', function (): void {
        $this->actingAs(User::factory()->admin()->create())
            ->get(route('targets.index'))
            ->assertOk();
    });

    it('allows editor to view targets', function (): void {
        $this->actingAs(User::factory()->editor()->create())
            ->get(route('targets.index'))
            ->assertOk();
    });

    it('allows viewer to view targets', function (): void {
        $this->actingAs(User::factory()->create())
            ->get(route('targets.index'))
            ->assertOk();
    });

    it('allows admin to create targets', function (): void {
        $this->actingAs(User::factory()->admin()->create())
            ->get(route('targets.create'))
            ->assertOk();
    });

    it('allows editor to create targets', function (): void {
        $this->actingAs(User::factory()->editor()->create())
            ->get(route('targets.create'))
            ->assertOk();
    });

    it('forbids viewer from creating targets', function (): void {
        $this->actingAs(User::factory()->create())
            ->get(route('targets.create'))
            ->assertForbidden();
    });

    it('forbids viewer from storing targets', function () use ($validSchema): void {
        $this->actingAs(User::factory()->create())
            ->post(route('targets.store'), [
                'title' => 'My Target',
                'url' => 'https://example.com',
                'schema' => $validSchema,
            ])
            ->assertForbidden();
    });

    it('allows admin to edit targets', function (): void {
        $target = Target::factory()->create();

        $this->actingAs(User::factory()->admin()->create())
            ->get(route('targets.edit', $target))
            ->assertOk();
    });

    it('allows editor to edit targets', function (): void {
        $target = Target::factory()->create();

        $this->actingAs(User::factory()->editor()->create())
            ->get(route('targets.edit', $target))
            ->assertOk();
    });

    it('forbids viewer from editing targets', function (): void {
        $target = Target::factory()->create();

        $this->actingAs(User::factory()->create())
            ->get(route('targets.edit', $target))
            ->assertForbidden();
    });

    it('forbids viewer from updating targets', function () use ($validSchema): void {
        $target = Target::factory()->create();

        $this->actingAs(User::factory()->create())
            ->put(route('targets.update', $target), [
                'title' => 'Updated',
                'url' => 'https://example.com',
                'schema' => $validSchema,
            ])
            ->assertForbidden();
    });
});

describe('targets edit', function (): void {
    it('shows the edit page with existing values', function (): void {
        $target = Target::factory()->create();

        $this->actingAs(User::factory()->admin()->create())
            ->get(route('targets.edit', $target))
            ->assertOk()
            ->assertInertia(fn ($page) => $page
                ->component('targets/edit')
                ->where('target.id', $target->id)
                ->where('target.title', $target->title)
                ->where('target.url', $target->url)
            );
    });

    it('requires authentication', function (): void {
        $target = Target::factory()->create();

        $this->get(route('targets.edit', $target))
            ->assertRedirect(route('login'));
    });
});

describe('targets update', function () use ($validSchema): void {
    it('updates the target', function () use ($validSchema): void {
        $target = Target::factory()->create();

        $this->actingAs(User::factory()->admin()->create())
            ->put(route('targets.update', $target), [
                'title' => 'Updated Title',
                'url' => 'https://updated.example.com',
                'schema' => $validSchema,
            ]);

        expect($target->fresh()->title)->toBe('Updated Title')
            ->and($target->fresh()->url)->toBe('https://updated.example.com');
    });

    it('stores the updated schema as structured data', function () use ($validSchema): void {
        $target = Target::factory()->create();

        $this->actingAs(User::factory()->admin()->create())
            ->put(route('targets.update', $target), [
                'title' => 'Updated Title',
                'url' => 'https://example.com',
                'schema' => $validSchema,
            ]);

        expect($target->fresh()->schema)->toBeArray()
            ->and($target->fresh()->schema['version'])->toBe('1');
    });

    it('redirects to the targets index after update', function () use ($validSchema): void {
        $target = Target::factory()->create();

        $this->actingAs(User::factory()->admin()->create())
            ->put(route('targets.update', $target), [
                'title' => 'Updated Title',
                'url' => 'https://example.com',
                'schema' => $validSchema,
            ])
            ->assertRedirect(route('targets.index'));
    });

    it('requires authentication', function () use ($validSchema): void {
        $target = Target::factory()->create();

        $this->put(route('targets.update', $target), [
            'title' => 'Updated Title',
            'url' => 'https://example.com',
            'schema' => $validSchema,
        ])->assertRedirect(route('login'));
    });

    it('validates the same rules as store', function (): void {
        $target = Target::factory()->create();
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->put(route('targets.update', $target), [])
            ->assertSessionHasErrors(['title', 'url', 'schema']);
    });
});
