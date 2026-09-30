import { useQuery } from '@tanstack/react-query';
import { useLocation, useNavigate, useParams } from 'react-router-dom';
import { getProjectTabParams } from '@/lib/paramsStore';
import type { HttpClient } from '@/services/core/httpClient';
import { fetchProjectById } from '@/services/projects';
import { PROVIDER_CONFIG } from './social-providers';

export function useSocialConnectionContext(httpClient: HttpClient) {
  const { project_id, provider } = useParams<{
    project_id: string;
    provider: string;
  }>();
  const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL as string;
  const navigate = useNavigate();
  const location = useLocation();

  const navigationState = location.state as { from?: string } | null;
  const paramsFrom =
    getProjectTabParams(project_id ?? '', 'settings') || navigationState?.from || '';

  const providerConfig = provider ? PROVIDER_CONFIG[provider] : undefined;

  // ─── Project data ───────────────────────────────────────────────────────────
  const projectQuery = useQuery({
    queryKey: ['project', project_id],
    queryFn: () => fetchProjectById(httpClient, AUTH_BASE_URL, project_id!),
    enabled: !!project_id,
    retry: 1,
  });

  return { project_id, provider, AUTH_BASE_URL, navigate, paramsFrom, providerConfig, projectQuery };
}
