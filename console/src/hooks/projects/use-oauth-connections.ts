import { fetchOAuthConnections, type OAuthConnection } from '@/services/projects/oauth-connections';
import type { HttpClient } from '@/services/core/httpClient';
import { useQuery } from '@tanstack/react-query';

interface Options {
  httpClient: HttpClient;
  baseUrl: string;
  projectId?: string;
}

export function useOAuthConnections({ httpClient, baseUrl, projectId }: Options) {
  const query = useQuery({
    queryKey: ['project', projectId, 'oauth-connections'],
    queryFn: () => fetchOAuthConnections(httpClient, baseUrl, projectId!),
    enabled: !!projectId,
    retry: 1,
  });

  return {
    connections: query.data?.data ?? ([] as OAuthConnection[]),
    isLoading: query.isLoading,
    error: query.error,
    refetch: query.refetch,
  };
}
