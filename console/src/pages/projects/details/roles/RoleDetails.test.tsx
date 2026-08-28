import { TooltipProvider } from '@/components/ui/tooltip';
import type { HttpClient } from '@/services/core/httpClient';
import * as projectService from '@/services/projects';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import RoleDetails from './RoleDetails';

vi.mock('@/services/projects', () => ({
  fetchProjectById: vi.fn(),
  fetchRoleById: vi.fn(),
  fetchProjectResources: vi.fn(),
  updateProjectRole: vi.fn(),
  deleteProjectRole: vi.fn(),
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

describe('RoleDetails Page', () => {
  beforeEach(() => {
    vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082');
  });
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
  });

  it('renders role data and pre-selects permissions', async () => {
    const role = {
      id: 'role-1',
      name: 'Admin Role',
      code: 'admin_role',
      description: 'Administrator role',
      permission_ids: ['perm-1'],
    };

    const resources = [
      { id: 'res-1', name: 'accounts', permissions: [{ id: 'perm-1', name: 'create', description: 'Create' }] },
    ];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p1', name: 'Test Project' } as any);
    vi.mocked((projectService as any).fetchRoleById).mockResolvedValue(role as any);
    vi.mocked(projectService.fetchProjectResources).mockResolvedValue({ data: resources, pagination: { page_num: 1, page_size: 10, total_data: 1 } } as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/roles/role-1/details', state: { projectId: 'p1' } }]}>
        <Routes>
          <Route path="/projects/:project_id/roles/:role_id/details" element={<RoleDetails httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByDisplayValue('Admin Role')).toBeInTheDocument());
    expect(screen.getByDisplayValue('admin_role')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Administrator role')).toBeInTheDocument();

    // changing name updates name input
    const nameInput = screen.getByDisplayValue('Admin Role') as HTMLInputElement;
    fireEvent.change(nameInput, { target: { value: 'New Role Name' } });
    await waitFor(() => expect(nameInput.value).toBe('New Role Name'));

    // permission checkbox present and checked
    const checkbox = screen.getByLabelText('permission-perm-1') as HTMLInputElement;
    expect(checkbox).toBeChecked();
  });

  it('updates role when Save is clicked', async () => {
    const role = {
      id: 'role-2',
      name: 'User Role',
      code: 'user_role',
      description: 'User role',
      permission_ids: ['perm-2'],
    } as any;

    const resources = [
      { id: 'res-2', name: 'users', permissions: [{ id: 'perm-2', name: 'read', description: 'Read' }] },
    ];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p2', code: 'roled-console', name: 'Roled Console' } as any);
    vi.mocked((projectService as any).fetchRoleById).mockResolvedValue(role as any);
    vi.mocked(projectService.fetchProjectResources).mockResolvedValue({ data: resources, pagination: { page_num: 1, page_size: 10, total_data: 1 } } as any);

    const updateMock = vi.fn().mockResolvedValue({ ...role, name: 'User Role Updated' });
    vi.mocked(projectService.updateProjectRole).mockImplementation(updateMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p2/roles/role-2/details', state: { projectId: 'p2' } }]}>
        <Routes>
          <Route path="/projects/:project_id/roles/:role_id/details" element={<RoleDetails httpClient={httpClient} />} />
          <Route path="/projects/:project_id/details" element={<div>PROJECT DETAILS</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByDisplayValue('User Role')).toBeInTheDocument());

    fireEvent.click(screen.getByText('Save'));

    await waitFor(() => expect(updateMock).toHaveBeenCalled());
    await waitFor(() => expect(screen.getByText('PROJECT DETAILS')).toBeInTheDocument());
  });

  it('displays loading state while fetching role', async () => {
    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p1', name: 'Test Project' } as any);
    vi.mocked((projectService as any).fetchRoleById).mockImplementation(() => new Promise(() => { }));

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/roles/role-1/details' }]}>
        <Routes>
          <Route path="/projects/:project_id/roles/:role_id/details" element={<RoleDetails httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByTestId('role-loading')).toBeInTheDocument());
  });

  it('displays error state when role fetch fails', async () => {
    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p1', name: 'Test Project' } as any);
    vi.mocked((projectService as any).fetchRoleById).mockRejectedValue(new Error('Failed to fetch role'));

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/roles/role-1/details' }]}>
        <Routes>
          <Route path="/projects/:project_id/roles/:role_id/details" element={<RoleDetails httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByTestId('role-error')).toBeInTheDocument());
    expect(screen.getByText('Failed to load role')).toBeInTheDocument();
  });

  it('opens confirm dialog when Remove is clicked', async () => {
    const role = {
      id: 'role-3',
      name: 'Manager Role',
      code: 'manager_role',
      description: 'Manager role',
      permission_ids: [],
    } as any;

    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p3', name: 'Test Project' } as any);
    vi.mocked((projectService as any).fetchRoleById).mockResolvedValue(role as any);
    vi.mocked(projectService.fetchProjectResources).mockResolvedValue({ data: [], pagination: { page_num: 1, page_size: 10, total_data: 0 } } as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p3/roles/role-3/details', state: { projectId: 'p3' } }]}>
        <Routes>
          <Route path="/projects/:project_id/roles/:role_id/details" element={<RoleDetails httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByDisplayValue('Manager Role')).toBeInTheDocument());

    fireEvent.click(screen.getByText('Remove'));

    await waitFor(() => expect(screen.getByText('Remove Role')).toBeInTheDocument());
    expect(screen.getByText(/Are you sure you want to remove/)).toBeInTheDocument();
  });

  it('deletes role and navigates when confirmed', async () => {
    const role = {
      id: 'role-4',
      name: 'Editor Role',
      code: 'editor_role',
      description: 'Editor role',
      permission_ids: [],
    } as any;

    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p4', name: 'Test Project' } as any);
    vi.mocked((projectService as any).fetchRoleById).mockResolvedValue(role as any);
    vi.mocked(projectService.fetchProjectResources).mockResolvedValue({ data: [], pagination: { page_num: 1, page_size: 10, total_data: 0 } } as any);

    const deleteMock = vi.fn().mockResolvedValue(undefined);
    vi.mocked(projectService.deleteProjectRole).mockImplementation(deleteMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p4/roles/role-4/details', state: { projectId: 'p4' } }]}>
        <Routes>
          <Route path="/projects/:project_id/roles/:role_id/details" element={<RoleDetails httpClient={httpClient} />} />
          <Route path="/projects/:project_id/details" element={<div>PROJECT DETAILS PAGE</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByDisplayValue('Editor Role')).toBeInTheDocument());

    fireEvent.click(screen.getByText('Remove'));

    await waitFor(() => expect(screen.getByText('Remove Role')).toBeInTheDocument());

    // Find the confirm button in the dialog (it's a Button component with "Remove" text)
    const removeButtons = screen.getAllByRole('button', { name: /Remove/i });
    // The last Remove button should be in the dialog footer
    const confirmButton = removeButtons[removeButtons.length - 1];
    fireEvent.click(confirmButton);

    await waitFor(() => expect(deleteMock).toHaveBeenCalledWith(httpClient, 'http://localhost:8082', 'p4', 'role-4'));
  });

  it('includes updated permission_ids in update request', async () => {
    const role = {
      id: 'role-5',
      name: 'Test Role',
      code: 'test_role',
      description: 'Test description',
      permission_ids: ['perm-1'],
    } as any;

    const resources = [
      { id: 'res-1', name: 'accounts', permissions: [{ id: 'perm-1', name: 'create', description: 'Create' }, { id: 'perm-2', name: 'read', description: 'Read' }] },
    ];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p5', name: 'Test Project' } as any);
    vi.mocked((projectService as any).fetchRoleById).mockResolvedValue(role as any);
    vi.mocked(projectService.fetchProjectResources).mockResolvedValue({ data: resources, pagination: { page_num: 1, page_size: 10, total_data: 1 } } as any);

    const updateMock = vi.fn().mockResolvedValue({ ...role, permission_ids: ['perm-1', 'perm-2'] });
    vi.mocked(projectService.updateProjectRole).mockImplementation(updateMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p5/roles/role-5/details', state: { projectId: 'p5' } }]}>
        <Routes>
          <Route path="/projects/:project_id/roles/:role_id/details" element={<RoleDetails httpClient={httpClient} />} />
          <Route path="/projects/:project_id/details" element={<div>PROJECT DETAILS</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByDisplayValue('Test Role')).toBeInTheDocument());

    // Select additional permission
    const perm2Checkbox = screen.getByLabelText('permission-perm-2') as HTMLInputElement;
    fireEvent.click(perm2Checkbox);

    fireEvent.click(screen.getByText('Save'));

    await waitFor(() => {
      expect(updateMock).toHaveBeenCalledWith(
        httpClient,
        'http://localhost:8082',
        'p5',
        'role-5',
        expect.objectContaining({
          name: 'Test Role',
          code: 'test_role',
          description: 'Test description',
          permission_ids: expect.arrayContaining(['perm-1', 'perm-2']),
        })
      );
    });
  });
});
