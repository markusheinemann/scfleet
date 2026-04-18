<?php

namespace App\Policies;

use App\Enums\Permission;
use App\Models\Target;
use App\Models\User;

class TargetPolicy
{
    public function viewAny(User $user): bool
    {
        return $user->hasPermission(Permission::ViewTargets);
    }

    public function create(User $user): bool
    {
        return $user->hasPermission(Permission::ManageTargets);
    }

    public function update(User $user, Target $target): bool
    {
        return $user->hasPermission(Permission::ManageTargets);
    }
}
