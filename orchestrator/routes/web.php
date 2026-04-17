<?php

use App\Http\Controllers\SetupController;
use Illuminate\Support\Facades\Route;
use Inertia\Inertia;

Route::get('/setup', [SetupController::class, 'create'])->name('setup');
Route::post('/setup', [SetupController::class, 'store'])->name('setup.store');

Route::middleware('auth')->group(function (): void {
    Route::get('/', function () {
        return Inertia::render('index');
    })->name('dashboard');
});
