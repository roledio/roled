import ApiPagination from '@/components/ApiPagination';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useToast } from '@/hooks/use-toast';
import { getProjectTabParams } from '@/lib/paramsStore';
import { slugify } from '@/lib/slugify';
import {
    ROLE_VALIDATION,
    truncateToMaxLength,
    validateRoleForm,
    type RoleValidationResult,
} from '@/lib/validation';
import type { HttpClient } from '@/services/core/httpClient';
import { createProjectRole, fetchProjectById, fetchProjectResources } from '@/services/projects';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { telemetry } from '@/lib/telemetry';
import { AlertCircle, ArrowLeft, Loader2, Search } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';

interface Props { httpClient: HttpClient }

export default function NewRole({ httpClient }: Props) {
    const { project_id } = useParams<{ project_id: string }>();
    const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL as string;
    const navigate = useNavigate();
    const location = useLocation();
    const queryClient = useQueryClient();
    const { toast } = useToast();

    const stateAny = (location.state || {}) as any;
    const paramsFrom = getProjectTabParams(project_id ?? '', 'roles') || stateAny?.from || '';

    const projectQuery = useQuery({
        queryKey: ['project', project_id],
        queryFn: () => fetchProjectById(httpClient, AUTH_BASE_URL, project_id!),
        enabled: !!project_id,
    });

    const project = projectQuery.data;
    const projectLoading = projectQuery.isLoading;

    const [name, setName] = useState('');
    const [code, setCode] = useState('');
    const [description, setDescription] = useState('');

    // Validation state
    const [nameTouched, setNameTouched] = useState(false);
    const [codeTouched, setCodeTouched] = useState(false);
    const [descriptionTouched, setDescriptionTouched] = useState(false);

    const [resourcePage, setResourcePage] = useState(1);
    const [resourceSearchInput, setResourceSearchInput] = useState('');
    const [resourceSearch, setResourceSearch] = useState('');
    const [isDefaultFilter, setIsDefaultFilter] = useState<'all' | 'default' | 'custom'>('all');
    const resourcePageSize = 10;
    const [selectedPermissionIds, setSelectedPermissionIds] = useState<string[]>([]);

    useEffect(() => { const t = setTimeout(() => setResourceSearch(resourceSearchInput), 300); return () => clearTimeout(t); }, [resourceSearchInput]);

    const resourcesQuery = useQuery({
        queryKey: ['project', project?.id, 'resources', resourcePage, resourceSearch, resourcePageSize, isDefaultFilter],
        queryFn: () => {
            const isDefaultParam = isDefaultFilter === 'all' ? null : (isDefaultFilter === 'default');
            return fetchProjectResources(httpClient, AUTH_BASE_URL, project!.id, resourcePage, resourcePageSize, 'name', 'asc', resourceSearch, isDefaultParam);
        },
        enabled: !!project?.id,
    });

    useEffect(() => {
        if (project) {
            setName('');
            setCode('');
            setDescription('');
            setSelectedPermissionIds([]);
            // Reset touched state
            setNameTouched(false);
            setCodeTouched(false);
            setDescriptionTouched(false);
        }
    }, [project]);

    // Validation logic using shared utilities
    const validation = useMemo<RoleValidationResult>(
        () => validateRoleForm(name, code, description),
        [name, code, description]
    );

    // Handle name change with validation
    const handleNameChange = (value: string) => {
        setName(truncateToMaxLength(value, ROLE_VALIDATION.name.maxLength));
    };

    // Handle code change with validation
    const handleCodeChange = (value: string) => {
        setCode(truncateToMaxLength(value, ROLE_VALIDATION.code.maxLength));
    };

    // Handle description change with validation
    const handleDescriptionChange = (value: string) => {
        setDescription(truncateToMaxLength(value, ROLE_VALIDATION.description.maxLength));
    };

    const createMutation = useMutation({
        mutationFn: (payload: { name: string; code?: string; description?: string; permission_ids: string[] }) => createProjectRole(httpClient, AUTH_BASE_URL, project!.id, payload),
        onSuccess: () => {
            telemetry.trackEvent('role_created', {
                project_id: project?.id,
                permission_count: selectedPermissionIds.length,
            });
            queryClient.invalidateQueries({ queryKey: ['project', project?.id, 'roles'] });
            toast({ title: 'Role created', description: 'New role has been added', variant: 'default' });
            navigate(`/projects/${project?.id}/details${paramsFrom}`, { state: { from: paramsFrom } });
        },
        onError: (err: any) => toast({ title: 'Create role failed', description: err?.message ?? 'Unable to create role', variant: 'destructive' }),
    });

    const togglePermission = (id: string) => setSelectedPermissionIds((prev) => prev.includes(id) ? prev.filter(p => p !== id) : [...prev, id]);

    if (projectLoading) {
        return (
            <div className="flex items-center justify-center py-20" data-testid="newrole-loading">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                <span className="ml-2 text-sm text-muted-foreground">Loading…</span>
            </div>
        );
    }

    if (projectQuery.isError || !project) {
        return (
            <div className="flex flex-col items-center justify-center py-20 space-y-2" data-testid="newrole-error">
                <AlertCircle className="h-8 w-8 text-destructive" />
                <p className="text-sm text-destructive font-medium">Failed to load project</p>
                <p className="text-xs text-muted-foreground">{projectQuery.error?.message ?? 'Project data is unavailable'}</p>
            </div>
        );
    }

    const resources = resourcesQuery.data?.data ?? [] as any[];

    const creating = createMutation.status === 'pending';

    return (
        <div className="space-y-6 max-w-4xl">
            <div>
                <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => {
                        navigate(`/projects/${project?.id}/details${paramsFrom}`);
                    }}
                    className="px-3"
                >
                    <ArrowLeft className="h-4 w-4 mr-2" />
                    Back to {project?.name}
                </Button>
            </div>

            <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                    {project?.logo_url ? (
                        <img src={project.logo_url} alt={project.name} className="h-12 w-12 rounded object-cover" />
                    ) : (
                        <div className="h-12 w-12 rounded bg-muted flex items-center justify-center text-sm font-bold text-muted-foreground">{project?.name?.charAt(0)}</div>
                    )}
                    <div>
                        <h1 className="text-lg font-medium">Add New Role</h1>
                        <div className="text-sm text-muted-foreground">Create a new role for {project?.name}</div>
                    </div>
                </div>
            </div>

            <div className="border rounded p-4 bg-card w-full">
                <h3 className="text-lg font-medium">Role Information</h3>
                <div className="text-sm text-muted-foreground">Set the basic information for the new role and assign the necessary permissions</div>
                <div className="grid grid-cols-12 gap-4 mt-4">
                    <div className="col-span-4 flex items-center">
                        <Label htmlFor="role-name">Name</Label>
                    </div>
                    <div className="col-span-8">
                        <Input
                            id="role-name"
                            placeholder="Enter role name"
                            className={`mt-1.5 w-full ${nameTouched && validation.errors.name ? 'border-destructive' : ''}`}
                            value={name}
                            onChange={(e) => {
                                const v = e.target.value;
                                handleNameChange(v);
                                setCode(slugify(v));
                                if (!nameTouched) setNameTouched(true);
                            }}
                            onBlur={() => setNameTouched(true)}
                            maxLength={ROLE_VALIDATION.name.maxLength}
                            aria-invalid={nameTouched && !!validation.errors.name}
                        />
                        {nameTouched && validation.errors.name && (
                            <div className="text-xs text-destructive mt-1">{validation.errors.name}</div>
                        )}
                    </div>

                    <div className="col-span-4 flex items-center">
                        <Label htmlFor="role-code">Code</Label>
                    </div>
                    <div className="col-span-8">
                        <Input
                            id="role-code"
                            placeholder="Enter role code"
                            className={`mt-1.5 w-full ${codeTouched && validation.errors.code ? 'border-destructive' : ''}`}
                            value={code}
                            onChange={(e) => {
                                handleCodeChange(e.target.value);
                                if (!codeTouched) setCodeTouched(true);
                            }}
                            onBlur={() => setCodeTouched(true)}
                            maxLength={ROLE_VALIDATION.code.maxLength}
                            aria-invalid={codeTouched && !!validation.errors.code}
                        />
                        {codeTouched && validation.errors.code && (
                            <div className="text-xs text-destructive mt-1">{validation.errors.code}</div>
                        )}
                    </div>

                    <div className="col-span-4 flex items-center">
                        <Label htmlFor="role-desc">Description</Label>
                    </div>
                    <div className="col-span-8">
                        <Input
                            id="role-desc"
                            placeholder="Enter role description"
                            className={`mt-1.5 w-full ${descriptionTouched && validation.errors.description ? 'border-destructive' : ''}`}
                            value={description}
                            onChange={(e) => {
                                handleDescriptionChange(e.target.value);
                                if (!descriptionTouched) setDescriptionTouched(true);
                            }}
                            onBlur={() => setDescriptionTouched(true)}
                            maxLength={ROLE_VALIDATION.description.maxLength}
                            aria-invalid={descriptionTouched && !!validation.errors.description}
                        />
                        {descriptionTouched && validation.errors.description && (
                            <div className="text-xs text-destructive mt-1">{validation.errors.description}</div>
                        )}
                    </div>

                    <div className="col-span-12 flex items-center pt-4">
                        <Label htmlFor="permissions-search">Permissions</Label>
                    </div>
                    <div className="col-span-12">
                        <div className="flex items-center gap-2">
                            <div className="relative w-full sm:max-w-xs">
                                <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                                <Input id="permissions-search" role="searchbox" className="pl-8 pr-10" placeholder="Search permissions..." value={resourceSearchInput} onChange={(e) => { setResourceSearchInput(e.target.value); setResourcePage(1); }} />
                            </div>

                            <Select value={isDefaultFilter} onValueChange={(v) => { setIsDefaultFilter(v as any); setResourcePage(1); }}>
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
                        <div className="mt-3">
                            <div className="border rounded mt-1.5">
                                {resourcesQuery.isFetching ? (
                                    <div className="flex items-center gap-2 p-3 text-sm text-muted-foreground">
                                        <Loader2 className="h-4 w-4 animate-spin" />
                                        <span>Loading permissions…</span>
                                    </div>
                                ) : resources.length === 0 ? (
                                    <div className="col-span-full p-3 text-sm text-muted-foreground">No permissions found</div>
                                ) : (
                                    resources.map((res: any) => {
                                        const permIds = (res.permissions ?? []).map((pp: any) => pp.id);
                                        const allSelected = permIds.length > 0 && permIds.every((id: string) => selectedPermissionIds.includes(id));
                                        const someSelected = permIds.some((id: string) => selectedPermissionIds.includes(id));

                                        return (
                                            <div key={res.id} className="border-b last:border-b-0">
                                                <div
                                                    role="button"
                                                    tabIndex={0}
                                                    onKeyDown={(e) => {
                                                        if (e.key === 'Enter' || e.key === ' ') {
                                                            e.preventDefault();
                                                            const clickTarget = (e.target as HTMLElement);
                                                            if ((clickTarget).closest('[data-resource-checkbox]')) return;
                                                            if (allSelected) setSelectedPermissionIds((prev) => prev.filter((id) => !permIds.includes(id)));
                                                            else setSelectedPermissionIds((prev) => Array.from(new Set([...prev, ...permIds])));
                                                        }
                                                    }}
                                                    onClick={(e) => {
                                                        if ((e.target as HTMLElement).closest('[data-resource-checkbox]')) return;
                                                        if (allSelected) setSelectedPermissionIds((prev) => prev.filter((id) => !permIds.includes(id)));
                                                        else setSelectedPermissionIds((prev) => Array.from(new Set([...prev, ...permIds])));
                                                    }}
                                                    className="flex items-center justify-between p-2 cursor-pointer hover:bg-muted/5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring transition-colors rounded"
                                                >
                                                    <div className="flex items-center gap-3">
                                                        <div>
                                                            <input
                                                                data-resource-checkbox
                                                                id={`resource-${res.id}`}
                                                                aria-label={`resource-${res.id}`}
                                                                type="checkbox"
                                                                className={process.env.NODE_ENV === 'test' ? 'h-4 w-4' : 'sr-only'}
                                                                checked={allSelected}
                                                                onChange={() => {
                                                                    if (allSelected) {
                                                                        setSelectedPermissionIds((prev) => prev.filter((id) => !permIds.includes(id)));
                                                                    } else {
                                                                        setSelectedPermissionIds((prev) => Array.from(new Set([...prev, ...permIds])));
                                                                    }
                                                                }}
                                                            />
                                                            <Checkbox
                                                                checked={someSelected && !allSelected ? 'indeterminate' as any : allSelected}
                                                                onCheckedChange={() => {
                                                                    if (allSelected) setSelectedPermissionIds((prev) => prev.filter((id) => !permIds.includes(id)));
                                                                    else setSelectedPermissionIds((prev) => Array.from(new Set([...prev, ...permIds])));
                                                                }}
                                                            />
                                                        </div>
                                                        <div>
                                                            <div className="flex items-center gap-2">
                                                                <div className="text-sm font-medium">{res.name}</div>
                                                                {res.is_default ? <Badge variant="secondary">Default</Badge> : null}
                                                            </div>
                                                            <div className="text-xs text-muted-foreground">{res.description ?? ''}</div>
                                                        </div>
                                                    </div>
                                                </div>

                                                <div className="pl-10 pr-3 pb-2">
                                                    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2">
                                                        {(res.permissions ?? []).map((perm: any) => {
                                                            const isSelected = selectedPermissionIds.includes(perm.id);
                                                            return (
                                                                <div
                                                                    key={perm.id}
                                                                    role="button"
                                                                    tabIndex={0}
                                                                    onKeyDown={(e) => {
                                                                        if (e.key === 'Enter' || e.key === ' ') {
                                                                            e.preventDefault();
                                                                            if ((e.target as HTMLElement).closest('[data-perm-checkbox]')) return;
                                                                            togglePermission(perm.id);
                                                                        }
                                                                    }}
                                                                    onClick={(e) => { if ((e.target as HTMLElement).closest('[data-perm-checkbox]')) return; togglePermission(perm.id); }}
                                                                    className={`flex items-start gap-3 p-2 transition-colors cursor-pointer ${isSelected ? 'bg-primary/10' : 'bg-muted/5 hover:bg-muted/100'}`}
                                                                >
                                                                    <div className="pt-1">
                                                                        <input
                                                                            data-perm-checkbox
                                                                            id={`perm-${perm.id}`}
                                                                            aria-label={`permission-${perm.id}`}
                                                                            type="checkbox"
                                                                            className={process.env.NODE_ENV === 'test' ? 'h-4 w-4' : 'sr-only'}
                                                                            checked={isSelected}
                                                                            onChange={() => togglePermission(perm.id)}
                                                                        />
                                                                        <Checkbox checked={isSelected} onCheckedChange={() => togglePermission(perm.id)} />
                                                                    </div>
                                                                    <div className="flex-1">
                                                                        <div className="text-sm font-medium">{perm.name}</div>
                                                                        <div className="text-xs text-muted-foreground mt-1">{perm.description ?? ''}</div>
                                                                    </div>
                                                                </div>
                                                            );
                                                        })}
                                                    </div>
                                                </div>
                                            </div>
                                        );
                                    })
                                )}
                            </div>
                        </div>

                        <div className="flex items-center justify-center mt-2">
                            <ApiPagination
                                pagination={resourcesQuery.data?.pagination ?? null}
                                pageSize={resourcePageSize}
                                onPageChange={setResourcePage}
                            />
                        </div>
                    </div>

                    <div className="col-span-4" />
                    <div className="col-span-12 flex justify-end">
                        <div className="flex items-center gap-2">
                            <Button
                                size="sm"
                                onClick={() => {
                                    // Mark all fields as touched to show validation errors
                                    setNameTouched(true);
                                    setCodeTouched(true);
                                    setDescriptionTouched(true);
                                    if (validation.isValid) {
                                        createMutation.mutate({ name, code: code || slugify(name), description: description || undefined, permission_ids: selectedPermissionIds });
                                    }
                                }}
                                disabled={creating || !validation.isValid}
                            >
                                {creating ? (
                                    <>
                                        <Loader2 className="h-4 w-4 animate-spin mr-2" />
                                        Creating...
                                    </>
                                ) : (
                                    <>
                                        Create
                                    </>
                                )}
                            </Button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
