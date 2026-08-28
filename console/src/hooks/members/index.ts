import { buildSigninCallbackRedirect } from '@/lib/redirect';
import type { HttpClient } from '@/services/core/httpClient';
import { deleteMember, fetchMembers, inviteMember, updateMember } from '@/services/members';
import type { UpdateMemberBody } from '@/services/members';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

type MembersParams = {
  httpClient: HttpClient;
  baseUrl: string;
  accountId: string | undefined | null;
  pageNum?: number;
  pageSize?: number;
  search?: string;
  isVerified?: string | null;
  isActive?: string | null;
  isAdmin?: string | null;
  sortBy?: string | null;
  sortDir?: 'asc' | 'desc';
};

export function useMembers({
  httpClient,
  baseUrl,
  accountId,
  pageNum = 1,
  pageSize = 5,
  search = '',
  isVerified = null,
  isActive = null,
  isAdmin = null,
  sortBy = 'updated_at',
  sortDir = 'desc',
}: MembersParams) {
  const query = useQuery({
    queryKey: ['account', accountId, 'members', pageNum, pageSize, search, isVerified, isActive, isAdmin, sortBy, sortDir],
    queryFn: () => {
      const params: Record<string, any> = { page_size: pageSize, page_num: pageNum };
      if (search) params.search = search;
      params.is_verified = isVerified === null ? null : (isVerified === 'true');
      params.is_active = isActive === null ? null : (isActive === 'true');
      params.is_admin = isAdmin === null ? null : (isAdmin === 'true');
      if (sortBy) {
        params.sort_by = sortBy;
        params.sort_dir = sortDir;
      }
      return fetchMembers(httpClient, baseUrl, params);
    },
    enabled: !!accountId,
    retry: 1,
  });

  return query;
}

export function useInviteMember({ httpClient, baseUrl, accountId }: { httpClient: HttpClient; baseUrl: string; accountId?: string | null }) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (email: string) => {
      const redirect = buildSigninCallbackRedirect();
      return inviteMember(httpClient, baseUrl, email, redirect);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['account', accountId, 'members'] });
    },
  });
}

export function useDeleteMember({ httpClient, baseUrl, accountId }: { httpClient: HttpClient; baseUrl: string; accountId?: string | null }) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (memberId: string) => deleteMember(httpClient, baseUrl, memberId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['account', accountId, 'members'] });
    },
  });
}

export function useUpdateMember({ httpClient, baseUrl, accountId }: { httpClient: HttpClient; baseUrl: string; accountId?: string | null }) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ memberId, body }: { memberId: string; body: UpdateMemberBody }) =>
      updateMember(httpClient, baseUrl, memberId, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['account', accountId, 'members'] });
    },
  });
}
