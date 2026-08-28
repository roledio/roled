import { useQuery } from '@tanstack/react-query';
import type { HttpClient } from '@/services/core/httpClient';
import { fetchCurrentUser, type CurrentUser } from '@/services/users';

type Options = { httpClient: HttpClient; baseUrl: string; enabled?: boolean };

export function useCurrentUser({ httpClient, baseUrl, enabled = true }: Options) {
  const query = useQuery<CurrentUser, Error>({
    queryKey: ['user', 'current'],
    queryFn: () => fetchCurrentUser(httpClient, baseUrl),
    retry: 1,
    enabled,
    staleTime: 5 * 60 * 1000,
  });

  return { user: query.data ?? null, isLoading: query.isLoading, error: query.error } as const;
}
