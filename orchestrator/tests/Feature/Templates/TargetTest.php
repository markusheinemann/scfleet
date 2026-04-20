<?php

use App\Models\Template;
use App\Models\User;

$validTemplate = json_encode([
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

describe('templates index', function (): void {
    it('shows the templates page', function (): void {
        $this->actingAs(User::factory()->create())
            ->get(route('templates.index'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page->component('templates/index'));
    });

    it('requires authentication', function (): void {
        $this->get(route('templates.index'))
            ->assertRedirect(route('login'));
    });

    it('lists all templates', function (): void {
        Template::factory(3)->create();

        $this->actingAs(User::factory()->create())
            ->get(route('templates.index'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page
                ->component('templates/index')
                ->has('templates', 3)
            );
    });
});

describe('templates create', function (): void {
    it('shows the create template page', function (): void {
        $this->actingAs(User::factory()->admin()->create())
            ->get(route('templates.create'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page->component('templates/create'));
    });

    it('requires authentication', function (): void {
        $this->get(route('templates.create'))
            ->assertRedirect(route('login'));
    });

    it('forbids viewer from accessing create page', function (): void {
        $this->actingAs(User::factory()->create())
            ->get(route('templates.create'))
            ->assertForbidden();
    });
});

describe('templates store', function () use ($validTemplate): void {
    it('creates a new template', function () use ($validTemplate): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)->post(route('templates.store'), [
            'title' => 'My Template',
            'template' => $validTemplate,
        ]);

        expect(Template::count())->toBe(1)
            ->and(Template::first()->title)->toBe('My Template')
            ->and(Template::first()->user_id)->toBe($user->id);
    });

    it('stores the template as structured data', function () use ($validTemplate): void {
        $this->actingAs(User::factory()->admin()->create())->post(route('templates.store'), [
            'title' => 'My Template',
            'template' => $validTemplate,
        ]);

        expect(Template::first()->template)->toBeArray()
            ->and(Template::first()->template['version'])->toBe('1');
    });

    it('redirects to the templates index after creation', function () use ($validTemplate): void {
        $this->actingAs(User::factory()->admin()->create())
            ->post(route('templates.store'), [
                'title' => 'My Template',
                'template' => $validTemplate,
            ])
            ->assertRedirect(route('templates.index'));
    });

    it('requires authentication', function () use ($validTemplate): void {
        $this->post(route('templates.store'), [
            'title' => 'My Template',
            'template' => $validTemplate,
        ])->assertRedirect(route('login'));
    });

    it('requires a title', function () use ($validTemplate): void {
        $this->actingAs(User::factory()->admin()->create())
            ->post(route('templates.store'), [
                'template' => $validTemplate,
            ])
            ->assertSessionHasErrors('title');
    });

    it('requires a template', function (): void {
        $this->actingAs(User::factory()->admin()->create())
            ->post(route('templates.store'), [
                'title' => 'My Template',
            ])
            ->assertSessionHasErrors('template');
    });

    it('requires template to be valid json', function (): void {
        $this->actingAs(User::factory()->admin()->create())
            ->post(route('templates.store'), [
                'title' => 'My Template',
                'template' => 'not-json',
            ])
            ->assertSessionHasErrors('template');
    });
});

describe('templates template validation', function (): void {
    it('rejects template with wrong version', function (): void {
        $template = json_encode(['version' => '2', 'fields' => [['name' => 'title', 'type' => 'string', 'extractors' => [['strategy' => 'css', 'selector' => 'h1']]]]]);

        $this->actingAs(User::factory()->admin()->create())
            ->post(route('templates.store'), [
                'title' => 'My Template',
                'template' => $template,
            ])
            ->assertSessionHasErrors('template');
    });

    it('rejects template with empty fields array', function (): void {
        $template = json_encode(['version' => '1', 'fields' => []]);

        $this->actingAs(User::factory()->admin()->create())
            ->post(route('templates.store'), [
                'title' => 'My Template',
                'template' => $template,
            ])
            ->assertSessionHasErrors('template');
    });

    it('rejects template with invalid field name', function (): void {
        $template = json_encode([
            'version' => '1',
            'fields' => [
                ['name' => 'Invalid Name!', 'type' => 'string', 'extractors' => [['strategy' => 'css', 'selector' => 'h1']]],
            ],
        ]);

        $this->actingAs(User::factory()->admin()->create())
            ->post(route('templates.store'), [
                'title' => 'My Template',
                'template' => $template,
            ])
            ->assertSessionHasErrors('template');
    });

    it('rejects template with unknown top-level key', function (): void {
        $template = json_encode([
            'version' => '1',
            'fields' => [['name' => 'title', 'type' => 'string', 'extractors' => [['strategy' => 'css', 'selector' => 'h1']]]],
            'unknown_key' => 'value',
        ]);

        $this->actingAs(User::factory()->admin()->create())
            ->post(route('templates.store'), [
                'title' => 'My Template',
                'template' => $template,
            ])
            ->assertSessionHasErrors('template');
    });

    it('rejects template with empty extractors', function (): void {
        $template = json_encode([
            'version' => '1',
            'fields' => [
                ['name' => 'title', 'type' => 'string', 'extractors' => []],
            ],
        ]);

        $this->actingAs(User::factory()->admin()->create())
            ->post(route('templates.store'), [
                'title' => 'My Template',
                'template' => $template,
            ])
            ->assertSessionHasErrors('template');
    });
});

describe('template permissions', function () use ($validTemplate): void {
    it('allows admin to view templates', function (): void {
        $this->actingAs(User::factory()->admin()->create())
            ->get(route('templates.index'))
            ->assertOk();
    });

    it('allows editor to view templates', function (): void {
        $this->actingAs(User::factory()->editor()->create())
            ->get(route('templates.index'))
            ->assertOk();
    });

    it('allows viewer to view templates', function (): void {
        $this->actingAs(User::factory()->create())
            ->get(route('templates.index'))
            ->assertOk();
    });

    it('allows admin to create templates', function (): void {
        $this->actingAs(User::factory()->admin()->create())
            ->get(route('templates.create'))
            ->assertOk();
    });

    it('allows editor to create templates', function (): void {
        $this->actingAs(User::factory()->editor()->create())
            ->get(route('templates.create'))
            ->assertOk();
    });

    it('forbids viewer from creating templates', function (): void {
        $this->actingAs(User::factory()->create())
            ->get(route('templates.create'))
            ->assertForbidden();
    });

    it('forbids viewer from storing templates', function () use ($validTemplate): void {
        $this->actingAs(User::factory()->create())
            ->post(route('templates.store'), [
                'title' => 'My Template',
                'template' => $validTemplate,
            ])
            ->assertForbidden();
    });

    it('allows admin to edit templates', function (): void {
        $template = Template::factory()->create();

        $this->actingAs(User::factory()->admin()->create())
            ->get(route('templates.edit', $template))
            ->assertOk();
    });

    it('allows editor to edit templates', function (): void {
        $template = Template::factory()->create();

        $this->actingAs(User::factory()->editor()->create())
            ->get(route('templates.edit', $template))
            ->assertOk();
    });

    it('forbids viewer from editing templates', function (): void {
        $template = Template::factory()->create();

        $this->actingAs(User::factory()->create())
            ->get(route('templates.edit', $template))
            ->assertForbidden();
    });

    it('forbids viewer from updating templates', function () use ($validTemplate): void {
        $template = Template::factory()->create();

        $this->actingAs(User::factory()->create())
            ->put(route('templates.update', $template), [
                'title' => 'Updated',
                'template' => $validTemplate,
            ])
            ->assertForbidden();
    });
});

describe('templates edit', function (): void {
    it('shows the edit page with existing values', function (): void {
        $template = Template::factory()->create();

        $this->actingAs(User::factory()->admin()->create())
            ->get(route('templates.edit', $template))
            ->assertOk()
            ->assertInertia(fn ($page) => $page
                ->component('templates/edit')
                ->where('template.id', $template->id)
                ->where('template.title', $template->title)
            );
    });

    it('requires authentication', function (): void {
        $template = Template::factory()->create();

        $this->get(route('templates.edit', $template))
            ->assertRedirect(route('login'));
    });
});

describe('templates update', function () use ($validTemplate): void {
    it('updates the template', function () use ($validTemplate): void {
        $template = Template::factory()->create();

        $this->actingAs(User::factory()->admin()->create())
            ->put(route('templates.update', $template), [
                'title' => 'Updated Title',
                'template' => $validTemplate,
            ]);

        expect($template->fresh()->title)->toBe('Updated Title');
    });

    it('stores the updated template as structured data', function () use ($validTemplate): void {
        $template = Template::factory()->create();

        $this->actingAs(User::factory()->admin()->create())
            ->put(route('templates.update', $template), [
                'title' => 'Updated Title',
                'template' => $validTemplate,
            ]);

        expect($template->fresh()->template)->toBeArray()
            ->and($template->fresh()->template['version'])->toBe('1');
    });

    it('redirects to the templates index after update', function () use ($validTemplate): void {
        $template = Template::factory()->create();

        $this->actingAs(User::factory()->admin()->create())
            ->put(route('templates.update', $template), [
                'title' => 'Updated Title',
                'template' => $validTemplate,
            ])
            ->assertRedirect(route('templates.index'));
    });

    it('requires authentication', function () use ($validTemplate): void {
        $template = Template::factory()->create();

        $this->put(route('templates.update', $template), [
            'title' => 'Updated Title',
            'template' => $validTemplate,
        ])->assertRedirect(route('login'));
    });

    it('validates the same rules as store', function (): void {
        $template = Template::factory()->create();
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->put(route('templates.update', $template), [])
            ->assertSessionHasErrors(['title', 'template']);
    });
});
