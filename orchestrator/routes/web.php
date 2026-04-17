<?php

use App\Http\Controllers\AgentController;
use App\Http\Controllers\SetupController;
use Illuminate\Support\Facades\Route;
use Inertia\Inertia;

Route::get('/setup', [SetupController::class, 'create'])->name('setup');
Route::post('/setup', [SetupController::class, 'store'])->name('setup.store');

Route::middleware('auth')->group(function (): void {
    Route::get('/', function () {
        return Inertia::render('index');
    })->name('dashboard');

    Route::get('agents/create', [AgentController::class, 'create'])->name('agents.create');
    Route::get('agents', [AgentController::class, 'index'])->name('agents.index');
    Route::post('agents', [AgentController::class, 'store'])->name('agents.store');
    Route::get('agents/{agent}', [AgentController::class, 'show'])->name('agents.show');
    Route::post('agents/{agent}/regenerate-token', [AgentController::class, 'regenerateToken'])
        ->middleware('throttle:10,1')
        ->name('agents.regenerate-token');
});
