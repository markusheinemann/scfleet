<?php

use App\Enums\Role;
use App\Models\User;

describe('setup redirect', function (): void {
    it('redirects to setup when no users exist', function (): void {
        $this->get(route('login'))
            ->assertRedirect(route('setup'));
    });

    it('does not redirect to setup when users exist', function (): void {
        User::factory()->create();

        $this->get(route('login'))
            ->assertOk();
    });

    it('shows the setup page when no users exist', function (): void {
        $this->get(route('setup'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page->component('auth/setup'));
    });

    it('redirects authenticated users away from setup to login', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->get(route('setup'))
            ->assertRedirect(route('login'));
    });
});

describe('setup store', function (): void {
    it('creates the first admin user and logs them in', function (): void {
        $this->post(route('setup.store'), [
            'username' => 'admin',
            'email' => 'admin@example.com',
            'password' => 'password123',
            'password_confirmation' => 'password123',
        ])->assertRedirect(route('dashboard'));

        $this->assertAuthenticated();

        $user = User::first();
        expect($user)->not->toBeNull()
            ->and($user->username)->toBe('admin')
            ->and($user->email)->toBe('admin@example.com')
            ->and($user->role)->toBe(Role::Admin);
    });

    it('requires all fields', function (): void {
        $this->post(route('setup.store'), [])
            ->assertSessionHasErrors(['username', 'email', 'password']);
    });

    it('requires password confirmation to match', function (): void {
        $this->post(route('setup.store'), [
            'username' => 'admin',
            'email' => 'admin@example.com',
            'password' => 'password123',
            'password_confirmation' => 'different',
        ])->assertSessionHasErrors('password');

        expect(User::count())->toBe(0);
    });

    it('requires a minimum password length of 8', function (): void {
        $this->post(route('setup.store'), [
            'username' => 'admin',
            'email' => 'admin@example.com',
            'password' => 'short',
            'password_confirmation' => 'short',
        ])->assertSessionHasErrors('password');
    });

    it('requires a valid email address', function (): void {
        $this->post(route('setup.store'), [
            'username' => 'admin',
            'email' => 'not-an-email',
            'password' => 'password123',
            'password_confirmation' => 'password123',
        ])->assertSessionHasErrors('email');
    });

    it('blocks setup when users already exist', function (): void {
        User::factory()->create();

        $this->post(route('setup.store'), [
            'username' => 'admin2',
            'email' => 'admin2@example.com',
            'password' => 'password123',
            'password_confirmation' => 'password123',
        ]);

        expect(User::count())->toBe(1);
    });
});
