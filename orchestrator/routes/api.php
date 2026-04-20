<?php

use App\Http\Controllers\Api\V1\HeartbeatController;
use App\Http\Controllers\Api\V1\JobArtifactsController;
use App\Http\Controllers\Api\V1\JobClaimController;
use App\Http\Controllers\Api\V1\JobCompleteController;
use App\Http\Controllers\Api\V1\JobFailController;
use App\Http\Controllers\Api\V1\RegisterController;
use App\Http\Controllers\Api\V1\ScrapeController;
use App\Http\Controllers\Api\V1\TemplateErrorController;
use App\Http\Middleware\AgentTokenMiddleware;
use Illuminate\Support\Facades\Route;

Route::middleware(AgentTokenMiddleware::class)->group(function (): void {
    Route::post('register', RegisterController::class);
    Route::post('heartbeat', HeartbeatController::class);

    Route::post('jobs/claim', JobClaimController::class);
    Route::post('jobs/{scrapeJob}/complete', JobCompleteController::class);
    Route::post('jobs/{scrapeJob}/fail', JobFailController::class);
    Route::post('jobs/{scrapeJob}/artifacts', JobArtifactsController::class);
});

Route::middleware('api.key')->group(function (): void {
    Route::post('scrape', [ScrapeController::class, 'store'])->name('scrape.store');
    Route::get('scrape/{scrapeJob:ulid}', [ScrapeController::class, 'show'])->name('scrape.show');
    Route::get('scrape/templates/errors', TemplateErrorController::class)->name('scrape.templates.errors');
});
