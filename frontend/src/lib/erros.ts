/** Mensagem tratada do envelope de erro da API, nunca o payload cru. */
export function mensagemDeErro(erro: unknown): string {
  if (erro && typeof erro === 'object') {
    const corpo = erro as { message?: unknown; code?: unknown }
    if (typeof corpo.message === 'string' && corpo.message) {
      return corpo.message
    }
  }
  return 'Não foi possível concluir a operação. Tente novamente.'
}

/** Erros por campo de um 422, indexados pelo field_id. */
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
