import { Box } from '@mui/material'

import { cores, mono } from '../theme'

/**
 * O elemento assinatura da interface.
 *
 * Publicar é a virada central do produto: o formulário para de ser seu e ganha
 * um link no mundo. Um Chip cinza escrito "Publicado" comunica o dado e
 * nenhuma da consequência. O carimbo comunica as duas coisas.
 *
 * A contenção é o que separa isto de kitsch: violeta dessaturado, borda dupla
 * fina, versalete espaçado, rotação pequena e assimétrica, e nada de textura
 * ou sombra. É a única peça ousada da tela — tudo em volta fica quieto.
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
        // O fio interno é o que faz ler como carimbo de repartição, e não
        // como um badge qualquer.
        boxShadow: `inset 0 0 0 1px ${cores.carta}, inset 0 0 0 2.5px ${cores.carimbo}`,
        borderRadius: 0.5,
        px: small ? 0.9 : 1.2,
        py: small ? 0.35 : 0.5,
        // Rotação pequena e ímpar: carimbo batido à mão nunca sai reto, e um
        // ângulo redondo (5°, 10°) leria como decoração deliberada.
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

/**
 * O contraponto do carimbo: rascunho não recebe marca nenhuma, só um rótulo
 * apagado. A ausência é o ponto — ainda não aconteceu nada digno de carimbo.
 */
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
