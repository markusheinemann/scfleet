<?php

namespace Database\Seeders;

use App\Enums\Role;
use App\Models\Agent;
use App\Models\User;
use Illuminate\Database\Console\Seeds\WithoutModelEvents;
use Illuminate\Database\Seeder;
use Illuminate\Support\Facades\Hash;

class DatabaseSeeder extends Seeder
{
    use WithoutModelEvents;

    public function run(): void
    {
        $user = User::firstOrCreate(
            ['email' => 'test@example.com'],
            ['username' => 'admin', 'password' => Hash::make('password'), 'role' => Role::Admin],
        );

        $plainToken = env('DEV_AGENT_TOKEN', 'dev-agent-token');

        Agent::firstOrCreate(
            ['name' => 'Default Dev Agent'],
            ['user_id' => $user->id, 'token' => hash('sha256', $plainToken)],
        );
    }
}
