import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  CircularProgress,
  Pagination,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
} from '@mui/material'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Link as RouterLink, useParams } from 'react-router-dom'

import { getFormOptions, listSubmissionsOptions } from '../api/gen/@tanstack/react-query.gen'
import { mensagemDeErro } from '../lib/erros'
import { cores, mono } from '../theme'

const POR_PAGINA = 20

export default function SubmissionsPage() {
  const { formId = '' } = useParams()
  const [pagina, setPagina] = useState(1)

  const { data: form } = useQuery(getFormOptions({ path: { formId } }))
  const { data, isPending, error } = useQuery(
    listSubmissionsOptions({
      path: { formId },
      query: { limit: POR_PAGINA, offset: (pagina - 1) * POR_PAGINA },
    }),
  )

  if (isPending) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', mt: 8 }}>
        <CircularProgress />
      </Box>
    )
  }
  if (error) {
    return <Alert severity="error">{mensagemDeErro(error)}</Alert>
  }

  // As colunas vêm dos rótulos gravados NAS RESPOSTAS, não da definição atual
  // do formulário: se um campo foi renomeado depois, a tabela mostra o texto
  // que o visitante realmente viu.
  const colunas: string[] = []
  for (const s of data.items) {
    for (const a of s.answers) {
      if (!colunas.includes(a.field_label)) colunas.push(a.field_label)
    }
  }

  const totalPaginas = Math.ceil(data.total / POR_PAGINA)

  return (
    <Stack spacing={3}>
      <Box>
        <Button
          component={RouterLink}
          to={`/forms/${formId}`}
          size="small"
          sx={{ ml: -2, mb: 2, color: cores.tintaFraca }}
        >
          ← Voltar ao formulário
        </Button>

        <Box sx={{ borderTop: `2px solid ${cores.tinta}`, pt: 2.5 }}>
          <Typography
            sx={{
              fontFamily: mono,
              fontSize: 11.5,
              letterSpacing: '0.14em',
              textTransform: 'uppercase',
              color: cores.tintaFraca,
              mb: 1.25,
            }}
          >
            {data.total} resposta{data.total === 1 ? '' : 's'}
          </Typography>
          <Typography variant="h4">{form?.title ?? 'Respostas'}</Typography>
        </Box>
      </Box>

      {data.total === 0 ? (
        <Card sx={{ borderStyle: 'dashed', borderColor: cores.grafite }}>
          <CardContent sx={{ py: 6, textAlign: 'center' }}>
            <Typography variant="h6" gutterBottom>
              Ainda sem respostas
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Compartilhe o link público e as respostas aparecem aqui.
            </Typography>
          </CardContent>
        </Card>
      ) : (
        <>
          <TableContainer component={Card} variant="outlined">
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Enviada em</TableCell>
                  {colunas.map((c) => (
                    <TableCell key={c}>{c}</TableCell>
                  ))}
                </TableRow>
              </TableHead>
              <TableBody>
                {data.items.map((s) => {
                  const porRotulo = Object.fromEntries(
                    s.answers.map((a) => [a.field_label, a.value]),
                  )
                  return (
                    <TableRow key={s.id}>
                      <TableCell
                        sx={{
                          whiteSpace: 'nowrap',
                          fontFamily: mono,
                          fontSize: 12.5,
                          color: cores.tintaFraca,
                        }}
                      >
                        {new Date(s.submitted_at).toLocaleString('pt-BR')}
                      </TableCell>
                      {colunas.map((c) => (
                        // {porRotulo[c]} passa por JSX, então o React escapa:
                        // um <script> enviado por um visitante anônimo aparece
                        // como TEXTO nesta tela, que roda numa sessão
                        // privilegiada. É aqui que uma XSS armazenada morderia.
                        <TableCell key={c}>{porRotulo[c] ?? '—'}</TableCell>
                      ))}
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </TableContainer>

          {totalPaginas > 1 && (
            <Pagination
              count={totalPaginas}
              page={pagina}
              onChange={(_, p) => setPagina(p)}
              sx={{ alignSelf: 'center' }}
            />
          )}
        </>
      )}
    </Stack>
  )
}
