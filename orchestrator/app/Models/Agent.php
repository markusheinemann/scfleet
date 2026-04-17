<?php

namespace App\Models;

use Database\Factories\AgentFactory;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Attributes\Hidden;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Support\Carbon;

#[Fillable(['user_id', 'name', 'token'])]
#[Hidden(['token'])]
class Agent extends Model
{
    /** @use HasFactory<AgentFactory> */
    use HasFactory;

    protected $appends = ['is_online'];

    protected function casts(): array
    {
        return [
            'last_heartbeat_at' => 'datetime',
            'registered_at' => 'datetime',
        ];
    }

    public function getIsOnlineAttribute(): bool
    {
        if ($this->last_heartbeat_at === null) {
            return false;
        }

        $offlineAfter = (int) config('agent.offline_after', 120);

        return $this->last_heartbeat_at->gt(Carbon::now()->subSeconds($offlineAfter));
    }

    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }
}
