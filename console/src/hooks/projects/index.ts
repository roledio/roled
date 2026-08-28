import type { HttpClient } from '@/services/core/httpClient';
import { fetchProjectById, type Project } from '@/services/projects';
import { useQuery } from '@tanstack/react-query';

type Options = { httpClient: HttpClient; baseUrl: string; projectId?: string; projectCode?: string };

export function useProject({ httpClient, baseUrl, projectId, projectCode }: Options) {
  const query = useQuery<Project | null, Error>({
    queryKey: ['project', projectId ?? projectCode ?? 'unknown'],
    queryFn: async () => {
      if (projectId) return fetchProjectById(httpClient, baseUrl, projectId);
      throw new Error('missing project identifier');
    },
    retry: 1,
  });

  return { project: query.data ?? null, isLoading: query.isLoading, error: query.error } as const;
}

export { useProjectClients } from './use-project-clients';
export { useProjectSettings } from './use-project-settings';

