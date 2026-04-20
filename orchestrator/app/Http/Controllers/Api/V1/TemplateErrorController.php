<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Models\ScrapeJob;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class TemplateErrorController extends Controller
{
    public function __invoke(Request $request): JsonResponse
    {
        $rows = ScrapeJob::query()
            ->selectRaw('
                template_hash,
                COUNT(*) AS total_jobs,
                SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS failed_jobs,
                MAX(created_at) AS last_seen_at
            ', ['failed'])
            ->whereNotNull('template_hash')
            ->groupBy('template_hash')
            ->havingRaw('SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) > 0', ['failed'])
            ->orderByRaw('failed_jobs DESC')
            ->get();

        $templateHashes = $rows->pluck('template_hash')->all();

        // Aggregate field_errors and error_types per template in PHP to avoid complex JSON SQL aggregation
        $fieldErrorCounts = [];
        $errorTypeCounts = [];

        ScrapeJob::query()
            ->whereIn('template_hash', $templateHashes)
            ->whereIn('status', ['failed', 'completed'])
            ->whereNotNull('field_errors')
            ->select(['template_hash', 'field_errors', 'error_type', 'status'])
            ->each(function (ScrapeJob $job) use (&$fieldErrorCounts, &$errorTypeCounts): void {
                $hash = $job->template_hash;

                if (is_array($job->field_errors)) {
                    foreach (array_keys($job->field_errors) as $field) {
                        $fieldErrorCounts[$hash][$field] = ($fieldErrorCounts[$hash][$field] ?? 0) + 1;
                    }
                }

                if ($job->status === 'failed' && $job->error_type) {
                    $errorTypeCounts[$hash][$job->error_type] = ($errorTypeCounts[$hash][$job->error_type] ?? 0) + 1;
                }
            });

        $data = $rows->map(function (object $row) use ($fieldErrorCounts, $errorTypeCounts): array {
            $hash = $row->template_hash;
            $topFieldErrors = $fieldErrorCounts[$hash] ?? [];
            arsort($topFieldErrors);

            return [
                'template_hash' => $hash,
                'total_jobs' => (int) $row->total_jobs,
                'failed_jobs' => (int) $row->failed_jobs,
                'failure_rate' => $row->total_jobs > 0 ? round($row->failed_jobs / $row->total_jobs, 4) : 0,
                'top_field_errors' => $topFieldErrors,
                'error_types' => $errorTypeCounts[$hash] ?? [],
                'last_seen_at' => $row->last_seen_at,
            ];
        });

        return response()->json(['data' => $data]);
    }
}
