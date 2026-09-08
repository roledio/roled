import type { HttpClient } from '@/services/core/httpClient';
import * as projectService from '@/services/projects';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import SettingsTab from './SettingsTab';

vi.mock('@/services/projects');

const toastMock = vi.fn();
vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({ toast: toastMock }),
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
  return function Wrapper({ children }: { children: any }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

const mockProject = { id: 'p1', code: 'test', name: 'Test' } as any;

const mockSettings = {
  project_id: 'p1',
  is_signup_enabled: false,
  default_signup_role_id: null,
  is_signup_verify_email: false,
  is_forgot_password_enabled: true,
  is_allow_temp_email: false,
};

const mockRoles = [
  {
    id: 'role_1',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    code: 'ADMIN',
    name: 'Admin',
    description: 'Admin role',
  },
];

const mockOAuthConnections = [
  {
    id: 'oauth_1',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    project_id: 'p1',
    provider: 'google',
    credential_type: 'default',
    enabled: true,
  },
];

describe('SettingsTab', () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
  });

  it('renders social connections section', async () => {
    vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082');
    vi.spyOn(projectService, 'fetchProjectSettings').mockResolvedValue({
      data: mockSettings,
    } as any);
    vi.spyOn(projectService, 'fetchProjectRoles').mockResolvedValue({
      data: mockRoles,
      pagination: { page_num: 1, page_size: 100, total_data: 1 },
    } as any);
    vi.spyOn(projectService, 'fetchOAuthConnections').mockResolvedValue({
      data: mockOAuthConnections,
    } as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={['/projects/p1']}>
        <Routes>
          <Route path="/projects/:project_id" element={<SettingsTab httpClient={httpClient} project={mockProject} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Social Connections')).toBeInTheDocument());
  });

  it('displays empty connections state when no connections exist', async () => {
    vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082');
    vi.spyOn(projectService, 'fetchProjectSettings').mockResolvedValue({
      data: mockSettings,
    } as any);
    vi.spyOn(projectService, 'fetchProjectRoles').mockResolvedValue({
      data: mockRoles,
      pagination: { page_num: 1, page_size: 100, total_data: 1 },
    } as any);
    vi.spyOn(projectService, 'fetchOAuthConnections').mockResolvedValue({
      data: [],
    } as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={['/projects/p1']}>
        <Routes>
          <Route path="/projects/:project_id" element={<SettingsTab httpClient={httpClient} project={mockProject} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => {
      expect(screen.queryByText('Showing 0 connections')).toBeInTheDocument();
    });
  });

  it('renders project settings and social connections sections', async () => {
    vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082');
    vi.spyOn(projectService, 'fetchProjectSettings').mockResolvedValue({
      data: mockSettings,
    } as any);
    vi.spyOn(projectService, 'fetchProjectRoles').mockResolvedValue({
      data: mockRoles,
      pagination: { page_num: 1, page_size: 100, total_data: 1 },
    } as any);
    vi.spyOn(projectService, 'fetchOAuthConnections').mockResolvedValue({
      data: mockOAuthConnections,
    } as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={['/projects/p1']}>
        <Routes>
          <Route path="/projects/:project_id" element={<SettingsTab httpClient={httpClient} project={mockProject} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => {
      expect(screen.queryByText('Project Settings')).toBeInTheDocument();
      expect(screen.queryByText('Social Connections')).toBeInTheDocument();
    });
  });
});
