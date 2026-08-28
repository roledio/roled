import { TooltipProvider } from '@/components/ui/tooltip';
import * as accountService from '@/services/accounts';
import type { HttpClient } from '@/services/core/httpClient';
import type { TokenService } from '@/services/core/tokenService';
import * as memberService from '@/services/members';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import * as authService from '@/services/core/authService';
import Account from './Account';

// Mock the accountService module
vi.mock('@/services/accounts', () => ({
    fetchCurrentAccount: vi.fn(),
    updateAccount: vi.fn(),
}));

// Mock the memberService
vi.mock('@/services/members', () => ({
    fetchMembers: vi.fn(),
    inviteMember: vi.fn(),
    deleteMember: vi.fn(),
    updateMember: vi.fn(),
}));

// Mock the DropdownMenu components to render inline for testing
vi.mock('@/components/ui/dropdown-menu', () => ({
    DropdownMenu: ({ children }: any) => <div>{children}</div>,
    DropdownMenuTrigger: ({ children, asChild }: any) => asChild ? children : <button>{children}</button>,
    DropdownMenuContent: ({ children }: any) => <div data-testid="dropdown-menu-content">{children}</div>,
    DropdownMenuItem: ({ children, onSelect }: any) => (
        <button onClick={() => onSelect && onSelect({} as any)}>{children}</button>
    ),
}));

// Mock the authService
vi.mock('@/services/core/authService', () => ({
    fetchCurrentTokenInfo: vi.fn(),
    revokeCurrentToken: vi.fn(),
}));

// Mock useToast to capture toast calls
const toastMock = vi.fn();
vi.mock('@/hooks/use-toast', () => ({
    useToast: () => ({ toast: toastMock }),
}));

const mockAccountData: accountService.Account = {
    id: '2JMeGXjYdW27uh6T4cbCSt',
    name: 'Roled (System)',
    description: 'Roled system account',
    is_active: true,
    created_at: '2026-02-25T12:05:01.9446Z',
    updated_at: '2026-02-26T08:30:00.0000Z',
};

function createMockTokenService(): TokenService {
    let cachedToken: any = null;
    let cachedMember: any = undefined;
    return {
        clear: vi.fn(),
        getRefreshToken: vi.fn().mockReturnValue('refresh123'),
        getAccessToken: vi.fn().mockReturnValue('access123'),
        isAccessTokenValid: vi.fn().mockReturnValue(true),
        getCurrentTokenInfo: vi.fn().mockImplementation(() => cachedToken),
        setCurrentTokenInfo: vi.fn().mockImplementation((info) => { cachedToken = info; }),
        getCurrentMember: vi.fn().mockImplementation(() => cachedMember),
        setCurrentMember: vi.fn().mockImplementation((member) => { cachedMember = member; }),
    } as unknown as TokenService;
}

function createMockHttpClient(tokenService?: TokenService): HttpClient {
    const ts = tokenService || createMockTokenService();
    return {
        instanceRef: {
            get: vi.fn(),
            post: vi.fn(),
            put: vi.fn(),
            delete: vi.fn(),
        },
        tokenServiceRef: ts,
    } as unknown as HttpClient;
}

function createWrapper() {
    const queryClient = new QueryClient({
        defaultOptions: {
            queries: {
                retry: false,
                gcTime: 0,
            },
        },
    });

    return function Wrapper({ children }: { children: React.ReactNode }) {
        return (
            <QueryClientProvider client={queryClient}>
                <TooltipProvider>
                    {children}
                </TooltipProvider>
            </QueryClientProvider>
        );
    };
}

describe('Account Page', () => {
    let mockMembersList: any[] = [];
    let currentUserIsAdmin = true;

    beforeEach(() => {
        vi.stubEnv('VITE_AUTH_BASE_URL', 'http://localhost:8082');
        mockMembersList = [];
        currentUserIsAdmin = true;

        vi.mocked(authService.fetchCurrentTokenInfo).mockResolvedValue({
            id: 'tok-123',
            issued_at: '2026-05-19T15:19:36.5887Z',
            expires_at: '2026-05-20T15:19:36.5887Z',
            project: { id: 'p1', name: 'Roled Console', description: '', logo_url: '' },
            client: { id: 'c1', name: 'Roled Client' },
            user: {
                id: 'usr-current',
                email: 'current_user@example.com',
                display_name: 'Current User',
            },
            role: { id: 'r1', code: 'admin', name: 'Admin', description: '' },
            permissions: [],
        });

        vi.mocked(memberService.fetchMembers).mockImplementation(async (httpClient, baseUrl, params) => {
            if (params && params.search === 'current_user@example.com') {
                return {
                    data: [
                        {
                            id: 'usr-current-member',
                            email: 'current_user@example.com',
                            display_name: 'Current User',
                            is_active: true,
                            is_verified: true,
                            is_admin: currentUserIsAdmin,
                            created_at: '',
                            updated_at: '',
                        }
                    ],
                };
            }
            return {
                data: mockMembersList,
                pagination: { page_num: 1, page_size: 5, total_data: mockMembersList.length }
            };
        });
    });

    afterEach(() => {
        vi.restoreAllMocks();
        vi.unstubAllEnvs();
    });

    it('shows loading state while fetching account', () => {
        // Make fetchCurrentAccount hang (never resolve)
        vi.mocked(accountService.fetchCurrentAccount).mockReturnValue(new Promise(() => { }));

        const tokenService = createMockTokenService();
        const httpClient = createMockHttpClient(tokenService);
        render(<Account httpClient={httpClient} tokenService={tokenService} />, { wrapper: createWrapper() });

        expect(screen.getByTestId('account-loading')).toBeInTheDocument();
        expect(screen.getByText('Loading account…')).toBeInTheDocument();
    });

    it('shows error state when API call fails', async () => {
        vi.mocked(accountService.fetchCurrentAccount).mockRejectedValue(
            new Error('Failed to fetch account: 500 Internal Server Error'),
        );

        const tokenService = createMockTokenService();
        const httpClient = createMockHttpClient(tokenService);
        render(<Account httpClient={httpClient} tokenService={tokenService} />, { wrapper: createWrapper() });

        await waitFor(
            () => {
                expect(screen.getByTestId('account-error')).toBeInTheDocument();
            },
            { timeout: 3000 },
        );

        expect(screen.getByText('Failed to load account')).toBeInTheDocument();
        expect(screen.getByText('Failed to fetch account: 500 Internal Server Error')).toBeInTheDocument();
    });

    it('renders account details after successful fetch', async () => {
        vi.mocked(accountService.fetchCurrentAccount).mockResolvedValue(mockAccountData);

        const tokenService = createMockTokenService();
        const httpClient = createMockHttpClient(tokenService);
        render(<Account httpClient={httpClient} tokenService={tokenService} />, { wrapper: createWrapper() });

        await waitFor(() => {
            expect(screen.getByText('Roled (System)')).toBeInTheDocument();
        });

        // Account name and description
        expect(screen.getByText('Roled system account')).toBeInTheDocument();

        // Status badge within the account info section
        const statusLabels = screen.getAllByText('Status');
        const statusSection = statusLabels[0].closest('div')!;
        expect(statusSection.textContent).toContain('Active');

        // No timestamp assertions in the current UI; ensure main fields render
        expect(screen.getByText('Roled (System)')).toBeInTheDocument();
        expect(screen.getByText('Roled system account')).toBeInTheDocument();
    });

    it('calls fetchCurrentAccount with httpClient and base URL', async () => {
        vi.mocked(accountService.fetchCurrentAccount).mockResolvedValue(mockAccountData);

        const tokenService = createMockTokenService();
        const httpClient = createMockHttpClient(tokenService);
        render(<Account httpClient={httpClient} tokenService={tokenService} />, { wrapper: createWrapper() });

        await waitFor(() => {
            expect(accountService.fetchCurrentAccount).toHaveBeenCalledWith(
                httpClient,
                'http://localhost:8082',
            );
        });
    });

    it('enters edit mode and allows editing name and description', async () => {
        vi.mocked(accountService.fetchCurrentAccount).mockResolvedValue(mockAccountData);

        const tokenService = createMockTokenService();
        const httpClient = createMockHttpClient(tokenService);
        render(<Account httpClient={httpClient} tokenService={tokenService} />, { wrapper: createWrapper() });

        // Wait for data to load
        await waitFor(() => {
            expect(screen.getByText('Roled (System)')).toBeInTheDocument();
        });

        // Click Edit button
        fireEvent.click(screen.getByRole('button', { name: /edit/i }));

        // Name input should be visible with current value
        const nameInput = screen.getByLabelText('Name') as HTMLInputElement;
        expect(nameInput.value).toBe('Roled (System)');

        // Description textarea should be visible with current value
        const descInput = screen.getByLabelText('Description') as HTMLTextAreaElement;
        expect(descInput.value).toBe('Roled system account');

        // Edit the name
        fireEvent.change(nameInput, { target: { value: 'Updated Name' } });
        expect(nameInput.value).toBe('Updated Name');

        // Cancel should revert
        fireEvent.click(screen.getByRole('button', { name: /cancel/i }));

        // Should show original values again (not in edit mode)
        expect(screen.getByText('Roled (System)')).toBeInTheDocument();
    });

    it('updates account on save', async () => {
        vi.mocked(accountService.fetchCurrentAccount).mockResolvedValue(mockAccountData);
        const updated = { ...mockAccountData, name: 'New Name', description: 'New desc' };
        vi.mocked(accountService.updateAccount).mockResolvedValue(updated as any);

        const tokenService = createMockTokenService();
        const httpClient = createMockHttpClient(tokenService);
        render(<Account httpClient={httpClient} tokenService={tokenService} />, { wrapper: createWrapper() });

        // Wait for data to load
        await waitFor(() => expect(screen.getByText('Roled (System)')).toBeInTheDocument());

        // Enter edit mode
        fireEvent.click(screen.getByRole('button', { name: /edit/i }));

        const nameInput = screen.getByLabelText('Name') as HTMLInputElement;
        const descInput = screen.getByLabelText('Description') as HTMLTextAreaElement;

        fireEvent.change(nameInput, { target: { value: 'New Name' } });
        fireEvent.change(descInput, { target: { value: 'New desc' } });

        // Click Save
        fireEvent.click(screen.getByRole('button', { name: /save/i }));

        await waitFor(() => {
            expect(accountService.updateAccount).toHaveBeenCalledWith(
                httpClient,
                'http://localhost:8082',
                mockAccountData.id,
                { name: 'New Name', description: 'New desc' },
            );
        });

        // After save, new values should be displayed and not in edit mode
        await waitFor(() => expect(screen.getByText('New Name')).toBeInTheDocument());
        expect(screen.getByText('New desc')).toBeInTheDocument();
    });

    it('shows API error message in toast when update fails', async () => {
        vi.mocked(accountService.fetchCurrentAccount).mockResolvedValue(mockAccountData);
        vi.mocked(accountService.updateAccount).mockRejectedValue({
            response: { data: { error: { message: 'System account modification not allowed' } } },
        });

        const tokenService = createMockTokenService();
        const httpClient = createMockHttpClient(tokenService);
        render(<Account httpClient={httpClient} tokenService={tokenService} />, { wrapper: createWrapper() });

        await waitFor(() => expect(screen.getByText('Roled (System)')).toBeInTheDocument());

        // Enter edit mode and change name
        fireEvent.click(screen.getByRole('button', { name: /edit/i }));
        const nameInput = screen.getByLabelText('Name') as HTMLInputElement;
        fireEvent.change(nameInput, { target: { value: 'Will Fail' } });

        // Click Save which will trigger updateAccount rejection
        fireEvent.click(screen.getByRole('button', { name: /save/i }));

        await waitFor(() => {
            expect(toastMock).toHaveBeenCalled();
        });

        const callArg = toastMock.mock.calls[toastMock.mock.calls.length - 1][0];
        expect(callArg.description).toContain('System account modification not allowed');
    });

    it('shows error when not authenticated', async () => {
        vi.mocked(accountService.fetchCurrentAccount).mockRejectedValue(
            new Error('Not authenticated'),
        );

        const tokenService = createMockTokenService();
        const httpClient = createMockHttpClient(tokenService);
        render(<Account httpClient={httpClient} tokenService={tokenService} />, { wrapper: createWrapper() });

        await waitFor(
            () => {
                expect(screen.getByTestId('account-error')).toBeInTheDocument();
            },
            { timeout: 3000 },
        );

        expect(screen.getByText('Failed to load account')).toBeInTheDocument();
    });

    it('displays the members table section', async () => {
        vi.mocked(accountService.fetchCurrentAccount).mockResolvedValue(mockAccountData);
        const membersData = [
            { id: 'm1', email: 'a@example.com', display_name: 'Member A', is_active: true, is_verified: true, is_admin: false, created_at: '', updated_at: '' },
            { id: 'm2', email: 'b@example.com', display_name: 'Member B', is_active: false, is_verified: false, is_admin: false, created_at: '', updated_at: '' },
        ];
        mockMembersList = membersData;

        const tokenService = createMockTokenService();
        const httpClient = createMockHttpClient(tokenService);
        render(<Account httpClient={httpClient} tokenService={tokenService} />, { wrapper: createWrapper() });

        await waitFor(() => expect(screen.getByText('Members')).toBeInTheDocument());

        // Check member names from API are rendered
        await waitFor(() => expect(screen.getByText('Member A')).toBeInTheDocument());
        expect(screen.getByText('Member B')).toBeInTheDocument();
        // total/summary displayed (UI shows "Showing ..." text)
        expect(screen.getByText(/Showing/)).toBeInTheDocument();
    });

    it('invites a member and refreshes the list', async () => {
        vi.mocked(accountService.fetchCurrentAccount).mockResolvedValue(mockAccountData);
        const initial = [
            { id: 'm1', email: 'a@example.com', display_name: 'Member A', is_active: true, is_verified: true, is_admin: false, created_at: '', updated_at: '' },
        ];
        const after = [
            ...initial,
            { id: 'm2', email: 'new@example.com', display_name: 'New Member', is_active: true, is_verified: false, is_admin: false, created_at: '', updated_at: '' },
        ];
        mockMembersList = initial;

        vi.mocked(memberService.inviteMember).mockImplementation(async () => {
            mockMembersList = after;
            return after[1] as any;
        });

        const tokenService = createMockTokenService();
        const httpClient = createMockHttpClient(tokenService);
        render(<Account httpClient={httpClient} tokenService={tokenService} />, { wrapper: createWrapper() });

        await waitFor(() => expect(screen.getByText('Member A')).toBeInTheDocument());

        // Open invite dialog
        fireEvent.click(screen.getByRole('button', { name: /invite member/i }));

        const emailInput = await screen.findByLabelText('Email');
        fireEvent.change(emailInput, { target: { value: 'new@example.com' } });

        // Click Invite
        fireEvent.click(screen.getByRole('button', { name: /invite/i }));

        await waitFor(() => {
            expect(memberService.inviteMember).toHaveBeenCalledWith(httpClient, 'http://localhost:8082', 'new@example.com', expect.any(String));
        });

        // After invite, members should be refreshed and new member rendered
        await waitFor(() => expect(screen.getByText('New Member')).toBeInTheDocument());
    });

    it('removes a member after confirmation and refreshes the list', async () => {
        vi.mocked(accountService.fetchCurrentAccount).mockResolvedValue(mockAccountData);
        const initial = [
            { id: 'm1', email: 'a@example.com', display_name: 'Member A', is_active: true, is_verified: true, is_admin: true, created_at: '', updated_at: '' },
            { id: 'm2', email: 'b@example.com', display_name: 'Member B', is_active: true, is_verified: true, is_admin: true, created_at: '', updated_at: '' },
        ];
        const after = [
            { id: 'm2', email: 'b@example.com', display_name: 'Member B', is_active: true, is_verified: true, is_admin: true, created_at: '', updated_at: '' },
        ];
        mockMembersList = initial;

        vi.mocked(memberService.deleteMember).mockImplementation(async () => {
            mockMembersList = after;
            return undefined as any;
        });

        const tokenService = createMockTokenService();
        const httpClient = createMockHttpClient(tokenService);
        render(<Account httpClient={httpClient} tokenService={tokenService} />, { wrapper: createWrapper() });

        await waitFor(() => expect(screen.getByText('Member A')).toBeInTheDocument());

        // Click Actions on Member A row
        const row = screen.getByText('Member A').closest('tr')!;
        fireEvent.click(within(row).getByRole('button', { name: /Actions for Member A/i }));

        // Click Remove on dropdown menu inside that row
        fireEvent.click(within(row).getByText('Remove'));

        // Confirm dialog shown, click Remove confirm
        const confirmBtn = await screen.findByRole('button', { name: /^remove$/i });
        fireEvent.click(confirmBtn);

        await waitFor(() => {
            expect(memberService.deleteMember).toHaveBeenCalledWith(httpClient, 'http://localhost:8082', 'm1');
        });

        // After delete, refreshed list should show only Member B
        await waitFor(() => expect(screen.getByText('Member B')).toBeInTheDocument());
        expect(screen.queryByText('Member A')).not.toBeInTheDocument();
    });

    it('displays remove button for other members when logged in user is admin, but not for self', async () => {
        vi.mocked(accountService.fetchCurrentAccount).mockResolvedValue(mockAccountData);
        currentUserIsAdmin = true;

        const membersData = [
            // Same email as currently logged in user (self) -> should NOT show remove button
            { id: 'm1', email: 'current_user@example.com', display_name: 'Self Admin', is_active: true, is_verified: true, is_admin: true, created_at: '', updated_at: '' },
            // Admin, different email -> should SHOW remove button
            { id: 'm2', email: 'other_admin@example.com', display_name: 'Other Admin', is_active: true, is_verified: true, is_admin: true, created_at: '', updated_at: '' },
            // Non-admin, different email -> should SHOW remove button (since we only care if the logged-in user is admin, not the listed member's role)
            { id: 'm3', email: 'member@example.com', display_name: 'Regular Member', is_active: true, is_verified: true, is_admin: false, created_at: '', updated_at: '' },
        ];
        mockMembersList = membersData;

        const tokenService = createMockTokenService();
        const httpClient = createMockHttpClient(tokenService);
        render(<Account httpClient={httpClient} tokenService={tokenService} />, { wrapper: createWrapper() });

        // Wait for page to load
        await waitFor(() => expect(screen.getByText('Self Admin')).toBeInTheDocument());

        // Other Admin and Regular Member should have an Actions button
        expect(screen.getByRole('button', { name: /Actions for Other Admin/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Actions for Regular Member/i })).toBeInTheDocument();

        // Self Admin should NOT have an Actions button
        expect(screen.queryByRole('button', { name: /Actions for Self Admin/i })).not.toBeInTheDocument();
    });

    it('does not display remove button for any member when logged in user is not admin', async () => {
        vi.mocked(accountService.fetchCurrentAccount).mockResolvedValue(mockAccountData);
        currentUserIsAdmin = false;

        const membersData = [
            { id: 'm1', email: 'current_user@example.com', display_name: 'Self Member', is_active: true, is_verified: true, is_admin: false, created_at: '', updated_at: '' },
            { id: 'm2', email: 'other_admin@example.com', display_name: 'Other Admin', is_active: true, is_verified: true, is_admin: true, created_at: '', updated_at: '' },
            { id: 'm3', email: 'member@example.com', display_name: 'Regular Member', is_active: true, is_verified: true, is_admin: false, created_at: '', updated_at: '' },
        ];
        mockMembersList = membersData;

        const tokenService = createMockTokenService();
        const httpClient = createMockHttpClient(tokenService);
        render(<Account httpClient={httpClient} tokenService={tokenService} />, { wrapper: createWrapper() });

        // Wait for page to load
        await waitFor(() => expect(screen.getByText('Self Member')).toBeInTheDocument());

        // No Actions buttons should be displayed at all
        expect(screen.queryByRole('button', { name: /Actions for Other Admin/i })).not.toBeInTheDocument();
        expect(screen.queryByRole('button', { name: /Actions for Regular Member/i })).not.toBeInTheDocument();
        expect(screen.queryByRole('button', { name: /Actions for Self Member/i })).not.toBeInTheDocument();
    });

    it('allows changing admin privileges for other members', async () => {
        vi.mocked(accountService.fetchCurrentAccount).mockResolvedValue(mockAccountData);
        const membersData = [
            { id: 'm1', email: 'other@example.com', display_name: 'Other User', is_active: true, is_verified: true, is_admin: false, created_at: '', updated_at: '' },
        ];
        mockMembersList = membersData;

        vi.mocked(memberService.updateMember).mockResolvedValue({
            id: 'm1',
            account_id: 'acc1',
            user_id: 'u1',
            is_admin: true,
            created_at: '',
            updated_at: '',
        });

        const tokenService = createMockTokenService();
        const httpClient = createMockHttpClient(tokenService);
        render(<Account httpClient={httpClient} tokenService={tokenService} />, { wrapper: createWrapper() });

        await waitFor(() => expect(screen.getByText('Other User')).toBeInTheDocument());

        const row = screen.getByText('Other User').closest('tr')!;

        // Click Actions dropdown
        fireEvent.click(within(row).getByRole('button', { name: /Actions for Other User/i }));

        // Click Set as admin
        fireEvent.click(within(row).getByText('Set as admin'));

        await waitFor(() => {
            expect(memberService.updateMember).toHaveBeenCalledWith(
                httpClient,
                'http://localhost:8082',
                'm1',
                { is_admin: true }
            );
        });

        expect(toastMock).toHaveBeenCalledWith(expect.objectContaining({
            title: 'Role updated',
            description: 'Other User is now an admin.',
        }));
    });

    it('allows removing admin privileges from other admin members', async () => {
        vi.mocked(accountService.fetchCurrentAccount).mockResolvedValue(mockAccountData);
        const membersData = [
            { id: 'm1', email: 'other_admin@example.com', display_name: 'Other Admin', is_active: true, is_verified: true, is_admin: true, created_at: '', updated_at: '' },
        ];
        mockMembersList = membersData;

        vi.mocked(memberService.updateMember).mockResolvedValue({
            id: 'm1',
            account_id: 'acc1',
            user_id: 'u1',
            is_admin: false,
            created_at: '',
            updated_at: '',
        });

        const tokenService = createMockTokenService();
        const httpClient = createMockHttpClient(tokenService);
        render(<Account httpClient={httpClient} tokenService={tokenService} />, { wrapper: createWrapper() });

        await waitFor(() => expect(screen.getByText('Other Admin')).toBeInTheDocument());

        const row = screen.getByText('Other Admin').closest('tr')!;

        // Click Actions dropdown
        fireEvent.click(within(row).getByRole('button', { name: /Actions for Other Admin/i }));

        // Click Remove as admin
        fireEvent.click(within(row).getByText('Remove as admin'));

        await waitFor(() => {
            expect(memberService.updateMember).toHaveBeenCalledWith(
                httpClient,
                'http://localhost:8082',
                'm1',
                { is_admin: false }
            );
        });

        expect(toastMock).toHaveBeenCalledWith(expect.objectContaining({
            title: 'Role updated',
            description: 'Other Admin is no longer an admin.',
        }));
    });
});
