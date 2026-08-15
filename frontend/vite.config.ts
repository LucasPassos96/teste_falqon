import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // Navegador e API ficam na mesma origem em desenvolvimento, o que
      // elimina CORS e os problemas de cookie que vêm junto.
      '/api': 'http://localhost:8080',
    },
  },
})
