import { TooltipProvider } from '@/components/ui/tooltip';
import { formatDate } from '@/lib/date';
import type { HttpClient } from '@/services/core/httpClient';
import * as projectService from '@/services/projects';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import Projects from './Projects';

vi.mock('@/services/projects', () => ({
  fetchProjects: vi.fn(),
  deleteProject: vi.fn(),
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

describe('Projects Page', () => {
  beforeEach(() => {
    vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082');
  });
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
  });

  it('shows loading and then renders projects with summary', async () => {
    const projects = [
      {
        id: 'p1',
        created_at: '2026-02-25T12:05:01.9478Z',
        updated_at: '2026-02-26T08:30:00.0000Z',
        code: 'roled-console',
        name: 'Roled Console',
        description: 'A very long description that should be truncated in the UI for display purposes',
        logo_url: 'https://example.com/logo.png',
        is_active: true,
      },
    ];
    vi.mocked(projectService.fetchProjects).mockResolvedValue({ data: projects, pagination: { page_num: 1, page_size: 10, total_data: 1 } });

    const httpClient = createMockHttpClient();
    render(<Projects httpClient={httpClient} />, { wrapper: (props) => (<MemoryRouter>{(createWrapper() as any)(props)}</MemoryRouter>) });

    // Wait until project name appears
    await waitFor(() => expect(screen.getByText('Roled Console')).toBeInTheDocument());

    // Summary text
    expect(screen.getByText(/Showing/)).toBeInTheDocument();

    // Logo rendered
    const img = screen.getByAltText('Roled Console') as HTMLImageElement;
    expect(img.src).toContain('logo.png');

    // Created formatted (localized)
    const expected = formatDate(projects[0].created_at as any);
    expect(screen.getByText(expected)).toBeInTheDocument();
  });

  it('calls fetchProjects with expected params including fixed page_size and null is_active for All', async () => {
    vi.mocked(projectService.fetchProjects).mockResolvedValue({ data: [], pagination: { page_num: 1, page_size: 15, total_data: 0 } });
    const httpClient = createMockHttpClient();
    render(<Projects httpClient={httpClient} />, { wrapper: (props) => (<MemoryRouter>{(createWrapper() as any)(props)}</MemoryRouter>) });

    await waitFor(() => expect(projectService.fetchProjects).toHaveBeenCalled());
    const callArgs = vi.mocked(projectService.fetchProjects).mock.calls[0];
    // args: httpClient, baseUrl, params
    expect(callArgs[0]).toBe(httpClient);
    expect(callArgs[1]).toBe('http://localhost:8082');
    const params = callArgs[2];
    expect(params.page_size).toBe(15);
    expect(params.is_active).toBeNull();
    expect(params.page_num).toBe(1);
  });

  it('deletes a project after user confirms name', async () => {
    const projects = [
      {
        id: 'p1',
        created_at: '2026-02-25T12:05:01.9478Z',
        updated_at: '2026-02-26T08:30:00.0000Z',
        code: 'roled-console',
        name: 'Roled Console',
        description: 'A very long description that should be truncated in the UI for display purposes',
        logo_url: 'https://example.com/logo.png',
        is_active: true,
      },
    ];
    // first call returns the project, subsequent call (after delete) returns empty
    vi.mocked(projectService.fetchProjects).mockResolvedValueOnce({ data: projects, pagination: { page_num: 1, page_size: 15, total_data: 1 } });
    vi.mocked(projectService.fetchProjects).mockResolvedValueOnce({ data: [], pagination: { page_num: 1, page_size: 15, total_data: 0 } });
    vi.mocked(projectService.deleteProject).mockResolvedValue(undefined);

    const httpClient = createMockHttpClient();
    const { container } = render(<Projects httpClient={httpClient} />, { wrapper: (props) => (<MemoryRouter>{(createWrapper() as any)(props)}</MemoryRouter>) });

    await waitFor(() => expect(screen.getByText('Roled Console')).toBeInTheDocument());

    // open actions menu and click Remove
    const actionsBtn = screen.getByLabelText('Actions for Roled Console');
    await userEvent.click(actionsBtn);
    const removeItem = await screen.findByText((content, node) => node?.textContent?.trim() === 'Remove');
    await userEvent.click(removeItem);

    // dialog appears - input and confirm
    const input = await screen.findByLabelText('confirm-project-name');
    // type name in different case to ensure case-insensitive match
    fireEvent.change(input, { target: { value: 'roled console' } });

    const confirmBtn = screen.getByRole('button', { name: 'Remove' });
    fireEvent.click(confirmBtn);

    await waitFor(() => expect(projectService.deleteProject).toHaveBeenCalled());
    // project should be removed from DOM
    await waitFor(() => expect(screen.queryByText('Roled Console')).not.toBeInTheDocument());
  });

  it('navigates to new project page when Add Project card is clicked', async () => {
    vi.mocked(projectService.fetchProjects).mockResolvedValue({ data: [], pagination: { page_num: 1, page_size: 10, total_data: 0 } });
    const httpClient = createMockHttpClient();
    render(
      <MemoryRouter initialEntries={['/projects']}>
        <Routes>
          <Route path="/projects" element={<Projects httpClient={httpClient} />} />
          <Route path="/projects/new" element={<div>NEW PROJECT PAGE</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() }
    );

    // Wait until Add Project card renders
    const addCard = await screen.findByLabelText('Add project card');
    await userEvent.click(addCard);

    // Verify navigation
    await waitFor(() => expect(screen.getByText('NEW PROJECT PAGE')).toBeInTheDocument());
  });
});

