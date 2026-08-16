import {
  Checkbox,
  FormControl,
  FormControlLabel,
  FormHelperText,
  InputLabel,
  MenuItem,
  Select,
  TextField,
} from '@mui/material'

import type { FieldType, PublicField } from '../../api/gen'

/**
 * Renderiza um campo a partir da definição vinda da API.
 *
 * Usado no preview do builder E na página pública: um switch num arquivo só
 * decide como cada tipo é exibido no sistema inteiro. Se os dois lugares
 * tivessem cópias, o preview mentiria sobre o formulário real.
 *
 * Recebe `PublicField` porque é o menor denominador comum entre os dois: o
 * builder tem mais dados (position), mas nada que a exibição precise.
 */
type Props = {
  field: PublicField
  value: string
  onChange: (value: string) => void
  error?: string
  disabled?: boolean
}

export function FieldRenderer({ field, value, onChange, error, disabled }: Props) {
  const helper = error ?? field.help_text ?? ''
  const common = {
    fullWidth: true,
    error: !!error,
    disabled,
    helperText: helper || undefined,
  }

  switch (field.type as FieldType) {
    case 'long_text':
      return (
        <TextField
          {...common}
          label={field.label}
          required={field.required}
          multiline
          minRows={3}
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
      )

    case 'email':
      return (
        <TextField
          {...common}
          label={field.label}
          required={field.required}
          // type="email" só liga o teclado certo no celular e a dica do
          // navegador. Quem decide se o e-mail vale é o backend.
          type="email"
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
      )

    case 'number':
      return (
        <TextField
          {...common}
          label={field.label}
          required={field.required}
          type="number"
          slotProps={{ htmlInput: { min: field.config?.min, max: field.config?.max } }}
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
      )

    case 'select':
      return (
        <FormControl fullWidth error={!!error} disabled={disabled} required={field.required}>
          <InputLabel id={`label-${field.id}`}>{field.label}</InputLabel>
          <Select
            labelId={`label-${field.id}`}
            label={field.label}
            value={value}
            onChange={(e) => onChange(e.target.value)}
          >
            {(field.config?.options ?? []).map((opt) => (
              // O value enviado é opt.value; o texto exibido é opt.label.
              <MenuItem key={opt.value} value={opt.value}>
                {opt.label}
              </MenuItem>
            ))}
          </Select>
          {helper && <FormHelperText>{helper}</FormHelperText>}
        </FormControl>
      )

    case 'checkbox':
      return (
        <FormControl error={!!error} disabled={disabled} required={field.required}>
          <FormControlLabel
            control={
              <Checkbox
                checked={value === 'true'}
                onChange={(e) => onChange(e.target.checked ? 'true' : 'false')}
              />
            }
            label={field.label}
          />
          {helper && <FormHelperText>{helper}</FormHelperText>}
        </FormControl>
      )

    case 'short_text':
    default:
      return (
        <TextField
          {...common}
          label={field.label}
          required={field.required}
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
      )
  }
}
