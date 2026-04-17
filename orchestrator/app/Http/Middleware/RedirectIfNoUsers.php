<?php

namespace App\Http\Middleware;

use App\Models\User;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class RedirectIfNoUsers
{
    /**
     * @param  Closure(Request): (Response)  $next
     */
    public function handle(Request $request, Closure $next): Response
    {
        if (User::exists()) {
            return $next($request);
        }

        if ($request->routeIs('setup*')) {
            return $next($request);
        }

        return redirect()->route('setup');
    }
}
