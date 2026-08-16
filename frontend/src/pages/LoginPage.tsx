import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Divider,
  Stack,
  Tab,
  Tabs,
  TextField,
  Typography,
} from '@mui/material'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Navigate, useLocation, useNavigate, useSearchParams } from 'react-router-dom'

import { loginMutation, registerMutation } from '../api/gen/@tanstack/react-query.gen'
import { useSession } from '../auth/useSession'
import { mensagemDeErro } from '../lib/erros'
import { cores } from '../theme'

/**
 * Motivos que o callback do Google pode devolver em ?erro=. Traduzir aqui, e
 * não exibir o código cru, é o que mantém a tela legível para quem não faz
 * ideia do que é OAuth.
 */
const ERROS_GOOGLE: Record<string, string> = {
  google_nao_configurado:
    'Login com Google não está configurado neste ambiente. Use e-mail e senha.',
  acesso_negado: 'Você cancelou o login com o Google.',
  state_invalido: 'A tentativa de login expirou ou veio de outra origem. Tente novamente.',
  email_nao_verificado:
    'Sua conta Google não tem o e-mail verificado, então não é possível vinculá-la.',
  resposta_invalida: 'O Google devolveu uma resposta inesperada. Tente novamente.',
}

export default function LoginPage() {
  const [searchParams] = useSearchParams()
  const erroGoogle = searchParams.get('erro')
  const [aba, setAba] = useState<'login' | 'cadastro'>('login')
  const [nome, setNome] = useState('')
  const [email, setEmail] = useState('')
  const [senha, setSenha] = useState('')

  const navigate = useNavigate()
  const location = useLocation()
  const queryClient = useQueryClient()
  const { isAuthenticated, isLoading } = useSession()

  const destino = (location.state as { from?: string } | null)?.from ?? '/'

  const aoEntrar = async () => {
    // Invalida a sessão para o useSession refazer GET /auth/me e o guard
    // liberar a rota com o usuário novo.
    await queryClient.invalidateQueries()
    navigate(destino, { replace: true })
  }

  const login = useMutation({ ...loginMutation(), onSuccess: aoEntrar })
  const cadastro = useMutation({ ...registerMutation(), onSuccess: aoEntrar })

  const emAndamento = login.isPending || cadastro.isPending
  const erro = login.error ?? cadastro.error

  if (!isLoading && isAuthenticated) {
    return <Navigate to="/" replace />
  }

  const enviar = (e: React.FormEvent) => {
    e.preventDefault()
    if (aba === 'login') {
      login.mutate({ body: { email, password: senha } })
    } else {
      cadastro.mutate({ body: { name: nome, email, password: senha } })
    }
  }

  return (
    <Box sx={{ maxWidth: 420, mx: 'auto', mt: { xs: 6, sm: 12 }, px: 2, pb: 8 }}>
      <Box sx={{ borderTop: `2px solid ${cores.tinta}`, pt: 2.5, mb: 4 }}>
        <Typography variant="h4">Form Builder</Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.75 }}>
          Monte formulários, publique um link, receba respostas.
        </Typography>
      </Box>

      <Card>
        <Tabs
          value={aba}
          onChange={(_, v) => setAba(v)}
          variant="fullWidth"
          sx={{ borderBottom: 1, borderColor: 'divider' }}
        >
          <Tab label="Entrar" value="login" />
          <Tab label="Criar conta" value="cadastro" />
        </Tabs>

        <CardContent sx={{ p: 3 }}>
          <form onSubmit={enviar}>
            <Stack spacing={2.5}>
              {aba === 'cadastro' && (
                <TextField
                  label="Nome"
                  value={nome}
                  onChange={(e) => setNome(e.target.value)}
                  required
                  fullWidth
                  autoFocus
                />
              )}

              <TextField
                label="E-mail"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                fullWidth
                autoFocus={aba === 'login'}
              />

              <TextField
                label="Senha"
                type="password"
                value={senha}
                onChange={(e) => setSenha(e.target.value)}
                required
                fullWidth
                helperText={aba === 'cadastro' ? 'Mínimo de 8 caracteres' : undefined}
              />

              {/* A mensagem exibida é a tratada pela API, nunca o payload cru:
                  nada de stack trace ou detalhe de banco na tela. */}
              {erroGoogle && (
                <Alert severity="warning">
                  {ERROS_GOOGLE[erroGoogle] ?? 'Não foi possível entrar com o Google.'}
                </Alert>
              )}

              {erro && <Alert severity="error">{mensagemDeErro(erro)}</Alert>}

              <Button type="submit" variant="contained" size="large" disabled={emAndamento}>
                {aba === 'login' ? 'Entrar' : 'Criar conta'}
              </Button>
            </Stack>
          </form>

          <Divider sx={{ my: 3 }}>ou</Divider>

          {/*
            Link comum, NÃO uma chamada do client TypeScript.

            O fluxo OAuth é uma navegação de navegador: o backend redireciona
            para o Google, o Google redireciona de volta para o backend, e só
            então o navegador volta para cá com o cookie de sessão. Um fetch
            não conseguiria seguir esses redirecionamentos entre origens — e é
            justamente por isso que o client_secret nunca precisa existir aqui.
          */}
          <Button
            component="a"
            href="/api/v1/auth/google"
            variant="outlined"
            size="large"
            fullWidth
          >
            Entrar com Google
          </Button>
        </CardContent>
      </Card>
    </Box>
  )
}
