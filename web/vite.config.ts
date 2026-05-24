import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  base: './',
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@philiaspace/ui-primitives': path.resolve(__dirname, '../../../libs/phi-ui-primitives/src/index.ts'),
      '@philiaspace/phi-dashboard': path.resolve(__dirname, '../../../libs/phi-dashboard/src/index.ts'),
    },
  },
});
