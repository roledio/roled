import type { HttpClient } from '@/services/core/httpClient';
import type { ApiResponse, PaginationInfo } from './types';

export type OAuthConnection = {
  id: string;
  created_at: string;
  updated_at: string;
  project_id: string;
  provider: 'google' | string;
  credential_type: 'default' | 'custom';
  enabled: boolean;
};

export async function fetchOAuthConnections(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
): Promise<{ data: OAuthConnection[]; pagination?: PaginationInfo }> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/oauth-connections`;
  try {
    const res = await httpClient.instanceRef.get<ApiResponse<OAuthConnection[]>>(url, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Invalid oauth connections response';
      throw new Error(msg);
    }
    return { data: res.data.data ?? [], pagination: res.data.pagination };
  } catch (err: any) {
    const msg =
      err?.response?.data?.error?.message ??
      err?.response?.data?.message ??
      err?.message ??
      'Failed to fetch oauth connections';
    throw new Error(msg);
  }
}

export async function createOAuthConnection(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  payload: { provider: string; credential_type?: 'default' | 'custom' },
): Promise<OAuthConnection> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/oauth-connections`;
  try {
    const res = await httpClient.instanceRef.post<ApiResponse<OAuthConnection>>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to create oauth connection';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg =
      err?.response?.data?.error?.message ??
      err?.response?.data?.message ??
      err?.message ??
      'Failed to create oauth connection';
    throw new Error(msg);
  }
}

export async function deleteOAuthConnection(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  provider: string,
): Promise<void> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/oauth-connections/${provider}`;
  try {
    const res = await httpClient.instanceRef.delete<ApiResponse<void>>(url, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to delete oauth connection';
      throw new Error(msg);
    }
  } catch (err: any) {
    const msg =
      err?.response?.data?.error?.message ??
      err?.response?.data?.message ??
      err?.message ??
      'Failed to delete oauth connection';
    throw new Error(msg);
  }
}

export type { OAuthConnection as OAuthConnectionType };
