import { TooltipProvider } from '@/components/ui/tooltip';
import type { HttpClient } from '@/services/core/httpClient';
import * as memberService from '@/services/members';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useInviteMember, useMembers } from '.';

vi.mock('@/services/members', () => ({
  fetchMembers: vi.fn(),
  inviteMember: vi.fn(),
  deleteMember: vi.fn(),
}));

function createMockHttpClient(): HttpClient {
  return {
    instanceRef: {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
    },
  } as unknown as HttpClient;
}

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: 0 } } });

  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <TooltipProvider>
          {children}
        </TooltipProvider>
      </QueryClientProvider>
    );
  };
}

function MembersList({ httpClient, baseUrl, accountId }: { httpClient: HttpClient; baseUrl: string; accountId: string }) {
  const q = useMembers({ httpClient, baseUrl, accountId });
  if (q.isLoading) return <div>loading-members</div>;
  return (
    <div>
      {(q.data?.data ?? []).map((m: any) => (
        <div key={m.id}>{m.display_name ?? m.email}</div>
      ))}
    </div>
  );
}

describe('use-members hook', () => {
  beforeEach(() => {
    vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082');
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
  });

  it('fetches and renders members', async () => {
    const httpClient = createMockHttpClient();
    const baseUrl = 'http://localhost:8082';
    const accountId = 'acc1';

    const members = [
      { id: 'm1', email: 'a@example.com', display_name: 'Member A', is_active: true, is_verified: true, is_admin: false, created_at: '', updated_at: '' },
    ];

    vi.mocked(memberService.fetchMembers).mockResolvedValue({ data: members, pagination: { page_num: 1, page_size: 5, total_data: 1 } });

    render(<MembersList httpClient={httpClient} baseUrl={baseUrl} accountId={accountId} />, { wrapper: createWrapper() });

    await waitFor(() => expect(screen.getByText('Member A')).toBeInTheDocument());
  });

  it('invite mutation calls inviteMember', async () => {
    const httpClient = createMockHttpClient();
    const baseUrl = 'http://localhost:8082';
    const accountId = 'acc1';

    function InviteComp() {
      const m = useInviteMember({ httpClient, baseUrl, accountId });
      return (
        <div>
          <button onClick={() => m.mutateAsync('new@example.com')}>invite</button>
        </div>
      );
    }

    vi.mocked(memberService.inviteMember).mockResolvedValue({ id: 'm2', email: 'new@example.com' } as any);

    render(<InviteComp />, { wrapper: createWrapper() });

    fireEvent.click(screen.getByRole('button', { name: /invite/i }));

    await waitFor(() => expect(memberService.inviteMember).toHaveBeenCalledWith(httpClient, baseUrl, 'new@example.com', expect.any(String)));
  });
});
