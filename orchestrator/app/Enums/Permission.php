<?php

namespace App\Enums;

enum Permission: string
{
    case ManageUsers = 'manage-users';
    case ViewUsers = 'view-users';
}
