<?php

namespace App\Http\Middleware;

use App\Enums\Permission;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class PermissionMiddleware
{
    /**
     * @param  Closure(Request): (Response)  $next
     */
    public function handle(Request $request, Closure $next, string ...$permissions): Response
    {
        if (! $request->user()) {
            abort(401);
        }

        foreach ($permissions as $permission) {
            if ($request->user()->hasPermission(Permission::from($permission))) {
                return $next($request);
            }
        }

        abort(403);
    }
}
