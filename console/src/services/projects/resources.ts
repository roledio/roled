import type { HttpClient } from '@/services/core/httpClient';
import type { ApiResponse, PaginationInfo, Resource } from './types';

export type { Resource };

export async function fetchProjectResources(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  pageNum = 1,
  pageSize = 10,
  sortBy = '',
  sortDir = '',
  search = '',
  isDefault: boolean | null = null,
): Promise<{ data: Resource[]; pagination?: PaginationInfo }> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/resources`;
  try {
    const res = await httpClient.instanceRef.get<ApiResponse<Resource[]>>(url, {
      params: { page_num: pageNum, page_size: pageSize, sort_by: sortBy === '' ? '' : sortBy, sort_dir: sortDir === '' ? '' : sortDir, search: search || undefined, is_default: isDefault },
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to fetch resources';
      throw new Error(msg);
    }
    return { data: res.data.data ?? [], pagination: res.data.pagination };
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to fetch resources';
    throw new Error(msg);
  }
}

export async function fetchProjectResourceById(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  resourceId: string,
): Promise<any> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/resources/${resourceId}`;
  try {
    const res = await httpClient.instanceRef.get<ApiResponse<any>>(url, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to fetch resource';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to fetch resource';
    throw new Error(msg);
  }
}

export async function createProjectResource(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  payload: { name: string; code?: string; description?: string; permissions?: Array<{ name: string; code: string; description?: string }> },
): Promise<any> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/resources`;
  try {
    const res = await httpClient.instanceRef.post<ApiResponse<any>>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to create resource';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to create resource';
    throw new Error(msg);
  }
}

export async function updateProjectResource(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  resourceId: string,
  payload: { name?: string; code?: string; description?: string; permissions?: Array<{ name: string; code: string; description?: string }> },
): Promise<any> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/resources/${resourceId}`;
  try {
    const res = await httpClient.instanceRef.put<ApiResponse<any>>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to update resource';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to update resource';
    throw new Error(msg);
  }
}

export async function deleteProjectResource(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  resourceId: string,
): Promise<void> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/resources/${resourceId}`;
  try {
    const res = await httpClient.instanceRef.delete<ApiResponse<void>>(url, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = (res as any)?.data?.error?.message ?? 'Failed to delete resource';
      throw new Error(msg);
    }
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to delete resource';
    throw new Error(msg);
  }
}
