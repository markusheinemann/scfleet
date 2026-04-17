<?php

namespace App\Enums;

enum Role: string
{
    case Admin = 'admin';
    case Editor = 'editor';
    case Viewer = 'viewer';

    /** @return Permission[] */
    public function permissions(): array
    {
        return match ($this) {
            Role::Admin => [Permission::ManageUsers, Permission::ViewUsers],
            Role::Editor => [],
            Role::Viewer => [],
        };
    }
}
