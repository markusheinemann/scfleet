<?php

use App\Models\User;
use Inertia\Testing\AssertableInertia;

test('the application returns a successful response', function () {
    $user = User::factory()->create();

    $this->actingAs($user)
        ->get('/')
        ->assertStatus(200)
        ->assertInertia(function (AssertableInertia $page) {
            $page->component('index');
        });
});
