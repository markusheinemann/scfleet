<?php

use App\Models\User;

describe('login page', function (): void {
    it('shows the login page', function (): void {
        User::factory()->create();

        $this->get(route('login'))
            ->assertOk()
            ->assertInertia(fn ($page) => $page->component('auth/login'));
    });

    it('redirects authenticated users away from login', function (): void {
        $user = User::factory()->create();

        $this->actingAs($user)
            ->get(route('login'))
            ->assertRedirect(route('dashboard'));
    });
});

describe('login', function (): void {
    it('authenticates a user with valid credentials', function (): void {
        $user = User::factory()->create();

        $this->post(route('login.store'), [
            'email' => $user->email,
            'password' => 'password',
        ])->assertRedirect(route('dashboard'));

        $this->assertAuthenticatedAs($user);
    });

    it('fails with invalid password', function (): void {
        $user = User::factory()->create();

        $this->post(route('login.store'), [
            'email' => $user->email,
            'password' => 'wrong-password',
        ])->assertRedirect()
            ->assertSessionHasErrors('email');

        $this->assertGuest();
    });

    it('fails with unknown email', function (): void {
        User::factory()->create();

        $this->post(route('login.store'), [
            'email' => 'nobody@example.com',
            'password' => 'password',
        ])->assertRedirect()
            ->assertSessionHasErrors('email');

        $this->assertGuest();
    });

    it('requires email and password', function (): void {
        User::factory()->create();

        $this->post(route('login.store'), [])
            ->assertSessionHasErrors(['email', 'password']);
    });

    it('rejects a non-email value in the email field', function (): void {
        User::factory()->create();

        $this->post(route('login.store'), [
            'email' => 'not-an-email',
            'password' => 'password',
        ])->assertSessionHasErrors('email');
    });

    it('regenerates the session on login', function (): void {
        $user = User::factory()->create();

        $this->post(route('login.store'), [
            'email' => $user->email,
            'password' => 'password',
        ]);

        $this->assertAuthenticatedAs($user);
    });
});

describe('login throttle', function (): void {
    it('blocks login after 5 failed attempts', function (): void {
        $user = User::factory()->create();

        foreach (range(1, 5) as $attempt) {
            $this->post(route('login.store'), [
                'email' => $user->email,
                'password' => 'wrong-password',
            ])->assertRedirect();
        }

        $this->post(route('login.store'), [
            'email' => $user->email,
            'password' => 'wrong-password',
        ])->assertStatus(429);
    });

    it('still blocks after 5 attempts even with correct credentials', function (): void {
        $user = User::factory()->create();

        foreach (range(1, 5) as $attempt) {
            $this->post(route('login.store'), [
                'email' => $user->email,
                'password' => 'wrong-password',
            ]);
        }

        $this->post(route('login.store'), [
            'email' => $user->email,
            'password' => 'password',
        ])->assertStatus(429);
    });
});

describe('logout', function (): void {
    it('logs out an authenticated user', function (): void {
        $user = User::factory()->create();

        $this->actingAs($user)
            ->post(route('logout'))
            ->assertRedirect(route('login'));

        $this->assertGuest();
    });

    it('redirects guests attempting to logout', function (): void {
        User::factory()->create();

        $this->post(route('logout'))
            ->assertRedirect(route('login'));
    });
});

describe('protected routes', function (): void {
    it('redirects unauthenticated users to login', function (): void {
        User::factory()->create();

        $this->get(route('dashboard'))
            ->assertRedirect(route('login'));
    });

    it('allows authenticated users to access the dashboard', function (): void {
        $user = User::factory()->create();

        $this->actingAs($user)
            ->get(route('dashboard'))
            ->assertOk();
    });
});
