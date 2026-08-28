import { TooltipProvider } from '@/components/ui/tooltip';
import type { HttpClient } from '@/services/core/httpClient';
import * as projectService from '@/services/projects';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import NewRole from './NewRole';

vi.mock('@/services/projects', () => ({
  fetchProjectById: vi.fn(),
  fetchProjectPermissions: vi.fn(),
  fetchProjectResources: vi.fn(),
  createProjectRole: vi.fn(),
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

describe('NewRole Page', () => {
  beforeEach(() => vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082'));
  afterEach(() => { vi.restoreAllMocks(); vi.unstubAllEnvs(); });

  it('fetches project and permissions and creates a role', async () => {
    const project = { id: 'p1', code: 'roled-console', name: 'Roled Console' } as any;
    const resources = [
      { id: 'r1', name: 'accounts', description: 'Accounts', permissions: [{ id: 'perm1', name: 'create', description: 'Create accounts' }] },
    ];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue(project);
    vi.mocked((projectService as any).fetchProjectResources).mockResolvedValue({ data: resources, pagination: { page_num: 1, page_size: 10, total_data: 1 } } as any);
    const createMock = vi.fn().mockResolvedValue({ id: 'r-new' });
    vi.mocked(projectService.createProjectRole).mockImplementation(createMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/roles/new', state: { projectId: 'p1' } }]}>
        <Routes>
          <Route path="/projects/:project_id/roles/new" element={<NewRole httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Role')).toBeInTheDocument());

    // fill in name input on NewRole page
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'New Role' } });

    const permCheckbox = screen.getByLabelText('permission-perm1');
    fireEvent.click(permCheckbox);

    fireEvent.click(screen.getByText('Create'));

    await waitFor(() => expect(createMock).toHaveBeenCalled());
  });

  it('displays loading state while fetching project', async () => {
    vi.mocked(projectService.fetchProjectById).mockImplementation(() => new Promise(() => { }));

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/roles/new' }]}>
        <Routes>
          <Route path="/projects/:project_id/roles/new" element={<NewRole httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByTestId('newrole-loading')).toBeInTheDocument());
  });

  it('displays error state when project fetch fails', async () => {
    vi.mocked(projectService.fetchProjectById).mockRejectedValue(new Error('Failed to fetch project'));

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/roles/new' }]}>
        <Routes>
          <Route path="/projects/:project_id/roles/new" element={<NewRole httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByTestId('newrole-error')).toBeInTheDocument());
    expect(screen.getByText('Failed to load project')).toBeInTheDocument();
  });

  it('auto-generates role code from name', async () => {
    const project = { id: 'p1', code: 'roled-console', name: 'Roled Console' } as any;
    const resources = [] as any[];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue(project);
    vi.mocked((projectService as any).fetchProjectResources).mockResolvedValue({ data: resources, pagination: { page_num: 1, page_size: 10, total_data: 0 } } as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/roles/new' }]}>
        <Routes>
          <Route path="/projects/:project_id/roles/new" element={<NewRole httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Role')).toBeInTheDocument());

    const nameInput = screen.getByLabelText('Name') as HTMLInputElement;
    fireEvent.change(nameInput, { target: { value: 'Super Admin' } });

    const codeInput = screen.getByLabelText('Code') as HTMLInputElement;
    expect(codeInput.value).toContain('super');
  });

  it('includes permission_ids in the create request', async () => {
    const project = { id: 'p1', code: 'roled-console', name: 'Roled Console' } as any;
    const resources = [
      { id: 'r1', name: 'accounts', description: 'Accounts', permissions: [{ id: 'perm1', name: 'create', description: 'Create accounts' }, { id: 'perm2', name: 'read', description: 'Read accounts' }] },
    ];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue(project);
    vi.mocked((projectService as any).fetchProjectResources).mockResolvedValue({ data: resources, pagination: { page_num: 1, page_size: 10, total_data: 1 } } as any);
    const createMock = vi.fn().mockResolvedValue({ id: 'r-new' });
    vi.mocked(projectService.createProjectRole).mockImplementation(createMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/roles/new' }]}>
        <Routes>
          <Route path="/projects/:project_id/roles/new" element={<NewRole httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Role')).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Admin' } });
    fireEvent.change(screen.getByLabelText('Description'), { target: { value: 'Admin role' } });

    // Select two permissions
    const perm1Checkbox = screen.getByLabelText('permission-perm1');
    const perm2Checkbox = screen.getByLabelText('permission-perm2');
    fireEvent.click(perm1Checkbox);
    fireEvent.click(perm2Checkbox);

    fireEvent.click(screen.getByText('Create'));

    await waitFor(() => {
      expect(createMock).toHaveBeenCalledWith(
        httpClient,
        'http://localhost:8082',
        'p1',
        expect.objectContaining({
          name: 'Admin',
          description: 'Admin role',
          permission_ids: expect.arrayContaining(['perm1', 'perm2']),
        })
      );
    });
  });

  it('disables create button when name is empty', async () => {
    const project = { id: 'p1', code: 'roled-console', name: 'Roled Console' } as any;
    const resources = [] as any[];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue(project);
    vi.mocked((projectService as any).fetchProjectResources).mockResolvedValue({ data: resources, pagination: { page_num: 1, page_size: 10, total_data: 0 } } as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/roles/new' }]}>
        <Routes>
          <Route path="/projects/:project_id/roles/new" element={<NewRole httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Role')).toBeInTheDocument());

    const createButton = screen.getByText('Create') as HTMLButtonElement;
    expect(createButton.disabled).toBe(true);
  });

  it('enables create button when name is filled', async () => {
    const project = { id: 'p1', code: 'roled-console', name: 'Roled Console' } as any;
    const resources = [] as any[];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue(project);
    vi.mocked((projectService as any).fetchProjectResources).mockResolvedValue({ data: resources, pagination: { page_num: 1, page_size: 10, total_data: 0 } } as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/roles/new' }]}>
        <Routes>
          <Route path="/projects/:project_id/roles/new" element={<NewRole httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Role')).toBeInTheDocument());

    const nameInput = screen.getByLabelText('Name') as HTMLInputElement;
    fireEvent.change(nameInput, { target: { value: 'New Role' } });

    const createButton = screen.getByText('Create') as HTMLButtonElement;
    expect(createButton.disabled).toBe(false);
  });
});
