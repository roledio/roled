import { useMutation, useQuery } from '@tanstack/react-query';
import type { HttpClient } from '@/services/core/httpClient';
import { fetchCurrentAccount, deleteAccount, type Account } from '@/services/accounts';

type UseCurrentAccountOptions = {
    httpClient: HttpClient;
    baseUrl: string;
};

export function useCurrentAccount({ httpClient, baseUrl }: UseCurrentAccountOptions) {
    const query = useQuery<Account, Error>({
        queryKey: ['account', 'current'],
        queryFn: () => fetchCurrentAccount(httpClient, baseUrl),
        retry: 1,
        staleTime: 5 * 60 * 1000, // 5 minutes
    });

    return {
        account: query.data ?? null,
        isLoading: query.isLoading,
        error: query.error,
    } as const;
}

export function useDeleteAccount({ httpClient, baseUrl, accountId }: { httpClient: HttpClient; baseUrl: string; accountId?: string | null }) {
    return useMutation({
        mutationFn: (payload: { password?: string }) => {
            if (!accountId) return Promise.reject(new Error('Account ID is missing'));
            return deleteAccount(httpClient, baseUrl, accountId, payload);
        },
    });
}
