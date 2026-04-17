<?php

use App\Enums\Permission;
use App\Enums\Role;
use App\Models\User;
use Illuminate\Support\Facades\Route;

beforeEach(function (): void {
    Route::middleware(['web', 'auth', 'permission:manage-users'])
        ->get('/test-manage-users', fn () => 'ok')
        ->name('test.manage-users');

    Route::middleware(['web', 'auth', 'permission:view-users'])
        ->get('/test-view-users', fn () => 'ok')
        ->name('test.view-users');

    Route::middleware(['web', 'auth', 'permission:manage-users,view-users'])
        ->get('/test-any-users-permission', fn () => 'ok')
        ->name('test.any-users-permission');
});

describe('permission middleware', function (): void {
    it('allows admin to access manage-users route', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->get('/test-manage-users')
            ->assertOk();
    });

    it('denies editor from manage-users route', function (): void {
        $user = User::factory()->editor()->create();

        $this->actingAs($user)
            ->get('/test-manage-users')
            ->assertForbidden();
    });

    it('denies viewer from manage-users route', function (): void {
        $user = User::factory()->create();

        $this->actingAs($user)
            ->get('/test-manage-users')
            ->assertForbidden();
    });

    it('allows admin to access view-users route', function (): void {
        $user = User::factory()->admin()->create();

        $this->actingAs($user)
            ->get('/test-view-users')
            ->assertOk();
    });

    it('denies editor from view-users route', function (): void {
        $user = User::factory()->editor()->create();

        $this->actingAs($user)
            ->get('/test-view-users')
            ->assertForbidden();
    });

    it('denies viewer from view-users route', function (): void {
        $user = User::factory()->create();

        $this->actingAs($user)
            ->get('/test-view-users')
            ->assertForbidden();
    });

    it('allows access when user has any of the required permissions', function (): void {
        $admin = User::factory()->admin()->create();

        $this->actingAs($admin)
            ->get('/test-any-users-permission')
            ->assertOk();
    });

    it('denies access when user has none of the required permissions', function (): void {
        $editor = User::factory()->editor()->create();

        $this->actingAs($editor)
            ->get('/test-any-users-permission')
            ->assertForbidden();
    });

    it('returns 401 for unauthenticated requests', function (): void {
        $this->get('/test-manage-users')
            ->assertRedirect(route('login'));
    });
});

describe('user model permission helpers', function (): void {
    it('hasPermission returns true when role has the permission', function (): void {
        $admin = User::factory()->admin()->make();

        expect($admin->hasPermission(Permission::ManageUsers))->toBeTrue();
        expect($admin->hasPermission(Permission::ViewUsers))->toBeTrue();
    });

    it('hasPermission returns false when role lacks the permission', function (): void {
        $editor = User::factory()->editor()->make();
        $viewer = User::factory()->make();

        expect($editor->hasPermission(Permission::ManageUsers))->toBeFalse();
        expect($viewer->hasPermission(Permission::ViewUsers))->toBeFalse();
    });

    it('hasPermission returns true when any of the given permissions match', function (): void {
        $admin = User::factory()->admin()->make();

        expect($admin->hasPermission(Permission::ManageUsers, Permission::ViewUsers))->toBeTrue();
    });
});

describe('role permissions mapping', function (): void {
    it('admin has all permissions', function (): void {
        expect(Role::Admin->permissions())->toBe([
            Permission::ManageUsers,
            Permission::ViewUsers,
            Permission::ManageAgents,
            Permission::ViewAgents,
            Permission::RegenerateAgentToken,
        ]);
    });

    it('editor has agent permissions', function (): void {
        expect(Role::Editor->permissions())->toBe([Permission::ManageAgents, Permission::ViewAgents]);
    });

    it('viewer can only view agents', function (): void {
        expect(Role::Viewer->permissions())->toBe([Permission::ViewAgents]);
    });
});
