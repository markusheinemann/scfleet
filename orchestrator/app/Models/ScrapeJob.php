<?php

namespace App\Models;

use Database\Factories\ScrapeJobFactory;
use Illuminate\Database\Eloquent\Attributes\Fillable;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\Concerns\HasUlids;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

#[Fillable([
    'url', 'template', 'template_hash', 'status', 'agent_id', 'template_id',
    'claimed_at', 'completed_at', 'result', 'field_errors',
    'error_type', 'error_message', 'attempts', 'max_attempts', 'timeout_at', 'has_artifacts',
])]
class ScrapeJob extends Model
{
    /** @use HasFactory<ScrapeJobFactory> */
    use HasFactory, HasUlids;

    protected $primaryKey = 'ulid';

    public $incrementing = false;

    protected $keyType = 'string';

    protected function casts(): array
    {
        return [
            'template' => 'array',
            'result' => 'array',
            'field_errors' => 'array',
            'claimed_at' => 'datetime',
            'completed_at' => 'datetime',
            'timeout_at' => 'datetime',
            'has_artifacts' => 'boolean',
        ];
    }

    protected static function boot(): void
    {
        parent::boot();

        static::creating(function (ScrapeJob $job): void {
            if ($job->template) {
                $job->template_hash = hash('sha256', json_encode(self::sortedKeys($job->template), \JSON_UNESCAPED_UNICODE));
            }
        });
    }

    private static function sortedKeys(array $value): array
    {
        ksort($value);

        foreach ($value as &$v) {
            if (is_array($v)) {
                $v = self::sortedKeys($v);
            }
        }

        return $value;
    }

    public function agent(): BelongsTo
    {
        return $this->belongsTo(Agent::class);
    }

    public function scrapeTemplate(): BelongsTo
    {
        return $this->belongsTo(Template::class);
    }

    public function scopePending(Builder $query): void
    {
        $query->where('status', 'pending');
    }

    public function scopeTimedOut(Builder $query): void
    {
        $query->where('status', 'processing')
            ->where('timeout_at', '<', now());
    }
}
