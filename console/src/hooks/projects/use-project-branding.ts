import { useQuery } from '@tanstack/react-query';
import type { HttpClient } from '@/services/core/httpClient';
import { fetchProjectBranding } from '@/services/projects/branding';

export const brandingQueryKey = (projectId: string) => ['project', projectId, 'branding'];

export function useProjectBranding(httpClient: HttpClient, projectId: string) {
  return useQuery({
    queryKey: brandingQueryKey(projectId),
    queryFn: () => fetchProjectBranding(httpClient, import.meta.env.VITE_AUTH_BASE_URL, projectId),
    enabled: Boolean(projectId),
  });
}
