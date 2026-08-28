import ApiPagination from '@/components/ApiPagination';
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { StatusBadge } from "@/components/StatusBadge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader } from '@/components/ui/card';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useToast } from '@/hooks/use-toast';
import { formatDate } from '@/lib/date';
import { saveProjectsParams } from '@/lib/paramsStore';
import type { HttpClient } from '@/services/core/httpClient';
import { deleteProject, fetchProjects, type Project as ServiceProject } from '@/services/projects';
import { useMutation, useQuery, useQueryClient, type UseQueryResult } from '@tanstack/react-query';
import { ArrowUpDown, Blocks, Clock, Eye, Loader2, MoreVertical, Plus, Search, Trash2 } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom';

type SortKey = "name" | "created_at";

type FilterKey = "all" | "active" | "inactive";

interface ProjectsProps {
  httpClient: HttpClient;
}

export default function Projects({ httpClient }: ProjectsProps) {
  const navigate = useNavigate();

  const [search, setSearch] = useState("");
  const [searchInput, setSearchInput] = useState("");
  const [filter, setFilter] = useState<FilterKey>("all");
  const [sortBy, setSortBy] = useState<SortKey>("created_at");
  const [removeTarget, setRemoveTarget] = useState<ServiceProject | null>(null);
  const [confirmName, setConfirmName] = useState('');
  const [pageNum, setPageNum] = useState<number>(1);
  const pageSize = 15;
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc');
  const { toast } = useToast();

  const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL as string;

  const queryClient = useQueryClient();

  const [searchParams, setSearchParams] = useSearchParams();
  const location = useLocation();

  // initialize from URL params
  useEffect(() => {
    const p = Number(searchParams.get('page_num') ?? pageNum);
    const s = searchParams.get('search') ?? '';
    const ia = searchParams.get('is_active') ?? 'all';
    const sb = (searchParams.get('sort_by') as SortKey) ?? 'created_at';
    const sd = (searchParams.get('sort_dir') as 'asc' | 'desc') ?? 'desc';
    setPageNum(Number.isNaN(p) ? 1 : p);
    setSearchInput(s);
    setSearch(s);
    setFilter((ia as FilterKey) ?? 'all');
    setSortBy(sb);
    setSortDir(sd);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // sync URL when filters change
  useEffect(() => {
    const next = new URLSearchParams(searchParams.toString());
    next.set('page_num', String(pageNum));
    if (search) next.set('search', search); else next.delete('search');
    next.set('is_active', filter);
    next.set('sort_by', sortBy);
    next.set('sort_dir', sortDir);
    setSearchParams(next, { replace: true, state: location.state });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pageNum, search, filter, sortBy, sortDir]);

  // debounce updating the actual `search` used by queries + URL
  useEffect(() => {
    const t = setTimeout(() => setSearch(searchInput), 300);
    return () => clearTimeout(t);
  }, [searchInput]);

  type ProjectsQueryResult = { data: ServiceProject[]; pagination?: { page_num: number; page_size: number; total_data: number } };
  type ProjectsQueryKey = (string | { search: string; filter: FilterKey; pageNum: number; pageSize: number; sortBy: SortKey; sortDir: 'asc' | 'desc' })[];

  const projectsQuery = useQuery({
    queryKey: ['projects', { search, filter, pageNum, pageSize, sortBy, sortDir }] as ProjectsQueryKey,
    queryFn: async () => {
      const params: Record<string, any> = { search: search || undefined, is_active: filter === 'all' ? null : (filter === 'active'), page_size: pageSize, page_num: pageNum };
      if (sortBy) {
        params.sort_by = sortBy;
        params.sort_dir = sortDir;
      }
      return fetchProjects(httpClient, AUTH_BASE_URL, params);
    },
    keepPreviousData: true,
  } as any) as UseQueryResult<ProjectsQueryResult, Error>;

  const serverPagination = projectsQuery.data?.pagination ?? null;

  const deleteMutation = useMutation({
    mutationFn: ({ projectId, name }: { projectId: string; name: string }) =>
      deleteProject(httpClient, AUTH_BASE_URL, projectId, { name }),
    onMutate: () => { },
    onSuccess: (_data, vars) => {
      // refresh projects list from server
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      setRemoveTarget(null);
      setConfirmName('');
      toast({ title: 'Project removed', description: 'Project was successfully removed.' });
    },
    onError: (err: any) => {
      const msg = err?.message ?? 'Failed to delete project';
      toast({ title: 'Failed to remove project', description: String(msg), variant: 'destructive' });
    },
  });

  const sortLabel = (() => {
    if (sortBy === 'name') return sortDir === 'asc' ? 'Name (A → Z)' : 'Name (Z → A)';
    return sortDir === 'asc' ? 'Created (Oldest)' : 'Created (Newest)';
  })();

  const filtered = useMemo(() => {
    // server returns already-filtered/sorted data based on query params
    return projectsQuery.data?.data ?? [];
  }, [projectsQuery.data]);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-semibold">Projects</h1>
        <p className="text-sm text-muted-foreground mt-1">Manage your projects</p>
      </div>

      {/* Toolbar */}
      <div className="flex flex-col sm:flex-row gap-3 items-start sm:items-center">
        <div className="relative flex-1 w-full sm:max-w-xs">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search projects…"
            value={searchInput}
            onChange={(e) => { setSearchInput(e.target.value); setPageNum(1); }}
            className="pl-8"
            aria-label="Search projects"
          />
        </div>

        <Select value={filter} onValueChange={(v) => setFilter(v as FilterKey)}>
          <SelectTrigger className="w-[140px]" aria-label="Filter by status">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All</SelectItem>
            <SelectItem value="active">Active</SelectItem>
            <SelectItem value="inactive">Inactive</SelectItem>
          </SelectContent>
        </Select>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" size="default" aria-label="Sort projects">
              <ArrowUpDown className="h-4 w-4 mr-1.5" />
              {sortLabel}
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => { setSortBy('name'); setSortDir('asc'); }}>
              Name (A → Z)
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => { setSortBy('name'); setSortDir('desc'); }}>
              Name (Z → A)
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => { setSortBy('created_at'); setSortDir('asc'); }}>
              Created (Oldest)
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => { setSortBy('created_at'); setSortDir('desc'); }}>
              Created (Newest)
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {/* Summary */}
      <div className="text-sm text-muted-foreground mb-2">
        {serverPagination ? (
          serverPagination.page_size === serverPagination.total_data
            ? `Showing ${serverPagination.page_size} projects`
            : `Showing ${serverPagination.page_size} of ${serverPagination.total_data} projects`
        ) : null}
      </div>

      {/* Cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 items-stretch">
        {!projectsQuery.isLoading && !projectsQuery.isError && (
          <Card
            onClick={() => {
              saveProjectsParams(location.search ?? '');
              navigate('/projects/new', { state: { from: location.search } });
            }}
            className="h-full min-h-[160px] flex flex-col items-center justify-center border-dashed border-2 cursor-pointer hover:border-primary hover:bg-accent/50 transition-all duration-200 group"
            aria-label="Add project card"
          >
            <div className="flex flex-col items-center gap-3">
              <div className="p-3 rounded-full bg-muted group-hover:bg-primary/10 group-hover:text-primary transition-all duration-200 shadow-inner">
                <Plus className="h-6 w-6 text-muted-foreground group-hover:text-primary transition-colors" />
              </div>
              <span className="text-sm font-semibold text-muted-foreground group-hover:text-primary transition-colors">
                Add Project
              </span>
            </div>
          </Card>
        )}
        {projectsQuery.isLoading ? (
          <div className="flex items-center">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            <span className="ml-2 text-sm text-muted-foreground">Loading projects…</span>
          </div>
        ) : projectsQuery.isError ? (
          <div className="col-span-full text-center py-6 text-destructive">{String(projectsQuery.error?.message ?? 'Failed to load projects')}</div>
        ) : filtered.map((project) => (
          <Card key={project.id} className="h-full flex flex-col shadow-sm hover:shadow transition-shadow">
            <CardHeader className="flex flex-row items-center justify-between gap-2 space-y-0">
              <div className="flex items-center gap-2.5 min-w-0">
                {project.logo_url ? (
                  <img src={project.logo_url} alt={project.name} className="h-9 w-9 rounded object-cover shrink-0" />
                ) : (
                  <div className="h-9 w-9 rounded bg-muted flex items-center justify-center text-xs font-bold text-muted-foreground shrink-0">
                    {project.name.charAt(0)}
                  </div>
                )}
                <div className="min-w-0">
                  <button
                    onClick={() => {
                      const params = new URLSearchParams();
                      params.set('tab', 'project');
                      // persist current projects page params so Back can restore them
                      saveProjectsParams(location.search ?? '');
                      navigate(`/projects/${project.id}/details?${params.toString()}`, { state: { from: location.search } });
                    }}
                    className="text-sm font-medium truncate text-left w-full"
                    aria-label={`Open ${project.name} details`}
                  >
                    {project.name}
                  </button>
                  <StatusBadge active={project.is_active} className="mt-0.5" />
                </div>
              </div>

              <DropdownMenu>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <DropdownMenuTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 shrink-0"
                        aria-label={`Actions for ${project.name}`}
                      >
                        <MoreVertical className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                  </TooltipTrigger>
                  <TooltipContent>Actions</TooltipContent>
                </Tooltip>

                <DropdownMenuContent align="end">
                  <DropdownMenuItem onClick={() => {
                    const params = new URLSearchParams();
                    params.set('tab', 'project');
                    // persist current projects page params so Back can restore them
                    saveProjectsParams(location.search ?? '');
                    navigate(`/projects/${project.id}/details?${params.toString()}`, { state: { from: location.search } });
                  }}>
                    <Eye className="h-4 w-4 mr-2" />
                    View Details
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    className="text-destructive focus:text-destructive"
                    onClick={() => setRemoveTarget(project as any)}
                  >
                    <Trash2 className="h-4 w-4 mr-2" />
                    Remove
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </CardHeader>

            <CardContent className="flex-1">
              <p className="text-xs text-muted-foreground line-clamp-2 leading-relaxed">
                {project.description}
              </p>
            </CardContent>

            <CardFooter className="justify-between">
              <div className="text-xs text-muted-foreground flex items-center gap-2">
                <Clock className="h-4 w-4" />
                <span>{project.created_at ? formatDate(project.created_at) : ''}</span>
              </div>
            </CardFooter>
          </Card>
        ))}
      </div>

      <div className="flex justify-center">
        <ApiPagination
          pagination={serverPagination}
          pageSize={pageSize}
          onPageChange={setPageNum}
        />
      </div>

      <ConfirmDialog
        open={!!removeTarget}
        onOpenChange={(open) => !open && (setConfirmName(''), setRemoveTarget(null))}
        title="Remove Project"
        description={<span>Are you sure you want to remove <span className="font-medium">{removeTarget?.name}</span>? This action cannot be undone.</span>}
        confirmLabel="Remove"
        destructive
        onConfirm={() => {
          if (!removeTarget) return;
          deleteMutation.mutate({ projectId: removeTarget.id, name: confirmName });
        }}
        disabled={deleteMutation.status === 'pending' || !(removeTarget && confirmName.trim().toLowerCase() === (removeTarget.name || '').trim().toLowerCase())}
      >
        <div className="mt-2">
          <p className="text-sm text-muted-foreground mb-2">Please type "<span className="font-medium">{removeTarget?.name}</span>" to confirm deletion</p>
          <Input
            aria-label="confirm-project-name"
            placeholder="Project name"
            value={confirmName}
            onChange={(e) => setConfirmName(e.target.value)}
          />
        </div>
      </ConfirmDialog>
    </div>
  );
}
