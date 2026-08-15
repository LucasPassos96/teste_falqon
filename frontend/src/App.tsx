import { useQuery } from '@tanstack/react-query'

import { getHealthOptions } from './api/gen/@tanstack/react-query.gen'

// Prova do pipeline: este hook e o tipo de `data` vêm de src/api/gen, que é
// gerado a partir de api/openapi.yaml — a mesma spec que gera o servidor Go.
export default function App() {
  const { data, isPending, error } = useQuery(getHealthOptions())

  return (
    <main style={{ fontFamily: 'system-ui, sans-serif', padding: '2rem' }}>
      <h1>Form Builder</h1>
      {isPending && <p>Consultando a API…</p>}
      {error && <p>API indisponível: {error.message}</p>}
      {data && (
        <p>
          A API respondeu: <strong>{data.status}</strong>
        </p>
      )}
    </main>
  )
}
