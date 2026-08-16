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

  // As colunas vêm dos rótulos gravados nas respostas, não da definição atual:
  // um campo renomeado depois não reescreve o passado.
  const colunas: string[] = []
  for (const submissao of data.items) {
    for (const resposta of submissao.answers) {
      if (!colunas.includes(resposta.field_label)) colunas.push(resposta.field_label)
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
                  {colunas.map((coluna) => (
                    <TableCell key={coluna}>{coluna}</TableCell>
                  ))}
                </TableRow>
              </TableHead>
              <TableBody>
                {data.items.map((submissao) => {
                  const porRotulo = Object.fromEntries(
                    submissao.answers.map((resposta) => [resposta.field_label, resposta.value]),
                  )
                  return (
                    <TableRow key={submissao.id}>
                      <TableCell
                        sx={{
                          whiteSpace: 'nowrap',
                          fontFamily: mono,
                          fontSize: 12.5,
                          color: cores.tintaFraca,
                        }}
                      >
                        {new Date(submissao.submitted_at).toLocaleString('pt-BR')}
                      </TableCell>
                      {colunas.map((coluna) => (
                        // Passa por JSX, então o React escapa: conteúdo de
                        // visitante anônimo aparece como texto nesta tela.
                        <TableCell key={coluna}>{porRotulo[coluna] ?? '—'}</TableCell>
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
