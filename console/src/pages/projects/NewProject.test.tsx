import { TooltipProvider } from '@/components/ui/tooltip';
import type { HttpClient } from '@/services/core/httpClient';
import * as projectService from '@/services/projects';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import NewProject from './NewProject';

vi.mock('@/services/projects', () => ({
  createProject: vi.fn(),
  uploadProjectLogo: vi.fn(),
}));

// Mock useToast to capture toast calls
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
    return (
      <QueryClientProvider client={queryClient}>
        <TooltipProvider>{children}</TooltipProvider>
      </QueryClientProvider>
    );
  };
}

describe('NewProject Page', () => {
  beforeEach(() => vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082'));
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
  });

  it('renders the new project form', async () => {
    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}>
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());
    expect(screen.getByText("Let's create a new project to integrate with Roled")).toBeInTheDocument();
    expect(screen.getByText('Project Information')).toBeInTheDocument();
    expect(screen.getByLabelText('Name')).toBeInTheDocument();
    expect(screen.getByLabelText('Description')).toBeInTheDocument();
    // Initial empty redirect URI row should be present
    expect(screen.getByPlaceholderText(/Redirect URI/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Login URL (optional)')).toBeInTheDocument();
  });

  it('creates a project with basic information', async () => {
    const newProject = { id: 'p-new', name: 'Test Project', code: 'test-project' } as any;

    const createMock = vi.fn().mockResolvedValue(newProject);
    vi.mocked(projectService.createProject).mockImplementation(createMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}>
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
          <Route path="/projects/:project_id/details" element={<div>PROJECT DETAILS</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    // Fill in name
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Test Project' } });
    fireEvent.change(screen.getByLabelText('Description'), { target: { value: 'A test project' } });

    // Fill in redirect URI (required)
    fireEvent.change(screen.getByPlaceholderText(/Redirect URI/i), { target: { value: 'http://localhost:4000/callback' } });

    fireEvent.click(screen.getByText('Create'));

    await waitFor(() => expect(createMock).toHaveBeenCalled());
    expect(createMock).toHaveBeenCalledWith(
      httpClient,
      'http://localhost:8082',
      expect.objectContaining({
        name: 'Test Project',
        description: 'A test project',
        redirect_uris: expect.arrayContaining([
          expect.objectContaining({
            redirect_uri: 'http://localhost:4000/callback',
          }),
        ]),
      })
    );
  });

  it('creates a project with multiple redirect URIs', async () => {
    const newProject = { id: 'p-new', name: 'Test Project', code: 'test-project' } as any;

    const createMock = vi.fn().mockResolvedValue(newProject);
    vi.mocked(projectService.createProject).mockImplementation(createMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}>
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
          <Route path="/projects/:project_id/details" element={<div>PROJECT DETAILS</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    // Fill in name
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Test Project' } });

    // Fill in the initial redirect URI
    fireEvent.change(screen.getByPlaceholderText(/Redirect URI/i), { target: { value: 'http://localhost:4000/callback' } });
    fireEvent.change(screen.getByPlaceholderText('Login URL (optional)'), { target: { value: 'http://localhost:4000/signin' } });

    // Add another redirect URI
    fireEvent.click(screen.getByText('Add Redirect URI'));

    const redirectInputs = screen.getAllByPlaceholderText(/Redirect URI/i);
    expect(redirectInputs).toHaveLength(2);
    fireEvent.change(redirectInputs[1], { target: { value: 'https://example.com/callback' } });

    const loginUrlInputs = screen.getAllByPlaceholderText('Login URL (optional)');
    fireEvent.change(loginUrlInputs[1], { target: { value: 'https://example.com/login' } });

    fireEvent.click(screen.getByText('Create'));

    await waitFor(() => expect(createMock).toHaveBeenCalled());
    expect(createMock).toHaveBeenCalledWith(
      httpClient,
      'http://localhost:8082',
      expect.objectContaining({
        name: 'Test Project',
        redirect_uris: expect.arrayContaining([
          expect.objectContaining({
            redirect_uri: 'http://localhost:4000/callback',
            login_url: 'http://localhost:4000/signin',
          }),
          expect.objectContaining({
            redirect_uri: 'https://example.com/callback',
            login_url: 'https://example.com/login',
          }),
        ]),
      })
    );
  });

  it('redirects to project details after successful creation', async () => {
    const newProject = { id: 'p-new', name: 'Test Project', code: 'test-project' } as any;

    vi.mocked(projectService.createProject).mockResolvedValue(newProject);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}>
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
          <Route path="/projects/:project_id/details" element={<div>PROJECT DETAILS PAGE</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Test Project' } });
    // Fill in required redirect URI
    fireEvent.change(screen.getByPlaceholderText(/Redirect URI/i), { target: { value: 'http://localhost:4000/callback' } });
    fireEvent.click(screen.getByText('Create'));

    await waitFor(() => expect(screen.getByText('PROJECT DETAILS PAGE')).toBeInTheDocument());
  });

  it('disables create button when required fields are empty', async () => {
    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}>
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    const createButton = screen.getByText('Create') as HTMLButtonElement;
    // Button should be disabled because both name and redirect URI are empty
    expect(createButton.disabled).toBe(true);
  });

it('enables create button when the name is provided even without redirect URIs', async () => {
    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}> 
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    const nameInput = screen.getByLabelText('Name') as HTMLInputElement;
    fireEvent.change(nameInput, { target: { value: 'Test Project' } });

    const createButton = screen.getByText('Create') as HTMLButtonElement;
    expect(createButton.disabled).toBe(false);
  });

  it('navigates back to projects when back button is clicked', async () => {
    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new', state: { from: '?page=1' } }]}>
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
          <Route path="/projects" element={<div>PROJECTS LIST PAGE</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    fireEvent.click(screen.getByText('Back to Projects'));

    await waitFor(() => expect(screen.getByText('PROJECTS LIST PAGE')).toBeInTheDocument());
  });

  it('shows loading state while creating project', async () => {
    const httpClient = createMockHttpClient();

    vi.mocked(projectService.createProject).mockImplementation(
      () => new Promise(() => { }) // Never resolves to keep pending state
    );

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}>
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Test Project' } });
    fireEvent.change(screen.getByPlaceholderText(/Redirect URI/i), { target: { value: 'http://localhost:4000/callback' } });
    fireEvent.click(screen.getByText('Create'));

    await waitFor(() => expect(screen.getByText('Creating...')).toBeInTheDocument());
  });

  it('removes redirect URI when delete button is clicked but keeps at least one row', async () => {
    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}>
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    // Add another redirect URI row (so we have 2)
    fireEvent.click(screen.getByText('Add Redirect URI'));

    const redirectInputs = screen.getAllByPlaceholderText(/Redirect URI/i);
    expect(redirectInputs).toHaveLength(2);

    // Fill in the second redirect URI
    fireEvent.change(redirectInputs[1], { target: { value: 'http://example.com/callback' } });

    // Find and click the delete button on the second row
    const deleteButtons = screen.getAllByLabelText(/^Delete redirect/);
    fireEvent.click(deleteButtons[1]);

    // Verify the second input is removed but we still have one
    await waitFor(() => {
      const remainingInputs = screen.getAllByPlaceholderText(/Redirect URI/i);
      expect(remainingInputs).toHaveLength(1);
    });
  });

it('keeps the delete button enabled when only one redirect URI row exists', async () => {
    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}> 
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    const deleteButton = screen.getByLabelText('Delete redirect 0');
    expect((deleteButton as HTMLButtonElement).disabled).toBe(false);
  });

  it('shows error toast when project creation fails', async () => {
    const httpClient = createMockHttpClient();

    vi.mocked(projectService.createProject).mockRejectedValue(new Error('Project name already exists'));

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}>
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Existing Project' } });
    fireEvent.change(screen.getByPlaceholderText(/Redirect URI/i), { target: { value: 'http://localhost:4000/callback' } });
    fireEvent.click(screen.getByText('Create'));

    await waitFor(() => {
      expect(toastMock).toHaveBeenCalled();
    });

    const callArg = toastMock.mock.calls[toastMock.mock.calls.length - 1][0];
    expect(callArg.title).toBe('Create project failed');
    expect(callArg.description).toContain('Project name already exists');
    expect(callArg.variant).toBe('destructive');
  });

  it('enforces max length on name input', async () => {
    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}>
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    const nameInput = screen.getByLabelText('Name') as HTMLInputElement;

    // Verify maxLength attribute is set
    expect(nameInput).toHaveAttribute('maxLength', '50');

    // Type a name that is within the limit
    fireEvent.change(nameInput, { target: { value: 'A'.repeat(50) } });
    expect(nameInput.value).toHaveLength(50);

    // Verify input is truncated at maxLength
    fireEvent.change(nameInput, { target: { value: 'A'.repeat(100) } });
    expect(nameInput.value).toHaveLength(50);
  });

  it('enforces max length on description input', async () => {
    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}>
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    const descInput = screen.getByLabelText('Description') as HTMLTextAreaElement;

    // Verify maxLength attribute is set to 400
    expect(descInput).toHaveAttribute('maxLength', '400');

    // Type a description that is within the limit
    fireEvent.change(descInput, { target: { value: 'A'.repeat(400) } });
    expect(descInput.value).toHaveLength(400);

    // Verify input is truncated at maxLength
    fireEvent.change(descInput, { target: { value: 'A'.repeat(500) } });
    expect(descInput.value).toHaveLength(400);
  });

  it('shows character count for name and description', async () => {
    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}>
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    // Check name maxLength attribute
    expect(screen.getByLabelText('Name')).toHaveAttribute('maxLength', '50');

    // Check description maxLength attribute
    expect(screen.getByLabelText('Description')).toHaveAttribute('maxLength', '400');
  });

it('allows creating a project without redirect URIs', async () => {
    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}> 
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Test Project' } });

    const createButton = screen.getByText('Create') as HTMLButtonElement;
    expect(createButton.disabled).toBe(false);
  });

  it('validates redirect URI format', async () => {
    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}>
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    // Fill in name
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Test Project' } });

    // Enter invalid URI format
    const redirectInput = screen.getByPlaceholderText(/Redirect URI/i);
    fireEvent.change(redirectInput, { target: { value: 'invalid-uri' } });
    fireEvent.blur(redirectInput);

    // Button should be disabled because URI format is invalid
    const createButton = screen.getByText('Create') as HTMLButtonElement;
    expect(createButton.disabled).toBe(true);
  });

  it('accepts valid absolute URI format for redirect URI', async () => {
    const newProject = { id: 'p-new', name: 'Test Project', code: 'test-project' } as any;
    const createMock = vi.fn().mockResolvedValue(newProject);
    vi.mocked(projectService.createProject).mockImplementation(createMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}>
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
          <Route path="/projects/:project_id/details" element={<div>PROJECT DETAILS</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Test Project' } });
    fireEvent.change(screen.getByPlaceholderText(/Redirect URI/i), { target: { value: 'https://example.com/callback' } });

    fireEvent.click(screen.getByText('Create'));

    await waitFor(() => expect(createMock).toHaveBeenCalled());
    expect(createMock).toHaveBeenCalledWith(
      httpClient,
      'http://localhost:8082',
      expect.objectContaining({
        name: 'Test Project',
        redirect_uris: expect.arrayContaining([
          expect.objectContaining({
            redirect_uri: 'https://example.com/callback',
          }),
        ]),
      })
    );
  });

  it('accepts relative path format for redirect URI', async () => {
    const newProject = { id: 'p-new', name: 'Test Project', code: 'test-project' } as any;
    const createMock = vi.fn().mockResolvedValue(newProject);
    vi.mocked(projectService.createProject).mockImplementation(createMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}>
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
          <Route path="/projects/:project_id/details" element={<div>PROJECT DETAILS</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Test Project' } });
    fireEvent.change(screen.getByPlaceholderText(/Redirect URI/i), { target: { value: '/callback' } });

    fireEvent.click(screen.getByText('Create'));

    await waitFor(() => expect(createMock).toHaveBeenCalled());
    expect(createMock).toHaveBeenCalledWith(
      httpClient,
      'http://localhost:8082',
      expect.objectContaining({
        name: 'Test Project',
        redirect_uris: expect.arrayContaining([
          expect.objectContaining({
            redirect_uri: '/callback',
          }),
        ]),
      })
    );
  });

  it('validates login URL format when provided', async () => {
    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}>
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    // Fill in name and valid redirect URI
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Test Project' } });
    fireEvent.change(screen.getByPlaceholderText(/Redirect URI/i), { target: { value: 'https://example.com/callback' } });

    // Enter invalid login URL (missing protocol)
    const loginUrlInput = screen.getByPlaceholderText('Login URL (optional)');
    fireEvent.change(loginUrlInput, { target: { value: 'example.com/login' } });
    fireEvent.blur(loginUrlInput);

    // Button should be disabled because login URL format is invalid
    const createButton = screen.getByText('Create') as HTMLButtonElement;
    expect(createButton.disabled).toBe(true);
  });

  it('accepts valid login URL format', async () => {
    const newProject = { id: 'p-new', name: 'Test Project', code: 'test-project' } as any;
    const createMock = vi.fn().mockResolvedValue(newProject);
    vi.mocked(projectService.createProject).mockImplementation(createMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/new' }]}>
        <Routes>
          <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
          <Route path="/projects/:project_id/details" element={<div>PROJECT DETAILS</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByText('Add New Project')).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Test Project' } });
    fireEvent.change(screen.getByPlaceholderText(/Redirect URI/i), { target: { value: 'https://example.com/callback' } });
    fireEvent.change(screen.getByPlaceholderText('Login URL (optional)'), { target: { value: 'https://example.com/login' } });

    fireEvent.click(screen.getByText('Create'));

    await waitFor(() => expect(createMock).toHaveBeenCalled());
    expect(createMock).toHaveBeenCalledWith(
      httpClient,
      'http://localhost:8082',
      expect.objectContaining({
        name: 'Test Project',
        redirect_uris: expect.arrayContaining([
          expect.objectContaining({
            redirect_uri: 'https://example.com/callback',
            login_url: 'https://example.com/login',
          }),
        ]),
      })
    );
  });
});
