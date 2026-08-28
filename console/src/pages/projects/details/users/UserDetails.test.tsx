import { TooltipProvider } from '@/components/ui/tooltip';
import type { HttpClient } from '@/services/core/httpClient';
import * as projectService from '@/services/projects';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import UserDetails from './UserDetails';

vi.mock('@/services/projects', () => ({
  fetchProjectById: vi.fn(),
  fetchUserById: vi.fn(),
  fetchProjectRoles: vi.fn(),
  updateProjectUser: vi.fn(),
  deleteProjectUser: vi.fn(),
  uploadUserAvatar: vi.fn(),
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

describe('UserDetails Page', () => {
  beforeEach(() => {
    vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082');
  });
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
  });

  it('renders user data and pre-selected role', async () => {
    const user = {
      id: 'user-1',
      display_name: 'John Doe',
      email: 'john@example.com',
      external_user_id: 'ext-123',
      is_active: true,
      role_id: 'role-1',
    };

    const roles = [
      { id: 'role-1', name: 'Admin' },
      { id: 'role-2', name: 'User' },
    ];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p1', name: 'Test Project' } as any);
    vi.mocked((projectService as any).fetchUserById).mockResolvedValue(user as any);
    vi.mocked(projectService.fetchProjectRoles).mockResolvedValue({ data: roles, pagination: { page_num: 1, page_size: 100, total_data: 2 } } as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/users/user-1/details', state: { projectId: 'p1' } }]}>
        <Routes>
          <Route path="/projects/:project_id/users/:user_id/details" element={<UserDetails httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByDisplayValue('John Doe')).toBeInTheDocument());
    expect(screen.getByDisplayValue('john@example.com')).toBeInTheDocument();
    expect(screen.getByDisplayValue('ext-123')).toBeInTheDocument();
  });

  it('updates user when Save is clicked', async () => {
    const user = {
      id: 'user-2',
      display_name: 'Jane Doe',
      email: 'jane@example.com',
      external_user_id: '',
      is_active: false,
      role_id: 'role-2',
    } as any;

    const roles = [{ id: 'role-2', name: 'User' }];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p2', name: 'Test Project' } as any);
    vi.mocked((projectService as any).fetchUserById).mockResolvedValue(user as any);
    vi.mocked(projectService.fetchProjectRoles).mockResolvedValue({ data: roles, pagination: { page_num: 1, page_size: 100, total_data: 1 } } as any);

    const updateMock = vi.fn().mockResolvedValue({ ...user, display_name: 'Jane Doe Updated' });
    vi.mocked(projectService.updateProjectUser).mockImplementation(updateMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p2/users/user-2/details', state: { projectId: 'p2' } }]}>
        <Routes>
          <Route path="/projects/:project_id/users/:user_id/details" element={<UserDetails httpClient={httpClient} />} />
          <Route path="/projects/:project_id/details" element={<div>PROJECT DETAILS</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByDisplayValue('Jane Doe')).toBeInTheDocument());

    fireEvent.click(screen.getByText('Save'));

    await waitFor(() => expect(updateMock).toHaveBeenCalled());
    await waitFor(() => expect(screen.getByText('PROJECT DETAILS')).toBeInTheDocument());
  });

  it('displays loading state while fetching user', async () => {
    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p1', name: 'Test Project' } as any);
    vi.mocked((projectService as any).fetchUserById).mockImplementation(() => new Promise(() => { }));
    vi.mocked(projectService.fetchProjectRoles).mockResolvedValue({ data: [], pagination: { page_num: 1, page_size: 100, total_data: 0 } } as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/users/user-1/details' }]}>
        <Routes>
          <Route path="/projects/:project_id/users/:user_id/details" element={<UserDetails httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByTestId('user-loading')).toBeInTheDocument());
  });

  it('displays error state when user fetch fails', async () => {
    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p1', name: 'Test Project' } as any);
    vi.mocked((projectService as any).fetchUserById).mockRejectedValue(new Error('Failed to fetch user'));

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/users/user-1/details' }]}>
        <Routes>
          <Route path="/projects/:project_id/users/:user_id/details" element={<UserDetails httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByTestId('user-error')).toBeInTheDocument());
    expect(screen.getByText('Failed to load user')).toBeInTheDocument();
  });

  it('opens confirm dialog when Remove is clicked', async () => {
    const user = {
      id: 'user-3',
      display_name: 'Bob Smith',
      email: 'bob@example.com',
      external_user_id: '',
      is_active: true,
      role_id: '',
    } as any;

    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p3', name: 'Test Project' } as any);
    vi.mocked((projectService as any).fetchUserById).mockResolvedValue(user as any);
    vi.mocked(projectService.fetchProjectRoles).mockResolvedValue({ data: [], pagination: { page_num: 1, page_size: 100, total_data: 0 } } as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p3/users/user-3/details', state: { projectId: 'p3' } }]}>
        <Routes>
          <Route path="/projects/:project_id/users/:user_id/details" element={<UserDetails httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByDisplayValue('Bob Smith')).toBeInTheDocument());

    fireEvent.click(screen.getByText('Remove'));

    await waitFor(() => expect(screen.getByText('Remove User')).toBeInTheDocument());
    expect(screen.getByText(/Are you sure you want to remove/)).toBeInTheDocument();
  });

  it('deletes user when confirmed', async () => {
    const user = {
      id: 'user-4',
      display_name: 'Alice Johnson',
      email: 'alice@example.com',
      external_user_id: '',
      is_active: true,
      role_id: '',
    } as any;

    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p4', name: 'Test Project' } as any);
    vi.mocked((projectService as any).fetchUserById).mockResolvedValue(user as any);
    vi.mocked(projectService.fetchProjectRoles).mockResolvedValue({ data: [], pagination: { page_num: 1, page_size: 100, total_data: 0 } } as any);

    const deleteMock = vi.fn().mockResolvedValue(undefined);
    vi.mocked(projectService.deleteProjectUser).mockImplementation(deleteMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p4/users/user-4/details', state: { projectId: 'p4' } }]}>
        <Routes>
          <Route path="/projects/:project_id/users/:user_id/details" element={<UserDetails httpClient={httpClient} />} />
          <Route path="/projects/:project_id/details" element={<div>PROJECT DETAILS PAGE</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByDisplayValue('Alice Johnson')).toBeInTheDocument());

    fireEvent.click(screen.getByText('Remove'));

    await waitFor(() => expect(screen.getByText('Remove User')).toBeInTheDocument());

    // Find the confirm button in the dialog
    const removeButtons = screen.getAllByRole('button', { name: /Remove/i });
    const confirmButton = removeButtons[removeButtons.length - 1];
    fireEvent.click(confirmButton);

    await waitFor(() => expect(deleteMock).toHaveBeenCalledWith(httpClient, 'http://localhost:8082', 'p4', 'user-4'));
  });

  it('toggles password visibility', async () => {
    const user = {
      id: 'user-5',
      display_name: 'Test User',
      email: 'test@example.com',
      external_user_id: '',
      is_active: true,
      role_id: '',
    } as any;

    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p5', name: 'Test Project' } as any);
    vi.mocked((projectService as any).fetchUserById).mockResolvedValue(user as any);
    vi.mocked(projectService.fetchProjectRoles).mockResolvedValue({ data: [], pagination: { page_num: 1, page_size: 100, total_data: 0 } } as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p5/users/user-5/details', state: { projectId: 'p5' } }]}>
        <Routes>
          <Route path="/projects/:project_id/users/:user_id/details" element={<UserDetails httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByDisplayValue('Test User')).toBeInTheDocument());

    const passwordInput = screen.getByPlaceholderText('Leave blank to keep current password') as HTMLInputElement;
    expect(passwordInput.type).toBe('password');

    // Click show password button
    const showButton = screen.getByLabelText('Show password');
    fireEvent.click(showButton);

    expect(passwordInput.type).toBe('text');

    // Click hide password button
    const hideButton = screen.getByLabelText('Hide password');
    fireEvent.click(hideButton);

    expect(passwordInput.type).toBe('password');
  });

  it('includes password in update request when provided', async () => {
    const user = {
      id: 'user-6',
      display_name: 'Test User',
      email: 'test@example.com',
      external_user_id: '',
      is_active: true,
      role_id: 'role-1',
    } as any;

    const roles = [{ id: 'role-1', name: 'Admin' }];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p6', name: 'Test Project' } as any);
    vi.mocked((projectService as any).fetchUserById).mockResolvedValue(user as any);
    vi.mocked(projectService.fetchProjectRoles).mockResolvedValue({ data: roles, pagination: { page_num: 1, page_size: 100, total_data: 1 } } as any);

    const updateMock = vi.fn().mockResolvedValue(user);
    vi.mocked(projectService.updateProjectUser).mockImplementation(updateMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p6/users/user-6/details', state: { projectId: 'p6' } }]}>
        <Routes>
          <Route path="/projects/:project_id/users/:user_id/details" element={<UserDetails httpClient={httpClient} />} />
          <Route path="/projects/:project_id/details" element={<div>PROJECT DETAILS</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByDisplayValue('Test User')).toBeInTheDocument());

    // Enter password
    const passwordInput = screen.getByPlaceholderText('Leave blank to keep current password');
    fireEvent.change(passwordInput, { target: { value: 'newpassword123' } });

    fireEvent.click(screen.getByText('Save'));

    await waitFor(() => {
      expect(updateMock).toHaveBeenCalledWith(
        httpClient,
        'http://localhost:8082',
        'p6',
        'user-6',
        expect.objectContaining({
          display_name: 'Test User',
          email: 'test@example.com',
          password: 'newpassword123',
          is_active: true,
          role_id: 'role-1',
        })
      );
    });
  });
});
