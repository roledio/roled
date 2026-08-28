import { TooltipProvider } from '@/components/ui/tooltip';
import type { HttpClient } from '@/services/core/httpClient';
import * as projectService from '@/services/projects';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import NewResource from './NewResource';

vi.mock('@/services/projects', () => ({
    fetchProjectById: vi.fn(),
    createProjectResource: vi.fn(),
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

describe('NewResource Page', () => {
    beforeEach(() => vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082'));
    afterEach(() => { vi.restoreAllMocks(); vi.unstubAllEnvs(); });

    it('fetches project and creates a resource', async () => {
        const project = { id: 'p1', name: 'Test Project' } as any;

        vi.mocked(projectService.fetchProjectById).mockResolvedValue(project);
        const createMock = vi.fn().mockResolvedValue({ id: 'r-new' });
        vi.mocked(projectService.createProjectResource).mockImplementation(createMock);

        const httpClient = createMockHttpClient();

        render(
            <MemoryRouter initialEntries={[{ pathname: '/projects/p1/resources/new' }]}>
                <Routes>
                    <Route path="/projects/:project_id/resources/new" element={<NewResource httpClient={httpClient} />} />
                    <Route path="/projects/:project_id/details" element={<div>PROJECT DETAILS</div>} />
                </Routes>
            </MemoryRouter>,
            { wrapper: createWrapper() as any }
        );

        await waitFor(() => expect(screen.getByText('Add New Resource')).toBeInTheDocument());

        // fill in name input
        fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Orders' } });
        // code auto-generated
        await waitFor(() => expect(screen.getByDisplayValue('orders')).toBeInTheDocument());

        // add a permission
        fireEvent.click(screen.getByRole('button', { name: 'Add Permission' }));
        await waitFor(() => expect(screen.getByLabelText('Name (Action)')).toBeInTheDocument());
        fireEvent.change(screen.getByLabelText('Name (Action)'), { target: { value: 'Read' } });
        fireEvent.click(screen.getByText('Add'));

        // submit
        fireEvent.click(screen.getByText('Create'));

        await waitFor(() => expect(createMock).toHaveBeenCalled());
        await waitFor(() => expect(screen.getByText('PROJECT DETAILS')).toBeInTheDocument());
    });
});