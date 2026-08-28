import type { HttpClient } from '@/services/core/httpClient';

export type CurrentUser = {
  id: string;
  created_at: string;
  updated_at: string;
  email: string;
  external_user_id?: string | null;
  display_name: string;
  avatar_url?: string | null;
  is_active: boolean;
  is_email_verified: boolean;
  role_id?: string;
  role_name?: string;
};

export type UpdateCurrentUserPayload = {
  display_name?: string;
  email?: string;
  password?: string;
  avatar_url?: string | null;
};

type CurrentUserResponse = {
  success: boolean;
  data: CurrentUser;
  error?: { message: string };
};

export async function fetchCurrentUser(httpClient: HttpClient, baseUrl: string): Promise<CurrentUser> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/users/current`;
  try {
    const res = await httpClient.instanceRef.get<CurrentUserResponse>(url, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success || !res.data?.data) {
      const msg = (res as any)?.data?.error?.message ?? 'Invalid user response';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to fetch current user';
    throw new Error(msg);
  }
}

type UpdateCurrentUserResponse = {
  success: boolean;
  data: CurrentUser;
  error?: { message: string };
};

export async function updateCurrentUser(
  httpClient: HttpClient,
  baseUrl: string,
  payload: UpdateCurrentUserPayload,
): Promise<CurrentUser> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/users/current`;
  try {
    const res = await httpClient.instanceRef.put<UpdateCurrentUserResponse>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success || !res.data?.data) {
      const msg = res.data?.error?.message ?? 'Failed to update current user';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to update current user';
    throw new Error(msg);
  }
}
