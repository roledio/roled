import { useQuery } from '@tanstack/react-query';
import type { HttpClient } from '@/services/core/httpClient';
import { fetchProjectSettings, type ProjectSettings } from '@/services/projects';

type Options = {
  httpClient: HttpClient;
  baseUrl: string;
  projectId?: string;
};

/**
 * Fetches and caches project settings for the given project.
 *
 * The query is disabled when `projectId` is not provided so the hook can be
 * called unconditionally even before the project ID is available.
 */
export function useProjectSettings({ httpClient, baseUrl, projectId }: Options) {
  const query = useQuery<ProjectSettings | null, Error>({
    queryKey: ['project', projectId ?? 'unknown', 'settings'],
    queryFn: async () => {
      if (!projectId) throw new Error('missing project id');
      return fetchProjectSettings(httpClient, baseUrl, projectId);
    },
    enabled: Boolean(projectId),
    retry: 1,
  });

  return {
    settings: query.data ?? null,
    isLoading: query.isLoading,
    error: query.error,
    refetch: query.refetch,
  } as const;
}
