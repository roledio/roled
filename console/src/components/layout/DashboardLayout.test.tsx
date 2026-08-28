import { TooltipProvider } from '@/components/ui/tooltip';
import type { HttpClient } from '@/services/core/httpClient';
import type { TokenService } from '@/services/core/tokenService';
import * as authService from '@/services/core/authService';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { DashboardLayout } from './DashboardLayout';
import { MemoryRouter } from 'react-router-dom';

vi.mock('@/services/core/authService', () => ({
  fetchCurrentTokenInfo: vi.fn(),
  revokeCurrentToken: vi.fn(),
}));

vi.mock('@/services/members', () => ({
  fetchMembers: vi.fn().mockResolvedValue({ data: [], pagination: {} }),
  inviteMember: vi.fn(),
  deleteMember: vi.fn(),
}));

// Mock AppSidebar to simplify sidebar layout dependencies in unit tests
vi.mock('./AppSidebar', () => ({
  AppSidebar: () => <div>sidebar-mock</div>,
}));

function createMockHttpClient(tokenService: TokenService): HttpClient {
  return {
    instanceRef: {
      get: vi.fn(),
      post: vi.fn(),
    },
    tokenServiceRef: tokenService,
  } as unknown as HttpClient;
}

function createMockTokenService(): TokenService {
  let cachedToken: any = null;
  let cachedMember: any = undefined;
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

describe('DashboardLayout Component', () => {
  beforeEach(() => {
    vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082');
    
    // Mock window.location.href to prevent JSDOM navigation errors
    const mockLocation = new URL('http://localhost');
    Object.defineProperty(mockLocation, 'href', {
      writable: true,
      value: 'http://localhost',
    });
    vi.stubGlobal('location', mockLocation);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
  });

  it('renders loading placeholder initially', async () => {
    const tokenService = createMockTokenService();
    const httpClient = createMockHttpClient(tokenService);
    vi.mocked(authService.fetchCurrentTokenInfo).mockReturnValue(new Promise(() => {}));

    const { container } = render(
      <MemoryRouter>
        <DashboardLayout httpClient={httpClient} tokenService={tokenService}>
          <div>content</div>
        </DashboardLayout>
      </MemoryRouter>,
      { wrapper: createWrapper() }
    );

    expect(container.querySelector('.animate-pulse')).toBeInTheDocument();
  });

  it('renders user initials as avatar fallback and opens menu on click', async () => {
    const tokenService = createMockTokenService();
    const httpClient = createMockHttpClient(tokenService);
    const mockInfo = {
      id: 'token123',
      client: { id: 'cid' },
      user: { display_name: 'John Doe', email: 'john@doe.com' },
    };

    vi.mocked(authService.fetchCurrentTokenInfo).mockResolvedValue(mockInfo as any);

    render(
      <MemoryRouter>
        <DashboardLayout httpClient={httpClient} tokenService={tokenService}>
          <div>content</div>
        </DashboardLayout>
      </MemoryRouter>,
      { wrapper: createWrapper() }
    );

    // Wait for initials
    await waitFor(() => expect(screen.getByText('J')).toBeInTheDocument());

    // Click user menu button to open dropdown
    const userMenuBtn = screen.getByLabelText('User menu');
    await userEvent.click(userMenuBtn);

    // User display name & email should be visible
    expect(await screen.findByText('John Doe')).toBeInTheDocument();
    expect(await screen.findByText('john@doe.com')).toBeInTheDocument();
  });

  it('calls revoke API and clears storage on Sign Out click', async () => {
    const tokenService = createMockTokenService();
    const httpClient = createMockHttpClient(tokenService);
    const mockInfo = {
      id: 'token123',
      client: { id: 'client123' },
      user: { display_name: 'Jane Smith', email: 'jane@smith.com' },
    };

    vi.mocked(authService.fetchCurrentTokenInfo).mockResolvedValue(mockInfo as any);
    vi.mocked(authService.revokeCurrentToken).mockResolvedValue(undefined);

    render(
      <MemoryRouter>
        <DashboardLayout httpClient={httpClient} tokenService={tokenService}>
          <div>content</div>
        </DashboardLayout>
      </MemoryRouter>,
      { wrapper: createWrapper() }
    );

    await waitFor(() => expect(screen.getByText('J')).toBeInTheDocument());

    const userMenuBtn = screen.getByLabelText('User menu');
    await userEvent.click(userMenuBtn);

    const signOutBtn = await screen.findByText('Sign Out');
    await userEvent.click(signOutBtn);

    await waitFor(() =>
      expect(authService.revokeCurrentToken).toHaveBeenCalledWith(httpClient, 'http://localhost:8082', {
        client_id: 'client123',
        refresh_token: 'refresh123',
      })
    );
    expect(tokenService.clear).toHaveBeenCalled();
  });
});
