import ApiPagination from '@/components/ApiPagination';
import { ConfirmDialog } from '@/components/ConfirmDialog';
import { StatusBadge } from '@/components/StatusBadge';
import { Badge } from "@/components/ui/badge";
import { Button } from '@/components/ui/button';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { useProject, useProjectClients } from '@/hooks/projects';
import { useToast } from '@/hooks/use-toast';
import { formatDate } from '@/lib/date';
import logger from '@/lib/logger';
import { saveProjectTabParams } from '@/lib/paramsStore';
import {
    PROJECT_VALIDATION,
    truncateToMaxLength,
    validateProjectForm,
    type ProjectValidationResult
} from '@/lib/validation';
import type { HttpClient } from '@/services/core/httpClient';
import { deleteClient, updateProject, uploadProjectLogo, type Project } from '@/services/projects';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ChevronDown, ChevronUp, CloudUpload, Eye, Loader2, MoreVertical, Search, Trash2, Copy, Check } from 'lucide-react';
import React, { useEffect, useMemo, useState } from 'react';
import { useLocation, useNavigate, useParams, useSearchParams } from 'react-router-dom';

interface Props { httpClient: HttpClient, project?: Project }

export default function ProjectTab({ httpClient, project: projectProp }: Props) {
    const { project_id } = useParams<{ project_id: string }>();

    const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL as string;

    const { project: projectFromHook, isLoading: projectLoadingFromHook } = useProject({ httpClient, baseUrl: AUTH_BASE_URL, projectId: project_id });
    const projectEffective = (projectProp ?? projectFromHook) as Project | undefined;
    const projectLoading = projectProp ? false : projectLoadingFromHook;

    const [pageNum, setPageNum] = useState(1);
    const [pageSize, setPageSize] = useState<number>(5);
    const [searchInput, setSearchInput] = useState('');
    const [search, setSearch] = useState('');
    const [isActive, setIsActive] = useState<string | null>(null);
    const [sortBy, setSortBy] = useState<string | null>('created_at');
    const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc');

    const [searchParams, setSearchParams] = useSearchParams();
    const location = useLocation();

    // initialize from URL params when present
    useEffect(() => {
        const p = Number(searchParams.get('page_num') ?? pageNum);
        const ps = Number(searchParams.get('page_size') ?? pageSize);
        const s = searchParams.get('search') ?? '';
        const ia = searchParams.get('is_active');
        const sb = searchParams.get('sort_by') ?? sortBy ?? undefined;
        const sd = (searchParams.get('sort_dir') as any) ?? sortDir;
        setSearchInput(s);
        setIsActive(ia === 'all' ? null : ia ?? null);
        setSortBy(sb ?? 'created_at');
        setSortDir(sd === 'asc' ? 'asc' : 'desc');
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    // sync state to URL params so filters are reflected
    useEffect(() => {
        const next = new URLSearchParams(searchParams.toString());
        next.set('page_num', String(pageNum));
        next.set('page_size', String(pageSize));
        if (search) next.set('search', search); else next.delete('search');
        next.set('is_active', isActive === null ? 'all' : isActive);
        next.set('sort_by', String(sortBy ?? ''));
        next.set('sort_dir', sortDir);
        setSearchParams(next, { replace: true, state: location.state });
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [pageNum, pageSize, search, isActive, sortBy, sortDir]);

    useEffect(() => { const t = setTimeout(() => setSearch(searchInput), 300); return () => clearTimeout(t); }, [searchInput]);

    const { clients, pagination, isLoading: clientsLoading } = useProjectClients({ httpClient, baseUrl: AUTH_BASE_URL, projectId: projectEffective?.id, pageNum, pageSize, search, isActive, sortBy, sortDir });

    const queryClient = useQueryClient();
    const [uploadingLogo, setUploadingLogo] = useState(false);
    const [showUploadDialog, setShowUploadDialog] = useState(false);
    const [selectedLogoFile, setSelectedLogoFile] = useState<File | null>(null);
    const { toast } = useToast();
    const [copiedId, setCopiedId] = useState<string | null>(null);

    const handleCopy = async (id: string) => {
        try {
            await navigator.clipboard.writeText(id);
            setCopiedId(id);
            toast({
                title: 'Copied',
                description: 'Client ID copied to clipboard',
            });
            setTimeout(() => setCopiedId(null), 1500);
        } catch (err) {
            toast({
                title: 'Copy failed',
                description: 'Failed to copy client ID to clipboard',
                variant: 'destructive',
            });
        }
    };
    const inputUploadRef = React.useRef<HTMLInputElement | null>(null);
    const [logoUrl, setLogoUrl] = useState<string | null>(null);
    const [removeClient, setRemoveClient] = useState<any | null>(null);
    const [removeClientLoading, setRemoveClientLoading] = useState(false);

    // Validation state
    const [nameTouched, setNameTouched] = useState(false);
    const [descriptionTouched, setDescriptionTouched] = useState(false);
    const [redirectUrisTouched, setRedirectUrisTouched] = useState(false);

    const navigate = useNavigate();

    // form state
    const [name, setName] = useState('');
    const [description, setDescription] = useState('');
    const [redirectUris, setRedirectUris] = useState<Array<{ redirect_uri: string; login_url?: string }>>([]);
    const [status, setStatus] = useState<'active' | 'inactive'>('active');

    useEffect(() => {
        if (projectEffective) {
            setName(projectEffective.name ?? '');
            setLogoUrl(projectEffective.logo_url ?? null);
            setDescription(projectEffective.description ?? '');
            // normalize redirect_uris: prefer object[] but handle legacy string[] explicitly
            const raw = projectEffective.redirect_uris ?? [];
            const ruris: Array<{ redirect_uri: string; login_url?: string }> = [];
            if (Array.isArray(raw)) {
                // if the server still returns array of strings (legacy), map them
                if (raw.length > 0 && typeof raw[0] === 'string') {
                    (raw as any[]).forEach((s) => ruris.push({ redirect_uri: s ?? '', login_url: '' }));
                } else {
                    (raw as any[]).forEach((it) => {
                        if (!it) {
                            ruris.push({ redirect_uri: '' });
                        } else if (typeof it === 'string') {
                            ruris.push({ redirect_uri: it, login_url: '' });
                        } else {
                            ruris.push({ redirect_uri: it.redirect_uri ?? '', login_url: it.login_url ?? '' });
                        }
                    });
                }
            }
            setRedirectUris(ruris.length > 0 ? ruris : []);
            setStatus(projectEffective.is_active ? 'active' : 'inactive');
        }
    }, [projectEffective]);

    // Intentionally do not notify parent about in-form edits. Header should
    // remain showing server-side values until an update completes and the
    // page is refreshed or re-fetched.

    const updateMutation = useMutation<Project, Error, Partial<any>, unknown>({
        mutationFn: (payload: Partial<any>) => {
            if (!projectEffective) throw new Error('Project not loaded');
            return updateProject(httpClient, AUTH_BASE_URL, projectEffective.id, payload);
        },
        onSuccess: (data) => {
            queryClient.setQueryData(['project', projectEffective?.id], data);
            queryClient.invalidateQueries({ queryKey: ['project'] });
            toast({ title: 'Saved', description: 'Project updated', variant: 'default' });
        },
        onError: (err: any) => {
            toast({ title: 'Save failed', description: err?.message ?? 'Unable to save project', variant: 'destructive' });
        },
    });

    // Validation logic using shared utilities
    const validation = useMemo<ProjectValidationResult>(
        () => validateProjectForm(name, description, redirectUris),
        [name, description, redirectUris]
    );

    const doSave = async () => {
        if (!projectEffective || !validation.isValid) return;
        const payload: any = {
            name,
            description,
            logo_url: logoUrl,
            redirect_uris: validation.validRedirectUris,
            is_active: status === 'active',
        };
        await updateMutation.mutateAsync(payload);
    };

    // Handle name change with validation
    const handleNameChange = (value: string) => {
        setName(truncateToMaxLength(value, PROJECT_VALIDATION.name.maxLength));
    };

    // Handle description change with validation
    const handleDescriptionChange = (value: string) => {
        setDescription(truncateToMaxLength(value, PROJECT_VALIDATION.description.maxLength));
    };

    const onLogoFile = async (file?: File) => {
        if (!projectEffective || !file) return;
        setUploadingLogo(true);
        try {
            const res = await uploadProjectLogo(httpClient, AUTH_BASE_URL, projectEffective.id, file);
            setLogoUrl(res.logo_url ?? null);
        } catch (err: any) {
            logger.error('Upload failed:', err);
            toast({ title: 'Upload failed', description: err?.message ?? 'Unable to upload logo', variant: 'destructive' });
        } finally {
            setUploadingLogo(false);
        }
    };

    const doUploadSelectedLogo = async () => {
        if (!selectedLogoFile) return;
        await onLogoFile(selectedLogoFile);
        setShowUploadDialog(false);
        setSelectedLogoFile(null);
    };

    const onRemoveLogo = async () => {
        if (!projectEffective) return;
        setLogoUrl(null);
    };

    const confirmRemoveClient = async () => {
        if (!removeClient) return;
        setRemoveClientLoading(true);
        try {
            if (!projectEffective) throw new Error('Project not loaded');
            await deleteClient(httpClient, AUTH_BASE_URL, projectEffective.id, removeClient.id);
            queryClient.invalidateQueries({ queryKey: ['project', projectEffective?.id, 'clients'] });
            toast({ title: 'Client removed', description: `${removeClient.name} has been removed`, variant: 'default' });
            setRemoveClient(null);
        } catch (err) {
            const msg = (err as any)?.message ?? 'Unable to remove client';
            toast({ title: 'Remove failed', description: msg, variant: 'destructive' });
        } finally {
            setRemoveClientLoading(false);
        }
    };

    if (projectLoading) {
        return (
            <div className="flex items-center">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                <span className="ml-2 text-sm text-muted-foreground">Loading…</span>
            </div>
        );
    }

    return (
        <div className="space-y-4">
            <section className="border rounded p-4 bg-card w-full">
                <h3 className="text-lg font-medium">Project Information</h3>
                <div className="text-sm text-muted-foreground">View or edit the project information</div>
                <div className="mt-4">
                    <div className="grid grid-cols-12 gap-4 items-center mb-4">
                        <div className="col-span-4 flex items-center">
                            <Label>Logo</Label>
                        </div>
                        <div className="col-span-8 flex items-center justify-start gap-3">
                            <div className="relative">
                                {logoUrl ? (
                                    <img src={logoUrl} alt={projectEffective?.name} className="h-20 w-20 rounded object-cover" />
                                ) : (
                                    <div className="h-20 w-20 rounded bg-muted flex items-center justify-center text-lg font-bold text-muted-foreground">{projectEffective?.name?.charAt(0) ?? ''}</div>
                                )}
                            </div>
                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <Button size="sm" variant="ghost" className="h-7 w-7 p-0">
                                        <MoreVertical className="h-4 w-4" />
                                    </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent align="end">
                                    <DropdownMenuItem onClick={() => { setSelectedLogoFile(null); setShowUploadDialog(true); }}>
                                        <CloudUpload className="h-4 w-4 mr-2" />
                                        Upload
                                    </DropdownMenuItem>
                                    <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => onRemoveLogo()}
                                        disabled={!logoUrl}>
                                        <Trash2 className="h-4 w-4 mr-2" />
                                        Remove
                                    </DropdownMenuItem>
                                </DropdownMenuContent>
                            </DropdownMenu>
                        </div>
                    </div>
                </div>

                <ConfirmDialog
                    open={showUploadDialog}
                    onOpenChange={(open) => { if (!open) { setShowUploadDialog(false); setSelectedLogoFile(null); } }}
                    closeOnConfirm={false}
                    title="Upload Project Logo"
                    description="Select an image file to use as the project logo."
                    confirmLabel="Upload"
                    onConfirm={doUploadSelectedLogo}
                    disabled={!selectedLogoFile || uploadingLogo}
                >
                    <div className="py-4">
                        {uploadingLogo ? (
                            <div className="flex items-center gap-3 justify-center py-6">
                                <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
                                <div className="text-sm">Uploading...</div>
                            </div>
                        ) : (
                            <div
                                onDragOver={(e) => e.preventDefault()}
                                onDrop={(e) => { e.preventDefault(); const f = e.dataTransfer?.files?.[0]; if (f) setSelectedLogoFile(f); }}
                                className="border-2 border-dashed rounded p-6 text-center"
                            >
                                {selectedLogoFile ? (
                                    <div className="space-y-2">
                                        <div className="font-medium">{selectedLogoFile.name}</div>
                                        <div className="text-sm text-muted-foreground">{Math.round(selectedLogoFile.size / 1024)} KB</div>
                                    </div>
                                ) : (
                                    <div className="text-sm text-muted-foreground">Drag & drop an image here, or click to select</div>
                                )}
                                <input id="logo-upload-input" ref={(el) => (inputUploadRef.current = el)} type="file" accept="image/*" className="hidden" onChange={(e) => setSelectedLogoFile(e.target.files?.[0] ?? null)} />
                                <div className="block mt-3">
                                    <Button size="sm" onClick={() => inputUploadRef.current?.click()}>Select File</Button>
                                </div>
                                <div className="mt-3 text-xs text-muted-foreground">
                                    Accepted formats: JPG, PNG, GIF
                                </div>
                            </div>
                        )}
                    </div>
                </ConfirmDialog>
                <div className="grid grid-cols-12 gap-4 w-full">
                    <div className="col-span-4 flex items-center">
                        <Label htmlFor="proj-name">Name</Label>
                    </div>
                    <div className="col-span-8">
                        <Input
                            id="proj-name"
                            value={name}
                            onChange={(e) => {
                                handleNameChange(e.target.value);
                                if (!nameTouched) setNameTouched(true);
                            }}
                            onBlur={() => setNameTouched(true)}
                            placeholder="Enter project name"
                            className={nameTouched && validation.errors.name ? 'border-destructive' : ''}
                            aria-invalid={nameTouched && !!validation.errors.name}
                            maxLength={PROJECT_VALIDATION.name.maxLength}
                        />
                        {nameTouched && validation.errors.name && (
                            <div className="text-xs text-destructive mt-1">{validation.errors.name}</div>
                        )}
                    </div>

                    <div className="col-span-4 flex items-start pt-1">
                        <Label htmlFor="proj-desc">Description</Label>
                    </div>
                    <div className="col-span-8">
                        <Textarea
                            id="proj-desc"
                            rows={3}
                            value={description}
                            onChange={(e) => {
                                handleDescriptionChange(e.target.value);
                                if (!descriptionTouched) setDescriptionTouched(true);
                            }}
                            onBlur={() => setDescriptionTouched(true)}
                            placeholder="Enter project description"
                            className={descriptionTouched && validation.errors.description ? 'border-destructive' : ''}
                            aria-invalid={descriptionTouched && !!validation.errors.description}
                            maxLength={PROJECT_VALIDATION.description.maxLength}
                        />
                        {descriptionTouched && validation.errors.description && (
                            <div className="text-xs text-destructive mt-1">{validation.errors.description}</div>
                        )}
                    </div>

                    <div className="col-span-4 flex items-start pt-1">
                        <div>
                            <Label>Redirect URIs</Label>
                            <div className="text-xs text-muted-foreground mt-1">
                                Configure allowed redirect URIs and their corresponding login pages.
                                Each redirect URI can have a login URL used for initiating authentication.
                            </div>
                        </div>
                    </div>
                    <div className="col-span-8 space-y-2">
                        {redirectUris.map((r, idx) => {
                            const rowError = validation.rowErrors?.[idx];
                            return (
                                <div key={idx} className="grid grid-cols-12 gap-2 items-start">
                                    <div className="col-span-6">
                                        <Input
                                            placeholder="Redirect URI"
                                            value={r.redirect_uri}
                                            onChange={(e) => {
                                                const newArr = [...redirectUris];
                                                newArr[idx] = { ...newArr[idx], redirect_uri: e.target.value };
                                                setRedirectUris(newArr);
                                                if (!redirectUrisTouched) setRedirectUrisTouched(true);
                                            }}
                                            onBlur={() => setRedirectUrisTouched(true)}
                                            className={rowError?.redirect_uri || (redirectUrisTouched && idx === 0 && !r.redirect_uri.trim()) ? 'border-destructive' : ''}
                                            aria-invalid={!!rowError?.redirect_uri}
                                            maxLength={PROJECT_VALIDATION.redirectUri.maxLength}
                                        />
                                        {rowError?.redirect_uri && (
                                            <div className="text-xs text-destructive mt-1">{rowError.redirect_uri}</div>
                                        )}
                                    </div>
                                    <div className="col-span-6 flex gap-2 items-start">
                                        <div className="flex-1">
                                            <Input
                                                placeholder="Login URL (optional)"
                                                value={r.login_url ?? ''}
                                                onChange={(e) => {
                                                    const newArr = [...redirectUris];
                                                    newArr[idx] = { ...newArr[idx], login_url: e.target.value };
                                                    setRedirectUris(newArr);
                                                }}
                                                className={rowError?.login_url ? 'border-destructive' : ''}
                                                aria-invalid={!!rowError?.login_url}
                                                maxLength={PROJECT_VALIDATION.loginUrl.maxLength}
                                            />
                                            {rowError?.login_url && (
                                                <div className="text-xs text-destructive mt-1">{rowError.login_url}</div>
                                            )}
                                        </div>
                                        <Button
                                            variant="outline"
                                            className="text-destructive bg-white hover:text-destructive border-destructive/30 hover:bg-destructive/10 shrink-0"
                                            onClick={() => {
                                                const newArr = [...redirectUris];
                                                newArr.splice(idx, 1);
                                                setRedirectUris(newArr);
                                            }}
                                            aria-label={`Delete redirect ${idx}`}
                                        >
                                            <Trash2 />
                                        </Button>
                                    </div>
                                </div>
                            );
                        })}
                        {redirectUrisTouched && validation.errors.redirectUris && (
                            <div className="text-xs text-destructive">{validation.errors.redirectUris}</div>
                        )}
                        <div>
                            <Button size="sm" onClick={() => setRedirectUris([...redirectUris, { redirect_uri: '', login_url: '' }])}>
                                Add Redirect URI
                            </Button>
                        </div>
                    </div>

                    <div className="col-span-4 flex items-center">
                        <Label>Status</Label>
                    </div>
                    <div className="col-span-8">
                        <Select value={status} onValueChange={(v) => setStatus(v as 'active' | 'inactive')}>
                            <SelectTrigger className="w-[160px]">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="active">Active</SelectItem>
                                <SelectItem value="inactive">Inactive</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>

                    <div className="col-span-8">
                        {projectEffective && (
                            <div className="text-xs text-muted-foreground flex flex-col gap-1">
                                <div>Created: {formatDate(projectEffective.created_at)}</div>
                                <div>Last Updated: {formatDate(projectEffective.updated_at)}</div>
                            </div>
                        )}
                    </div>
                    <div className="col-span-4 flex justify-end">
                        <Button
                            size="sm"
                            onClick={() => {
                                // Mark all fields as touched to show validation errors
                                setNameTouched(true);
                                setDescriptionTouched(true);
                                setRedirectUrisTouched(true);
                                if (validation.isValid) {
                                    doSave();
                                }
                            }}
                            disabled={updateMutation.status === 'pending' || !validation.isValid}
                        >
                            {updateMutation.status === 'pending' ? (
                                <>
                                    <Loader2 className="h-4 w-4 animate-spin mr-2" />
                                    Saving...
                                </>
                            ) : (
                                <>
                                    Save
                                </>
                            )}
                        </Button>
                    </div>
                </div>
            </section>
            <section className="border rounded p-4 bg-card w-full">
                <div className="mb-3">
                    <div className="flex items-start justify-between">
                        <div>
                            <h3 className="text-lg font-medium">Clients</h3>
                            <div className="text-sm text-muted-foreground">Manage the clients of this project. Open the details to view the client credentials and permissions.</div>
                        </div>
                    </div>

                    {/* Filters toolbar under title/subtitle - mirror members UI */}
                    <div className="flex flex-col sm:flex-row gap-3 items-start sm:items-center mt-4">
                        <div className="relative flex-1 w-full sm:max-w-xs">
                            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                            <Input
                                placeholder="Search clients…"
                                value={searchInput}
                                onChange={(e) => { setSearchInput(e.target.value); setPageNum(1); }}
                                className="pl-8"
                                aria-label="Search clients"
                            />
                        </div>

                        <Select value={isActive ?? 'all'} onValueChange={(v) => { setIsActive(v === 'all' ? null : v); setPageNum(1); }}>
                            <SelectTrigger className="w-[160px]" aria-label="Filter by active">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="all">All</SelectItem>
                                <SelectItem value="true">Active</SelectItem>
                                <SelectItem value="false">Inactive</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </div>

                <div className="flex items-center justify-between mb-3">
                    <div className="text-sm text-muted-foreground">
                        {pagination ? (
                            pagination.page_size === pagination.total_data
                                ? `Showing ${pagination.page_size} clients`
                                : `Showing ${pagination.page_size} of ${pagination.total_data} clients`
                        ) : null}
                    </div>
                    <Button size="sm" onClick={() => { if (projectEffective?.id) saveProjectTabParams(projectEffective.id, 'project', location.search ?? ''); navigate(`/projects/${projectEffective?.id}/clients/new`, { state: { from: location.search } }); }}>
                        Add Client
                    </Button>
                </div>

                <div className="mt-2 border rounded bg-card shadow-sm overflow-hidden">
                    <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                            <thead>
                                <tr className="border-b bg-muted/40">
                                    <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                                        <button onClick={() => {
                                            const key = 'name';
                                            if (sortBy !== key) { setSortBy(key); setSortDir('asc'); } else { setSortDir(sortDir === 'asc' ? 'desc' : 'asc'); }
                                            setPageNum(1);
                                        }} className="flex items-center gap-2">
                                            <span>Name</span>
                                            {sortBy === 'name' ? (sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                                        </button>
                                    </th>
                                    <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                                        <button onClick={() => {
                                            const key = 'id';
                                            if (sortBy !== key) { setSortBy(key); setSortDir('asc'); } else { setSortDir(sortDir === 'asc' ? 'desc' : 'asc'); }
                                            setPageNum(1);
                                        }} className="flex items-center gap-2">
                                            <span>Client ID</span>
                                            {sortBy === 'id' ? (sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                                        </button>
                                    </th>
                                    <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                                        <button onClick={() => {
                                            const key = 'is_active';
                                            if (sortBy !== key) { setSortBy(key); setSortDir('asc'); } else { setSortDir(sortDir === 'asc' ? 'desc' : 'asc'); }
                                            setPageNum(1);
                                        }} className="flex items-center gap-2">
                                            <span>Status</span>
                                            {sortBy === 'is_active' ? (sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                                        </button>
                                    </th>
                                    <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                                        <button onClick={() => {
                                            const key = 'created_at';
                                            if (sortBy !== key) { setSortBy(key); setSortDir('asc'); } else { setSortDir(sortDir === 'asc' ? 'desc' : 'asc'); }
                                            setPageNum(1);
                                        }} className="flex items-center gap-2">
                                            <span>Created</span>
                                            {sortBy === 'created_at' ? (sortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                                        </button>
                                    </th>
                                    <th className="w-10 px-4 py-2.5"> </th>
                                </tr>
                            </thead>
                            <tbody>
                                {clientsLoading ? (
                                    <tr>
                                        <td colSpan={5} className="px-4 py-3">
                                            <div className="flex items-center gap-2">
                                                <Loader2 className="h-4 w-4 animate-spin" />
                                                <span>Loading clients…</span>
                                            </div>
                                        </td>
                                    </tr>
                                ) : clients.length === 0 ? (
                                    <tr>
                                        <td colSpan={5} className="px-4 py-3 text-sm text-muted-foreground">No clients found</td>
                                    </tr>
                                ) : (
                                    clients.map((c: any) => (
                                        <tr key={c.id} className="border-b last:border-0 hover:bg-muted/20 transition-colors">
                                            <td className="px-4 py-3">
                                                <div className="">
                                                    <div className="flex items-center gap-2">
                                                        <button
                                                            onClick={() => {
                                                                if (projectEffective?.id) saveProjectTabParams(projectEffective.id, 'project', location.search ?? '');
                                                                navigate(`/projects/${projectEffective?.id}/clients/${c.id}/details`, { state: { from: location.search } });
                                                            }}
                                                            className="text-sm font-medium truncate text-left"
                                                            aria-label={`Open ${c.name} details`}
                                                        >
                                                            {c.name}
                                                        </button>
                                                        {c.is_default && <Badge variant="secondary" className="text-xs capitalize">Default</Badge>}
                                                    </div>
                                                    <div className="text-xs text-muted-foreground mt-1">{c.description ?? ''}</div>
                                                </div>
                                            </td>
                                            <td className="px-4 py-3">
                                                <div className="flex items-center gap-1.5">
                                                    <span className="inline-flex items-center px-2 py-0.5 font-mono rounded bg-muted text-muted-foreground text-xs">{c.id}</span>
                                                    <Button
                                                        size="sm"
                                                        variant="ghost"
                                                        className="h-6 w-6 p-0 hover:bg-muted shrink-0"
                                                        onClick={() => handleCopy(c.id)}
                                                        aria-label={`Copy client ID for ${c.name}`}
                                                    >
                                                        {copiedId === c.id ? (
                                                            <Check className="h-3.5 w-3.5 text-green-500 animate-in fade-in zoom-in duration-200" />
                                                        ) : (
                                                            <Copy className="h-3.5 w-3.5 text-muted-foreground hover:text-foreground" />
                                                        )}
                                                    </Button>
                                                </div>
                                            </td>
                                            <td className="px-4 py-3"><StatusBadge active={c.is_active} /></td>
                                            <td className="px-4 py-3 text-xs text-muted-foreground">{formatDate(c.created_at)}</td>
                                            <td className="px-4 py-3">
                                                <div className="flex items-center">
                                                    <DropdownMenu>
                                                        <DropdownMenuTrigger asChild>
                                                            <Button size="sm" variant="ghost" className="h-7 w-7 p-0">
                                                                <MoreVertical className="h-4 w-4" />
                                                            </Button>
                                                        </DropdownMenuTrigger>
                                                        <DropdownMenuContent align="end">
                                                            <DropdownMenuItem onClick={() => { if (projectEffective?.id) saveProjectTabParams(projectEffective.id, 'project', location.search ?? ''); navigate(`/projects/${projectEffective?.id}/clients/${c.id}/details`, { state: { from: location.search } }); }}>
                                                                <Eye className="h-4 w-4 mr-2" />
                                                                View Details
                                                            </DropdownMenuItem>
                                                            {!c.is_default && (
                                                                <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => setRemoveClient(c)}>
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
                                {[5, 25, 50].map((s) => <SelectItem key={s} value={String(s)}>{String(s)}</SelectItem>)}
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
            </section>
            <ConfirmDialog
                open={!!removeClient}
                onOpenChange={(open) => { if (!open) setRemoveClient(null); }}
                closeOnConfirm={false}
                title="Remove Client"
                description={<span>Are you sure you want to remove <span className="font-medium">{removeClient?.name}</span>? This action cannot be undone.</span>}
                confirmLabel={removeClientLoading ? (<><Loader2 className="h-4 w-4 animate-spin mr-2" />Removing...</>) : 'Remove'}
                destructive
                disabled={removeClientLoading}
                onConfirm={confirmRemoveClient}
            />
        </div>
    );
}
