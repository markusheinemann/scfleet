<?php

namespace App\Console\Commands;

use App\Models\Agent;
use App\Models\ScrapeJob;
use Illuminate\Console\Attributes\Description;
use Illuminate\Console\Attributes\Signature;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\DB;

#[Signature('scrape:requeue-timed-out')]
#[Description('Requeue scrape jobs whose agents have timed out')]
class RequeueTimedOutJobs extends Command
{
    public function handle(): int
    {
        DB::transaction(function (): void {
            $jobs = ScrapeJob::timedOut()
                ->lockForUpdate()
                ->get();

            foreach ($jobs as $job) {
                $agentId = $job->agent_id;

                if ($job->attempts >= $job->max_attempts) {
                    $job->forceFill([
                        'status' => 'failed',
                        'error_type' => 'agent_timeout',
                        'error_message' => 'Agent stopped responding before completing the job.',
                        'completed_at' => now(),
                    ])->save();
                } else {
                    $job->forceFill([
                        'status' => 'pending',
                        'agent_id' => null,
                        'claimed_at' => null,
                        'timeout_at' => null,
                    ])->save();
                }

                if ($agentId) {
                    Agent::whereKey($agentId)
                        ->where('status', 'processing')
                        ->update(['status' => 'ready']);
                }
            }
        });

        return Command::SUCCESS;
    }
}
