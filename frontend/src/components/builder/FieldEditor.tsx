import {
  Box,
  Button,
  Card,
  CardContent,
  Checkbox,
  FormControlLabel,
  IconButton,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from '@mui/material'

import type { FieldInput, FieldType } from '../../api/gen'

export const TIPOS: { valor: FieldType; rotulo: string }[] = [
  { valor: 'short_text', rotulo: 'Texto curto' },
  { valor: 'long_text', rotulo: 'Texto longo' },
  { valor: 'email', rotulo: 'E-mail' },
  { valor: 'number', rotulo: 'Número' },
  { valor: 'select', rotulo: 'Seleção' },
  { valor: 'checkbox', rotulo: 'Caixa de seleção' },
]

type Props = {
  campo: FieldInput
  indice: number
  total: number
  onChange: (campo: FieldInput) => void
  onRemove: () => void
  onMover: (direcao: -1 | 1) => void
}

export function FieldEditor({ campo, indice, total, onChange, onRemove, onMover }: Props) {
  const alterar = (patch: Partial<FieldInput>) => onChange({ ...campo, ...patch })
  const alterarConfig = (patch: Partial<NonNullable<FieldInput['config']>>) =>
    onChange({ ...campo, config: { ...campo.config, ...patch } })

  const opcoes = campo.config?.options ?? []

  return (
    <Card variant="outlined">
      <CardContent>
        <Stack spacing={2}>
          <Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
            <Typography variant="caption" color="text.secondary" sx={{ minWidth: 24 }}>
              {indice + 1}
            </Typography>

            <TextField
              label="Rótulo"
              value={campo.label}
              onChange={(e) => alterar({ label: e.target.value })}
              size="small"
              sx={{ flexGrow: 1 }}
              required
            />

            <TextField
              label="Tipo"
              select
              size="small"
              sx={{ minWidth: 170 }}
              value={campo.type}
              onChange={(e) => {
                // Trocar o tipo limpa a config, que não faz sentido no novo.
                alterar({ type: e.target.value as FieldType, config: {} })
              }}
            >
              {TIPOS.map((tipo) => (
                <MenuItem key={tipo.valor} value={tipo.valor}>
                  {tipo.rotulo}
                </MenuItem>
              ))}
            </TextField>

            {/* Setas em vez de arrastar: acessível por teclado e sem
                biblioteca de drag-and-drop. */}
            <IconButton size="small" disabled={indice === 0} onClick={() => onMover(-1)}>
              <span aria-hidden>↑</span>
            </IconButton>
            <IconButton size="small" disabled={indice === total - 1} onClick={() => onMover(1)}>
              <span aria-hidden>↓</span>
            </IconButton>
            <IconButton size="small" color="error" onClick={onRemove}>
              <span aria-hidden>✕</span>
            </IconButton>
          </Stack>

          <TextField
            label="Texto de ajuda"
            value={campo.help_text ?? ''}
            onChange={(e) => alterar({ help_text: e.target.value })}
            size="small"
            fullWidth
          />

          <FormControlLabel
            control={
              <Checkbox
                checked={campo.required}
                onChange={(e) => alterar({ required: e.target.checked })}
              />
            }
            label={
              campo.type === 'checkbox'
                ? 'Obrigatório (precisa estar marcado)'
                : 'Obrigatório'
            }
          />

          {(campo.type === 'short_text' || campo.type === 'long_text') && (
            <Stack direction="row" spacing={2}>
              <TextField
                label="Mínimo de caracteres"
                type="number"
                size="small"
                value={campo.config?.min_length ?? ''}
                onChange={(e) =>
                  alterarConfig({
                    min_length: e.target.value === '' ? undefined : Number(e.target.value),
                  })
                }
              />
              <TextField
                label="Máximo de caracteres"
                type="number"
                size="small"
                value={campo.config?.max_length ?? ''}
                onChange={(e) =>
                  alterarConfig({
                    max_length: e.target.value === '' ? undefined : Number(e.target.value),
                  })
                }
              />
            </Stack>
          )}

          {campo.type === 'number' && (
            <Stack direction="row" spacing={2}>
              <TextField
                label="Valor mínimo"
                type="number"
                size="small"
                value={campo.config?.min ?? ''}
                onChange={(e) =>
                  alterarConfig({ min: e.target.value === '' ? undefined : Number(e.target.value) })
                }
              />
              <TextField
                label="Valor máximo"
                type="number"
                size="small"
                value={campo.config?.max ?? ''}
                onChange={(e) =>
                  alterarConfig({ max: e.target.value === '' ? undefined : Number(e.target.value) })
                }
              />
            </Stack>
          )}

          {campo.type === 'select' && (
            <Box>
              <Typography variant="caption" color="text.secondary">
                Opções
              </Typography>

              <Stack spacing={1} sx={{ mt: 1 }}>
                {opcoes.map((opcao, indice) => (
                  <Stack key={indice} direction="row" spacing={1}>
                    <TextField
                      label="Valor"
                      size="small"
                      value={opcao.value}
                      onChange={(e) => {
                        const novas = [...opcoes]
                        novas[indice] = { ...opcao, value: e.target.value }
                        alterarConfig({ options: novas })
                      }}
                    />
                    <TextField
                      label="Rótulo exibido"
                      size="small"
                      sx={{ flexGrow: 1 }}
                      value={opcao.label}
                      onChange={(e) => {
                        const novas = [...opcoes]
                        novas[indice] = { ...opcao, label: e.target.value }
                        alterarConfig({ options: novas })
                      }}
                    />
                    <IconButton
                      size="small"
                      color="error"
                      onClick={() =>
                        alterarConfig({ options: opcoes.filter((_, outro) => outro !== indice) })
                      }
                    >
                      <span aria-hidden>✕</span>
                    </IconButton>
                  </Stack>
                ))}

                <Button
                  size="small"
                  onClick={() =>
                    alterarConfig({ options: [...opcoes, { value: '', label: '' }] })
                  }
                  sx={{ alignSelf: 'flex-start' }}
                >
                  Adicionar opção
                </Button>
              </Stack>
            </Box>
          )}
        </Stack>
      </CardContent>
    </Card>
  )
}
