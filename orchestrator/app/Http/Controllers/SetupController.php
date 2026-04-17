<?php

namespace App\Http\Controllers;

use App\Enums\Role;
use App\Http\Requests\SetupRequest;
use App\Models\User;
use Illuminate\Http\RedirectResponse;
use Illuminate\Support\Facades\Auth;
use Inertia\Inertia;
use Inertia\Response;

class SetupController extends Controller
{
    public function create(): Response|RedirectResponse
    {
        if (User::exists()) {
            return redirect()->route('login');
        }

        return Inertia::render('auth/setup');
    }

    public function store(SetupRequest $request): RedirectResponse
    {
        if (User::exists()) {
            return redirect()->route('login');
        }

        $user = User::create([
            'username' => $request->validated('username'),
            'email' => $request->validated('email'),
            'password' => $request->validated('password'),
            'role' => Role::Admin->value,
        ]);

        Auth::login($user);

        $request->session()->regenerate();

        return redirect()->route('dashboard');
    }
}
