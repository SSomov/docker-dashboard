import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { fileURLToPath } from 'url';
import { resolve, dirname } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

export default defineConfig({
  plugins: [svelte()],
  root: './src',
  build: {
    outDir: resolve(__dirname, 'public'),
    emptyOutDir: true
  },
  base: './'
});

