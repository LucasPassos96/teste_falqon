import { useQuery } from '@tanstack/react-query'

import { getCurrentUserOptions } from '../api/gen/@tanstack/react-query.gen'

/**
 * O front não sabe se está logado — ele pergunta.
 *
 * O cookie de sessão é HttpOnly, então o JavaScript não consegue lê-lo nem
 * decodificar o token, o que é exatamente o objetivo. O estado de sessão vem
 * de GET /auth/me: enquanto carrega mostramos splash, e 401 redireciona.
 */
export function useSession() {
  const query = useQuery({
    ...getCurrentUserOptions(),
    // 401 é uma resposta válida ("não está logado"), não uma falha de rede:
    // repetir a requisição três vezes só atrasaria o redirecionamento.
    retry: false,
  })

  return {
    user: query.data,
    isLoading: query.isPending,
    isAuthenticated: !!query.data,
  }
}
