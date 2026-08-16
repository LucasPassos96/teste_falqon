import { Box, CircularProgress } from '@mui/material'
import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'

import { useSession } from './useSession'

/**
 * Guard de rota — UX, não segurança. Evita a tela quebrada e o flash de
 * interface antes do redirect; quem protege é o middleware no Go.
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
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }

  return <>{children}</>
}
