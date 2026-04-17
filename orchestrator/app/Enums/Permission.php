<?php

namespace App\Enums;

enum Permission: string
{
    case ManageUsers = 'manage-users';
    case ViewUsers = 'view-users';
    case ManageAgents = 'manage-agents';
    case ViewAgents = 'view-agents';
    case RegenerateAgentToken = 'regenerate-agent-token';
}
