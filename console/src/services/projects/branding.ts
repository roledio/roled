import { isAxiosError } from 'axios';
import type { HttpClient } from '@/services/core/httpClient';
import type { ApiResponse } from '@/services/projects/types';

export type BrandingSettings = {
  logo_url: string | null;
  favicon_url: string | null;
  primary_color: string;
  rounding: 'sharp' | 'small' | 'medium' | 'large';
  enable_shadow: boolean;
  enable_border: boolean;
};

export type ProjectBranding = BrandingSettings & {
  project_id: string;
  source_project_id: string;
  is_default: boolean;
};

export function brandingError(error: unknown): string {
  if (isAxiosError<ApiResponse<never>>(error)) {
    return error.response?.data?.error?.message ?? error.message;
  }
  return error instanceof Error ? error.message : 'Unable to update branding';
}

export async function fetchProjectBranding(httpClient: HttpClient, baseUrl: string, projectId: string): Promise<ProjectBranding> {
  const response = await httpClient.instanceRef.get<ApiResponse<ProjectBranding>>(`${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/branding`);
  if (!response.data.success) throw new Error(response.data.error?.message ?? 'Unable to load branding');
  return response.data.data;
}

export async function updateProjectBranding(httpClient: HttpClient, baseUrl: string, projectId: string, settings: BrandingSettings): Promise<ProjectBranding> {
  const response = await httpClient.instanceRef.put<ApiResponse<ProjectBranding>>(`${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/branding`, settings);
  if (!response.data.success) throw new Error(response.data.error?.message ?? 'Unable to save branding');
  return response.data.data;
}

export async function uploadBrandingAsset(httpClient: HttpClient, baseUrl: string, file: File, type: 'project-logo' | 'favicon'): Promise<string> {
  const body = new FormData();
  body.append('file', file);
  body.append('type', type);
  const response = await httpClient.instanceRef.post<ApiResponse<{ url: string }>>(`${baseUrl.replace(/\/$/, '')}/api/v1/uploads`, body, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  if (!response.data.success) throw new Error(response.data.error?.message ?? 'Unable to upload image');
  return response.data.data.url;
}
