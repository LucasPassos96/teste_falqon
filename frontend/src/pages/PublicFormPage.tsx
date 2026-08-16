import { Alert, Box, Button, CircularProgress, Container, Stack, Typography } from '@mui/material'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { useParams } from 'react-router-dom'

import {
  getPublicFormOptions,
  submitPublicFormMutation,
} from '../api/gen/@tanstack/react-query.gen'
import { FieldRenderer } from '../components/render/FieldRenderer'
import { errosPorCampo, mensagemDeErro } from '../lib/erros'
import { cores, mono } from '../theme'

export default function PublicFormPage() {
  const { slug = '' } = useParams()
  const [valores, setValores] = useState<Record<string, string>>({})
  const [enviado, setEnviado] = useState(false)

  const {
    data: form,
    isPending,
    error,
  } = useQuery({ ...getPublicFormOptions({ path: { slug } }), retry: false })

  const enviar = useMutation({
    ...submitPublicFormMutation(),
    onSuccess: () => {
      setEnviado(true)
      window.scrollTo({ top: 0 })
    },
  })

  if (isPending) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', mt: 12 }}>
        <CircularProgress size={28} />
      </Box>
    )
  }

  // Inexistente e rascunho respondem 404 igualmente.
  if (error || !form) {
    return (
      <Container maxWidth="sm" sx={{ mt: 12 }}>
        <Typography variant="h5" gutterBottom>
          Formulário indisponível
        </Typography>
        <Typography color="text.secondary">
          Este link não está ativo. Confirme o endereço com quem o enviou.
        </Typography>
      </Container>
    )
  }

  if (enviado) {
    return (
      <Container maxWidth="sm" sx={{ py: 12 }}>
        <Box sx={{ borderTop: `2px solid ${cores.carimbo}`, pt: 3 }}>
          <Typography variant="h4" gutterBottom>
            Resposta registrada
          </Typography>
          <Typography color="text.secondary">
            Obrigado por preencher {form.title}. Você já pode fechar esta página.
          </Typography>
        </Box>
      </Container>
    )
  }

  // Os erros por campo vêm do 422 do backend, indexados por field_id.
  const erros = errosPorCampo(enviar.error)
  const total = form.fields.length

  return (
    <Container maxWidth="sm" sx={{ py: { xs: 5, sm: 9 } }}>
      {/* A contagem vem antes do título: quem abre um formulário quer saber
          de cara o quanto vai levar. */}
      <Box sx={{ borderTop: `2px solid ${cores.tinta}`, pt: 2.5, mb: 5 }}>
        <Typography
          sx={{
            fontFamily: mono,
            fontSize: 11.5,
            letterSpacing: '0.14em',
            textTransform: 'uppercase',
            color: cores.tintaFraca,
            mb: 1.5,
          }}
        >
          {total} pergunta{total === 1 ? '' : 's'}
        </Typography>

        <Typography variant="h4" sx={{ mb: form.description ? 1.5 : 0 }}>
          {form.title}
        </Typography>

        {form.description && (
          <Typography color="text.secondary" sx={{ maxWidth: '46ch' }}>
            {form.description}
          </Typography>
        )}
      </Box>

      <form
        onSubmit={(e) => {
          e.preventDefault()
          enviar.mutate({
            path: { slug },
            body: {
              answers: form.fields.map((campo) => ({
                field_id: campo.id,
                value: valores[campo.id] ?? '',
              })),
            },
          })
        }}
      >
        <Stack spacing={4}>
          {form.fields.map((campo, indice) => (
            // O número mora na margem, fora do fluxo de leitura, e some no
            // celular.
            <Stack key={campo.id} direction="row" spacing={2.5}>
              <Typography
                aria-hidden
                sx={{
                  display: { xs: 'none', sm: 'block' },
                  fontFamily: mono,
                  fontSize: 12.5,
                  color: erros[campo.id] ? cores.erro : cores.grafite,
                  pt: 1.6,
                  width: 22,
                  flexShrink: 0,
                  textAlign: 'right',
                }}
              >
                {String(indice + 1).padStart(2, '0')}
              </Typography>

              <Box sx={{ flexGrow: 1, minWidth: 0 }}>
                <FieldRenderer
                  field={campo}
                  value={valores[campo.id] ?? ''}
                  onChange={(valor) => setValores((atual) => ({ ...atual, [campo.id]: valor }))}
                  error={erros[campo.id]}
                />
              </Box>
            </Stack>
          ))}

          {Object.keys(erros).length > 0 && (
            <Alert severity="error">
              {Object.keys(erros).length} campo
              {Object.keys(erros).length === 1 ? '' : 's'} para corrigir acima.
            </Alert>
          )}

          {enviar.error && Object.keys(erros).length === 0 && (
            <Alert severity="error">{mensagemDeErro(enviar.error)}</Alert>
          )}

          <Box sx={{ borderTop: `1px solid ${cores.fio}`, pt: 3 }}>
            <Button
              type="submit"
              variant="contained"
              size="large"
              disabled={enviar.isPending}
              sx={{ px: 4 }}
            >
              {enviar.isPending ? 'Enviando' : 'Enviar resposta'}
            </Button>
          </Box>
        </Stack>
      </form>
    </Container>
  )
}
