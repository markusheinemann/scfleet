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
            Role::Admin => [Permission::ManageUsers, Permission::ViewUsers, Permission::ManageAgents, Permission::ViewAgents, Permission::RegenerateAgentToken, Permission::ManageTemplates, Permission::ViewTemplates],
            Role::Editor => [Permission::ManageAgents, Permission::ViewAgents, Permission::ManageTemplates, Permission::ViewTemplates],
            Role::Viewer => [Permission::ViewAgents, Permission::ViewTemplates],
        };
    }
}
