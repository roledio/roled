import ApiPagination from '@/components/ApiPagination';
import { ConfirmDialog } from '@/components/ConfirmDialog';
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
import { deleteProjectRole, fetchProjectById, fetchProjectResources, fetchRoleById, updateProjectRole } from '@/services/projects';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { telemetry } from '@/lib/telemetry';
import { AlertCircle, ArrowLeft, Loader2, Search } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';

interface Props { httpClient: HttpClient }

export default function RoleDetails({ httpClient }: Props) {
  const { project_id, role_id } = useParams<{ project_id: string; role_id: string }>();
  const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL as string;
  const navigate = useNavigate();
  const { toast } = useToast();

  const location = useLocation();
  const stateAny = (location.state || {}) as any;

  const getParamsFrom = () => {
    const resolvedPid = (projectQuery && (projectQuery.data as any)?.id) ?? projectId ?? project_id;
    if (resolvedPid) {
      const stored = getProjectTabParams(resolvedPid as string, 'roles');
      if (stored) return stored;
    }
    return (stateAny?.from as string) || '';
  };

  const [projectId, setProjectId] = useState<string | undefined>(project_id);

  // load project first
  const projectQuery = useQuery<any, Error>({
    queryKey: ['project', projectId ?? project_id],
    queryFn: () => fetchProjectById(httpClient, AUTH_BASE_URL, projectId ?? project_id!),
    enabled: !!(projectId ?? project_id),
  });

  useEffect(() => {
    if (projectQuery.data) setProjectId((projectQuery.data as any).id);
  }, [projectQuery.data]);

  const [name, setName] = useState('');
  const [code, setCode] = useState('');
  const [description, setDescription] = useState('');

  // Validation state
  const [nameTouched, setNameTouched] = useState(false);
  const [codeTouched, setCodeTouched] = useState(false);
  const [descriptionTouched, setDescriptionTouched] = useState(false);

  const [roleId, setRoleId] = useState('');
  const [selectedPermissionIds, setSelectedPermissionIds] = useState<string[]>([]);

  const roleQuery = useQuery<any, Error>({
    queryKey: ['role', projectId ?? project_id, roleId || role_id],
    queryFn: async () => {
      const pid = projectId ?? projectQuery.data?.id ?? project_id;
      if (!pid) throw new Error('project id missing');
      const rid = roleId || role_id;
      if (!rid) throw new Error('role id missing');
      return fetchRoleById(httpClient, AUTH_BASE_URL, pid, rid!);
    },
    enabled: !!(projectQuery.data?.id || projectId || project_id) && !!(roleId || role_id),
  });

  // pagination + search state for resources
  const [resourcePage, setResourcePage] = useState(1);
  const [resourcePageSize] = useState(10);
  const [resourceSearchInput, setResourceSearchInput] = useState('');
  const [resourceSearch, setResourceSearch] = useState('');
  const [isDefaultFilter, setIsDefaultFilter] = useState<'all' | 'default' | 'custom'>('all');

  useEffect(() => {
    if (roleQuery.data) {
      setName(roleQuery.data.name ?? '');
      setCode(roleQuery.data.code ?? '');
      setDescription(roleQuery.data.description ?? '');
      setRoleId(roleQuery.data.id ?? '');
      setSelectedPermissionIds(roleQuery.data.permission_ids ?? []);
      // Reset touched state when data loads
      setNameTouched(false);
      setCodeTouched(false);
      setDescriptionTouched(false);
    }
  }, [roleQuery.data]);

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

  useEffect(() => { const t = setTimeout(() => setResourceSearch(resourceSearchInput), 300); return () => clearTimeout(t); }, [resourceSearchInput]);

  const resourcesPagedQuery = useQuery<any, Error>({
    queryKey: ['project', projectId, 'resources', resourcePage, resourceSearch, resourcePageSize, isDefaultFilter],
    queryFn: () => {
      const isDefaultParam = isDefaultFilter === 'all' ? null : (isDefaultFilter === 'default');
      return fetchProjectResources(httpClient, AUTH_BASE_URL, projectId!, resourcePage, resourcePageSize, 'name', 'asc', resourceSearch, isDefaultParam);
    },
    enabled: !!projectId,
  });

  const queryClient = useQueryClient();

  const updateMutation = useMutation({
    mutationFn: (payload: { name: string; code?: string; description?: string; permission_ids: string[] }) => {
      const pid = projectId ?? projectQuery.data?.id;
      const rid = roleId;
      if (!pid || !rid) return Promise.reject(new Error('project or role id missing'));
      return updateProjectRole(httpClient, AUTH_BASE_URL, pid, rid, payload);
    },
    onMutate: () => { },
    onSuccess: (data) => {
      const pid = projectId ?? projectQuery.data?.id;
      telemetry.trackEvent('role_updated', {
        role_id: roleId,
        project_id: pid,
        permission_count: selectedPermissionIds.length,
      });
      queryClient.setQueryData(['role', pid, roleId], data);
      toast({ title: 'Saved', description: 'Role updated', variant: 'default' });
      // redirect to project details after successful update (preserve query state)
      if (pid) {
        navigate(`/projects/${pid}/details${getParamsFrom()}`);
      }
    },
    onError: (err: any) => {
      toast({ title: 'Save failed', description: err?.message ?? 'Unable to update role', variant: 'destructive' });
    },
  });

  const [showRemoveDialog, setShowRemoveDialog] = useState(false);
  const [removeLoading, setRemoveLoading] = useState(false);

  const confirmRemoveRole = async () => {
    const pid = projectId ?? projectQuery.data?.id;
    if (!pid) {
      toast({ title: 'Remove failed', description: 'Project id missing', variant: 'destructive' });
      return;
    }
    setRemoveLoading(true);
    try {
      await deleteProjectRole(httpClient, AUTH_BASE_URL, pid, roleId);
      telemetry.trackEvent('role_deleted', { role_id: roleId, project_id: pid });
      queryClient.invalidateQueries({ queryKey: ['project', pid, 'roles'] });
      toast({ title: 'Role removed', description: 'Role has been removed', variant: 'default' });
      navigate(`/projects/${pid}/details${getParamsFrom()}`);
    } catch (err: any) {
      toast({ title: 'Remove failed', description: err?.message ?? 'Unable to remove role', variant: 'destructive' });
    } finally {
      setRemoveLoading(false);
      setShowRemoveDialog(false);
    }
  };

  if (roleQuery.isLoading || projectQuery.isLoading) {
    return (
      <div className="flex items-center justify-center py-20" data-testid="role-loading">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        <span className="ml-2 text-sm text-muted-foreground">Loading…</span>
      </div>
    );
  }

  if (roleQuery.isError || projectQuery.isError || !projectQuery.data) {
    return (
      <div className="flex flex-col items-center justify-center py-20 space-y-2" data-testid="role-error">
        <AlertCircle className="h-8 w-8 text-destructive" />
        <p className="text-sm text-destructive font-medium">Failed to load role</p>
        <p className="text-xs text-muted-foreground">{roleQuery.error?.message ?? projectQuery.error?.message ?? 'Role data is unavailable'}</p>
      </div>
    );
  }

  const resources = resourcesPagedQuery.data?.data ?? [] as any[];

  return (
    <div className="space-y-6 max-w-4xl">
      <div>
        <Button
          variant="secondary"
          size="sm"
          onClick={() => {
            navigate(`/projects/${projectQuery.data?.id}/details${getParamsFrom()}`);
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
            <h1 className="text-lg font-medium">{name || 'Role'}</h1>
            <div className="text-sm text-muted-foreground">{description ?? ''}</div>
          </div>
        </div>
        <div />
      </div>

      <div className="border rounded p-4 bg-card w-full">
        <h3 className="text-lg font-medium">Role Information</h3>
        <div className="text-sm text-muted-foreground">View or edit the role information and manage the permissions for this role</div>
        <div className="grid grid-cols-12 gap-4 mt-4">
          <div className="col-span-4 flex items-center"><Label htmlFor="role-name">Name</Label></div>
          <div className="col-span-8">
            <Input
              id="role-name"
              placeholder="Enter role name"
              value={name}
              onChange={(e) => {
                const v = e.target.value;
                handleNameChange(v);
                setCode(slugify(v));
                if (!nameTouched) setNameTouched(true);
              }}
              onBlur={() => setNameTouched(true)}
              className={nameTouched && validation.errors.name ? 'border-destructive' : ''}
              maxLength={ROLE_VALIDATION.name.maxLength}
              aria-invalid={nameTouched && !!validation.errors.name}
            />
            {nameTouched && validation.errors.name && (
              <div className="text-xs text-destructive mt-1">{validation.errors.name}</div>
            )}
          </div>
          <div className="col-span-4 flex items-center"><Label htmlFor="role-code">Code</Label></div>
          <div className="col-span-8">
            <Input
              id="role-code"
              placeholder="Enter role code"
              value={code}
              onChange={(e) => {
                handleCodeChange(e.target.value);
                if (!codeTouched) setCodeTouched(true);
              }}
              onBlur={() => setCodeTouched(true)}
              className={codeTouched && validation.errors.code ? 'border-destructive' : ''}
              maxLength={ROLE_VALIDATION.code.maxLength}
              aria-invalid={codeTouched && !!validation.errors.code}
            />
            {codeTouched && validation.errors.code && (
              <div className="text-xs text-destructive mt-1">{validation.errors.code}</div>
            )}
          </div>
          <div className="col-span-4 flex items-center"><Label htmlFor="role-desc">Description</Label></div>
          <div className="col-span-8">
            <Input
              id="role-desc"
              placeholder="Enter role description"
              value={description}
              onChange={(e) => {
                handleDescriptionChange(e.target.value);
                if (!descriptionTouched) setDescriptionTouched(true);
              }}
              onBlur={() => setDescriptionTouched(true)}
              className={descriptionTouched && validation.errors.description ? 'border-destructive' : ''}
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
                {resourcesPagedQuery.isFetching ? (
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
                                      setSelectedPermissionIds((prev) => prev.includes(perm.id) ? prev.filter(p => p !== perm.id) : [...prev, perm.id]);
                                    }
                                  }}
                                  onClick={(e) => { if ((e.target as HTMLElement).closest('[data-perm-checkbox]')) return; setSelectedPermissionIds((prev) => prev.includes(perm.id) ? prev.filter(p => p !== perm.id) : [...prev, perm.id]); }}
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
                                      onChange={() => setSelectedPermissionIds((prev) => prev.includes(perm.id) ? prev.filter(p => p !== perm.id) : [...prev, perm.id])}
                                    />
                                    <Checkbox checked={isSelected} onCheckedChange={() => setSelectedPermissionIds((prev) => prev.includes(perm.id) ? prev.filter(p => p !== perm.id) : [...prev, perm.id])} />
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
                pagination={resourcesPagedQuery.data?.pagination ?? null}
                pageSize={resourcePageSize}
                onPageChange={setResourcePage}
              />
            </div>
          </div>

          <div className="col-span-12 flex justify-end">
            <div className="flex items-center gap-2">
              <Button size="sm" variant="destructive" onClick={() => setShowRemoveDialog(true)} disabled={removeLoading}>
                Remove
              </Button>
              <Button
                size="sm"
                onClick={() => {
                  // Mark all fields as touched to show validation errors
                  setNameTouched(true);
                  setCodeTouched(true);
                  setDescriptionTouched(true);
                  if (validation.isValid) {
                    updateMutation.mutate({ name, code: code || slugify(name), description: description || undefined, permission_ids: selectedPermissionIds });
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
        </div>
      </div>
      <ConfirmDialog
        open={showRemoveDialog}
        onOpenChange={(open) => { if (!open) setShowRemoveDialog(false); }}
        closeOnConfirm={false}
        title="Remove Role"
        description={<span>Are you sure you want to remove <span className="font-medium">{name || roleQuery.data?.name}</span>? This action cannot be undone.</span>}
        confirmLabel={removeLoading ? (<><Loader2 className="h-4 w-4 animate-spin mr-2" />Removing...</>) : 'Remove'}
        destructive
        disabled={removeLoading}
        onConfirm={confirmRemoveRole}
      />
    </div>
  );
}
