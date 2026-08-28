import ApiPagination from '@/components/ApiPagination';
import { ConfirmDialog } from '@/components/ConfirmDialog';
import { StatusBadge } from '@/components/StatusBadge';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useToast } from '@/hooks/use-toast';
import { formatDate } from '@/lib/date';
import { saveProjectTabParams } from '@/lib/paramsStore';
import type { HttpClient } from '@/services/core/httpClient';
import { deleteProjectUser, fetchProjectRoles, fetchProjectUsers, requestProjectUserPasswordReset, resendProjectUserVerificationEmail, type Project } from '@/services/projects';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { AlertCircle, CheckCircle2, ChevronDown, ChevronUp, Eye, KeyRound, Loader2, Mail, MoreVertical, Search, Trash2, XCircle } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom';

type UserSortKey = 'display_name' | 'is_active' | 'role_name' | 'created_at';
type UserFilterKey = 'all' | 'active' | 'inactive';
type UserActionType = 'resend_verification' | 'password_reset' | null;

interface Props { httpClient: HttpClient; project?: Project | null }

export default function UsersTab({ httpClient, project }: Props) {
    const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL as string;
    const queryClient = useQueryClient();
    const { toast } = useToast();
    const navigate = useNavigate();
    const [searchParams, setSearchParams] = useSearchParams();
    const location = useLocation();

    const [pageNum, setPageNum] = useState(1);
    const [pageSize, setPageSize] = useState<number>(10);
    const [searchInput, setSearchInput] = useState('');
    const [search, setSearch] = useState('');
    const [isActiveFilter, setIsActiveFilter] = useState<UserFilterKey>('all');
    const [roleIdFilter, setRoleIdFilter] = useState<string>('');
    const [sortBy, setSortBy] = useState<UserSortKey>('created_at');
    const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc');

    // Helper to handle sort toggle
    const handleSort = (key: UserSortKey) => {
        if (sortBy !== key) {
            setSortBy(key);
            setSortDir('asc');
        } else {
            setSortDir(sortDir === 'asc' ? 'desc' : 'asc');
        }
        setPageNum(1);
    };
    const [removeTarget, setRemoveTarget] = useState<any | null>(null);
    const [actionState, setActionState] = useState<{ type: UserActionType; user: any | null }>({ type: null, user: null });
    const [selectedRedirectUri, setSelectedRedirectUri] = useState<string>('none');

    const handleOpenAction = (type: UserActionType, user: any) => {
        setActionState({ type, user });
        setSelectedRedirectUri('none');
    };

    useEffect(() => { const t = setTimeout(() => setSearch(searchInput), 300); return () => clearTimeout(t); }, [searchInput]);

    // initialize from URL params
    useEffect(() => {
        const p = Number(searchParams.get('page_num') ?? pageNum);
        const ps = Number(searchParams.get('page_size') ?? pageSize);
        const s = searchParams.get('search') ?? '';
        const ia = (searchParams.get('is_active') ?? isActiveFilter) as UserFilterKey;
        const rid = searchParams.get('role_id') ?? '';
        const sb = (searchParams.get('sort_by') as UserSortKey) ?? 'created_at';
        const sd = (searchParams.get('sort_dir') as 'asc' | 'desc') ?? 'desc';
        setPageNum(Number.isNaN(p) ? 1 : p);
        setPageSize(Number.isNaN(ps) ? 10 : ps);
        setSearchInput(s);
        setIsActiveFilter(ia as UserFilterKey);
        setRoleIdFilter(rid);
        setSortBy(sb);
        setSortDir(sd);
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    // sync URL when filters change
    useEffect(() => {
        const next = new URLSearchParams(searchParams.toString());
        next.set('page_num', String(pageNum));
        next.set('page_size', String(pageSize));
        if (search) next.set('search', String(search)); else next.delete('search');
        next.set('is_active', String(isActiveFilter));
        if (roleIdFilter) next.set('role_id', String(roleIdFilter)); else next.delete('role_id');
        next.set('sort_by', String(sortBy));
        next.set('sort_dir', String(sortDir));
        setSearchParams(next, { replace: true, state: location.state });
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [pageNum, pageSize, search, isActiveFilter, roleIdFilter, sortBy, sortDir]);

    const isActiveParam = isActiveFilter === 'all' ? null : (isActiveFilter === 'active');

    // Fetch roles for the filter dropdown
    const rolesQuery = useQuery({
        queryKey: ['project', project?.id, 'roles', 1, 100],
        queryFn: () => fetchProjectRoles(httpClient, AUTH_BASE_URL, project!.id, 1, 100, 'name', 'asc'),
        enabled: !!project?.id,
    });

    const usersQuery = useQuery({
        queryKey: ['project', project?.id, 'users', pageNum, pageSize, search, isActiveFilter, roleIdFilter, sortBy, sortDir],
        queryFn: () => fetchProjectUsers(httpClient, AUTH_BASE_URL, project!.id, pageNum, pageSize, sortBy, sortDir, search, isActiveParam, roleIdFilter),
        enabled: !!project?.id,
    });

    const roles = rolesQuery.data?.data ?? [];

    const deleteMutation = useMutation({
        mutationFn: (uid: string) => deleteProjectUser(httpClient, AUTH_BASE_URL, project!.id, uid),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['project', project?.id, 'users'] });
            toast({ title: 'User removed', description: 'User has been removed', variant: 'default' });
            setRemoveTarget(null);
        },
        onError: (err: any) => {
            toast({ title: 'Remove failed', description: err?.message ?? 'Unable to remove user', variant: 'destructive' });
        },
    });

    const resendVerificationMutation = useMutation({
        mutationFn: ({ userId, redirectUri }: { userId: string; redirectUri?: string }) =>
            resendProjectUserVerificationEmail(httpClient, AUTH_BASE_URL, project!.id, userId, redirectUri ? { redirect_uri: redirectUri } : undefined),
        onSuccess: () => {
            toast({ title: 'Verification email sent', description: 'Verification email has been sent successfully.', variant: 'default' });
            setActionState({ type: null, user: null });
        },
        onError: (err: any) => {
            toast({ title: 'Resend verification failed', description: err?.message ?? 'Unable to send verification email', variant: 'destructive' });
        },
    });

    const requestPasswordResetMutation = useMutation({
        mutationFn: ({ userId, redirectUri }: { userId: string; redirectUri?: string }) =>
            requestProjectUserPasswordReset(httpClient, AUTH_BASE_URL, project!.id, userId, redirectUri ? { redirect_uri: redirectUri } : undefined),
        onSuccess: () => {
            toast({ title: 'Password reset email sent', description: 'Password reset email has been sent successfully.', variant: 'default' });
            setActionState({ type: null, user: null });
        },
        onError: (err: any) => {
            toast({ title: 'Password reset failed', description: err?.message ?? 'Unable to send password reset email', variant: 'destructive' });
        },
    });

    const handleConfirmAction = () => {
        if (!actionState.user || !project?.id) return;
        const redirectUri = selectedRedirectUri === 'none' ? undefined : selectedRedirectUri;
        if (actionState.type === 'resend_verification') {
            resendVerificationMutation.mutate({ userId: actionState.user.id, redirectUri });
        } else if (actionState.type === 'password_reset') {
            requestPasswordResetMutation.mutate({ userId: actionState.user.id, redirectUri });
        }
    };

    if (usersQuery.isLoading) {
        return (
            <div className="flex items-center">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                <span className="ml-2 text-sm text-muted-foreground">Loading…</span>
            </div>
        );
    }

    if (usersQuery.isError) {
        return (
            <section className="border rounded p-4 bg-card w-full">
                <div className="flex flex-col items-center justify-center py-20 space-y-2" data-testid="users-error">
                    <AlertCircle className="h-8 w-8 text-destructive" />
                    <p className="text-sm text-destructive font-medium">Failed to load users</p>
                    <p className="text-xs text-muted-foreground">{(usersQuery.error as any)?.message ?? 'Users data is unavailable'}</p>
                </div>
            </section>
        );
    }

    const users = usersQuery.data?.data ?? [] as any[];
    const pagination = usersQuery.data?.pagination ?? null;

    return (
        <section className="border rounded p-4 bg-card w-full">
            <h3 className="text-lg font-medium">Users</h3>
            <div className="text-sm text-muted-foreground mb-4">Manage the users and their roles in this project</div>

            <div className="mb-3">
                <div className="flex flex-wrap items-center gap-3">
                    <div className="relative w-full sm:max-w-xs">
                        <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                        <Input id="users-search" role="searchbox" className="pl-8 pr-10" placeholder="Search users..." value={searchInput} onChange={(e) => { setSearchInput(e.target.value); setPageNum(1); }} />
                    </div>

                    <Select value={isActiveFilter} onValueChange={(v) => { setIsActiveFilter(v as UserFilterKey); setPageNum(1); }}>
                        <SelectTrigger className="w-[160px]">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">All</SelectItem>
                            <SelectItem value="active">Active</SelectItem>
                            <SelectItem value="inactive">Inactive</SelectItem>
                        </SelectContent>
                    </Select>

                    <Select value={roleIdFilter || 'all-roles'} onValueChange={(v) => { setRoleIdFilter(v === 'all-roles' ? '' : v); setPageNum(1); }}>
                        <SelectTrigger className="w-[180px]">
                            <SelectValue placeholder="All" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all-roles">All</SelectItem>
                            {roles.map((role: any) => (
                                <SelectItem key={role.id} value={role.id}>{role.name}</SelectItem>
                            ))}
                        </SelectContent>
                    </Select>

                </div>
            </div>

            <div className="flex items-center justify-between mb-3">
                <div className="text-sm text-muted-foreground">
                    {pagination ? (
                        pagination.page_size === pagination.total_data
                            ? `Showing ${pagination.page_size} user${pagination.page_size !== 1 ? 's' : ''}`
                            : `Showing ${pagination.page_size} of ${pagination.total_data} users`
                    ) : null}
                </div>

                <div>
                    <Button
                        size="sm"
                        onClick={() => { if (project?.id) saveProjectTabParams(project.id, 'users', location.search ?? ''); navigate(`/projects/${project?.id}/users/new`, { state: { from: location.search } }); }}
                    >
                        Add User
                    </Button>
                </div>
            </div>

            <div className="mt-2 border rounded bg-card shadow-sm overflow-hidden">
                <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="border-b bg-muted/40">
                                <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                                    <button onClick={() => handleSort('display_name')} className="flex items-center gap-2">
                                        <span>Name</span>
                                        {sortBy === 'display_name' ? (sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                                    </button>
                                </th>
                                <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                                    <button onClick={() => handleSort('is_active')} className="flex items-center gap-2">
                                        <span>Status</span>
                                        {sortBy === 'is_active' ? (sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                                    </button>
                                </th>
                                <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                                    <button onClick={() => handleSort('role_name')} className="flex items-center gap-2">
                                        <span>Role</span>
                                        {sortBy === 'role_name' ? (sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                                    </button>
                                </th>
                                <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                                    <button onClick={() => handleSort('created_at')} className="flex items-center gap-2">
                                        <span>Created</span>
                                        {sortBy === 'created_at' ? (sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                                    </button>
                                </th>
                                <th className="w-10 px-4 py-2.5"> </th>
                            </tr>
                        </thead>
                        <tbody>
                            {usersQuery.isFetching ? (
                                <tr>
                                    <td colSpan={5} className="px-4 py-3">
                                        <div className="flex items-center gap-2">
                                            <Loader2 className="h-4 w-4 animate-spin" />
                                            <span>Loading users…</span>
                                        </div>
                                    </td>
                                </tr>
                            ) : users.length === 0 ? (
                                <tr>
                                    <td colSpan={5} className="px-4 py-3 text-sm text-muted-foreground">No users found</td>
                                </tr>
                            ) : (
                                users.map((user: any) => (
                                    <tr key={user.id} className="border-b last:border-0 hover:bg-muted/20 transition-colors">
                                        <td className="px-4 py-3 align-top">
                                            <div className="flex items-center gap-3">
                                                {user.avatar_url ? (
                                                    <img src={user.avatar_url} alt={user.display_name} className="h-8 w-8 rounded object-cover border border-border" />
                                                ) : (
                                                    <div className="h-8 w-8 rounded bg-muted flex items-center justify-center text-xs font-bold text-muted-foreground">{(user.display_name ?? '')?.charAt(0).toUpperCase()}</div>
                                                )}
                                                <div>
                                                    <button
                                                        onClick={() => { if (project?.id) saveProjectTabParams(project.id, 'users', location.search ?? ''); navigate(`/projects/${project?.id}/users/${user.id}/details`, { state: { from: location.search } }); }}
                                                        className="text-sm font-medium truncate text-left block"
                                                        aria-label={`Open ${user.display_name} details`}
                                                    >
                                                        {user.display_name}
                                                    </button>
                                                    <div className="text-xs text-muted-foreground mt-1">
                                                        {user.email}
                                                        {user.email ? (
                                                            user.is_email_verified ? (
                                                                <span className="inline-flex items-center gap-1 text-success text-xs ml-2">
                                                                    <CheckCircle2 className="h-3.5 w-3.5" /> Verified
                                                                </span>
                                                            ) : (
                                                                <span className="inline-flex items-center gap-1 text-muted-foreground text-xs ml-2">
                                                                    <XCircle className="h-3.5 w-3.5" /> Not Verified
                                                                </span>
                                                            )
                                                        ) : null}
                                                    </div>
                                                </div>
                                            </div>
                                        </td>
                                        <td className="px-4 py-3 align-center">
                                            <StatusBadge active={user.is_active} />
                                        </td>
                                        <td className="px-4 py-3 align-center">
                                            {user.role_name ? (
                                                <Badge variant="secondary" className="text-xs">{user.role_name}</Badge>
                                            ) : (
                                                <span className="text-xs text-muted-foreground">—</span>
                                            )}
                                        </td>
                                        <td className="px-4 py-3 text-xs text-muted-foreground">
                                            {formatDate(user.created_at)}
                                        </td>
                                        <td className="px-4 py-3">
                                            <div className="flex items-center justify-end">
                                                <DropdownMenu>
                                                    <DropdownMenuTrigger asChild>
                                                        <Button size="sm" variant="ghost" className="h-7 w-7 p-0">
                                                            <MoreVertical className="h-4 w-4" />
                                                        </Button>
                                                    </DropdownMenuTrigger>
                                                    <DropdownMenuContent align="end">
                                                        <DropdownMenuItem onClick={() => { if (project?.id) { saveProjectTabParams(project.id, 'users', location.search ?? ''); navigate(`/projects/${project?.id}/users/${user.id}/details`, { state: { from: location.search } }); } }}>
                                                            <Eye className="h-4 w-4 mr-2" />
                                                            View Details
                                                        </DropdownMenuItem>
                                                        {!user.is_email_verified && user.is_active && (
                                                            <DropdownMenuItem onClick={() => handleOpenAction('resend_verification', user)}>
                                                                <Mail className="h-4 w-4 mr-2" />
                                                                Resend Verification Email
                                                            </DropdownMenuItem>
                                                        )}
                                                        {user.is_active && (
                                                            <DropdownMenuItem onClick={() => handleOpenAction('password_reset', user)}>
                                                                <KeyRound className="h-4 w-4 mr-2" />
                                                                Request Password Reset
                                                            </DropdownMenuItem>
                                                        )}
                                                        <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => setRemoveTarget(user)}>
                                                            <Trash2 className="h-4 w-4 mr-2" />
                                                            Remove
                                                        </DropdownMenuItem>
                                                    </DropdownMenuContent>
                                                </DropdownMenu>
                                            </div>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>
            </div>

            <div className="flex items-center justify-between mt-3">
                <div className="flex items-center gap-3">
                    <Select value={String(pageSize)} onValueChange={(v) => { setPageSize(Number(v)); setPageNum(1); }}>
                        <SelectTrigger className="w-[96px]">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            {[10, 25, 50].map((s) => <SelectItem key={s} value={String(s)}>{String(s)}</SelectItem>)}
                        </SelectContent>
                    </Select>
                </div>
                <div>
                    <ApiPagination
                        pagination={pagination}
                        pageSize={pageSize}
                        onPageChange={setPageNum}
                    />
                </div>
            </div>

            <ConfirmDialog
                open={!!removeTarget}
                onOpenChange={(open) => { if (!open) setRemoveTarget(null); }}
                title="Remove User"
                description={`Are you sure you want to remove ${removeTarget?.display_name}? This action cannot be undone.`}
                confirmLabel={deleteMutation.isPending ? 'Removing...' : 'Remove'}
                destructive
                disabled={deleteMutation.isPending}
                onConfirm={() => {
                    if (removeTarget?.id) {
                        deleteMutation.mutate(removeTarget.id);
                    }
                }}
            />

            <Dialog open={!!actionState.type} onOpenChange={(open) => { if (!open) setActionState({ type: null, user: null }); }}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>
                            {actionState.type === 'resend_verification' ? 'Resend Verification Email' : 'Request Password Reset'}
                        </DialogTitle>
                        <DialogDescription>
                            {actionState.type === 'resend_verification'
                                ? `Send email verification link to ${actionState.user?.display_name || 'user'}.`
                                : `Send password reset link to ${actionState.user?.display_name || 'user'}.`}
                        </DialogDescription>
                    </DialogHeader>

                    <div className="space-y-4 py-2">
                        <div className="text-sm text-muted-foreground bg-muted/50 p-3 rounded border">
                            Select your project's login URL from the list to be displayed to the user after completing the action.
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="action-redirect-uri-select">Login URL (Optional)</Label>
                            <Select value={selectedRedirectUri} onValueChange={setSelectedRedirectUri}>
                                <SelectTrigger id="action-redirect-uri-select" className="w-full">
                                    <SelectValue placeholder="Select login URL..." />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="none">
                                        <span className="text-muted-foreground">None</span>
                                        </SelectItem>
                                    {(project?.redirect_uris ?? []).map((item, idx) => (
                                        <SelectItem key={`${item.redirect_uri || 'uri'}-${idx}`} value={item.redirect_uri}>
                                            {item.login_url && item.login_url.trim() ? item.login_url : item.redirect_uri}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                    </div>

                    <DialogFooter>
                        <Button
                            variant="outline"
                            onClick={() => setActionState({ type: null, user: null })}
                            disabled={resendVerificationMutation.isPending || requestPasswordResetMutation.isPending}
                        >
                            Cancel
                        </Button>
                        <Button
                            onClick={handleConfirmAction}
                            disabled={resendVerificationMutation.isPending || requestPasswordResetMutation.isPending}
                        >
                            {(resendVerificationMutation.isPending || requestPasswordResetMutation.isPending) ? (
                                <>
                                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                                    Sending...
                                </>
                            ) : (
                                actionState.type === 'resend_verification' ? 'Send Verification Email' : 'Request Password Reset'
                            )}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </section>
    );
}
