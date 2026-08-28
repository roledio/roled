import type { HttpClient } from '@/services/core/httpClient';
import type { ApiResponse, PaginationInfo, PermissionItem } from './types';

export type { PermissionItem };

export async function fetchProjectPermissions(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  pageNum = 1,
  pageSize = 100,
  search = '',
): Promise<{ data: PermissionItem[]; pagination?: PaginationInfo }> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/permissions`;
  try {
    const res = await httpClient.instanceRef.get<ApiResponse<PermissionItem[]>>(url, {
      params: { page_num: pageNum, page_size: pageSize, search: search || undefined },
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to fetch permissions';
      throw new Error(msg);
    }
    return { data: res.data.data ?? [], pagination: res.data.pagination };
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to fetch permissions';
    throw new Error(msg);
  }
}
