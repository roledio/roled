import { TooltipProvider } from '@/components/ui/tooltip';
import type { HttpClient } from '@/services/core/httpClient';
import * as projectService from '@/services/projects';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor, within, cleanup } from '@testing-library/react';
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

describe('ResourceDetails permission editor', () => {
 beforeEach(() => { vi.clearAllMocks(); vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082'); });
 afterEach(() => { cleanup(); vi.unstubAllEnvs(); });
 function mount() {
  vi.mocked(projectService.fetchProjectById).mockResolvedValue({ id: 'p', name: 'Project', logo_url: 'https://example.com/logo.png' } as Awaited<ReturnType<typeof projectService.fetchProjectById>>);
  vi.mocked(projectService.fetchProjectResourceById).mockResolvedValue({ id: 'r', name: 'Documents', code: 'documents', description: 'Files', is_default: false, permissions: [{id:'read',name:'Read',code:'read',description:'Read files'}, {id:'write',name:'Write',code:'write',description:'Write files'}] } as Awaited<ReturnType<typeof projectService.fetchProjectResourceById>>);
  const client = createMockHttpClient();
  render(<MemoryRouter initialEntries={['/projects/p/resources/r/details']}><Routes><Route path="/projects/:project_id/resources/:resource_id/details" element={<ResourceDetails httpClient={client} />} /><Route path="/projects/:project_id/details" element={<div>Returned to project</div>} /></Routes></MemoryRouter>, { wrapper: createWrapper() });
  return client;
 }
 it('adds, validates, sorts and removes permissions before saving the edited payload', async () => {
  const client = mount(); await screen.findByDisplayValue('Documents');
  const table = screen.getByRole('table');
  fireEvent.click(screen.getByRole('button', { name: 'Name (Action)' }));
  expect(within(table).getAllByRole('row')[1]).toHaveTextContent('Write');
  fireEvent.click(screen.getByRole('button', { name: 'Code', exact: true }));
  expect(within(table).getAllByRole('row')[1]).toHaveTextContent('Read');
  fireEvent.click(screen.getByRole('button', { name: 'Code', exact: true }));
  expect(within(table).getAllByRole('row')[1]).toHaveTextContent('Write');
  fireEvent.click(screen.getByRole('button', { name: 'Add Permission' }));
  const dialog = screen.getByRole('dialog');
  fireEvent.blur(within(dialog).getByLabelText('Name (Action)')); expect(within(dialog).getByLabelText('Name (Action)')).toHaveAttribute('aria-invalid', 'true');
  fireEvent.change(within(dialog).getByLabelText('Name (Action)'), { target: { value: 'Read' } });
  fireEvent.blur(within(dialog).getByLabelText('Code')); expect(within(dialog).getByLabelText('Code')).toHaveAttribute('aria-invalid', 'true');
  fireEvent.change(within(dialog).getByLabelText('Name (Action)'), { target: { value: 'Archive' } });
  fireEvent.change(within(dialog).getByLabelText('Code'), { target: { value: 'archive-record' } });
  fireEvent.change(within(dialog).getByLabelText('Description'), { target: { value: 'Archive old files' } });
  fireEvent.blur(within(dialog).getByLabelText('Description'));
  fireEvent.click(within(dialog).getByRole('button', { name: 'Add', exact: true }));
  await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  expect(screen.getByText('documents:archive-record')).toBeInTheDocument();
  const row = screen.getByText('Read files').closest('tr')!;
  fireEvent.click(within(row).getByRole('button'));
  expect(screen.queryByText('Read files')).not.toBeInTheDocument();
  fireEvent.change(screen.getByLabelText('Description'), { target: { value: 'Updated files' } });
  fireEvent.blur(screen.getByLabelText('Description'));
  vi.mocked(projectService.updateProjectResource).mockResolvedValue({} as Awaited<ReturnType<typeof projectService.updateProjectResource>>);
  fireEvent.click(screen.getByRole('button', {name:'Save'}));
  await screen.findByText('Returned to project');
  expect(projectService.updateProjectResource).toHaveBeenCalledWith(client,'http://localhost:8082','p','r',{name:'Documents',code:'documents',description:'Updated files',permissions:[{name:'Write',code:'write',description:'Write files'},{name:'Archive',code:'archive-record',description:'Archive old files'}]});
 });
 it('prevents invalid fields, cancels dialogs and returns using Back', async () => {
  mount(); await screen.findByDisplayValue('Documents');
  fireEvent.change(screen.getByLabelText('Name'), {target:{value:''}});fireEvent.blur(screen.getByLabelText('Name'));
  expect(screen.getByLabelText('Name')).toHaveAttribute('aria-invalid','true'); expect(screen.getByRole('button',{name:'Save'})).toBeDisabled();
  fireEvent.change(screen.getByLabelText('Name'), {target:{value:'Documents'}});
  fireEvent.change(screen.getByLabelText('Code'), {target:{value:''}});fireEvent.blur(screen.getByLabelText('Code'));
  expect(screen.getByLabelText('Code')).toHaveAttribute('aria-invalid','true');
  fireEvent.click(screen.getByRole('button',{name:'Add Permission'}));fireEvent.click(within(screen.getByRole('dialog')).getByRole('button',{name:'Cancel'}));
  fireEvent.click(screen.getByRole('button',{name:'Remove',exact:true}));fireEvent.click(within(screen.getByRole('alertdialog')).getByRole('button',{name:'Cancel'}));
  expect(projectService.deleteProjectResource).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole('button',{name:'Back to Project'}));await screen.findByText('Returned to project');
 });
 it('keeps edits visible when saving fails',async()=>{
  mount();await screen.findByDisplayValue('Documents');vi.mocked(projectService.updateProjectResource).mockRejectedValue(new Error('Offline'));
  fireEvent.click(screen.getByRole('button',{name:'Save'}));await waitFor(()=>expect(projectService.updateProjectResource).toHaveBeenCalled());
  await waitFor(()=>expect(screen.getByRole('button',{name:'Save'})).toBeEnabled());expect(screen.getByDisplayValue('Documents')).toBeInTheDocument();expect(screen.queryByText('Returned to project')).not.toBeInTheDocument();
 });
 it('renders a loading state until the project request settles',()=>{
  mount();expect(screen.getByText('Loading…')).toBeInTheDocument();
 });
});
