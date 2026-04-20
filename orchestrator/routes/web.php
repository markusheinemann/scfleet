<?php

use App\Http\Controllers\AgentController;
use App\Http\Controllers\ScrapeJobArtifactController;
use App\Http\Controllers\SetupController;
use App\Http\Controllers\TemplateController;
use App\Http\Controllers\TemplateJobController;
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

    Route::get('templates/create', [TemplateController::class, 'create'])->name('templates.create');
    Route::get('templates', [TemplateController::class, 'index'])->name('templates.index');
    Route::post('templates', [TemplateController::class, 'store'])->name('templates.store');
    Route::get('templates/{template}', [TemplateController::class, 'show'])->name('templates.show');
    Route::get('templates/{template}/edit', [TemplateController::class, 'edit'])->name('templates.edit');
    Route::put('templates/{template}', [TemplateController::class, 'update'])->name('templates.update');
    Route::post('templates/{template}/jobs', [TemplateJobController::class, 'store'])->name('templates.jobs.store');

    Route::get('scrape-jobs/{scrapeJob}/screenshot', [ScrapeJobArtifactController::class, 'screenshot'])->name('scrape-jobs.screenshot');
    Route::get('scrape-jobs/{scrapeJob}/html', [ScrapeJobArtifactController::class, 'html'])->name('scrape-jobs.html');
});
