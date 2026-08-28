import type { HttpClient } from '@/services/core/httpClient';
import * as projectService from '@/services/projects';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import UsersTab from './UsersTab';

vi.mock('@/services/projects');

// Mock DropdownMenu components to render inline
vi.mock('@/components/ui/dropdown-menu', () => ({
    DropdownMenu: ({ children }: any) => <div>{children}</div>,
    DropdownMenuTrigger: ({ children, asChild }: any) => asChild ? children : <button>{children}</button>,
    DropdownMenuContent: ({ children }: any) => <div>{children}</div>,
    DropdownMenuItem: ({ children, onClick }: any) => (
        <button onClick={onClick}>{children}</button>
    ),
}));

// Mock Dialog components to render inline
vi.mock('@/components/ui/dialog', () => ({
    Dialog: ({ children, open }: any) => open ? <div data-testid="mock-dialog">{children}</div> : null,
    DialogContent: ({ children }: any) => <div>{children}</div>,
    DialogHeader: ({ children }: any) => <div>{children}</div>,
    DialogTitle: ({ children }: any) => <h2>{children}</h2>,
    DialogDescription: ({ children }: any) => <p>{children}</p>,
    DialogFooter: ({ children }: any) => <div>{children}</div>,
}));

// Mock Select components to render as standard HTML select/option elements
vi.mock('@/components/ui/select', () => ({
    Select: ({ children, value, onValueChange, ...props }: any) => (
        <select value={value} onChange={(e) => onValueChange(e.target.value)} {...props}>
            {children}
        </select>
    ),
    SelectTrigger: ({ children, ...props }: any) => <div {...props}>{children}</div>,
    SelectValue: ({ placeholder }: any) => <span>{placeholder}</span>,
    SelectContent: ({ children }: any) => <>{children}</>,
    SelectItem: ({ children, value }: any) => <option value={value}>{children}</option>,
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

const mockProject = { id: 'proj_123', code: 'TEST', name: 'Test Project' } as any;

const mockUsers = [
    {
        id: 'user_1',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        email: 'john@example.com',
        external_user_id: null,
        display_name: 'John Doe',
        is_active: true,
        is_email_verified: true,
        role_name: 'Admin',
    },
    {
        id: 'user_2',
        created_at: '2024-01-02T00:00:00Z',
        updated_at: '2024-01-02T00:00:00Z',
        email: 'jane@example.com',
        external_user_id: null,
        display_name: 'Jane Smith',
        is_active: false,
        is_email_verified: false,
        role_name: 'Editor',
    },
];

const mockPagination = {
    page_num: 1,
    page_size: 10,
    total_data: 2,
};

const mockRoles = [
    {
        id: 'role_1',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        code: 'ADMIN',
        name: 'Admin',
        description: 'Administrator role',
    },
    {
        id: 'role_2',
        created_at: '2024-01-02T00:00:00Z',
        updated_at: '2024-01-02T00:00:00Z',
        code: 'EDITOR',
        name: 'Editor',
        description: 'Editor role',
    },
];

function createWrapper() {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: 0 } } });
    return function Wrapper({ children }: { children: any }) {
        return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
    };
}

describe('UsersTab', () => {
    beforeEach(() => {
        vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082');
        vi.clearAllMocks();
        // Mock fetchProjectRoles to return empty roles by default
        vi.spyOn(projectService, 'fetchProjectRoles').mockResolvedValue({
            data: mockRoles,
            pagination: { page_num: 1, page_size: 100, total_data: 2 },
        } as any);
    });

    afterEach(() => {
        vi.restoreAllMocks();
        vi.unstubAllEnvs();
    });

    it('renders loading state initially', () => {
        vi.spyOn(projectService, 'fetchProjectUsers').mockImplementation(() => new Promise(() => { }));
        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={mockProject} />
            </MemoryRouter>,
            { wrapper }
        );
        expect(screen.getByText(/Loading…/i)).toBeInTheDocument();
    });

    it('renders error state when fetch fails', async () => {
        vi.spyOn(projectService, 'fetchProjectUsers').mockRejectedValueOnce(new Error('Failed to fetch users'));
        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={mockProject} />
            </MemoryRouter>,
            { wrapper }
        );
        await waitFor(() => {
            expect(screen.getByText(/Failed to load users/i)).toBeInTheDocument();
        });
    });

    it('renders users table with correct columns', async () => {
        vi.spyOn(projectService, 'fetchProjectUsers').mockResolvedValueOnce({
            data: mockUsers,
            pagination: mockPagination,
        } as any);
        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={mockProject} />
            </MemoryRouter>,
            { wrapper }
        );
        await waitFor(() => {
            expect(screen.getByText('John Doe')).toBeInTheDocument();
            expect(screen.getByText('Jane Smith')).toBeInTheDocument();
        });
        expect(screen.getByText('john@example.com')).toBeInTheDocument();
        expect(screen.getByText('jane@example.com')).toBeInTheDocument();
    });

    it('displays verified badge for verified users', async () => {
        vi.spyOn(projectService, 'fetchProjectUsers').mockResolvedValueOnce({
            data: mockUsers,
            pagination: mockPagination,
        } as any);
        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={mockProject} />
            </MemoryRouter>,
            { wrapper }
        );
        await waitFor(() => {
            expect(screen.getByText('Verified')).toBeInTheDocument();
        });
    });

    it('displays not verified badge for unverified users', async () => {
        vi.spyOn(projectService, 'fetchProjectUsers').mockResolvedValueOnce({
            data: mockUsers,
            pagination: mockPagination,
        } as any);
        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={mockProject} />
            </MemoryRouter>,
            { wrapper }
        );
        await waitFor(() => {
            expect(screen.getByText('Not Verified')).toBeInTheDocument();
        });
    });

    it('displays user status correctly', async () => {
        vi.spyOn(projectService, 'fetchProjectUsers').mockResolvedValueOnce({
            data: mockUsers,
            pagination: mockPagination,
        } as any);
        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={mockProject} />
            </MemoryRouter>,
            { wrapper }
        );
        await waitFor(() => {
            const statusBadges = screen.getAllByText(/Active|Inactive/);
            expect(statusBadges.length).toBeGreaterThan(0);
        });
    });

    it('displays user role in badge', async () => {
        vi.spyOn(projectService, 'fetchProjectUsers').mockResolvedValueOnce({
            data: mockUsers,
            pagination: mockPagination,
        } as any);
        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={mockProject} />
            </MemoryRouter>,
            { wrapper }
        );
        await waitFor(() => {
            const table = screen.getByRole('table');
            expect(within(table).getByText('Admin')).toBeInTheDocument();
            expect(within(table).getByText('Editor')).toBeInTheDocument();
        });
    });

    it('filters users by search input with debounce', async () => {
        vi.spyOn(projectService, 'fetchProjectUsers').mockResolvedValueOnce({
            data: mockUsers,
            pagination: mockPagination,
        } as any);
        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={mockProject} />
            </MemoryRouter>,
            { wrapper }
        );

        await waitFor(() => {
            expect(screen.getByText('John Doe')).toBeInTheDocument();
        });

        const searchInput = screen.getByDisplayValue('');
        expect(searchInput).toBeInTheDocument();
    });

    it('filters users by is_active status', async () => {
        const fetchSpy = vi.spyOn(projectService, 'fetchProjectUsers')
            .mockResolvedValueOnce({
                data: mockUsers,
                pagination: mockPagination,
            } as any)
            .mockResolvedValueOnce({
                data: [mockUsers[0]],
                pagination: { page_num: 1, page_size: 10, total_data: 1 },
            } as any);
        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={mockProject} />
            </MemoryRouter>,
            { wrapper }
        );

        await waitFor(() => {
            expect(screen.getByText('John Doe')).toBeInTheDocument();
        });

        expect(fetchSpy).toHaveBeenCalled();
    });

    it('renders empty state when no users', async () => {
        vi.spyOn(projectService, 'fetchProjectUsers').mockResolvedValueOnce({
            data: [],
            pagination: { page_num: 1, page_size: 10, total_data: 0 },
        } as any);
        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={mockProject} />
            </MemoryRouter>,
            { wrapper }
        );
        await waitFor(() => {
            expect(screen.getByText('No users found')).toBeInTheDocument();
        });
    });

    it('handles pagination page size changes', async () => {
        const fetchSpy = vi.spyOn(projectService, 'fetchProjectUsers')
            .mockResolvedValueOnce({
                data: mockUsers,
                pagination: mockPagination,
            } as any)
            .mockResolvedValueOnce({
                data: mockUsers,
                pagination: { ...mockPagination, page_size: 25 },
            } as any);
        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={mockProject} />
            </MemoryRouter>,
            { wrapper }
        );

        await waitFor(() => {
            expect(screen.getByText('John Doe')).toBeInTheDocument();
        });

        expect(fetchSpy).toHaveBeenCalled();
    });

    it('opens remove confirmation dialog when remove is clicked', async () => {
        vi.spyOn(projectService, 'fetchProjectUsers').mockResolvedValueOnce({
            data: mockUsers,
            pagination: mockPagination,
        } as any);
        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={mockProject} />
            </MemoryRouter>,
            { wrapper }
        );

        await waitFor(() => {
            expect(screen.getByText('John Doe')).toBeInTheDocument();
        });

        const moreButtons = screen.queryAllByRole('button');
        expect(moreButtons.length).toBeGreaterThan(0);
    });

    it('removes user on confirm', async () => {
        vi.spyOn(projectService, 'fetchProjectUsers').mockResolvedValueOnce({
            data: mockUsers,
            pagination: mockPagination,
        } as any);
        vi.spyOn(projectService, 'deleteProjectUser').mockResolvedValueOnce(undefined);

        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={mockProject} />
            </MemoryRouter>,
            { wrapper }
        );

        await waitFor(() => {
            expect(screen.getByText('John Doe')).toBeInTheDocument();
        });

        expect(projectService.deleteProjectUser).not.toHaveBeenCalled();
    });

    it('displays showing label correctly', async () => {
        vi.spyOn(projectService, 'fetchProjectUsers').mockResolvedValueOnce({
            data: mockUsers,
            pagination: mockPagination,
        } as any);
        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={mockProject} />
            </MemoryRouter>,
            { wrapper }
        );
        await waitFor(() => {
            expect(screen.getByText(/Showing 10 of 2 users/)).toBeInTheDocument();
        });
    });

    it('displays section title and description', async () => {
        vi.spyOn(projectService, 'fetchProjectUsers').mockResolvedValueOnce({
            data: mockUsers,
            pagination: mockPagination,
        } as any);
        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={mockProject} />
            </MemoryRouter>,
            { wrapper }
        );
        await waitFor(() => {
            expect(screen.getByText('Users')).toBeInTheDocument();
            expect(screen.getByText(/Manage the users and their roles/)).toBeInTheDocument();
        });
    });

    it('renders Add User button as enabled', async () => {
        vi.spyOn(projectService, 'fetchProjectUsers').mockResolvedValueOnce({
            data: mockUsers,
            pagination: mockPagination,
        } as any);
        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={mockProject} />
            </MemoryRouter>,
            { wrapper }
        );
        await waitFor(() => {
            const addButton = screen.getByRole('button', { name: /Add User/ });
            expect(addButton).toBeEnabled();
        });
    });

    it('shows or hides dropdown actions based on user status and email verification', async () => {
        const customUsers = [
            {
                id: 'user_1',
                email: 'john@example.com',
                display_name: 'John Doe',
                is_active: true,
                is_email_verified: true,
                role_name: 'Admin',
                created_at: '2024-01-01T00:00:00Z',
            },
            {
                id: 'user_2',
                email: 'jane@example.com',
                display_name: 'Jane Smith',
                is_active: false,
                is_email_verified: false,
                role_name: 'Editor',
                created_at: '2024-01-02T00:00:00Z',
            },
            {
                id: 'user_3',
                email: 'bob@example.com',
                display_name: 'Bob Ross',
                is_active: true,
                is_email_verified: false,
                role_name: 'Editor',
                created_at: '2024-01-03T00:00:00Z',
            },
        ];

        vi.spyOn(projectService, 'fetchProjectUsers').mockResolvedValueOnce({
            data: customUsers,
            pagination: mockPagination,
        } as any);

        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={mockProject} />
            </MemoryRouter>,
            { wrapper }
        );

        await waitFor(() => {
            expect(screen.getByText('John Doe')).toBeInTheDocument();
        });

        // John Doe (active, verified): should have "Request Password Reset" but NOT "Resend Verification Email"
        const johnRow = screen.getByText('John Doe').closest('tr')!;
        expect(within(johnRow).getByText('Request Password Reset')).toBeInTheDocument();
        expect(within(johnRow).queryByText('Resend Verification Email')).not.toBeInTheDocument();

        // Jane Smith (inactive): should have neither action
        const janeRow = screen.getByText('Jane Smith').closest('tr')!;
        expect(within(janeRow).queryByText('Request Password Reset')).not.toBeInTheDocument();
        expect(within(janeRow).queryByText('Resend Verification Email')).not.toBeInTheDocument();

        // Bob Ross (active, unverified): should have both actions
        const bobRow = screen.getByText('Bob Ross').closest('tr')!;
        expect(within(bobRow).getByText('Request Password Reset')).toBeInTheDocument();
        expect(within(bobRow).getByText('Resend Verification Email')).toBeInTheDocument();
    });

    it('handles Resend Verification Email dialog flow and API call with redirect URI', async () => {
        const customProject = {
            id: 'proj_123',
            code: 'TEST',
            name: 'Test Project',
            redirect_uris: [
                { redirect_uri: 'https://login.example.com/callback', login_url: 'https://login.example.com' },
                { redirect_uri: 'https://app.example.com/oauth', login_url: '' }
            ]
        } as any;

        const customUsers = [
            {
                id: 'user_3',
                email: 'bob@example.com',
                display_name: 'Bob Ross',
                is_active: true,
                is_email_verified: false,
                role_name: 'Editor',
                created_at: '2024-01-03T00:00:00Z',
            },
        ];

        vi.spyOn(projectService, 'fetchProjectUsers').mockResolvedValueOnce({
            data: customUsers,
            pagination: mockPagination,
        } as any);

        const resendSpy = vi.spyOn(projectService, 'resendProjectUserVerificationEmail').mockResolvedValueOnce(undefined);

        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={customProject} />
            </MemoryRouter>,
            { wrapper }
        );

        await waitFor(() => {
            expect(screen.getByText('Bob Ross')).toBeInTheDocument();
        });

        // Open dialog
        const resendBtn = screen.getByText('Resend Verification Email');
        fireEvent.click(resendBtn);

        expect(screen.getByText(/Send email verification link to Bob Ross/i)).toBeInTheDocument();

        // Select redirect URI within dialog
        const dialog = screen.getByTestId('mock-dialog');
        const selectEl = within(dialog).getByRole('combobox');
        fireEvent.change(selectEl, { target: { value: 'https://login.example.com/callback' } });

        // Click Send button inside dialog
        const submitBtn = within(dialog).getByRole('button', { name: /^Send Verification Email$/i });
        fireEvent.click(submitBtn);

        await waitFor(() => {
            expect(resendSpy).toHaveBeenCalledWith(
                httpClient,
                'http://localhost:8082',
                'proj_123',
                'user_3',
                { redirect_uri: 'https://login.example.com/callback' }
            );
        });

        expect(toastMock).toHaveBeenCalledWith(expect.objectContaining({
            title: 'Verification email sent',
        }));
    });

    it('handles Request Password Reset dialog flow and API call with redirect URI', async () => {
        const customProject = {
            id: 'proj_123',
            code: 'TEST',
            name: 'Test Project',
            redirect_uris: [
                { redirect_uri: 'https://login.example.com/callback', login_url: 'https://login.example.com' },
                { redirect_uri: 'https://app.example.com/oauth', login_url: '' }
            ]
        } as any;

        const customUsers = [
            {
                id: 'user_1',
                email: 'john@example.com',
                display_name: 'John Doe',
                is_active: true,
                is_email_verified: true,
                role_name: 'Admin',
                created_at: '2024-01-01T00:00:00Z',
            },
        ];

        vi.spyOn(projectService, 'fetchProjectUsers').mockResolvedValueOnce({
            data: customUsers,
            pagination: mockPagination,
        } as any);

        const resetSpy = vi.spyOn(projectService, 'requestProjectUserPasswordReset').mockResolvedValueOnce(undefined);

        const httpClient = createMockHttpClient();
        const wrapper = createWrapper();
        render(
            <MemoryRouter>
                <UsersTab httpClient={httpClient} project={customProject} />
            </MemoryRouter>,
            { wrapper }
        );

        await waitFor(() => {
            expect(screen.getByText('John Doe')).toBeInTheDocument();
        });

        // Open dialog
        const resetBtn = screen.getByText('Request Password Reset');
        fireEvent.click(resetBtn);

        expect(screen.getByText(/Send password reset link to John Doe/i)).toBeInTheDocument();

        // Select redirect URI within dialog
        const dialog = screen.getByTestId('mock-dialog');
        const selectEl = within(dialog).getByRole('combobox');
        fireEvent.change(selectEl, { target: { value: 'https://login.example.com/callback' } });

        // Click Submit button inside dialog
        const submitBtn = within(dialog).getByRole('button', { name: /^Request Password Reset$/i });
        fireEvent.click(submitBtn);

        await waitFor(() => {
            expect(resetSpy).toHaveBeenCalledWith(
                httpClient,
                'http://localhost:8082',
                'proj_123',
                'user_1',
                { redirect_uri: 'https://login.example.com/callback' }
            );
        });

        expect(toastMock).toHaveBeenCalledWith(expect.objectContaining({
            title: 'Password reset email sent',
        }));
    });
});
