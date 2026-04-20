<?php

namespace App\Policies;

use App\Enums\Permission;
use App\Models\ApiKey;
use App\Models\User;

class ApiKeyPolicy
{
    public function viewAny(User $user): bool
    {
        return $user->hasPermission(Permission::ManageApiKeys);
    }

    public function create(User $user): bool
    {
        return $user->hasPermission(Permission::ManageApiKeys);
    }

    public function delete(User $user, ApiKey $apiKey): bool
    {
        return $user->hasPermission(Permission::ManageApiKeys) && $apiKey->user_id === $user->id;
    }
}
