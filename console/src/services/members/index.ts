import type { HttpClient } from '@/services/core/httpClient';

export type Member = {
  id: string;
  email: string;
  display_name: string;
  is_active: boolean;
  is_verified: boolean;
  is_admin: boolean;
  created_at: string;
  updated_at: string;
  avatar_url?: string | null;
};

type MembersResponse = {
  success: boolean;
  data: Member[];
  pagination?: {
    page_num: number;
    page_size: number;
    total_data: number;
  };
};

export async function fetchMembers(
  httpClient: HttpClient,
  baseUrl: string,
  params: Record<string, any> = {},
): Promise<{ data: Member[]; pagination?: MembersResponse['pagination'] }> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/members`;
  try {
    const res = await httpClient.instanceRef.get<MembersResponse>(url, {
      params,
      headers: { 'Content-Type': 'application/json' },
    });

    if (!res.data?.success) {
      const msg = (res as any)?.data?.error?.message ?? 'Invalid members response';
      throw new Error(msg);
    }

    return { data: res.data.data ?? [], pagination: res.data.pagination };
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to fetch members';
    throw new Error(msg);
  }
}

export async function inviteMember(
  httpClient: HttpClient,
  baseUrl: string,
  email: string,
  redirectUri?: string,
): Promise<Member> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/members`;

  const body: Record<string, any> = { email };
  if (redirectUri) body.redirect_uri = redirectUri;

  try {
    const res = await httpClient.instanceRef.post<{ success: boolean; data: Member }>(url, body, {
      headers: { 'Content-Type': 'application/json' },
    });

    if (!res.data?.success) {
      const msg = (res as any)?.data?.error?.message ?? 'Failed to invite member';
      throw new Error(msg);
    }

    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to invite member';
    throw new Error(msg);
  }
}

export async function deleteMember(
  httpClient: HttpClient,
  baseUrl: string,
  memberId: string,
): Promise<void> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/members/${encodeURIComponent(memberId)}`;
  try {
    const res = await httpClient.instanceRef.delete<{ success: boolean }>(url, {
      headers: { 'Content-Type': 'application/json' },
    });

    if (!res.data?.success) {
      const msg = (res as any)?.data?.error?.message ?? 'Failed to delete member';
      throw new Error(msg);
    }

    return;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to delete member';
    throw new Error(msg);
  }
}

export type UpdateMemberBody = {
  is_admin?: boolean;
  account_id?: string;
};

export type UpdateMemberResult = {
  id: string;
  account_id: string;
  user_id: string;
  is_admin: boolean;
  created_at: string;
  updated_at: string;
};

export async function updateMember(
  httpClient: HttpClient,
  baseUrl: string,
  memberId: string,
  body: UpdateMemberBody,
): Promise<UpdateMemberResult> {
  const url = `${baseUrl.replace(/\/$/, '')}/api/v1/members/${encodeURIComponent(memberId)}`;
  try {
    const res = await httpClient.instanceRef.patch<{ success: boolean; data: UpdateMemberResult }>(url, body, {
      headers: { 'Content-Type': 'application/json' },
    });

    if (!res.data?.success) {
      const msg = (res as any)?.data?.error?.message ?? 'Failed to update member';
      throw new Error(msg);
    }

    return res.data.data;
  } catch (err: any) {
    const msg = err?.response?.data?.error?.message ?? err?.response?.data?.message ?? err?.message ?? 'Failed to update member';
    throw new Error(msg);
  }
}
