import type { HttpClient } from '@/services/core/httpClient';

export type Account = {
  id: string;
  name: string;
  description: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

type AccountResponse = {
  success: boolean;
  data: Account;
};

export async function fetchCurrentAccount(
  httpClient: HttpClient,
  baseUrl: string,
): Promise<Account> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/accounts/current`;
  try {
    const res = await httpClient.instanceRef.get<AccountResponse>(url);
    if (!res.data?.success || !res.data?.data) {
      const msg = (res as any)?.data?.error?.message ?? 'Invalid account response';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to fetch account';
    throw new Error(msg);
  }
}

export async function updateAccount(
  httpClient: HttpClient,
  baseUrl: string,
  id: string,
  payload: { name?: string; description?: string },
): Promise<Account> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/accounts/${id}`;
  try {
    const res = await httpClient.instanceRef.put<AccountResponse>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success || !res.data?.data) {
      const msg = (res as any)?.data?.error?.message ?? 'Invalid account response';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to update account';
    throw new Error(msg);
  }
}

export async function deleteAccount(
  httpClient: HttpClient,
  baseUrl: string,
  id: string,
  payload: { password?: string },
): Promise<void> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/accounts/${id}/delete`;
  try {
    const res = await httpClient.instanceRef.post<{ success: boolean; error?: { message: string } }>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res.data?.error?.message ?? 'Failed to delete account';
      throw new Error(msg);
    }
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to delete account';
    throw new Error(msg);
  }
}
