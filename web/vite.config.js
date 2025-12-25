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
            const content = readFileSync(filePath, 'utf-8');
            
            if (file.endsWith('.css')) {
              // Встраиваем CSS - ищем пути с ./assets/ или assets/
              const escapedFile = file.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
              const linkRegex = new RegExp(`<link[^>]*href=["']\\.?/?assets/${escapedFile}["'][^>]*>`, 'gi');
              if (linkRegex.test(inlinedHtml)) {
                inlinedHtml = inlinedHtml.replace(linkRegex, `<style>${content}</style>`);
                console.log(`Inlined CSS: ${file}`);
              }
            } else if (file.endsWith('.js')) {
              // Встраиваем JS - ищем пути с ./assets/ или assets/
              const escapedFile = file.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
              
              // СНАЧАЛА находим и заменяем оригинальный тег script в HTML
              // Пробуем найти тег script с закрывающим тегом
              let scriptRegex = new RegExp(`<script[^>]*src=["']\\.?/?assets/${escapedFile}["'][^>]*>\\s*<\\/script>`, 'gi');
              let matches = inlinedHtml.match(scriptRegex);
              let found = false;
              
              if (matches && matches.length > 0) {
                found = true;
              } else {
                // Пробуем найти тег script без закрывающего тега (самозакрывающийся)
                scriptRegex = new RegExp(`<script[^>]*src=["']\\.?/?assets/${escapedFile}["'][^>]*>`, 'gi');
                matches = inlinedHtml.match(scriptRegex);
                if (matches && matches.length > 0) {
                  found = true;
                } else {
                  // Последняя попытка - ищем любой тег script, который содержит имя файла
                  scriptRegex = new RegExp(`<script[^>]*src=["'][^"']*${escapedFile}[^"']*["'][^>]*>`, 'gi');
                  matches = inlinedHtml.match(scriptRegex);
                  if (matches && matches.length > 0) {
                    found = true;
                  }
                }
              }
              
              if (found) {
                // Теперь экранируем содержимое JavaScript ПЕРЕД встраиванием
                // Экранируем </script> чтобы он не интерпретировался как закрывающий тег
                const escapedContent = content.replace(/<\/script>/g, '<\\/script>');
                
                // Заменяем все найденные теги на встроенный JavaScript
                // Используем replaceAll для гарантированной замены всех вхождений
                scriptRegex.lastIndex = 0; // Сбрасываем индекс
                let replaced = false;
                inlinedHtml = inlinedHtml.replace(scriptRegex, (match) => {
                  replaced = true;
                  return `<script type="module">${escapedContent}</script>`;
                });
                
                if (replaced) {
                  console.log(`Inlined JS: ${file}`);
                } else {
                  console.warn(`Found tag but replacement failed for ${file}`);
                }
                
                // Проверяем, что замена произошла
                const remainingTags = (inlinedHtml.match(new RegExp(`<script[^>]*src=["'][^"']*${escapedFile}[^"']*["'][^>]*>`, 'gi')) || []).length;
                if (remainingTags > 0) {
                  console.warn(`Warning: ${remainingTags} script tags with src still remain for ${file}`);
                }
              } else {
                console.warn(`Could not find script tag for ${file}`);
              }
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
        console.log('Removed assets directory');
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

