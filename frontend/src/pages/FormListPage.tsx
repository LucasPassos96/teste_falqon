import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  TextField,
  Typography,
} from '@mui/material'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link as RouterLink } from 'react-router-dom'

import type { Form } from '../api/gen'
import {
  createFormMutation,
  deleteFormMutation,
  listFormsOptions,
  listFormsQueryKey,
} from '../api/gen/@tanstack/react-query.gen'
import { Carimbo, RotuloRascunho } from '../components/Carimbo'
import { mensagemDeErro } from '../lib/erros'
import { cores, mono } from '../theme'

export default function FormListPage() {
  const queryClient = useQueryClient()
  const { data: formularios, isPending, error } = useQuery(listFormsOptions())

  const [criando, setCriando] = useState(false)
  const [titulo, setTitulo] = useState('')
  const [descricao, setDescricao] = useState('')

  const recarregar = () => queryClient.invalidateQueries({ queryKey: listFormsQueryKey() })

  const criar = useMutation({
    ...createFormMutation(),
    onSuccess: () => {
      setCriando(false)
      setTitulo('')
      setDescricao('')
      recarregar()
    },
  })

  const remover = useMutation({ ...deleteFormMutation(), onSuccess: recarregar })

  if (isPending) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
        <CircularProgress size={28} />
      </Box>
    )
  }

  if (error) {
    return <Alert severity="error">{mensagemDeErro(error)}</Alert>
  }

  return (
    <Stack spacing={4}>
      <Stack
        direction={{ xs: 'column', sm: 'row' }}
        spacing={2}
        sx={{ alignItems: { sm: 'baseline' }, justifyContent: 'space-between' }}
      >
        <Box>
          <Typography variant="h4">Formulários</Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
            {formularios.length === 0
              ? 'Nada por aqui ainda'
              : `${formularios.length} no total · ${
                  formularios.filter((f) => f.status === 'published').length
                } publicado${
                  formularios.filter((f) => f.status === 'published').length === 1 ? '' : 's'
                }`}
          </Typography>
        </Box>

        <Button variant="contained" onClick={() => setCriando(true)}>
          Criar formulário
        </Button>
      </Stack>

      {formularios.length === 0 && (
        <Card sx={{ borderStyle: 'dashed', borderColor: cores.grafite }}>
          <CardContent sx={{ py: 6, textAlign: 'center' }}>
            <Typography variant="h6" gutterBottom>
              Crie seu primeiro formulário
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ maxWidth: 420, mx: 'auto' }}>
              Monte as perguntas, publique, e a aplicação gera um link que qualquer pessoa
              pode preencher sem precisar de conta.
            </Typography>
          </CardContent>
        </Card>
      )}

      <Stack spacing={2}>
        {formularios.map((f) => (
          <CartaoFormulario
            key={f.id}
            form={f}
            onRemover={() => {
              if (confirm(`Remover "${f.title}" e todas as respostas recebidas?`)) {
                remover.mutate({ path: { formId: f.id } })
              }
            }}
          />
        ))}
      </Stack>

      <Dialog open={criando} onClose={() => setCriando(false)} fullWidth maxWidth="sm">
        <DialogTitle sx={{ fontWeight: 600, letterSpacing: '-0.02em' }}>
          Criar formulário
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2.5} sx={{ mt: 1 }}>
            <TextField
              label="Título"
              value={titulo}
              onChange={(e) => setTitulo(e.target.value)}
              required
              fullWidth
              autoFocus
            />
            <TextField
              label="Descrição"
              value={descricao}
              onChange={(e) => setDescricao(e.target.value)}
              fullWidth
              multiline
              minRows={2}
              helperText="Aparece no topo da página que o visitante vê"
            />
            {criar.error && <Alert severity="error">{mensagemDeErro(criar.error)}</Alert>}
          </Stack>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={() => setCriando(false)}>Cancelar</Button>
          <Button
            variant="contained"
            disabled={!titulo.trim() || criar.isPending}
            onClick={() => criar.mutate({ body: { title: titulo, description: descricao } })}
          >
            Criar
          </Button>
        </DialogActions>
      </Dialog>
    </Stack>
  )
}

function CartaoFormulario({ form, onRemover }: { form: Form; onRemover: () => void }) {
  const [copiado, setCopiado] = useState(false)
  const publicado = form.status === 'published'

  return (
    <Card
      sx={{
        // Tracejado: a mesma informação do rótulo, dita pela forma.
        borderStyle: publicado ? 'solid' : 'dashed',
        borderColor: publicado ? cores.fio : '#C9CFD8',
        transition: 'border-color 120ms ease',
        '&:hover': { borderColor: publicado ? cores.tintaFraca : cores.grafite },
      }}
    >
      <CardContent sx={{ p: 3, '&:last-child': { pb: 3 } }}>
        <Stack
          direction={{ xs: 'column', sm: 'row' }}
          spacing={2}
          sx={{ justifyContent: 'space-between' }}
        >
          <Box sx={{ minWidth: 0, flexGrow: 1 }}>
            <Stack direction="row" spacing={1.5} sx={{ alignItems: 'center', mb: 0.75 }}>
              <Typography
                variant="h6"
                component={RouterLink}
                to={`/forms/${form.id}`}
                sx={{
                  color: 'text.primary',
                  textDecoration: 'none',
                  '&:hover': { textDecoration: 'underline' },
                }}
              >
                {form.title}
              </Typography>
              {publicado ? <Carimbo small /> : <RotuloRascunho />}
            </Stack>

            {form.description && (
              <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                {form.description}
              </Typography>
            )}

            <Typography
              sx={{ fontFamily: mono, fontSize: 12.5, color: cores.tintaFraca, letterSpacing: 0 }}
            >
              {form.fields?.length ?? 0} campos · {form.submission_count} resposta
              {form.submission_count === 1 ? '' : 's'}
            </Typography>

            {form.public_url && (
              <Stack
                direction="row"
                spacing={1}
                sx={{ alignItems: 'center', mt: 1.5, flexWrap: 'wrap' }}
              >
                <Box
                  component="a"
                  href={form.public_url}
                  target="_blank"
                  rel="noreferrer"
                  sx={{
                    fontFamily: mono,
                    fontSize: 12.5,
                    color: cores.carimbo,
                    bgcolor: cores.carimboFraco,
                    px: 1,
                    py: 0.4,
                    borderRadius: 0.5,
                    textDecoration: 'none',
                    wordBreak: 'break-all',
                    '&:hover': { textDecoration: 'underline' },
                  }}
                >
                  {form.public_url.replace(/^https?:\/\//, '')}
                </Box>
                <Button
                  size="small"
                  onClick={() => {
                    navigator.clipboard.writeText(form.public_url ?? '')
                    setCopiado(true)
                    setTimeout(() => setCopiado(false), 2000)
                  }}
                  sx={{ minWidth: 0, color: cores.tintaFraca }}
                >
                  {copiado ? 'Copiado' : 'Copiar'}
                </Button>
              </Stack>
            )}
          </Box>

          <Stack
            direction={{ xs: 'row', sm: 'column' }}
            spacing={1}
            sx={{ flexShrink: 0, alignItems: { sm: 'stretch' } }}
          >
            <Button size="small" variant="outlined" component={RouterLink} to={`/forms/${form.id}`}>
              Editar
            </Button>
            <Button
              size="small"
              component={RouterLink}
              to={`/forms/${form.id}/submissions`}
              sx={{ color: cores.tintaFraca }}
            >
              Respostas
            </Button>
            <Button size="small" color="error" onClick={onRemover}>
              Remover
            </Button>
          </Stack>
        </Stack>
      </CardContent>
    </Card>
  )
}
