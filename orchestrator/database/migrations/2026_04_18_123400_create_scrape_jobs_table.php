<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('scrape_jobs', function (Blueprint $table) {
            $table->id();
            $table->char('ulid', 26)->unique();
            $table->string('url', 2048);
            $table->json('template');
            $table->char('template_hash', 64)->nullable()->index();
            $table->string('status', 20)->default('pending');
            $table->foreignId('agent_id')->nullable()->constrained()->nullOnDelete();
            $table->timestamp('claimed_at')->nullable();
            $table->timestamp('completed_at')->nullable();
            $table->json('result')->nullable();
            $table->json('field_errors')->nullable();
            $table->string('error_type', 50)->nullable();
            $table->text('error_message')->nullable();
            $table->unsignedTinyInteger('attempts')->default(0);
            $table->unsignedTinyInteger('max_attempts')->default(3);
            $table->timestamp('timeout_at')->nullable();
            $table->timestamps();
        });

        if (DB::getDriverName() === 'pgsql') {
            DB::statement("CREATE INDEX idx_scrape_jobs_pending ON scrape_jobs (created_at ASC) WHERE status = 'pending'");
            DB::statement("CREATE INDEX idx_scrape_jobs_timeout ON scrape_jobs (timeout_at ASC) WHERE status = 'processing'");
        } else {
            // MySQL does not support partial indexes; use plain composite indexes instead.
            DB::statement('CREATE INDEX idx_scrape_jobs_pending ON scrape_jobs (status, created_at)');
            DB::statement('CREATE INDEX idx_scrape_jobs_timeout ON scrape_jobs (status, timeout_at)');
        }
    }

    public function down(): void
    {
        Schema::dropIfExists('scrape_jobs');
    }
};
