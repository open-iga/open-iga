import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { tanstackRouter } from '@tanstack/router-plugin/vite';
import tailwindcss from '@tailwindcss/vite';
import { fileURLToPath } from 'node:url';

export default defineConfig({
    resolve: {
        // force a single React instance so base-ui subpath deps don't bundle their own copy
        dedupe: ['react', 'react-dom'],
        alias: {
            '@': fileURLToPath(new URL('./src', import.meta.url)),
        },
    },
    optimizeDeps: {
        include: ['@base-ui/react/collapsible'],
    },
    plugins: [
        tailwindcss(),
        // Please make sure that '@tanstack/router-plugin' is passed before '@vitejs/plugin-react'
        tanstackRouter({
            target: 'react',
            autoCodeSplitting: true,
        }),
        react(),
    ],
    server: {
        port: 3000,
        proxy: {
            '/api': {
                target: 'http://localhost:8080',
                changeOrigin: true,
                autoRewrite: true,
                selfHandleResponse: false,
            },
        },
    },
});
