import { TooltipProvider } from '@/components/ui/tooltip';
import type { HttpClient } from '@/services/core/httpClient';
import * as projectService from '@/services/projects';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import NewUser from './NewUser';

vi.mock('@/services/projects', () => ({
  fetchProjectById: vi.fn(),
  fetchProjectRoles: vi.fn(),
  createProjectUser: vi.fn(),
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

describe('NewUser Page', () => {
  beforeEach(() => vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082'));
  afterEach(() => { vi.restoreAllMocks(); vi.unstubAllEnvs(); });

  it('fetches project and roles and creates a user', async () => {
    const project = { id: 'p1', code: 'test-project', name: 'Test Project' } as any;
    const roles = [
      { id: 'role-1', name: 'Admin' },
      { id: 'role-2', name: 'User' },
    ];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue(project);
    vi.mocked(projectService.fetchProjectRoles).mockResolvedValue({ data: roles, pagination: { page_num: 1, page_size: 100, total_data: 2 } } as any);
    const createMock = vi.fn().mockResolvedValue({ id: 'user-new' });
    vi.mocked(projectService.createProjectUser).mockImplementation(createMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/users/new', state: { projectId: 'p1' } }]}>
        <Routes>
          <Route path="/projects/:project_id/users/new" element={<NewUser httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New User')).toBeInTheDocument());

    // fill in name input
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'John Doe' } });
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'john@example.com' } });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'password123' } });

    fireEvent.click(screen.getByText('Create'));

    await waitFor(() => expect(createMock).toHaveBeenCalled());
  });

  it('displays loading state while fetching project', async () => {
    vi.mocked(projectService.fetchProjectById).mockImplementation(() => new Promise(() => { }));

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/users/new' }]}>
        <Routes>
          <Route path="/projects/:project_id/users/new" element={<NewUser httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByTestId('newuser-loading')).toBeInTheDocument());
  });

  it('displays error state when project fetch fails', async () => {
    vi.mocked(projectService.fetchProjectById).mockRejectedValue(new Error('Failed to fetch project'));

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/users/new' }]}>
        <Routes>
          <Route path="/projects/:project_id/users/new" element={<NewUser httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByTestId('newuser-error')).toBeInTheDocument());
    expect(screen.getByText('Failed to load project')).toBeInTheDocument();
  });

  it('includes all fields in the create request', async () => {
    const project = { id: 'p1', code: 'test-project', name: 'Test Project' } as any;
    const roles = [
      { id: 'role-1', name: 'Admin' },
    ];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue(project);
    vi.mocked(projectService.fetchProjectRoles).mockResolvedValue({ data: roles, pagination: { page_num: 1, page_size: 100, total_data: 1 } } as any);
    const createMock = vi.fn().mockResolvedValue({ id: 'user-new' });
    vi.mocked(projectService.createProjectUser).mockImplementation(createMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/users/new' }]}>
        <Routes>
          <Route path="/projects/:project_id/users/new" element={<NewUser httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New User')).toBeInTheDocument());

    // fill in all fields
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Jane Doe' } });
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'jane@example.com' } });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'securepassword' } });
    fireEvent.change(screen.getByLabelText('External User ID'), { target: { value: 'ext-123' } });

    fireEvent.click(screen.getByText('Create'));

    await waitFor(() => {
      expect(createMock).toHaveBeenCalledWith(
        httpClient,
        'http://localhost:8082',
        'p1',
        expect.objectContaining({
          display_name: 'Jane Doe',
          email: 'jane@example.com',
          password: 'securepassword',
          external_user_id: 'ext-123',
        })
      );
    });
  });

  it('disables create button when name is empty', async () => {
    const project = { id: 'p1', code: 'test-project', name: 'Test Project' } as any;
    const roles = [] as any[];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue(project);
    vi.mocked(projectService.fetchProjectRoles).mockResolvedValue({ data: roles, pagination: { page_num: 1, page_size: 100, total_data: 0 } } as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/users/new' }]}>
        <Routes>
          <Route path="/projects/:project_id/users/new" element={<NewUser httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New User')).toBeInTheDocument());

    const createButton = screen.getByText('Create') as HTMLButtonElement;
    expect(createButton.disabled).toBe(true);
  });

  it('enables create button when all required fields are filled', async () => {
    const project = { id: 'p1', code: 'test-project', name: 'Test Project' } as any;
    const roles = [] as any[];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue(project);
    vi.mocked(projectService.fetchProjectRoles).mockResolvedValue({ data: roles, pagination: { page_num: 1, page_size: 100, total_data: 0 } } as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/users/new' }]}>
        <Routes>
          <Route path="/projects/:project_id/users/new" element={<NewUser httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New User')).toBeInTheDocument());

    const nameInput = screen.getByLabelText('Name') as HTMLInputElement;
    fireEvent.change(nameInput, { target: { value: 'John Doe' } });

    // Button should still be disabled - need email + password or external_user_id
    const createButton = screen.getByText('Create') as HTMLButtonElement;
    expect(createButton.disabled).toBe(true);

    // Fill in email and password
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'john@example.com' } });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'password123' } });

    expect(createButton.disabled).toBe(false);
  });

  it('redirects to project details after successful creation', async () => {
    const project = { id: 'p1', code: 'test-project', name: 'Test Project' } as any;
    const roles = [] as any[];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue(project);
    vi.mocked(projectService.fetchProjectRoles).mockResolvedValue({ data: roles, pagination: { page_num: 1, page_size: 100, total_data: 0 } } as any);
    vi.mocked(projectService.createProjectUser).mockResolvedValue({ id: 'user-new' });

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/users/new' }]}>
        <Routes>
          <Route path="/projects/:project_id/users/new" element={<NewUser httpClient={httpClient} />} />
          <Route path="/projects/:project_id/details" element={<div>PROJECT DETAILS</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New User')).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'John Doe' } });
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'john@example.com' } });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'password123' } });
    fireEvent.click(screen.getByText('Create'));

    await waitFor(() => expect(screen.getByText('PROJECT DETAILS')).toBeInTheDocument());
  });

  it('displays show/hide password toggle', async () => {
    const project = { id: 'p1', code: 'test-project', name: 'Test Project' } as any;
    const roles = [] as any[];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue(project);
    vi.mocked(projectService.fetchProjectRoles).mockResolvedValue({ data: roles, pagination: { page_num: 1, page_size: 100, total_data: 0 } } as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/users/new' }]}>
        <Routes>
          <Route path="/projects/:project_id/users/new" element={<NewUser httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New User')).toBeInTheDocument());

    const passwordInput = screen.getByLabelText('Password') as HTMLInputElement;
    fireEvent.change(passwordInput, { target: { value: 'secretpassword' } });

    // Password should be hidden by default
    expect(passwordInput.type).toBe('password');

    // Click show password button
    const showButton = screen.getByLabelText('Show password');
    fireEvent.click(showButton);

    // Password should now be visible
    expect(passwordInput.type).toBe('text');

    // Click hide password button
    const hideButton = screen.getByLabelText('Hide password');
    fireEvent.click(hideButton);

    // Password should be hidden again
    expect(passwordInput.type).toBe('password');
  });
});
