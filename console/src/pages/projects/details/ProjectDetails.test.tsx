import { TooltipProvider } from '@/components/ui/tooltip';
import type { HttpClient } from '@/services/core/httpClient';
import * as projectService from '@/services/projects';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import NewClient from './clients/NewClient';
import ProjectDetails from './ProjectDetails';

vi.mock('@/services/projects', () => ({
  fetchProjectById: vi.fn(),
  fetchProjectClients: vi.fn(),
  fetchProjectResources: vi.fn(),
  fetchProjectPermissions: vi.fn(),
  createClient: vi.fn(),
  fetchProjects: vi.fn(),
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

describe('ProjectDetails Page', () => {
  beforeEach(() => {
    vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082');
  });
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
  });

  it('loads project and clients and renders them', async () => {
    const project = {
      id: 'p1',
      created_at: '2026-02-25T12:05:01.9478Z',
      updated_at: '2026-02-26T08:30:00.0000Z',
      code: 'roled-console',
      name: 'Roled Console',
      description: 'Project description',
      logo_url: 'https://example.com/logo.png',
      redirect_uris: [{ redirect_uri: 'https://example.com/callback', login_url: '' }],
      is_active: true,
    };

    const clients = [
      { id: 'c1', created_at: '2026-02-26T01:00:00Z', updated_at: '2026-02-26T01:00:00Z', code: 'cli1', name: 'Client 1', project_id: 'p1', is_default: true, is_active: true },
    ];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue(project as any);
    vi.mocked(projectService.fetchProjectClients).mockResolvedValue({ data: clients, pagination: { page_num: 1, page_size: 10, total_data: 1 } });

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/details' }]}>
        <Routes>
          <Route path="/projects/:project_id/details" element={<ProjectDetails httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Roled Console')).toBeInTheDocument());
    expect(screen.getByText('Project Information')).toBeInTheDocument();
    expect(screen.getByText('Clients')).toBeInTheDocument();
    // client row
    expect(screen.getByText('Client 1')).toBeInTheDocument();
  });

  it('adds a client via the NewClient page', async () => {
    const project = {
      id: 'p1',
      created_at: '2026-02-25T12:05:01.9478Z',
      updated_at: '2026-02-26T08:30:00.0000Z',
      code: 'roled-console',
      name: 'Roled Console',
      description: 'Project description',
      logo_url: 'https://example.com/logo.png',
      redirect_uris: [{ redirect_uri: 'https://example.com/callback', login_url: '' }],
      is_active: true,
    };

    const resources = [
      { id: 'r1', name: 'accounts', description: 'Accounts', permissions: [{ id: 'p1', name: 'create', description: 'Create accounts', is_default: true }] },
    ];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue(project as any);
    vi.mocked((projectService as any).fetchProjectResources).mockResolvedValue({ data: resources, pagination: { page_num: 1, page_size: 10, total_data: 1 } });
    const createMock = vi.fn().mockResolvedValue({ id: 'c-new' });
    vi.mocked((projectService as any).createClient).mockImplementation(createMock);

    const httpClient = createMockHttpClient();

    // Render NewClient route directly (simulate navigation) with projectId in state
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

    // select the permission checkbox
    const permCheckbox = screen.getByLabelText('permission-p1');
    fireEvent.click(permCheckbox);

    // submit
    fireEvent.click(screen.getByText('Create'));

    await waitFor(() => expect(createMock).toHaveBeenCalled());
  });
});
