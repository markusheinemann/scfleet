<?php

return [
    /*
     * How often agents are expected to send a heartbeat, in seconds.
     * Used to inform operators of the expected check-in interval.
     */
    'heartbeat_interval' => (int) env('AGENT_HEARTBEAT_INTERVAL', 30),

    /*
     * Number of seconds after the last heartbeat before an agent is
     * considered offline.
     */
    'offline_after' => (int) env('AGENT_OFFLINE_AFTER', 120),
];
