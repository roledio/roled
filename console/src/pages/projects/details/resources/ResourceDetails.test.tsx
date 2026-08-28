import { TooltipProvider } from '@/components/ui/tooltip';
import type { HttpClient } from '@/services/core/httpClient';
import * as projectService from '@/services/projects';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import ResourceDetails from './ResourceDetails';

vi.mock('@/services/projects', () => ({
    fetchProjectById: vi.fn(),
    fetchProjectResourceById: vi.fn(),
    updateProjectResource: vi.fn(),
    deleteProjectResource: vi.fn(),
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

describe('ResourceDetails Page', () => {
    beforeEach(() => vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082'));
    afterEach(() => { vi.restoreAllMocks(); vi.unstubAllEnvs(); });

    it('renders default resource data without remove button', async () => {
        const resource = {
            id: 'res-1',
            name: 'Accounts',
            code: 'accounts',
            description: 'Accounts resource',
            is_default: true,
            permissions: [{ id: 'perm-1', name: 'Read', code: 'read', description: 'Read accounts', is_default: true }],
        } as any;

        vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p1', name: 'Test' } as any);
        vi.mocked(projectService.fetchProjectResourceById).mockResolvedValue(resource as any);

        const httpClient = createMockHttpClient();

        render(
            <MemoryRouter initialEntries={[`/projects/p1/resources/res-1/details`]}>
                <Routes>
                    <Route path="/projects/:project_id/resources/:resource_id/details" element={<ResourceDetails httpClient={httpClient} />} />
                </Routes>
            </MemoryRouter>,
            { wrapper: createWrapper() as any }
        );

        await waitFor(() => expect(screen.getByDisplayValue('Accounts')).toBeInTheDocument());
        const nameInput = screen.getByDisplayValue('Accounts') as HTMLInputElement;
        const codeInput = screen.getByDisplayValue('accounts') as HTMLInputElement;
        const descInput = screen.getByDisplayValue('Accounts resource') as HTMLInputElement;
        // default resource -> inputs readonly and no delete button rendered
        expect(nameInput).toHaveAttribute('readonly');
        expect(codeInput).toHaveAttribute('readonly');
        expect(descInput).toHaveAttribute('readonly');
        expect(screen.queryByText('Remove')).toBeNull();
    });

    it('shows remove button for non-default resource and deletes on confirm', async () => {
        const resource = {
            id: 'res-3',
            name: 'Orders',
            code: 'orders',
            description: 'Orders resource',
            is_default: false,
            permissions: [],
        } as any;

        vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p3', name: 'Test' } as any);
        vi.mocked(projectService.fetchProjectResourceById).mockResolvedValue(resource as any);
        const deleteMock = vi.mocked(projectService.deleteProjectResource).mockResolvedValue(undefined as any);

        const httpClient = createMockHttpClient();

        render(
            <MemoryRouter initialEntries={[`/projects/p3/resources/res-3/details`]}>
                <Routes>
                    <Route path="/projects/:project_id/resources/:resource_id/details" element={<ResourceDetails httpClient={httpClient} />} />
                    <Route path="/projects/:project_id/details" element={<div>PROJECT DETAILS</div>} />
                </Routes>
            </MemoryRouter>,
            { wrapper: createWrapper() as any }
        );

        await waitFor(() => expect(screen.getByDisplayValue('Orders')).toBeInTheDocument());
        expect(screen.getByText('Save')).toBeInTheDocument();
        expect(screen.getByText('Remove')).toBeInTheDocument();

        fireEvent.click(screen.getByText('Remove'));
        await waitFor(() => expect(screen.getByText('Remove Resource')).toBeInTheDocument());
        fireEvent.click(screen.getAllByText('Remove')[1]);

        await waitFor(() => expect(deleteMock).toHaveBeenCalled());
        await waitFor(() => expect(screen.getByText('PROJECT DETAILS')).toBeInTheDocument());
    });

    it('saves changes and navigates back on save', async () => {
        const resource = {
            id: 'res-2', name: 'Orders', code: 'orders', description: 'Order resources', permissions: [],
        } as any;

        vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p2', name: 'Proj' } as any);
        vi.mocked(projectService.fetchProjectResourceById).mockResolvedValue(resource as any);
        const updateMock = vi.mocked(projectService.updateProjectResource).mockResolvedValue({ ...resource, name: 'Orders Updated' } as any);

        const httpClient = createMockHttpClient();

        render(
            <MemoryRouter initialEntries={[`/projects/p2/resources/res-2/details`]}>
                <Routes>
                    <Route path="/projects/:project_id/resources/:resource_id/details" element={<ResourceDetails httpClient={httpClient} />} />
                    <Route path="/projects/:project_id/details" element={<div>PROJECT DETAILS</div>} />
                </Routes>
            </MemoryRouter>,
            { wrapper: createWrapper() as any }
        );

        await waitFor(() => expect(screen.getByDisplayValue('Orders')).toBeInTheDocument());

        fireEvent.change(screen.getByDisplayValue('Orders'), { target: { value: 'Orders New' } });
        // code auto-generated
        await waitFor(() => expect(screen.getByDisplayValue('orders-new')).toBeInTheDocument());

        fireEvent.click(screen.getByText('Save'));

        await waitFor(() => expect(updateMock).toHaveBeenCalled());
        await waitFor(() => expect(screen.getByText('PROJECT DETAILS')).toBeInTheDocument());
    });
});
