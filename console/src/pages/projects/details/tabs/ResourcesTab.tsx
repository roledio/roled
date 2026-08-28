import ApiPagination from '@/components/ApiPagination';
import { ConfirmDialog } from '@/components/ConfirmDialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useToast } from '@/hooks/use-toast';
import { saveProjectTabParams } from '@/lib/paramsStore';
import type { HttpClient } from '@/services/core/httpClient';
import { deleteProjectResource, fetchProjectResources, type Project } from '@/services/projects';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { AlertCircle, ChevronDown, ChevronUp, Eye, Loader2, MoreVertical, Search, Trash2 } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom';

interface Props { httpClient: HttpClient; project?: Project | null }

export default function ResourcesTab({ httpClient, project }: Props) {
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
    const [isDefaultFilter, setIsDefaultFilter] = useState<'all' | 'default' | 'custom'>('all');
    const [sortBy, setSortBy] = useState<string>('name');
    const [sortDir, setSortDir] = useState<'asc' | 'desc'>('asc');
    const [removeTarget, setRemoveTarget] = useState<any | null>(null);

    // Helper to handle sort toggle
    const handleSort = () => {
        if (sortDir === 'asc') {
            setSortDir('desc');
        } else {
            setSortDir('asc');
        }
        setPageNum(1);
    };

    useEffect(() => { const t = setTimeout(() => setSearch(searchInput), 300); return () => clearTimeout(t); }, [searchInput]);

    // initialize from URL params
    useEffect(() => {
        const p = Number(searchParams.get('page_num') ?? pageNum);
        const ps = Number(searchParams.get('page_size') ?? pageSize);
        const s = searchParams.get('search') ?? '';
        const df = (searchParams.get('is_default') ?? isDefaultFilter) as any;
        const sb = searchParams.get('sort_by') ?? 'name';
        const sd = (searchParams.get('sort_dir') as 'asc' | 'desc') ?? 'asc';
        setPageNum(Number.isNaN(p) ? 1 : p);
        setPageSize(Number.isNaN(ps) ? 10 : ps);
        setSearchInput(s);
        setIsDefaultFilter(df as any);
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
        next.set('is_default', String(isDefaultFilter));
        next.set('sort_by', String(sortBy));
        next.set('sort_dir', String(sortDir));
        setSearchParams(next, { replace: true, state: location.state });
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [pageNum, pageSize, search, isDefaultFilter, sortBy, sortDir]);

    const isDefaultParam = isDefaultFilter === 'all' ? null : (isDefaultFilter === 'default');

    const resourcesQuery = useQuery({
        queryKey: ['project', project?.id, 'resources', pageNum, pageSize, search, isDefaultFilter, sortBy, sortDir],
        queryFn: () => fetchProjectResources(httpClient, AUTH_BASE_URL, project!.id, pageNum, pageSize, sortBy, sortDir, search, isDefaultParam),
        enabled: !!project?.id,
    });

    const deleteMutation = useMutation({
        mutationFn: (rid: string) => deleteProjectResource(httpClient, AUTH_BASE_URL, project!.id, rid),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['project', project?.id, 'resources'] });
            toast({ title: 'Resource removed', description: 'Resource has been removed', variant: 'default' });
        },
        onError: (err: any) => {
            toast({ title: 'Remove failed', description: err?.message ?? 'Unable to remove resource', variant: 'destructive' });
        },
    });

    if (resourcesQuery.isLoading) {
        return (
            <div className="flex items-center">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                <span className="ml-2 text-sm text-muted-foreground">Loading…</span>
            </div>
        );
    }

    if (resourcesQuery.isError) {
        return (
            <section className="border rounded p-4 bg-card w-full">
                <div className="flex flex-col items-center justify-center py-20 space-y-2" data-testid="resources-error">
                    <AlertCircle className="h-8 w-8 text-destructive" />
                    <p className="text-sm text-destructive font-medium">Failed to load resources</p>
                    <p className="text-xs text-muted-foreground">{(resourcesQuery.error as any)?.message ?? 'Resources data is unavailable'}</p>
                </div>
            </section>
        );
    }

    const resources = resourcesQuery.data?.data ?? [] as any[];
    const pagination = resourcesQuery.data?.pagination ?? null;

    return (
        <section className="border rounded p-4 bg-card w-full">
            <h3 className="text-lg font-medium">Resources</h3>
            <div className="text-sm text-muted-foreground mb-4">Manage the resources and permissions of this project</div>

            <div className="mb-3">
                <div className="flex items-center gap-3">
                    <div className="relative w-full sm:max-w-xs">
                        <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                        <Input id="resources-search" role="searchbox" className="pl-8 pr-10" placeholder="Search resources..." value={searchInput} onChange={(e) => { setSearchInput(e.target.value); setPageNum(1); }} />
                    </div>

                    <Select value={isDefaultFilter} onValueChange={(v) => { setIsDefaultFilter(v as any); setPageNum(1); }}>
                        <SelectTrigger className="w-[180px]">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">All</SelectItem>
                            <SelectItem value="default">Default Resource</SelectItem>
                            <SelectItem value="custom">Custom Resource</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
            </div>

            <div className="flex items-center justify-between mb-3">
                <div className="text-sm text-muted-foreground">
                    {pagination ? (
                        pagination.page_size === pagination.total_data
                            ? `Showing ${pagination.page_size} resources`
                            : `Showing ${pagination.page_size} of ${pagination.total_data} resources`
                    ) : null}
                </div>

                <div>
                    <Button size="sm" onClick={() => { if (project?.id) saveProjectTabParams(project.id, 'resources', location.search ?? ''); navigate(`/projects/${project?.id}/resources/new`, { state: { from: location.search } }); }}>
                        Add Resource
                    </Button>
                </div>
            </div>

            <div className="mt-2 border rounded bg-card shadow-sm overflow-hidden">
                <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="border-b bg-muted/40">
                                <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                                    <button onClick={handleSort} className="flex items-center gap-2">
                                        <span>Resource</span>
                                        {sortBy === 'name' ? (sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                                    </button>
                                </th>
                                <th className="text-left font-medium text-muted-foreground px-4 py-2.5">Permissions</th>
                                <th className="w-10 px-4 py-2.5"> </th>
                            </tr>
                        </thead>
                        <tbody>
                            {resourcesQuery.isFetching ? (
                                <tr>
                                    <td colSpan={3} className="px-4 py-3">
                                        <div className="flex items-center gap-2">
                                            <Loader2 className="h-4 w-4 animate-spin" />
                                            <span>Loading resources…</span>
                                        </div>
                                    </td>
                                </tr>
                            ) : resources.length === 0 ? (
                                <tr>
                                    <td colSpan={3} className="px-4 py-3 text-sm text-muted-foreground">No resources found</td>
                                </tr>
                            ) : (
                                resources.map((r: any) => (
                                    <tr key={r.id} className="border-b last:border-0 hover:bg-muted/20 transition-colors">
                                        <td className="w-48 px-4 py-3 align-top">
                                            <div className="font-medium flex items-center gap-2">
                                                <button
                                                    onClick={() => { if (project?.id) saveProjectTabParams(project.id, 'resources', location.search ?? ''); navigate(`/projects/${project?.id}/resources/${r.id}/details`, { state: { from: location.search } }); }}
                                                    className="text-sm font-medium truncate text-left"
                                                    aria-label={`Open ${r.name} details`}
                                                >
                                                    {r.name}
                                                </button>
                                                {r.is_default && <Badge variant="secondary" className="text-xs">Default</Badge>}
                                            </div>
                                            <div className="text-xs text-muted-foreground mt-1">{r.description ?? ''}</div>
                                        </td>
                                        <td className="px-4 py-3 align-top">
                                            <div className="flex flex-wrap gap-2">
                                                {(r.permissions ?? []).map((p: any) => (
                                                    <Card key={p.id} className="min-w-[140px] max-w-xs">
                                                        <CardContent className='p-3'>
                                                            <div className="text-sm font-medium">{p.name}</div>
                                                            <div className="text-xs text-muted-foreground mt-1">{p.description ?? ''}</div>
                                                        </CardContent>
                                                    </Card>
                                                ))}
                                            </div>
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
                                                        <DropdownMenuItem onClick={() => { if (project?.id) saveProjectTabParams(project.id, 'resources', location.search ?? ''); navigate(`/projects/${project?.id}/resources/${r.id}/details`, { state: { from: location.search } }); }}>
                                                            <Eye className="h-4 w-4 mr-2" />
                                                            View Details
                                                        </DropdownMenuItem>
                                                        {!r.is_default && (
                                                            <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => setRemoveTarget(r)}>
                                                                <Trash2 className="h-4 w-4 mr-2" />
                                                                Remove
                                                            </DropdownMenuItem>
                                                        )}
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
                closeOnConfirm={false}
                title="Remove Resource"
                description={<span>Are you sure you want to remove <span className="font-medium">{removeTarget?.name}</span>? This action cannot be undone.</span>}
                confirmLabel={deleteMutation.status === 'pending' ? (<><Loader2 className="h-4 w-4 animate-spin mr-2" />Removing...</>) : 'Remove'}
                destructive
                disabled={deleteMutation.status === 'pending'}
                onConfirm={async () => {
                    if (!removeTarget) return;
                    await deleteMutation.mutateAsync(removeTarget.id);
                    setRemoveTarget(null);
                }}
            />
        </section>
    );
}
