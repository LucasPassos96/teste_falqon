import { defineConfig } from '@hey-api/openapi-ts'

// Gera tipos, client e hooks do TanStack Query a partir da MESMA spec que
// gera o servidor Go. Nada em src/api/gen é escrito à mão.
export default defineConfig({
  input: '../api/openapi.yaml',
  output: 'src/api/gen',
  plugins: ['@hey-api/client-fetch', '@tanstack/react-query'],
})
