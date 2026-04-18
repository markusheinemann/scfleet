import process from "node:process";
import { defineConfig } from "vite";
import laravel from "laravel-vite-plugin";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import inertia from "@inertiajs/vite";

export default defineConfig({
    plugins: [
        laravel({
            input: ["resources/css/app.css", "resources/js/app.jsx"],
            refresh: true,
        }),
        tailwindcss(),
        react(),
        inertia(),
    ],
    server: {
        hmr: {
            // When running inside Docker the container's hostname is not reachable
            // from the browser. VITE_HMR_HOST overrides the host written into
            // public/hot so Laravel and the browser use the correct address.
            // Does not affect CORS (unlike server.origin).
            host: process.env.VITE_HMR_HOST,
        },
        watch: {
            // Enable polling only in Docker (VITE_USE_POLLING=true) because macOS
            // bind-mounts don't fire inotify events reliably inside containers.
            // Polling adds CPU overhead and is not needed for native dev.
            usePolling: process.env.VITE_USE_POLLING === "true",
            interval: 500,
            ignored: ["**/storage/framework/views/**"],
        },
    },
});
