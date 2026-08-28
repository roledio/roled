import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import * as projectService from '@/services/projects';
import ResourcesTab from './ResourcesTab';
import type { HttpClient } from '@/services/core/httpClient';

vi.mock('@/services/projects');

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

describe('ResourcesTab', () => {
  beforeEach(() => {
    vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082');
  });
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
  });

  it('renders resources and permission cards and hides Remove for default resource', async () => {
    const project = { id: 'p1', code: 'test', name: 'Test' } as any;
    const resources = [
      { id: 'r1', name: 'accounts', description: 'Accounts', is_default: true, permissions: [{ id: 'perm1', name: 'create', description: 'Create accounts' }] },
    ];

    vi.spyOn(projectService, 'fetchProjectResources').mockResolvedValue({ data: resources, pagination: { page_num: 1, page_size: 10, total_data: 1 } } as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter>
        <ResourcesTab httpClient={httpClient} project={project} />
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('accounts')).toBeInTheDocument());
    expect(screen.getByText('Resources')).toBeInTheDocument();
    expect(screen.getByText('Manage the resources and permissions of this project')).toBeInTheDocument();
    // permissions are shown in the Permissions column
    await waitFor(() => expect(screen.getByText('create')).toBeInTheDocument());

    // open menu and assert Remove not present for default resource
    const buttons = screen.getAllByRole('button');
    const menuButton = buttons.find((b) => b.getAttribute('class')?.includes('h-7'))!;
    await userEvent.click(menuButton);
    await waitFor(() => expect(screen.getByText('View Details')).toBeInTheDocument());
    expect(screen.queryByText('Remove')).toBeNull();
  });

  it('removes a non-default resource via deleteProjectResource', async () => {
    const project = { id: 'p1', code: 'test', name: 'Test' } as any;
    const resources = [
      { id: 'r2', name: 'users', description: 'Users', is_default: false, permissions: [] },
    ];

    vi.spyOn(projectService, 'fetchProjectResources').mockResolvedValue({ data: resources, pagination: { page_num: 1, page_size: 10, total_data: 1 } } as any);
    const deleteMock = vi.spyOn(projectService, 'deleteProjectResource').mockResolvedValue(undefined as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter>
        <ResourcesTab httpClient={httpClient} project={project} />
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('users')).toBeInTheDocument());

    // open menu and click Remove
    const buttons = screen.getAllByRole('button');
    // find the menu button (uses MoreVertical with a small button)
    const menuBtn = buttons.find((b) => b.getAttribute('class')?.includes('h-7'))!;
    await userEvent.click(menuBtn);

    await waitFor(() => expect(screen.queryByText('Remove')).not.toBeNull());
    fireEvent.click(screen.getByText('Remove'));

    // confirm dialog appears
    await waitFor(() => expect(screen.getByText('Remove Resource')).toBeInTheDocument());
    fireEvent.click(screen.getByText('Remove'));

    await waitFor(() => expect(deleteMock).toHaveBeenCalled());
  });
});
