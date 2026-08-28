import type { HttpClient } from '@/services/core/httpClient';
import type { ApiResponse, PaginationInfo, ProjectClient } from './types';

export type { ProjectClient };

export async function fetchProjectClients(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  pageNum = 1,
  pageSize = 10,
  sortBy = '',
  sortDir = '',
  search = '',
  isActive: boolean | null = null,
): Promise<{ data: ProjectClient[]; pagination?: PaginationInfo }> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/clients`;
  try {
    const res = await httpClient.instanceRef.get<ApiResponse<ProjectClient[]>>(url, {
      params: {
        page_num: pageNum,
        page_size: pageSize,
        sort_by: sortBy === '' ? '' : sortBy,
        sort_dir: sortDir === '' ? '' : sortDir,
        search: search || undefined,
        is_active: isActive,
      },
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = (res as any)?.data?.error?.message ?? 'Invalid clients response';
      throw new Error(msg);
    }
    return { data: res.data.data ?? [], pagination: res.data.pagination };
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to fetch clients';
    throw new Error(msg);
  }
}

export async function fetchClientById(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  clientId: string,
): Promise<any> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/clients/${clientId}`;
  try {
    const res = await httpClient.instanceRef.get<ApiResponse<any>>(url, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to fetch client';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to fetch client';
    throw new Error(msg);
  }
}

export async function createClient(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  payload: { name: string; description?: string; permission_ids: string[] },
): Promise<any> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/clients`;
  try {
    const res = await httpClient.instanceRef.post<ApiResponse<any>>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to create client';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to create client';
    throw new Error(msg);
  }
}

export async function updateClient(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  clientId: string,
  payload: { name?: string; description?: string; is_active?: boolean; permission_ids?: string[] },
): Promise<any> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/clients/${clientId}`;
  try {
    const res = await httpClient.instanceRef.put<ApiResponse<any>>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to update client';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to update client';
    throw new Error(msg);
  }
}

export async function deleteClient(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  clientId: string,
): Promise<void> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/clients/${clientId}`;
  try {
    const res = await httpClient.instanceRef.delete<ApiResponse<void>>(url, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = (res as any)?.data?.error?.message ?? 'Failed to delete client';
      throw new Error(msg);
    }
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to delete client';
    throw new Error(msg);
  }
}
