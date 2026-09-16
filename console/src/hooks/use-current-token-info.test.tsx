import { TooltipProvider } from '@/components/ui/tooltip';
import type { HttpClient } from '@/services/core/httpClient';
import type { TokenService } from '@/services/core/tokenService';
import type { CurrentTokenInfo } from '@/services/core/authService';
import type { Member } from '@/services/members';
import * as authService from '@/services/core/authService';
import * as memberService from '@/services/members';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useCurrentTokenInfo, useCurrentTokenAndMemberInfo, useRevokeToken } from './use-current-token-info';

vi.mock('@/services/core/authService', () => ({
  fetchCurrentTokenInfo: vi.fn(),
  revokeCurrentToken: vi.fn(),
}));

vi.mock('@/services/members', () => ({
  fetchMembers: vi.fn().mockResolvedValue({ data: [] }),
  inviteMember: vi.fn(),
  deleteMember: vi.fn(),
}));

function createMockTokenService(): TokenService {
  let cachedToken: CurrentTokenInfo | null = null;
  let cachedMember: Member | undefined = undefined;
  return {
    clear: vi.fn(),
    getRefreshToken: vi.fn().mockReturnValue('refresh123'),
    getAccessToken: vi.fn().mockReturnValue('access123'),
    isAccessTokenValid: vi.fn().mockReturnValue(true),
    getCurrentTokenInfo: vi.fn().mockImplementation(() => cachedToken),
    setCurrentTokenInfo: vi.fn().mockImplementation((info) => { cachedToken = info; }),
    getCurrentMember: vi.fn().mockImplementation(() => cachedMember),
    setCurrentMember: vi.fn().mockImplementation((member) => { cachedMember = member; }),
  } as unknown as TokenService;
}

function createMockHttpClient(tokenService?: TokenService): HttpClient {
  const ts = tokenService || createMockTokenService();
  return {
    instanceRef: {
      get: vi.fn(),
      post: vi.fn(),
    },
    tokenServiceRef: ts,
  } as unknown as HttpClient;
}

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: 0 } } });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <TooltipProvider>{children}</TooltipProvider>
      </QueryClientProvider>
    );
  };
}

function TokenInfoComp({ httpClient, authBaseUrl }: { httpClient: HttpClient; authBaseUrl: string }) {
  const q = useCurrentTokenInfo({ httpClient, authBaseUrl });
  if (q.isLoading) return <div>loading-token</div>;
  if (q.error) return <div>error-{q.error.message}</div>;
  return <div>{q.data?.user?.display_name}</div>;
}

describe('useCurrentTokenInfo & useRevokeToken hooks', () => {
  beforeEach(() => {
    vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082');
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
  });

  it('fetches and renders current token user info', async () => {
    const httpClient = createMockHttpClient();
    const mockInfo: CurrentTokenInfo = {
      id: 'token123',
      issued_at: new Date().toISOString(),
      expires_at: new Date(Date.now() + 3600000).toISOString(),
      project: { id: 'proj1', name: 'Test Project', description: 'Test', logo_url: '' },
      client: { id: 'client1', name: 'Test Client' },
      user: { id: 'user123', display_name: 'Test Admin', email: 'admin@local.id' },
      role: { id: 'role1', code: 'admin', name: 'Admin', description: 'Admin role' },
      permissions: [],
    };

    vi.mocked(authService.fetchCurrentTokenInfo).mockResolvedValue(mockInfo);

    render(<TokenInfoComp httpClient={httpClient} authBaseUrl="http://localhost:8082" />, { wrapper: createWrapper() });

    await waitFor(() => expect(screen.getByText('Test Admin')).toBeInTheDocument());
  });

  it('revokes current token successfully', async () => {
    const httpClient = createMockHttpClient();
    const onSuccess = vi.fn();

    function RevokeComp() {
      const mutation = useRevokeToken({
        httpClient,
        authBaseUrl: 'http://localhost:8082',
        onSuccess,
      });
      return (
        <button onClick={() => mutation.mutate({ client_id: 'cid', refresh_token: 'ref' })}>
          revoke
        </button>
      );
    }

    vi.mocked(authService.revokeCurrentToken).mockResolvedValue(undefined);

    render(<RevokeComp />, { wrapper: createWrapper() });

    fireEvent.click(screen.getByRole('button', { name: /revoke/i }));

    await waitFor(() => expect(authService.revokeCurrentToken).toHaveBeenCalledWith(httpClient, 'http://localhost:8082', {
      client_id: 'cid',
      refresh_token: 'ref',
    }));
    expect(onSuccess).toHaveBeenCalled();
  });

  it('fetches and renders combined token and member info', async () => {
    const tokenService = createMockTokenService();
    const httpClient = createMockHttpClient(tokenService);
    const mockInfo: CurrentTokenInfo = {
      id: 'token123',
      issued_at: new Date().toISOString(),
      expires_at: new Date(Date.now() + 3600000).toISOString(),
      project: { id: 'proj1', name: 'Test Project', description: 'Test', logo_url: '' },
      client: { id: 'client1', name: 'Test Client' },
      user: { id: 'user123', display_name: 'Test Admin', email: 'admin@local.id' },
      role: { id: 'role1', code: 'admin', name: 'Admin', description: 'Admin role' },
      permissions: [],
    };
    const mockMembers: Member[] = [
      { id: 'mem123', email: 'admin@local.id', display_name: 'Test Admin', is_admin: true, is_active: true, is_verified: true, created_at: '', updated_at: '' }
    ];

    vi.mocked(authService.fetchCurrentTokenInfo).mockResolvedValue(mockInfo);
    vi.mocked(memberService.fetchMembers).mockResolvedValue({ data: mockMembers });

    function CombinedComp() {
      const q = useCurrentTokenAndMemberInfo({ httpClient, authBaseUrl: 'http://localhost:8082' });
      if (q.isLoading) return <div>loading-combined</div>;
      return (
        <div>
          <span>{q.data?.tokenInfo?.user?.display_name}</span>
          <span>{q.data?.memberInfo?.is_admin ? 'is-admin' : 'not-admin'}</span>
        </div>
      );
    }

    render(<CombinedComp />, { wrapper: createWrapper() });

    await waitFor(() => expect(screen.getByText('Test Admin')).toBeInTheDocument());
    expect(screen.getByText('is-admin')).toBeInTheDocument();
  });
});
