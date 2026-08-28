import { TooltipProvider } from '@/components/ui/tooltip';
import type { HttpClient } from '@/services/core/httpClient';
import * as projectService from '@/services/projects';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import NewClient from './NewClient';

vi.mock('@/services/projects', () => ({
  fetchProjectById: vi.fn(),
  fetchProjectPermissions: vi.fn(),
  fetchProjectResources: vi.fn(),
  createClient: vi.fn(),
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
    return (
      <QueryClientProvider client={queryClient}>
        <TooltipProvider>{children}</TooltipProvider>
      </QueryClientProvider>
    );
  };
}

describe('NewClient Page', () => {
  beforeEach(() => vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082'));
  afterEach(() => { vi.restoreAllMocks(); vi.unstubAllEnvs(); });

  it('fetches project and permissions and creates a client', async () => {
    const project = { id: 'p1', code: 'roled-console', name: 'Roled Console' } as any;
    const resources = [
      { id: 'r1', name: 'accounts', description: 'Accounts', permissions: [{ id: 'perm1', name: 'create', description: 'Create accounts' }] },
    ];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue(project);
    vi.mocked((projectService as any).fetchProjectResources).mockResolvedValue({ data: resources, pagination: { page_num: 1, page_size: 10, total_data: 1 } } as any);
    const createMock = vi.fn().mockResolvedValue({ id: 'c-new' });
    vi.mocked(projectService.createClient).mockImplementation(createMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/clients/new', state: { projectId: 'p1' } }]}>
        <Routes>
          <Route path="/projects/:project_id/clients/new" element={<NewClient httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Client')).toBeInTheDocument());

    // fill in name input on NewClient page
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'New Client' } });

    const permCheckbox = screen.getByLabelText('permission-perm1');
    fireEvent.click(permCheckbox);

    fireEvent.click(screen.getByText('Create'));

    await waitFor(() => expect(createMock).toHaveBeenCalled());
  });
});
