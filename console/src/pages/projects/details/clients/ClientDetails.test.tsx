import { TooltipProvider } from '@/components/ui/tooltip';
import type { HttpClient } from '@/services/core/httpClient';
import * as projectService from '@/services/projects';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import ClientDetails from './ClientDetails';

vi.mock('@/services/projects', () => ({
  fetchProjectById: vi.fn(),
  fetchClientById: vi.fn(),
  fetchProjectResources: vi.fn(),
  updateClient: vi.fn(),
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

describe('ClientDetails Page', () => {
  beforeEach(() => {
    vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082');
  });
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
  });

  it('renders client data and pre-selects permissions', async () => {
    const client = {
      id: 'cli-1',
      name: 'Client One',
      code: 'client-one',
      secret: 's3cr3t',
      permission_ids: ['perm-1'],
      permissions: [{ id: 'perm-1', name: 'create', description: 'Create', is_default: false }],
    };

    const resources = [
      { id: 'res-1', name: 'accounts', is_default: false, permissions: [{ id: 'perm-1', name: 'create', description: 'Create', is_default: false }] },
    ];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p1' } as any);
    vi.mocked((projectService as any).fetchClientById).mockResolvedValue(client as any);
    vi.mocked(projectService.fetchProjectResources).mockResolvedValue({ data: resources, pagination: { page_num: 1, page_size: 10, total_data: 1 } } as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/clients/cli-1/details', state: { projectId: 'p1' } }]}>
        <Routes>
          <Route path="/projects/:project_id/clients/:client_id/details" element={<ClientDetails httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByDisplayValue('Client One')).toBeInTheDocument());
    expect(screen.getByDisplayValue('cli-1')).toBeInTheDocument();
    // changing name updates name input
    const nameInput = screen.getByDisplayValue('Client One') as HTMLInputElement;
    fireEvent.change(nameInput, { target: { value: 'New Client' } });
    await waitFor(() => expect(nameInput.value).toBe('New Client'));
    // permission checkbox present and checked
    const checkbox = screen.getByLabelText('permission-perm-1') as HTMLInputElement;
    expect(checkbox).toBeChecked();
  });

  it('updates client when Save is clicked', async () => {
    const client = {
      id: 'cli-2',
      name: 'Client Two',
      code: 'client-two',
      secret: 's2',
      permission_ids: ['perm-2'],
      is_active: true,
      permissions: [{ id: 'perm-2', name: 'read', description: 'Read', is_default: false }],
    } as any;

    const resources = [
      { id: 'res-2', name: 'users', is_default: false, permissions: [{ id: 'perm-2', name: 'read', description: 'Read', is_default: false }] },
    ];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p2', code: 'roled-console', name: 'Roled Console' } as any);
    vi.mocked((projectService as any).fetchClientById).mockResolvedValue(client as any);
    vi.mocked(projectService.fetchProjectResources).mockResolvedValue({ data: resources, pagination: { page_num: 1, page_size: 10, total_data: 1 } } as any);

    const updateMock = vi.fn().mockResolvedValue({ ...client, name: 'Client Two Updated' });
    vi.mocked(projectService.updateClient).mockImplementation(updateMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p2/clients/cli-2/details', state: { projectId: 'p2' } }]}>
        <Routes>
          <Route path="/projects/:project_id/clients/:client_id/details" element={<ClientDetails httpClient={httpClient} />} />
          <Route path="/projects/:project_id/details" element={<div>PROJECT DETAILS</div>} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByDisplayValue('Client Two')).toBeInTheDocument());

    fireEvent.click(screen.getByText('Save'));

    await waitFor(() => expect(updateMock).toHaveBeenCalled());
    await waitFor(() => expect(screen.getByText('PROJECT DETAILS')).toBeInTheDocument());
  });

  it('allows selecting custom permissions for default client while default permissions are disabled and checked', async () => {
    const client = {
      id: 'cli-default',
      name: 'Default Client',
      code: 'default-client',
      secret: 's3cr3t',
      permission_ids: ['perm-default-1', 'perm-custom-1'],
      is_default: true,
      permissions: [
        { id: 'perm-default-1', name: 'read', description: 'Read Default', is_default: true },
        { id: 'perm-custom-1', name: 'custom', description: 'Custom Permission', is_default: false },
      ],
    };

    const resources = [
      { 
        id: 'res-1', 
        name: 'accounts', 
        is_default: true,
        permissions: [
          { id: 'perm-default-1', name: 'read', description: 'Read Default', is_default: true },
        ] 
      },
      { 
        id: 'res-2', 
        name: 'users', 
        is_default: false,
        permissions: [
          { id: 'perm-custom-1', name: 'custom', description: 'Custom Permission', is_default: false },
        ] 
      },
    ];

    vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p1' } as any);
    vi.mocked((projectService as any).fetchClientById).mockResolvedValue(client as any);
    vi.mocked(projectService.fetchProjectResources).mockResolvedValue({ data: resources, pagination: { page_num: 1, page_size: 10, total_data: 2 } } as any);

    const updateMock = vi.fn().mockResolvedValue(client);
    vi.mocked(projectService.updateClient).mockImplementation(updateMock as any);

    const httpClient = createMockHttpClient();

    render(
      <MemoryRouter initialEntries={[{ pathname: '/projects/p1/clients/cli-default/details', state: { projectId: 'p1' } }]}>
        <Routes>
          <Route path="/projects/:project_id/clients/:client_id/details" element={<ClientDetails httpClient={httpClient} />} />
        </Routes>
      </MemoryRouter>,
      { wrapper: createWrapper() as any }
    );

    await waitFor(() => expect(screen.getByDisplayValue('Default Client')).toBeInTheDocument());

    // Default permission checkbox should be disabled and checked
    const defaultCheckbox = screen.getByLabelText('permission-perm-default-1') as HTMLInputElement;
    expect(defaultCheckbox).toBeDisabled();
    expect(defaultCheckbox).toBeChecked(); // Checked since it's in the client's permission_ids

    // Custom permission checkbox should be enabled and checked
    const customCheckbox = screen.getByLabelText('permission-perm-custom-1') as HTMLInputElement;
    expect(customCheckbox).not.toBeDisabled();
    expect(customCheckbox).toBeChecked();

    // Click Save and verify all permission IDs are sent (both default and custom)
    fireEvent.click(screen.getByText('Save'));

    await waitFor(() => expect(updateMock).toHaveBeenCalled());
    expect(updateMock).toHaveBeenCalledWith(
      expect.anything(),
      'http://localhost:8082',
      'p1',
      'cli-default',
      expect.objectContaining({
        permission_ids: ['perm-default-1', 'perm-custom-1'], // Both permissions sent
      })
    );
  });
});
