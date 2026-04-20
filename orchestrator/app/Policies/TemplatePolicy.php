<?php

namespace App\Policies;

use App\Enums\Permission;
use App\Models\Template;
use App\Models\User;

class TemplatePolicy
{
    public function viewAny(User $user): bool
    {
        return $user->hasPermission(Permission::ViewTemplates);
    }

    public function create(User $user): bool
    {
        return $user->hasPermission(Permission::ManageTemplates);
    }

    public function view(User $user, Template $template): bool
    {
        return $user->hasPermission(Permission::ViewTemplates);
    }

    public function update(User $user, Template $template): bool
    {
        return $user->hasPermission(Permission::ManageTemplates);
    }
}
