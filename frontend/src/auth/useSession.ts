import { useQuery } from '@tanstack/react-query'

import { getCurrentUserOptions } from '../api/gen/@tanstack/react-query.gen'

/**
 * O cookie de sessão é HttpOnly, então o JavaScript não consegue lê-lo — o
 * estado de sessão vem de GET /auth/me.
 */
export function useSession() {
  const query = useQuery({
    ...getCurrentUserOptions(),
    // 401 é resposta válida, não falha de rede: repetir só atrasa o redirect.
    retry: false,
  })

  return {
    user: query.data,
    isLoading: query.isPending,
    isAuthenticated: !!query.data,
  }
}
