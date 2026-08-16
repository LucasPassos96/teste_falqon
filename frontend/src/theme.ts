import { createTheme } from '@mui/material/styles'

/**
 * Direção visual: tinta e carimbo. Papel branco-frio, texto em azul-quase-preto,
 * e um violeta usado em exatamente dois lugares — a marca de publicado e a URL
 * pública. Botão primário é tinta: se o violeta virasse a cor dos botões,
 * deixaria de destacar qualquer coisa.
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

// A mono marca o que é valor de máquina — slug, URL, timestamp, contagem.
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

    // Sem sombra: a hierarquia vem de borda, espaço e peso de tipo.
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
