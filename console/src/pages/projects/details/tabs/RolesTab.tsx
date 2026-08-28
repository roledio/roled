import ApiPagination from '@/components/ApiPagination';
import { ConfirmDialog } from '@/components/ConfirmDialog';
import { Button } from '@/components/ui/button';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useToast } from '@/hooks/use-toast';
import { formatDate } from '@/lib/date';
import { saveProjectTabParams } from '@/lib/paramsStore';
import type { HttpClient } from '@/services/core/httpClient';
import { deleteProjectRole, fetchProjectRoles, setSignupRole, type Project } from '@/services/projects';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { AlertCircle, ChevronDown, ChevronUp, Eye, Loader2, MoreVertical, Search, Trash2, UserCheck } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom';

interface Props { httpClient: HttpClient; project?: Project | null }

export default function RolesTab({ httpClient, project }: Props) {
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
    const [removeTarget, setRemoveTarget] = useState<any | null>(null);
    const [sortBy, setSortBy] = useState<'name' | 'code' | 'created_at'>('created_at');
    const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc');

    // Helper to handle sort toggle
    const handleSort = (key: 'name' | 'code' | 'created_at') => {
        if (sortBy !== key) {
            setSortBy(key);
            setSortDir('asc');
        } else {
            setSortDir(sortDir === 'asc' ? 'desc' : 'asc');
        }
        setPageNum(1);
    };

    useEffect(() => { const t = setTimeout(() => setSearch(searchInput), 300); return () => clearTimeout(t); }, [searchInput]);

    // initialize from URL params
    useEffect(() => {
        const p = Number(searchParams.get('page_num') ?? pageNum);
        const ps = Number(searchParams.get('page_size') ?? pageSize);
        const s = searchParams.get('search') ?? '';
        const sb = (searchParams.get('sort_by') as 'name' | 'code' | 'created_at') ?? 'created_at';
        const sd = (searchParams.get('sort_dir') as 'asc' | 'desc') ?? 'desc';
        setPageNum(Number.isNaN(p) ? 1 : p);
        setPageSize(Number.isNaN(ps) ? 10 : ps);
        setSearchInput(s);
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
        next.set('sort_by', String(sortBy));
        next.set('sort_dir', String(sortDir));
        setSearchParams(next, { replace: true, state: location.state });
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [pageNum, pageSize, search, sortBy, sortDir]);

    const rolesQuery = useQuery({
        queryKey: ['project', project?.id, 'roles', pageNum, pageSize, search, sortBy, sortDir],
        queryFn: () => fetchProjectRoles(httpClient, AUTH_BASE_URL, project!.id, pageNum, pageSize, sortBy, sortDir, search),
        enabled: !!project?.id,
    });

    const deleteMutation = useMutation({
        mutationFn: (rid: string) => deleteProjectRole(httpClient, AUTH_BASE_URL, project!.id, rid),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['project', project?.id, 'roles'] });
            toast({ title: 'Role removed', description: 'Role has been removed', variant: 'default' });
            setRemoveTarget(null);
        },
        onError: (err: any) => {
            toast({ title: 'Remove failed', description: err?.message ?? 'Unable to remove role', variant: 'destructive' });
        },
    });

    const setSignupRoleMutation = useMutation({
        mutationFn: (rid: string) => setSignupRole(httpClient, AUTH_BASE_URL, project!.id, rid),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['project', project?.id, 'roles'] });
            toast({ title: 'Role updated', description: 'Successfully set as sign-up role', variant: 'default' });
        },
        onError: (err: any) => {
            toast({ title: 'Update failed', description: err?.message ?? 'Unable to set sign-up role', variant: 'destructive' });
        },
    });

    if (rolesQuery.isLoading) {
        return (
            <div className="flex items-center">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                <span className="ml-2 text-sm text-muted-foreground">Loading…</span>
            </div>
        );
    }

    if (rolesQuery.isError) {
        return (
            <section className="border rounded p-4 bg-card w-full">
                <div className="flex flex-col items-center justify-center py-20 space-y-2" data-testid="roles-error">
                    <AlertCircle className="h-8 w-8 text-destructive" />
                    <p className="text-sm text-destructive font-medium">Failed to load roles</p>
                    <p className="text-xs text-muted-foreground">{(rolesQuery.error as any)?.message ?? 'Roles data is unavailable'}</p>
                </div>
            </section>
        );
    }

    const roles = rolesQuery.data?.data ?? [] as any[];
    const pagination = rolesQuery.data?.pagination ?? null;

    return (
        <section className="border rounded p-4 bg-card w-full">
            <h3 className="text-lg font-medium">Roles</h3>
            <div className="text-sm text-muted-foreground mb-4">Manage the roles and permissions of this project</div>

            <div className="mb-3">
                <div className="relative w-full sm:max-w-xs">
                    <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                    <Input id="roles-search" role="searchbox" className="pl-8 pr-10" placeholder="Search roles..." value={searchInput} onChange={(e) => { setSearchInput(e.target.value); setPageNum(1); }} />
                </div>
            </div>

            <div className="flex items-center justify-between mb-3">
                <div className="text-sm text-muted-foreground">
                    {pagination ? (
                        pagination.page_size === pagination.total_data
                            ? `Showing ${pagination.page_size} role${pagination.page_size !== 1 ? 's' : ''}`
                            : `Showing ${pagination.page_size} of ${pagination.total_data} roles`
                    ) : null}
                </div>

                <div>
                    <Button size="sm" onClick={() => { if (project?.id) { saveProjectTabParams(project.id, 'roles', location.search ?? ''); navigate(`/projects/${project?.id}/roles/new`, { state: { from: location.search } }); } }}>
                        Add Role
                    </Button>
                </div>
            </div>

            <div className="mt-2 border rounded bg-card shadow-sm overflow-hidden">
                <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="border-b bg-muted/40">
                                <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                                    <button onClick={() => handleSort('name')} className="flex items-center gap-2">
                                        <span>Name</span>
                                        {sortBy === 'name' ? (sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                                    </button>
                                </th>
                                <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                                    <button onClick={() => handleSort('code')} className="flex items-center gap-2">
                                        <span>Code</span>
                                        {sortBy === 'code' ? (sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                                    </button>
                                </th>
                                <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                                    <button onClick={() => handleSort('created_at')} className="flex items-center gap-2">
                                        <span>Created</span>
                                        {sortBy === 'created_at' ? (sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                                    </button>
                                </th>
                                <th className="px-4 py-2.5"> </th>
                            </tr>
                        </thead>
                        <tbody>
                            {rolesQuery.isFetching ? (
                                <tr>
                                    <td colSpan={4} className="px-4 py-3">
                                        <div className="flex items-center gap-2">
                                            <Loader2 className="h-4 w-4 animate-spin" />
                                            <span>Loading roles…</span>
                                        </div>
                                    </td>
                                </tr>
                            ) : roles.length === 0 ? (
                                <tr>
                                    <td colSpan={4} className="px-4 py-3 text-sm text-muted-foreground">No roles found</td>
                                </tr>
                            ) : (
                                roles.map((role: any) => (
                                    <tr key={role.id} className="border-b last:border-0 hover:bg-muted/20 transition-colors">
                                        <td className="px-4 py-3 align-top">
                                            <div className="font-medium text-left flex items-center gap-2">
                                                <button
                                                    onClick={() => { if (project?.id) saveProjectTabParams(project.id, 'roles', location.search ?? ''); navigate(`/projects/${project?.id}/roles/${role.id}/details`, { state: { from: location.search } }); }}
                                                    className="text-sm font-medium truncate text-left"
                                                    aria-label={`Open ${role.name} details`}
                                                >
                                                    {role.name}
                                                </button>
                                                {role.is_default_signup ? (
                                                    <span className="inline-flex items-center p-1 rounded bg-success/10 text-success text-xs">
                                                        <UserCheck className="h-3.5 w-3.5" />
                                                    </span>
                                                ) : null}
                                            </div>
                                            <div className="text-xs text-muted-foreground mt-1">{role.description ?? ''}</div>
                                        </td>
                                        <td className="px-4 py-3 align-center">
                                            <span className="inline-flex items-center px-2 py-0.5 font-mono rounded bg-muted text-muted-foreground text-xs">{role.code}</span>
                                        </td>
                                        <td className="px-4 py-3 text-xs text-muted-foreground">
                                            {formatDate(role.created_at)}
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
                                                        <DropdownMenuItem onClick={() => { if (project?.id) saveProjectTabParams(project.id, 'roles', location.search ?? ''); navigate(`/projects/${project?.id}/roles/${role.id}/details`, { state: { from: location.search } }); }}>
                                                            <Eye className="h-4 w-4 mr-2" />
                                                            View Details
                                                        </DropdownMenuItem>
                                                        {!role.is_default_signup && (
                                                            <DropdownMenuItem onClick={() => setSignupRoleMutation.mutate(role.id)} disabled={setSignupRoleMutation.isPending}>
                                                                <UserCheck className="h-4 w-4 mr-2" />
                                                                Set as sign-up role
                                                            </DropdownMenuItem>
                                                        )}
                                                        <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => setRemoveTarget(role)}>
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

            {/* Remove Confirmation Dialog */}
            <ConfirmDialog
                open={!!removeTarget}
                onOpenChange={(open) => {
                    if (!open) {
                        setRemoveTarget(null);
                    }
                }}
                title="Remove Role"
                description={`Are you sure you want to remove the role "${removeTarget?.name}"? This action cannot be undone.`}
                confirmLabel={deleteMutation.isPending ? 'Removing...' : 'Remove'}
                destructive
                disabled={deleteMutation.isPending}
                onConfirm={() => {
                    if (removeTarget?.id) {
                        deleteMutation.mutate(removeTarget.id);
                    }
                }}
            />
        </section>
    );
}
