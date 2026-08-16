/**
 * Extrai a mensagem tratada que a API devolveu.
 *
 * O front mostra a mensagem do envelope de erro, nunca o payload cru nem o
 * objeto de exceção — nada de stack trace ou detalhe de banco na tela.
 */
export function mensagemDeErro(erro: unknown): string {
  if (erro && typeof erro === 'object') {
    const corpo = erro as { message?: unknown; code?: unknown }
    if (typeof corpo.message === 'string' && corpo.message) {
      return corpo.message
    }
  }
  return 'Não foi possível concluir a operação. Tente novamente.'
}

/**
 * Erros por campo, quando a API devolve 422 com field_errors.
 * A chave é o field_id, que é como a página pública localiza o campo.
 */
export function errosPorCampo(erro: unknown): Record<string, string> {
  const corpo = erro as { field_errors?: { field?: string; message?: string }[] } | null
  const lista = corpo?.field_errors
  if (!Array.isArray(lista)) return {}

  const mapa: Record<string, string> = {}
  for (const item of lista) {
    if (item.field && item.message) mapa[item.field] = item.message
  }
  return mapa
}
