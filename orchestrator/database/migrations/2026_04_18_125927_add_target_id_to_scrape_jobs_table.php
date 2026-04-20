<?php

use App\Models\Template;
use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     */
    public function up(): void
    {
        Schema::table('scrape_jobs', function (Blueprint $table) {
            $table->foreignId('template_id')->nullable()->after('agent_id')->constrained()->nullOnDelete();
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::table('scrape_jobs', function (Blueprint $table) {
            $table->dropForeignIdFor(Template::class);
            $table->dropColumn('template_id');
        });
    }
};
