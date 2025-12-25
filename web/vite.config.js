import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { fileURLToPath } from 'url';
import { resolve, dirname } from 'path';
import { readFileSync, writeFileSync, readdirSync, statSync, rmSync } from 'fs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Плагин для встраивания assets в HTML
function inlineAssets() {
  return {
    name: 'inline-assets',
    closeBundle() {
      const htmlPath = resolve(__dirname, 'public', 'index.html');
      const html = readFileSync(htmlPath, 'utf-8');
      const publicDir = resolve(__dirname, 'public');
      
      let inlinedHtml = html;
      
      // Находим все файлы в assets директории
      const assetsDir = resolve(publicDir, 'assets');
      try {
        const files = readdirSync(assetsDir);
        
        files.forEach(file => {
          const filePath = resolve(assetsDir, file);
          const stats = statSync(filePath);
          
          if (stats.isFile()) {
            const relativePath = `assets/${file}`;
            const content = readFileSync(filePath, 'utf-8');
            
            if (file.endsWith('.css')) {
              // Встраиваем CSS
              const linkRegex = new RegExp(`<link[^>]*href=["']${relativePath.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}["'][^>]*>`, 'gi');
              inlinedHtml = inlinedHtml.replace(linkRegex, `<style>${content}</style>`);
            } else if (file.endsWith('.js')) {
              // Встраиваем JS
              const scriptRegex = new RegExp(`<script[^>]*src=["']${relativePath.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}["'][^>]*><\\/script>`, 'gi');
              inlinedHtml = inlinedHtml.replace(scriptRegex, `<script type="module">${content}</script>`);
            }
          }
        });
      } catch (err) {
        console.warn('Could not inline assets:', err.message);
      }
      
      writeFileSync(htmlPath, inlinedHtml);
      
      // Удаляем директорию assets после встраивания
      try {
        rmSync(assetsDir, { recursive: true, force: true });
      } catch (err) {
        console.warn('Could not remove assets directory:', err.message);
      }
    }
  };
}

export default defineConfig({
  plugins: [svelte(), inlineAssets()],
  root: './src',
  build: {
    outDir: resolve(__dirname, 'public'),
    emptyOutDir: true
  },
  base: './'
});

