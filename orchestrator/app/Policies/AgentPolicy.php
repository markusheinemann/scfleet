<?php

namespace App\Policies;

use App\Enums\Permission;
use App\Models\Agent;
use App\Models\User;

class AgentPolicy
{
    public function viewAny(User $user): bool
    {
        return $user->hasPermission(Permission::ViewAgents);
    }

    public function view(User $user, Agent $agent): bool
    {
        return $user->hasPermission(Permission::ViewAgents);
    }

    public function create(User $user): bool
    {
        return $user->hasPermission(Permission::ManageAgents);
    }

    public function regenerateToken(User $user, Agent $agent): bool
    {
        return $user->hasPermission(Permission::RegenerateAgentToken);
    }
}
