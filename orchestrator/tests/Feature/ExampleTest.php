<?php

use Inertia\Testing\AssertableInertia;

test('the application returns a successful response', function () {
    $response = $this->get('/');

    $response->assertStatus(200)
        ->assertInertia(function (AssertableInertia $page) {
            $page->component('index');
        });
});
