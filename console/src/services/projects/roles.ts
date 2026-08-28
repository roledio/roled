import type { HttpClient } from '@/services/core/httpClient';
import type { ApiResponse, PaginationInfo, Role } from './types';

export type { Role };

export async function fetchProjectRoles(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  pageNum = 1,
  pageSize = 10,
  sortBy = '',
  sortDir = '',
  search = '',
): Promise<{ data: Role[]; pagination?: PaginationInfo }> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/roles`;
  try {
    const res = await httpClient.instanceRef.get<ApiResponse<Role[]>>(url, {
      params: {
        page_num: pageNum,
        page_size: pageSize,
        sort_by: sortBy === '' ? '' : sortBy,
        sort_dir: sortDir === '' ? '' : sortDir,
        search: search || undefined,
      },
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Invalid roles response';
      throw new Error(msg);
    }
    return { data: res.data.data ?? [], pagination: res.data.pagination };
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to fetch roles';
    throw new Error(msg);
  }
}

export async function fetchRoleById(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  roleId: string,
): Promise<any> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/roles/${roleId}`;
  try {
    const res = await httpClient.instanceRef.get<ApiResponse<any>>(url, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to fetch role';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to fetch role';
    throw new Error(msg);
  }
}

export async function createProjectRole(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  payload: { name: string; code?: string; description?: string; permission_ids?: string[] },
): Promise<Role> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/roles`;
  try {
    const res = await httpClient.instanceRef.post<ApiResponse<Role>>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to create role';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to create role';
    throw new Error(msg);
  }
}

export async function updateProjectRole(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  roleId: string,
  payload: { name?: string; code?: string; description?: string; permission_ids?: string[] },
): Promise<any> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/roles/${roleId}`;
  try {
    const res = await httpClient.instanceRef.put<ApiResponse<any>>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to update role';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to update role';
    throw new Error(msg);
  }
}

export async function deleteProjectRole(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  roleId: string,
): Promise<void> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/roles/${roleId}`;
  try {
    const res = await httpClient.instanceRef.delete<ApiResponse<void>>(url, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = (res as any)?.data?.error?.message ?? 'Failed to delete role';
      throw new Error(msg);
    }
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to delete role';
    throw new Error(msg);
  }
}

export async function setSignupRole(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  roleId: string,
): Promise<void> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/signup-role`;
  try {
    const res = await httpClient.instanceRef.patch<ApiResponse<void>>(url, { role_id: roleId }, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = (res as any)?.data?.error?.message ?? 'Failed to set sign-up role';
      throw new Error(msg);
    }
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to set sign-up role';
    throw new Error(msg);
  }
}
