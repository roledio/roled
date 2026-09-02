import type { HttpClient } from '@/services/core/httpClient';
import type { ApiResponse, PaginationInfo, User } from './types';

export type { User };

export async function fetchProjectUsers(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  pageNum = 1,
  pageSize = 10,
  sortBy = '',
  sortDir = '',
  search = '',
  isActive: boolean | null = null,
  roleId = '',
): Promise<{ data: User[]; pagination?: PaginationInfo }> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/users`;
  try {
    const params: Record<string, any> = {
      page_num: pageNum,
      page_size: pageSize,
      sort_by: sortBy === '' ? '' : sortBy,
      sort_dir: sortDir === '' ? '' : sortDir,
      search: search || undefined,
    };
    if (isActive !== null) {
      params.is_active = isActive;
    }
    if (roleId) {
      params.role_id = roleId;
    }
    const res = await httpClient.instanceRef.get<ApiResponse<User[]>>(url, {
      params,
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Invalid users response';
      throw new Error(msg);
    }
    return { data: res.data.data ?? [], pagination: res.data.pagination };
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to fetch users';
    throw new Error(msg);
  }
}

export async function fetchUserById(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  userId: string,
): Promise<any> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/users/${userId}`;
  try {
    const res = await httpClient.instanceRef.get<ApiResponse<any>>(url, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to fetch user';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to fetch user';
    throw new Error(msg);
  }
}

export async function createProjectUser(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  payload: { display_name: string; email?: string; password?: string; external_user_id?: string; role_id?: string; avatar_url?: string | null },
): Promise<any> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/users`;
  try {
    const res = await httpClient.instanceRef.post<ApiResponse<any>>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to create user';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to create user';
    throw new Error(msg);
  }
}

export async function updateProjectUser(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  userId: string,
  payload: { display_name?: string; email?: string; password?: string; external_user_id?: string; is_active?: boolean; role_id?: string; avatar_url?: string | null },
): Promise<any> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/users/${userId}`;
  try {
    const res = await httpClient.instanceRef.put<ApiResponse<any>>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to update user';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to update user';
    throw new Error(msg);
  }
}

export async function deleteProjectUser(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  userId: string,
): Promise<void> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/users/${userId}`;
  try {
    const res = await httpClient.instanceRef.delete<ApiResponse<void>>(url, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to delete user';
      throw new Error(msg);
    }
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to delete user';
    throw new Error(msg);
  }
}

export async function resendProjectUserVerificationEmail(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  userId: string,
  payload?: { redirect_uri?: string },
): Promise<void> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/users/${userId}/verification-email`;
  try {
    const res = await httpClient.instanceRef.post<ApiResponse<void>>(url, payload ?? {}, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to send verification email';
      throw new Error(msg);
    }
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to send verification email';
    throw new Error(msg);
  }
}

export async function requestProjectUserPasswordReset(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  userId: string,
  payload?: { redirect_uri?: string },
): Promise<void> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/users/${userId}/password-reset`;
  try {
    const res = await httpClient.instanceRef.post<ApiResponse<void>>(url, payload ?? {}, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to request password reset';
      throw new Error(msg);
    }
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to request password reset';
    throw new Error(msg);
  }
}

export async function inviteProjectUser(
  httpClient: HttpClient,
  baseUrl: string,
  projectId: string,
  payload: { email: string; role_id?: string; redirect_uri?: string },
): Promise<any> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/projects/${projectId}/users/invitation`;
  try {
    const res = await httpClient.instanceRef.post<ApiResponse<any>>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res?.data?.error?.message ?? 'Failed to invite user';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to invite user';
    throw new Error(msg);
  }
}
