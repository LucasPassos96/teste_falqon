import { Box, CircularProgress } from '@mui/material'
import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'

import { useSession } from './useSession'

/**
 * Guard de rota — que é UX, NÃO segurança.
 *
 * Ele evita que a pessoa veja uma tela quebrada e evita o flash de interface
 * administrativa antes do redirecionamento. Nada além disso: qualquer um edita
 * o JavaScript no navegador e "entra" na rota. Quem protege de verdade é o
 * middleware no Go — sem cookie válido, toda rota administrativa responde 401
 * e a tela renderiza vazia.
 */
export function RequireAuth({ children }: { children: ReactNode }) {
  const { isLoading, isAuthenticated } = useSession()
  const location = useLocation()

  if (isLoading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', mt: 8 }}>
        <CircularProgress />
      </Box>
    )
  }

  if (!isAuthenticated) {
    // `state` guarda de onde a pessoa veio, para voltar ali depois do login.
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }

  return <>{children}</>
}
