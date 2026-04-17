<?php

use App\Http\Controllers\Api\V1\HeartbeatController;
use App\Http\Controllers\Api\V1\RegisterController;
use App\Http\Middleware\AgentTokenMiddleware;
use Illuminate\Support\Facades\Route;

Route::middleware(AgentTokenMiddleware::class)->group(function (): void {
    Route::post('register', RegisterController::class);
    Route::post('heartbeat', HeartbeatController::class);
});
