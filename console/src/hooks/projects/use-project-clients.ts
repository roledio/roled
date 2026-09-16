import { useQuery } from '@tanstack/react-query';
import type { UseQueryResult } from '@tanstack/react-query';
import type { HttpClient } from '@/services/core/httpClient';
import { fetchProjectClients, type ProjectClient } from '@/services/projects';

type Pagination = {
  page_num: number;
  page_size: number;
  total_data: number;
};

type Options = {
  httpClient: HttpClient;
  baseUrl: string;
  projectId?: string;
  pageNum?: number;
  pageSize?: number;
  search?: string;
  isActive?: string | null; // 'true' | 'false' | null
  sortBy?: string | null;
  sortDir?: 'asc' | 'desc' | null;
};

export function useProjectClients({ httpClient, baseUrl, projectId, pageNum = 1, pageSize = 10, search = '', isActive = null, sortBy = null, sortDir = null }: Options) {
  const enabled = Boolean(projectId && projectId.length > 0);
  const isActiveBool = isActive === null ? null : isActive === 'true';

  const query = useQuery({
    queryKey: ['project', projectId ?? 'unknown', 'clients', pageNum, pageSize, search || '', isActive ?? 'all', sortBy ?? '', sortDir ?? ''],
    queryFn: () => fetchProjectClients(httpClient, baseUrl, projectId!, pageNum, pageSize, sortBy ?? '', sortDir ?? '', search ?? '', isActiveBool),
    enabled,
    placeholderData: (previousData) => previousData,
    retry: 1,
  }) as UseQueryResult<{ data: ProjectClient[]; pagination?: Pagination }, Error>;

  return {
    clients: query.data?.data ?? [],
    pagination: query.data?.pagination ?? null,
    isLoading: query.isLoading,
    error: query.error,
  } as const;
}
