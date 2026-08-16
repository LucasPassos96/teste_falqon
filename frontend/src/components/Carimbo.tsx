import { Box } from '@mui/material'

import { cores, mono } from '../theme'

/**
 * Marca de formulário publicado. É a única peça ousada da interface: publicar é
 * a virada central do produto, e tudo em volta fica quieto de propósito.
 */
export function Carimbo({ small = false }: { small?: boolean }) {
  return (
    <Box
      component="span"
      aria-label="Formulário publicado"
      sx={{
        display: 'inline-block',
        fontFamily: mono,
        fontWeight: 600,
        fontSize: small ? 10 : 11,
        letterSpacing: '0.18em',
        color: cores.carimbo,
        border: `1.5px solid ${cores.carimbo}`,
        // O fio interno é o que faz ler como carimbo, e não como badge.
        boxShadow: `inset 0 0 0 1px ${cores.carta}, inset 0 0 0 2.5px ${cores.carimbo}`,
        borderRadius: 0.5,
        px: small ? 0.9 : 1.2,
        py: small ? 0.35 : 0.5,
        // Ângulo ímpar: carimbo batido à mão nunca sai reto.
        transform: 'rotate(-3.5deg)',
        transformOrigin: 'center',
        whiteSpace: 'nowrap',
        userSelect: 'none',
      }}
    >
      PUBLICADO
    </Box>
  )
}

/** Contraponto do carimbo: rascunho recebe só um rótulo apagado. */
export function RotuloRascunho() {
  return (
    <Box
      component="span"
      sx={{
        fontFamily: mono,
        fontSize: 11,
        letterSpacing: '0.12em',
        color: cores.grafite,
        textTransform: 'uppercase',
        userSelect: 'none',
      }}
    >
      rascunho
    </Box>
  )
}
