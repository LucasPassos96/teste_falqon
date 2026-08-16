import { AppBar, Box, Button, Container, Stack, Toolbar, Typography } from '@mui/material'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'

import { logoutMutation } from '../api/gen/@tanstack/react-query.gen'
import { useSession } from '../auth/useSession'
import { cores, mono } from '../theme'

export function AppLayout({ children }: { children: ReactNode }) {
  const { user } = useSession()
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  const sair = useMutation({
    ...logoutMutation(),
    onSuccess: () => {
      // Sem isto o cache do TanStack Query sobrevive ao logout: o próximo
      // login na mesma máquina exibiria por um instante os formulários do
      // usuário anterior. É vazamento real entre contas.
      queryClient.clear()
      navigate('/login', { replace: true })
    },
  })

  return (
    <Box sx={{ minHeight: '100vh', bgcolor: 'background.default' }}>
      <AppBar position="static">
        <Toolbar sx={{ minHeight: { xs: 60, sm: 66 } }}>
          <Stack
            direction="row"
            spacing={1.25}
            sx={{ alignItems: 'baseline', flexGrow: 1, cursor: 'pointer' }}
            onClick={() => navigate('/')}
          >
            <Typography sx={{ fontWeight: 600, letterSpacing: '-0.02em', fontSize: '1.05rem' }}>
              Form Builder
            </Typography>
          </Stack>

          {user && (
            <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
              <Typography
                sx={{
                  fontFamily: mono,
                  fontSize: 12.5,
                  color: cores.tintaFraca,
                  display: { xs: 'none', sm: 'block' },
                }}
              >
                {user.email}
              </Typography>
              <Button onClick={() => sair.mutate({})} disabled={sair.isPending} size="small">
                Sair
              </Button>
            </Stack>
          )}
        </Toolbar>
      </AppBar>

      <Container maxWidth="md" sx={{ py: { xs: 4, sm: 6 } }}>
        {children}
      </Container>
    </Box>
  )
}
