import { ConfirmDialog } from '@/components/ConfirmDialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useToast } from '@/hooks/use-toast';
import { getProjectTabParams } from '@/lib/paramsStore';
import { slugify } from '@/lib/slugify';
import {
    RESOURCE_VALIDATION,
    truncateToMaxLength,
    validateResourceForm,
    validatePermissionForm,
    type ResourceValidationResult,
    type PermissionValidationResult,
} from '@/lib/validation';
import type { HttpClient } from '@/services/core/httpClient';
import { deleteProjectResource, fetchProjectById, fetchProjectResourceById, updateProjectResource } from '@/services/projects';
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { telemetry } from '@/lib/telemetry';
import { AlertCircle, ArrowLeft, ChevronDown, ChevronUp, Loader2, Trash2 } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';

interface Props { httpClient: HttpClient }

export default function ResourceDetails({ httpClient }: Props) {
    const { project_id, resource_id } = useParams<{ project_id: string; resource_id: string }>();
    const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL as string;
    const navigate = useNavigate();
    const { toast } = useToast();
    const queryClient = useQueryClient();

    const location = useLocation();
    const stateAny = (location.state || {}) as any;
    const paramsFrom = getProjectTabParams(project_id ?? '', 'resources') || stateAny?.from || '';

    const projectQuery = useQuery({ queryKey: ['project', project_id], queryFn: () => fetchProjectById(httpClient, AUTH_BASE_URL, project_id!), enabled: !!project_id });

    const resourceQuery = useQuery({
        queryKey: ['project', project_id, 'resource', resource_id],
        queryFn: () => fetchProjectResourceById(httpClient, AUTH_BASE_URL, project_id!, resource_id!),
        enabled: !!projectQuery.data && !!resource_id,
    });

    const [name, setName] = useState('');
    const [code, setCode] = useState('');
    const [description, setDescription] = useState('');

    // Validation state
    const [nameTouched, setNameTouched] = useState(false);
    const [codeTouched, setCodeTouched] = useState(false);
    const [descriptionTouched, setDescriptionTouched] = useState(false);

    const [permissions, setPermissions] = useState<Array<any>>([]);
    const [permSortBy, setPermSortBy] = useState<'name' | 'code'>('name');
    const [permSortDir, setPermSortDir] = useState<'asc' | 'desc'>('asc');
    const [showAddDialog, setShowAddDialog] = useState(false);
    const [showRemoveDialog, setShowRemoveDialog] = useState(false);

    // Helper to handle permission sort toggle
    const handlePermSort = (key: 'name' | 'code') => {
        if (permSortBy !== key) {
            setPermSortBy(key);
            setPermSortDir('asc');
        } else {
            setPermSortDir(permSortDir === 'asc' ? 'desc' : 'asc');
        }
    };

    // Sort permissions locally
    const sortedPermissions = useMemo(() => {
        return [...permissions].sort((a, b) => {
            let aValue: string;
            let bValue: string;
            if (permSortBy === 'name') {
                aValue = (a.name ?? '').toLowerCase();
                bValue = (b.name ?? '').toLowerCase();
            } else {
                // Sort by code suffix
                aValue = (a.code ?? '').toLowerCase();
                bValue = (b.code ?? '').toLowerCase();
            }
            if (aValue < bValue) return permSortDir === 'asc' ? -1 : 1;
            if (aValue > bValue) return permSortDir === 'asc' ? 1 : -1;
            return 0;
        });
    }, [permissions, permSortBy, permSortDir]);

    useEffect(() => {
        if (resourceQuery.data) {
            setName(resourceQuery.data.name ?? '');
            setCode(resourceQuery.data.code ?? '');
            setDescription(resourceQuery.data.description ?? '');
            setPermissions((resourceQuery.data.permissions ?? []).map((p: any) => ({ ...p })));
            // Reset touched state when data loads
            setNameTouched(false);
            setCodeTouched(false);
            setDescriptionTouched(false);
        }
    }, [resourceQuery.data]);

    const isDefault = !!resourceQuery.data?.is_default;

    useEffect(() => { setCode(slugify(name)); }, [name]);

    // Validation logic using shared utilities
    const validation = useMemo<ResourceValidationResult>(
        () => validateResourceForm(name, code, description),
        [name, code, description]
    );

    // Handle name change with validation
    const handleNameChange = (value: string) => {
        setName(truncateToMaxLength(value, RESOURCE_VALIDATION.name.maxLength));
    };

    // Handle code change with validation
    const handleCodeChange = (value: string) => {
        setCode(truncateToMaxLength(value, RESOURCE_VALIDATION.code.maxLength));
    };

    // Handle description change with validation
    const handleDescriptionChange = (value: string) => {
        setDescription(truncateToMaxLength(value, RESOURCE_VALIDATION.description.maxLength));
    };

    const updateMutation = useMutation({
        mutationFn: (payload: any) => updateProjectResource(httpClient, AUTH_BASE_URL, project_id!, resource_id!, payload),
        onSuccess: (data) => {
            telemetry.trackEvent('resource_updated', {
                resource_id,
                project_id,
                permission_count: permissions.length,
            });
            queryClient.invalidateQueries({ queryKey: ['project', project_id, 'resources'] });
            toast({ title: 'Saved', description: 'Resource updated', variant: 'default' });
            navigate(`/projects/${project_id}/details${paramsFrom}`);
        },
        onError: (err: any) => {
            toast({ title: 'Save failed', description: err?.message ?? 'Unable to update resource', variant: 'destructive' });
        },
    });

    const deleteMutation = useMutation({
        mutationFn: () => deleteProjectResource(httpClient, AUTH_BASE_URL, project_id!, resource_id!),
        onSuccess: () => {
            telemetry.trackEvent('resource_deleted', { resource_id, project_id });
            queryClient.invalidateQueries({ queryKey: ['project', project_id, 'resources'] });
            toast({ title: 'Resource removed', description: 'Resource has been removed', variant: 'default' });
            navigate(`/projects/${project_id}/details${paramsFrom}`);
        },
        onError: (err: any) => {
            toast({ title: 'Remove failed', description: err?.message ?? 'Unable to remove resource', variant: 'destructive' });
        },
    });

    const addPermission = (perm: { name: string; code: string; description?: string }) => {
        setPermissions((prev) => [...prev, { id: `tmp-${Date.now()}`, ...perm }]);
        setShowAddDialog(false);
    };

    const removePermission = (id: string) => setPermissions((prev) => prev.filter((p) => p.id !== id));

    if (resourceQuery.isLoading || projectQuery.isLoading) {
        return (
            <div className="flex items-center justify-center py-20">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                <span className="ml-2 text-sm text-muted-foreground">Loading…</span>
            </div>
        );
    }

    if (resourceQuery.isError || projectQuery.isError || !projectQuery.data) {
        return (
            <div className="flex flex-col items-center justify-center py-20 space-y-2">
                <AlertCircle className="h-8 w-8 text-destructive" />
                <p className="text-sm text-destructive font-medium">Failed to load resource</p>
                <p className="text-xs text-muted-foreground">{resourceQuery.error?.message ?? projectQuery.error?.message ?? 'Resource data is unavailable'}</p>
            </div>
        );
    }

    return (
        <div className="space-y-6 max-w-4xl">
            <div>
                <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => {
                        navigate(`/projects/${projectQuery.data?.id}/details${paramsFrom}`);
                    }}
                    className="px-3"
                >
                    <ArrowLeft className="h-4 w-4 mr-2" />
                    Back to {projectQuery.data?.name}
                </Button>
            </div>

            <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                    {projectQuery.data?.logo_url ? (
                        <img src={projectQuery.data.logo_url} alt={projectQuery.data?.name} className="h-12 w-12 rounded object-cover" />
                    ) : (
                        <div className="h-12 w-12 rounded bg-muted flex items-center justify-center text-sm font-bold text-muted-foreground">{(projectQuery.data?.name ?? '')?.charAt(0)}</div>
                    )}
                    <div>
                        <h1 className="text-lg font-medium">{name || 'Resource'}</h1>
                        <div className="text-sm text-muted-foreground">{description ?? ''}</div>
                    </div>
                </div>
                <div />
            </div>

            <div className="border rounded p-4 bg-card w-full">
                {isDefault && (
                    <div className="mb-3 p-3 rounded bg-yellow-50 text-sm text-yellow-800 border border-yellow-200">Default resource cannot be edited or removed.</div>
                )}

                <h3 className="text-lg font-medium">Resource Information</h3>
                <div className="text-sm text-muted-foreground">View or edit the resource and manage permissions</div>

                <div className="grid grid-cols-12 gap-4 mt-4">
                    <div className="col-span-4 flex items-center"><Label htmlFor="resource-name">Name</Label></div>
                    <div className="col-span-8">
                        <Input
                            id="resource-name"
                            placeholder="Enter resource name"
                            value={name}
                            onChange={(e) => {
                                const v = e.target.value;
                                handleNameChange(v);
                                setCode(slugify(v));
                                if (!nameTouched) setNameTouched(true);
                            }}
                            onBlur={() => setNameTouched(true)}
                            className={nameTouched && validation.errors.name ? 'border-destructive' : ''}
                            maxLength={RESOURCE_VALIDATION.name.maxLength}
                            aria-invalid={nameTouched && !!validation.errors.name}
                            readOnly={isDefault}
                        />
                        {nameTouched && validation.errors.name && (
                            <div className="text-xs text-destructive mt-1">{validation.errors.name}</div>
                        )}
                    </div>

                    <div className="col-span-4 flex items-center"><Label htmlFor="resource-code">Code</Label></div>
                    <div className="col-span-8">
                        <Input
                            id="resource-code"
                            placeholder="Enter resource code"
                            value={code}
                            onChange={(e) => {
                                handleCodeChange(e.target.value);
                                if (!codeTouched) setCodeTouched(true);
                            }}
                            onBlur={() => setCodeTouched(true)}
                            className={codeTouched && validation.errors.code ? 'border-destructive' : ''}
                            maxLength={RESOURCE_VALIDATION.code.maxLength}
                            aria-invalid={codeTouched && !!validation.errors.code}
                            readOnly={isDefault}
                        />
                        {codeTouched && validation.errors.code && (
                            <div className="text-xs text-destructive mt-1">{validation.errors.code}</div>
                        )}
                    </div>
                    <div className="col-span-4 flex items-center"><Label htmlFor="resource-description">Description</Label></div>
                    <div className="col-span-8">
                        <Input
                            id="resource-description"
                            placeholder="Enter resource description"
                            value={description}
                            onChange={(e) => {
                                handleDescriptionChange(e.target.value);
                                if (!descriptionTouched) setDescriptionTouched(true);
                            }}
                            onBlur={() => setDescriptionTouched(true)}
                            className={descriptionTouched && validation.errors.description ? 'border-destructive' : ''}
                            maxLength={RESOURCE_VALIDATION.description.maxLength}
                            aria-invalid={descriptionTouched && !!validation.errors.description}
                            readOnly={isDefault}
                        />
                        {descriptionTouched && validation.errors.description && (
                            <div className="text-xs text-destructive mt-1">{validation.errors.description}</div>
                        )}
                    </div>

                    <div className="col-span-12 flex items-center">
                        <div className="flex items-center justify-between w-full">
                            <Label>Permissions</Label>
                            {!isDefault && (
                                <>
                                    <div>
                                        <Button size="sm" onClick={() => setShowAddDialog(true)}>Add Permission</Button>
                                    </div>
                                </>
                            )}
                        </div>
                    </div>

                    <div className="col-span-12">
                        <div className="border rounded bg-card overflow-hidden">
                            <table className="w-full text-sm">
                                <thead>
                                    <tr className="border-b bg-muted/40">
                                        <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                                            <button onClick={() => handlePermSort('name')} className="flex items-center gap-2">
                                                <span>Name (Action)</span>
                                                {permSortBy === 'name' ? (permSortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                                            </button>
                                        </th>
                                        <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                                            <button onClick={() => handlePermSort('code')} className="flex items-center gap-2">
                                                <span>Code</span>
                                                {permSortBy === 'code' ? (permSortDir === 'asc' ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />) : null}
                                            </button>
                                        </th>
                                        <th className="w-20 px-4 py-2.5"> </th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {sortedPermissions.length === 0 ? (
                                        <tr><td colSpan={3} className="px-4 py-3 text-sm text-muted-foreground">No permissions</td></tr>
                                    ) : (
                                        sortedPermissions.map((p) => (
                                            <tr key={p.id} className="border-b last:border-0 hover:bg-muted/20 transition-colors">
                                                <td className="px-4 py-3 align-top">
                                                    <div className="text-sm font-medium">{p.name}</div>
                                                    <div className="text-xs text-muted-foreground mt-1">{p.description ?? ''}</div>
                                                </td>
                                                <td className="px-4 py-3 align-center">
                                                    <span className="inline-flex items-center px-2 py-0.5 font-mono rounded bg-muted text-muted-foreground text-xs">{`${code}:${p.code}`.replace(/^:/, '')}</span>
                                                </td>
                                                <td className="px-4 py-3 align-center">
                                                    {!isDefault && (
                                                        <>
                                                            <div className="flex items-center justify-end">
                                                                <Tooltip>
                                                                    <TooltipTrigger asChild>
                                                                        <Button size="sm" variant="outline" className="text-destructive bg-white hover:text-destructive border-destructive/30 hover:bg-destructive/10" onClick={() => removePermission(p.id)} disabled={isDefault && p.is_default}>
                                                                            <Trash2 />
                                                                        </Button>
                                                                    </TooltipTrigger>
                                                                    <TooltipContent>Remove</TooltipContent>
                                                                </Tooltip>
                                                            </div>
                                                        </>
                                                    )}
                                                </td>
                                            </tr>
                                        ))
                                    )}
                                </tbody>
                            </table>
                        </div>
                    </div>

                    {!isDefault && (
                        <div className="col-span-12 flex justify-end gap-2">
                            <Button
                                size="sm"
                                variant="destructive"
                                onClick={() => setShowRemoveDialog(true)}
                                disabled={deleteMutation.status === 'pending'}
                            >
                                {deleteMutation.status === 'pending' ? (<><Loader2 className="h-4 w-4 animate-spin mr-2" />Removing...</>) : 'Remove'}
                            </Button>
                            <Button
                                size="sm"
                                onClick={() => {
                                    // Mark all fields as touched to show validation errors
                                    setNameTouched(true);
                                    setCodeTouched(true);
                                    setDescriptionTouched(true);
                                    if (validation.isValid) {
                                        updateMutation.mutate({ name, code, description, permissions: permissions.map((pp) => ({ name: pp.name, code: pp.code, description: pp.description })) });
                                    }
                                }}
                                disabled={updateMutation.status === 'pending' || !validation.isValid}
                            >
                                {updateMutation.status === 'pending' ? (<><Loader2 className="h-4 w-4 animate-spin mr-2" />Saving...</>) : 'Save'}
                            </Button>
                        </div>
                    )}
                </div>
            </div>

            <Dialog open={showAddDialog} onOpenChange={(open) => { if (!open) setShowAddDialog(false); }}>
                <DialogContent onOpenAutoFocus={(e: any) => e.preventDefault()}>
                    <DialogHeader>
                        <DialogTitle>Add Permission</DialogTitle>
                        <DialogDescription>Define a new permission for this resource</DialogDescription>
                    </DialogHeader>
                    <PermissionForm resourceCode={code} existingPermissions={permissions} onCancel={() => setShowAddDialog(false)} onAdd={addPermission} />
                </DialogContent>
            </Dialog>

            <ConfirmDialog
                open={showRemoveDialog}
                onOpenChange={(open) => { if (!open) setShowRemoveDialog(false); }}
                closeOnConfirm={false}
                title="Remove Resource"
                description={<span>Are you sure you want to remove <span className="font-medium">{name}</span>? This action cannot be undone.</span>}
                confirmLabel={deleteMutation.status === 'pending' ? (<><Loader2 className="h-4 w-4 animate-spin mr-2" />Removing...</>) : 'Remove'}
                destructive
                disabled={deleteMutation.status === 'pending'}
                onConfirm={async () => {
                    await deleteMutation.mutateAsync();
                    setShowRemoveDialog(false);
                }}
            />
        </div>
    );
}

function PermissionForm({ resourceCode, existingPermissions, onAdd, onCancel }: { resourceCode: string; existingPermissions: Array<any>; onAdd: (p: any) => void; onCancel: () => void }) {
    const [name, setName] = useState('');
    const [codeSuffix, setCodeSuffix] = useState('');
    const [description, setDescription] = useState('');

    // Validation state
    const [nameTouched, setNameTouched] = useState(false);
    const [codeTouched, setCodeTouched] = useState(false);
    const [descriptionTouched, setDescriptionTouched] = useState(false);

    useEffect(() => { setCodeSuffix(slugify(name)); }, [name]);
    const codePrefix = resourceCode ? `${resourceCode}:` : '';

    // Get existing permission codes for duplicate checking
    const existingCodes = existingPermissions.map(p => p.code);

    // Validation logic
    const validation = useMemo<PermissionValidationResult>(
        () => validatePermissionForm(name, codeSuffix, description, existingCodes),
        [name, codeSuffix, description, existingCodes]
    );

    // Handle name change with validation
    const handleNameChange = (value: string) => {
        setName(truncateToMaxLength(value, RESOURCE_VALIDATION.permission.name.maxLength));
    };

    // Handle code change with validation
    const handleCodeChange = (value: string) => {
        setCodeSuffix(truncateToMaxLength(value, RESOURCE_VALIDATION.permission.code.maxLength));
    };

    // Handle description change with validation
    const handleDescriptionChange = (value: string) => {
        setDescription(truncateToMaxLength(value, RESOURCE_VALIDATION.permission.description.maxLength));
    };

    const handleAdd = () => {
        // Mark all fields as touched to show validation errors
        setNameTouched(true);
        setCodeTouched(true);
        setDescriptionTouched(true);
        if (validation.isValid) {
            onAdd({ name, code: codeSuffix, description });
        }
    };

    return (
        <div className="grid grid-cols-12 gap-4">
            <div className="col-span-12">
                <Label htmlFor="perm-name">Name (Action)</Label>
                <Input
                    className={`mt-1.5 ${nameTouched && validation.errors.name ? 'border-destructive' : ''}`}
                    id="perm-name"
                    placeholder="Enter action name"
                    value={name}
                    onChange={(e) => {
                        handleNameChange(e.target.value);
                        if (!nameTouched) setNameTouched(true);
                    }}
                    onBlur={() => setNameTouched(true)}
                    maxLength={RESOURCE_VALIDATION.permission.name.maxLength}
                    aria-invalid={nameTouched && !!validation.errors.name}
                />
                {nameTouched && validation.errors.name && (
                    <div className="text-xs text-destructive mt-1">{validation.errors.name}</div>
                )}
            </div>
            <div className="col-span-12">
                <Label htmlFor="perm-code">Code</Label>
                <div className={`mt-1.5 flex items-stretch overflow-hidden rounded-md border bg-background focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-2 focus-within:ring-offset-background ${codeTouched && validation.errors.code ? 'border-destructive' : 'border-input'}`}>
                    {codePrefix ? (
                        <span className="inline-flex max-w-[55%] shrink-0 items-center border-r border-input bg-muted px-3 py-2 text-sm text-muted-foreground whitespace-nowrap overflow-hidden text-ellipsis">
                            {codePrefix}
                        </span>
                    ) : null}
                    <Input
                        id="perm-code"
                        className="border-0 bg-transparent shadow-none focus-visible:ring-0 focus-visible:ring-offset-0 min-w-0"
                        placeholder="Enter code suffix"
                        value={codeSuffix}
                        onChange={(e) => {
                            handleCodeChange(e.target.value);
                            if (!codeTouched) setCodeTouched(true);
                        }}
                        onBlur={() => setCodeTouched(true)}
                        maxLength={RESOURCE_VALIDATION.permission.code.maxLength}
                        aria-invalid={codeTouched && !!validation.errors.code}
                    />
                </div>
                {codeTouched && validation.errors.code && (
                    <div className="text-xs text-destructive mt-1">{validation.errors.code}</div>
                )}
            </div>
            <div className="col-span-12">
                <Label htmlFor="perm-description">Description</Label>
                <Input
                    className={`mt-1.5 ${descriptionTouched && validation.errors.description ? 'border-destructive' : ''}`}
                    id="perm-description"
                    placeholder="Enter permission description"
                    value={description}
                    onChange={(e) => {
                        handleDescriptionChange(e.target.value);
                        if (!descriptionTouched) setDescriptionTouched(true);
                    }}
                    onBlur={() => setDescriptionTouched(true)}
                    maxLength={RESOURCE_VALIDATION.permission.description.maxLength}
                    aria-invalid={descriptionTouched && !!validation.errors.description}
                />
                {descriptionTouched && validation.errors.description && (
                    <div className="text-xs text-destructive mt-1">{validation.errors.description}</div>
                )}
            </div>
            <div className="col-span-12 flex justify-end gap-2 pt-4">
                <Button variant="outline" onClick={onCancel}>Cancel</Button>
                <Button onClick={handleAdd} disabled={!validation.isValid}>
                    Add
                </Button>
            </div>
        </div>
    );
}
