import type { HttpClient } from '@/services/core/httpClient';
import {
  fetchCurrentTokenInfo,
  revokeCurrentToken,
  type CurrentTokenInfo,
  type RevokeTokenPayload,
} from '@/services/core/authService';
import { fetchMembers, type Member } from '@/services/members';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

type Options = {
  httpClient: HttpClient;
  authBaseUrl: string;
  enabled?: boolean;
};

export function useCurrentTokenInfo({ httpClient, authBaseUrl, enabled = true }: Options) {
  return useQuery<CurrentTokenInfo, Error>({
    queryKey: ['currentTokenInfo'],
    queryFn: async () => {
      const tokenService = httpClient.tokenServiceRef;
      const cached = tokenService.getCurrentTokenInfo();
      if (cached) return cached;
      const info = await fetchCurrentTokenInfo(httpClient, authBaseUrl);
      tokenService.setCurrentTokenInfo(info);
      return info;
    },
    enabled,
    retry: 1,
    staleTime: 5 * 60 * 1000, // consider fresh for 5 minutes
  });
}

export type CurrentTokenAndMemberInfo = {
  tokenInfo: CurrentTokenInfo;
  memberInfo: Member | null;
};

export async function fetchCurrentTokenAndMemberInfo(
  httpClient: HttpClient,
  authBaseUrl: string
): Promise<CurrentTokenAndMemberInfo> {
  const tokenService = httpClient.tokenServiceRef;
  const cachedToken = tokenService.getCurrentTokenInfo();
  const cachedMember = tokenService.getCurrentMember();

  if (cachedToken && cachedMember !== undefined) {
    return {
      tokenInfo: cachedToken,
      memberInfo: cachedMember,
    };
  }

  const tokenInfo = await fetchCurrentTokenInfo(httpClient, authBaseUrl);
  let memberInfo: Member | null = null;

  if (tokenInfo?.user?.email) {
    try {
      const res = await fetchMembers(httpClient, authBaseUrl, {
        search: tokenInfo.user.email,
      });
      if (res && res.data) {
        memberInfo = res.data.find((m) => m.email === tokenInfo.user.email) || null;
      }
    } catch (err) {
      console.error('Failed to fetch member details for role check', err);
    }
  }

  tokenService.setCurrentTokenInfo(tokenInfo);
  tokenService.setCurrentMember(memberInfo);

  return { tokenInfo, memberInfo };
}

export function useCurrentTokenAndMemberInfo({ httpClient, authBaseUrl, enabled = true }: Options) {
  return useQuery<CurrentTokenAndMemberInfo, Error>({
    queryKey: ['currentTokenAndMemberInfo'],
    queryFn: () => fetchCurrentTokenAndMemberInfo(httpClient, authBaseUrl),
    enabled,
    retry: 1,
    staleTime: 5 * 60 * 1000, // consider fresh for 5 minutes
  });
}

type RevokeOptions = {
  httpClient: HttpClient;
  authBaseUrl: string;
  onSuccess?: () => void;
  onError?: (error: Error) => void;
};

export function useRevokeToken({ httpClient, authBaseUrl, onSuccess, onError }: RevokeOptions) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: RevokeTokenPayload) => revokeCurrentToken(httpClient, authBaseUrl, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['currentTokenInfo'] });
      queryClient.invalidateQueries({ queryKey: ['currentTokenAndMemberInfo'] });
      if (onSuccess) onSuccess();
    },
    onError: (error: Error) => {
      if (onError) onError(error);
    },
  });
}
