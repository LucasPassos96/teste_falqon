import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  CircularProgress,
  Divider,
  Stack,
  TextField,
  Typography,
} from '@mui/material'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { Link as RouterLink, useParams } from 'react-router-dom'

import type { Field, FieldInput } from '../api/gen'
import {
  getFormOptions,
  getFormQueryKey,
  listFormsQueryKey,
  publishFormMutation,
  replaceFormFieldsMutation,
  unpublishFormMutation,
  updateFormMutation,
} from '../api/gen/@tanstack/react-query.gen'
import { Carimbo, RotuloRascunho } from '../components/Carimbo'
import { FieldEditor } from '../components/builder/FieldEditor'
import { FieldRenderer } from '../components/render/FieldRenderer'
import { mensagemDeErro } from '../lib/erros'
import { cores, mono } from '../theme'

/** Field (da API, com id e position) vira FieldInput (o que o builder edita). */
function paraInput(f: Field): FieldInput {
  return {
    type: f.type,
    label: f.label,
    help_text: f.help_text ?? '',
    required: f.required,
    config: f.config ?? {},
  }
}

export default function FormBuilderPage() {
  const { formId = '' } = useParams()
  const queryClient = useQueryClient()

  const { data: form, isPending, error } = useQuery(getFormOptions({ path: { formId } }))

  // O array de campos é estado LOCAL, não estado de servidor: o usuário
  // reordena, renomeia e remove livremente, e só ao salvar isso vira uma
  // requisição. É a razão de o backend expor um PUT da lista inteira.
  const [campos, setCampos] = useState<FieldInput[]>([])
  const [titulo, setTitulo] = useState('')
  const [descricao, setDescricao] = useState('')
  const [preview, setPreview] = useState<Record<string, string>>({})
  const [salvo, setSalvo] = useState(false)

  // Sincroniza o estado local quando o formulário chega da API.
  useEffect(() => {
    if (!form) return
    setCampos((form.fields ?? []).map(paraInput))
    setTitulo(form.title)
    setDescricao(form.description)
  }, [form])

  const recarregar = () => {
    queryClient.invalidateQueries({ queryKey: getFormQueryKey({ path: { formId } }) })
    queryClient.invalidateQueries({ queryKey: listFormsQueryKey() })
  }

  const salvarCampos = useMutation({
    ...replaceFormFieldsMutation(),
    onSuccess: () => {
      setSalvo(true)
      recarregar()
    },
  })
  const salvarTitulo = useMutation({ ...updateFormMutation(), onSuccess: recarregar })
  const publicar = useMutation({ ...publishFormMutation(), onSuccess: recarregar })
  const despublicar = useMutation({ ...unpublishFormMutation(), onSuccess: recarregar })

  if (isPending) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', mt: 8 }}>
        <CircularProgress />
      </Box>
    )
  }
  if (error || !form) {
    return <Alert severity="error">{mensagemDeErro(error)}</Alert>
  }

  const publicado = form.status === 'published'

  const alterarCampo = (i: number, campo: FieldInput) =>
    setCampos((atual) => atual.map((c, j) => (j === i ? campo : c)))

  const moverCampo = (i: number, direcao: -1 | 1) =>
    setCampos((atual) => {
      const destino = i + direcao
      if (destino < 0 || destino >= atual.length) return atual
      const copia = [...atual]
      ;[copia[i], copia[destino]] = [copia[destino], copia[i]]
      return copia
    })

  const erroDeSalvar = salvarCampos.error ?? publicar.error ?? despublicar.error

  return (
    <Stack spacing={3}>
      <Box>
        <Button
          component={RouterLink}
          to="/"
          size="small"
          sx={{ ml: -2, mb: 2, color: cores.tintaFraca }}
        >
          ← Voltar aos formulários
        </Button>

        <Box sx={{ borderTop: `2px solid ${cores.tinta}`, pt: 2.5 }}>
          <Stack direction="row" spacing={2} sx={{ alignItems: 'center', mb: 1 }}>
            <Typography
              sx={{
                fontFamily: mono,
                fontSize: 11.5,
                letterSpacing: '0.14em',
                textTransform: 'uppercase',
                color: cores.tintaFraca,
              }}
            >
              {campos.length} campo{campos.length === 1 ? '' : 's'}
            </Typography>
            {publicado ? <Carimbo small /> : <RotuloRascunho />}
          </Stack>
          <Typography variant="h4">{form.title}</Typography>
        </Box>
      </Box>

      <Card variant="outlined">
        <CardContent>
          <Stack spacing={2}>
            <TextField
              label="Título"
              value={titulo}
              onChange={(e) => setTitulo(e.target.value)}
              fullWidth
            />
            <TextField
              label="Descrição"
              value={descricao}
              onChange={(e) => setDescricao(e.target.value)}
              fullWidth
              multiline
              minRows={2}
            />
            <Button
              sx={{ alignSelf: 'flex-start' }}
              disabled={
                salvarTitulo.isPending ||
                (titulo === form.title && descricao === form.description)
              }
              onClick={() =>
                salvarTitulo.mutate({
                  path: { formId },
                  body: { title: titulo, description: descricao },
                })
              }
            >
              Salvar título e descrição
            </Button>
          </Stack>
        </CardContent>
      </Card>

      {publicado && (
        <Alert severity="info">
          A estrutura fica travada enquanto o formulário estiver publicado, para não
          invalidar as respostas já recebidas. Despublique para editar os campos — o link
          para de funcionar nesse intervalo e volta a valer, o mesmo link, ao republicar.
        </Alert>
      )}

      {form.public_url && (
        <Box
          sx={{
            border: `1px solid ${cores.carimbo}33`,
            bgcolor: cores.carimboFraco,
            borderRadius: 1,
            px: 2,
            py: 1.75,
          }}
        >
          <Typography
            sx={{
              fontFamily: mono,
              fontSize: 10.5,
              letterSpacing: '0.14em',
              textTransform: 'uppercase',
              color: cores.carimbo,
              mb: 0.75,
            }}
          >
            Link público
          </Typography>
          <Box
            component="a"
            href={form.public_url}
            target="_blank"
            rel="noreferrer"
            sx={{
              fontFamily: mono,
              fontSize: 13.5,
              color: cores.carimbo,
              wordBreak: 'break-all',
            }}
          >
            {form.public_url}
          </Box>
        </Box>
      )}

      <Divider />

      <Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'space-between' }}>
        <Typography variant="h6">Campos</Typography>
        <Button
          variant="outlined"
          disabled={publicado || campos.length >= 50}
          onClick={() =>
            setCampos((atual) => [
              ...atual,
              { type: 'short_text', label: '', required: false, help_text: '', config: {} },
            ])
          }
        >
          Adicionar campo
        </Button>
      </Stack>

      {campos.length === 0 && (
        <Card sx={{ borderStyle: 'dashed', borderColor: cores.grafite }}>
          <CardContent sx={{ py: 5, textAlign: 'center' }}>
            <Typography variant="body2" color="text.secondary">
              Adicione ao menos um campo para poder publicar.
            </Typography>
          </CardContent>
        </Card>
      )}

      <Stack spacing={2}>
        {campos.map((campo, i) => (
          <Box key={i} sx={{ opacity: publicado ? 0.6 : 1, pointerEvents: publicado ? 'none' : 'auto' }}>
            <FieldEditor
              campo={campo}
              indice={i}
              total={campos.length}
              onChange={(c) => alterarCampo(i, c)}
              onRemove={() => setCampos((atual) => atual.filter((_, j) => j !== i))}
              onMover={(d) => moverCampo(i, d)}
            />
          </Box>
        ))}
      </Stack>

      {erroDeSalvar && <Alert severity="error">{mensagemDeErro(erroDeSalvar)}</Alert>}
      {salvo && !salvarCampos.isPending && <Alert severity="success">Campos salvos.</Alert>}

      <Stack direction="row" spacing={2}>
        <Button
          variant="contained"
          disabled={publicado || salvarCampos.isPending}
          onClick={() => {
            setSalvo(false)
            salvarCampos.mutate({ path: { formId }, body: campos })
          }}
        >
          Salvar campos
        </Button>

        {publicado ? (
          <Button
            variant="outlined"
            color="warning"
            disabled={despublicar.isPending}
            onClick={() => despublicar.mutate({ path: { formId } })}
          >
            Despublicar
          </Button>
        ) : (
          <Button
            variant="outlined"
            color="success"
            disabled={publicar.isPending || (form.fields ?? []).length === 0}
            onClick={() => publicar.mutate({ path: { formId } })}
          >
            Publicar
          </Button>
        )}

        <Button component={RouterLink} to={`/forms/${formId}/submissions`}>
          Ver respostas ({form.submission_count})
        </Button>
      </Stack>

      {(form.fields ?? []).length > 0 && (
        <>
          <Divider />
          <Box>
            <Typography variant="h6">Pré-visualização</Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
              O mesmo componente que desenha a página pública, com os campos já salvos.
            </Typography>
          </Box>

          <Card variant="outlined">
            <CardContent>
              <Stack spacing={2}>
                {(form.fields ?? []).map((f) => (
                  <FieldRenderer
                    key={f.id}
                    field={f}
                    value={preview[f.id] ?? ''}
                    onChange={(v) => setPreview((p) => ({ ...p, [f.id]: v }))}
                  />
                ))}
              </Stack>
            </CardContent>
          </Card>
        </>
      )}
    </Stack>
  )
}
