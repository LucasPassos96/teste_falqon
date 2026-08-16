import { createTheme } from '@mui/material/styles'

/**
 * Direção visual: tinta e carimbo.
 *
 * O conceito central do produto é a virada de rascunho para publicado — é o
 * momento em que o formulário deixa de ser seu e ganha um link no mundo. Todo
 * o sistema visual existe para tornar esse estado legível.
 *
 * Papel branco-frio (não creme), texto em azul-quase-preto de tinta
 * permanente, e um violeta de carimbo cartorial usado em EXATAMENTE dois
 * lugares: a marca de publicado e a URL pública. Botão primário é tinta, não
 * violeta — se a cor de destaque virasse a cor dos botões, ela deixaria de
 * destacar qualquer coisa.
 */
export const cores = {
  papel: '#F6F7F9',
  carta: '#FFFFFF',
  tinta: '#14243D',
  tintaFraca: '#5A6B84',
  fio: '#DDE3EA',
  grafite: '#8A93A3',
  carimbo: '#6B3FA0',
  carimboFraco: '#F1ECF8',
  erro: '#B3261E',
  ok: '#1E6B47',
} as const

// IBM Plex: desenhada para uma empresa cujo negócio eram registros e
// processamento de dados. A mono não é enfeite — ela marca o que é valor de
// máquina (slug, URL, timestamp, contagem), coisas que se copia e se confere.
const sans = '"IBM Plex Sans", system-ui, -apple-system, sans-serif'
export const mono = '"IBM Plex Mono", ui-monospace, monospace'

export const theme = createTheme({
  palette: {
    background: { default: cores.papel, paper: cores.carta },
    primary: { main: cores.tinta, contrastText: '#FFFFFF' },
    secondary: { main: cores.carimbo },
    text: { primary: cores.tinta, secondary: cores.tintaFraca },
    divider: cores.fio,
    error: { main: cores.erro },
    success: { main: cores.ok },
  },

  shape: { borderRadius: 4 },

  typography: {
    fontFamily: sans,
    // Tracking negativo nos títulos é o tratamento que dá densidade e
    // intenção ao display, em vez de deixá-lo ser só texto grande.
    h4: { fontWeight: 600, letterSpacing: '-0.025em', fontSize: '1.9rem' },
    h5: { fontWeight: 600, letterSpacing: '-0.02em', fontSize: '1.4rem' },
    h6: { fontWeight: 600, letterSpacing: '-0.015em', fontSize: '1.08rem' },
    body1: { lineHeight: 1.6 },
    body2: { lineHeight: 1.55 },
    button: { textTransform: 'none', fontWeight: 500, letterSpacing: 0 },
    caption: { color: cores.tintaFraca },
  },

  components: {
    MuiCssBaseline: {
      styleOverrides: {
        // Piso de qualidade, sem anunciar: foco de teclado sempre visível, e
        // quem pediu menos movimento no sistema operacional recebe menos.
        '*:focus-visible': {
          outline: `2px solid ${cores.carimbo}`,
          outlineOffset: 2,
        },
        '@media (prefers-reduced-motion: reduce)': {
          '*': {
            animationDuration: '0.01ms !important',
            transitionDuration: '0.01ms !important',
          },
        },
      },
    },

    // Cartão chapado com um fio: a sombra é o default que faz tudo parecer
    // igual. Aqui a hierarquia vem de borda, espaço e peso de tipo.
    MuiCard: {
      defaultProps: { elevation: 0, variant: 'outlined' },
      styleOverrides: {
        root: { borderColor: cores.fio },
      },
    },

    MuiButton: {
      defaultProps: { disableElevation: true },
      styleOverrides: {
        root: { paddingInline: 16 },
        outlined: { borderColor: cores.fio },
      },
    },

    MuiAppBar: {
      defaultProps: { elevation: 0, color: 'transparent' },
      styleOverrides: {
        root: {
          backgroundColor: cores.carta,
          borderBottom: `1px solid ${cores.fio}`,
        },
      },
    },

    MuiChip: {
      styleOverrides: {
        root: { fontWeight: 500 },
      },
    },

    MuiTableCell: {
      styleOverrides: {
        head: {
          fontWeight: 600,
          fontSize: '0.75rem',
          textTransform: 'uppercase',
          letterSpacing: '0.06em',
          color: cores.tintaFraca,
          backgroundColor: cores.papel,
        },
      },
    },

    MuiAlert: {
      defaultProps: { variant: 'outlined' },
    },

    MuiTextField: {
      defaultProps: { size: 'small' },
    },
  },
})
