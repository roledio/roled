import { ConfirmDialog } from '@/components/ConfirmDialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@/components/ui/dropdown-menu';
import { useToast } from '@/hooks/use-toast';
import { getProjectTabParams } from '@/lib/paramsStore';
import {
    USER_VALIDATION,
    truncateToMaxLength,
    validateUserForm,
    type UserValidationResult,
} from '@/lib/validation';
import type { HttpClient } from '@/services/core/httpClient';
import { deleteProjectUser, fetchProjectById, fetchProjectRoles, fetchUserById, updateProjectUser, uploadUserAvatar } from '@/services/projects';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { telemetry } from '@/lib/telemetry';
import { AlertCircle, ArrowLeft, Eye, EyeOff, Loader2, MoreVertical, CloudUpload, Trash2 } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';

interface Props { httpClient: HttpClient }

export default function UserDetails({ httpClient }: Props) {
  const { project_id, user_id } = useParams<{ project_id: string; user_id: string }>();
  const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL as string;
  const navigate = useNavigate();
  const { toast } = useToast();

  const location = useLocation();
  const stateAny = (location.state || {}) as any;

  const getParamsFrom = () => {
    const resolvedPid = (projectQuery && (projectQuery.data as any)?.id) ?? projectId ?? project_id;
    if (resolvedPid) {
      const stored = getProjectTabParams(resolvedPid as string, 'users');
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

  // fetch roles for dropdown
  const rolesQuery = useQuery({
    queryKey: ['project', projectId, 'roles'],
    queryFn: () => fetchProjectRoles(httpClient, AUTH_BASE_URL, projectId!, 1, 100, 'name', 'asc'),
    enabled: !!projectId,
  });

  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [externalUserId, setExternalUserId] = useState('');

  const [avatarPreview, setAvatarPreview] = useState<string | null>(null);
  const [showUploadDialog, setShowUploadDialog] = useState(false);
  const [selectedAvatarFile, setSelectedAvatarFile] = useState<File | null>(null);
  const inputUploadRef = useRef<HTMLInputElement | null>(null);
  const [uploadingAvatar, setUploadingAvatar] = useState(false);

  // Validation state
  const [displayNameTouched, setDisplayNameTouched] = useState(false);
  const [emailTouched, setEmailTouched] = useState(false);
  const [passwordTouched, setPasswordTouched] = useState(false);
  const [externalUserIdTouched, setExternalUserIdTouched] = useState(false);

  const [isActive, setIsActive] = useState<boolean>(true);
  const [roleId, setRoleId] = useState<string>('');
  const [userId, setUserId] = useState('');

  const userQuery = useQuery<any, Error>({
    queryKey: ['user', projectId ?? project_id, userId || user_id],
    queryFn: async () => {
      const pid = projectId ?? projectQuery.data?.id ?? project_id;
      if (!pid) throw new Error('project id missing');
      const uid = userId || user_id;
      if (!uid) throw new Error('user id missing');
      return fetchUserById(httpClient, AUTH_BASE_URL, pid, uid!);
    },
    enabled: !!(projectQuery.data?.id || projectId || project_id) && !!(userId || user_id),
  });

  useEffect(() => {
    if (userQuery.data) {
      setDisplayName(userQuery.data.display_name ?? '');
      setEmail(userQuery.data.email ?? '');
      setExternalUserId(userQuery.data.external_user_id ?? '');
      setIsActive(typeof userQuery.data.is_active === 'boolean' ? userQuery.data.is_active : userQuery.data.is_active === 1);
      setRoleId(userQuery.data.role_id ?? '');
      setUserId(userQuery.data.id ?? '');
      setAvatarPreview(userQuery.data.avatar_url ?? null);
      // Reset touched state when data loads
      setDisplayNameTouched(false);
      setEmailTouched(false);
      setPasswordTouched(false);
      setExternalUserIdTouched(false);
    }
  }, [userQuery.data]);

  // Validation logic using shared utilities (isUpdate=true for updates)
  const validation = useMemo<UserValidationResult>(
    () => validateUserForm(displayName, email, password, externalUserId, true),
    [displayName, email, password, externalUserId]
  );

  // Handle display name change with validation
  const handleDisplayNameChange = (value: string) => {
    setDisplayName(truncateToMaxLength(value, USER_VALIDATION.name.maxLength));
  };

  // Handle email change with validation
  const handleEmailChange = (value: string) => {
    setEmail(truncateToMaxLength(value, USER_VALIDATION.email.maxLength));
  };

  // Handle password change with validation
  const handlePasswordChange = (value: string) => {
    setPassword(value);
  };

  // Handle external user ID change with validation
  const handleExternalUserIdChange = (value: string) => {
    setExternalUserId(truncateToMaxLength(value, USER_VALIDATION.externalUserId.maxLength));
  };

  const queryClient = useQueryClient();

  const updateMutation = useMutation({
    mutationFn: (payload: { display_name: string; email: string; password?: string; external_user_id?: string; is_active: boolean; role_id?: string; avatar_url?: string | null }) => {
      const pid = projectId ?? projectQuery.data?.id;
      const uid = userId;
      if (!pid || !uid) return Promise.reject(new Error('project or user id missing'));
      return updateProjectUser(httpClient, AUTH_BASE_URL, pid, uid, payload);
    },
    onMutate: () => { },
    onSuccess: (data) => {
      const pid = projectId ?? projectQuery.data?.id;
      telemetry.trackEvent('user_updated', {
        user_id: userId,
        project_id: pid,
        has_role: Boolean(roleId),
        is_active: isActive,
      });
      queryClient.setQueryData(['user', pid, userId], data);
      toast({ title: 'Saved', description: 'User updated', variant: 'default' });
      // redirect to project details after successful update (preserve query state)
      if (pid) {
        navigate(`/projects/${pid}/details${getParamsFrom()}`);
      }
    },
    onError: (err: any) => {
      toast({ title: 'Save failed', description: err?.message ?? 'Unable to update user', variant: 'destructive' });
    },
  });

  const [showRemoveDialog, setShowRemoveDialog] = useState(false);
  const [removeLoading, setRemoveLoading] = useState(false);

  const confirmRemoveUser = async () => {
    const pid = projectId ?? projectQuery.data?.id;
    if (!pid) {
      toast({ title: 'Remove failed', description: 'Project id missing', variant: 'destructive' });
      return;
    }
    setRemoveLoading(true);
    try {
      await deleteProjectUser(httpClient, AUTH_BASE_URL, pid, userId);
      telemetry.trackEvent('user_deleted', { user_id: userId, project_id: pid });
      queryClient.invalidateQueries({ queryKey: ['project', pid, 'users'] });
      toast({ title: 'User removed', description: 'User has been removed', variant: 'default' });
      navigate(`/projects/${pid}/details${getParamsFrom()}`);
    } catch (err: any) {
      toast({ title: 'Remove failed', description: err?.message ?? 'Unable to remove user', variant: 'destructive' });
    } finally {
      setRemoveLoading(false);
      setShowRemoveDialog(false);
    }
  };

  const onAvatarFile = async (f: File | null) => {
    if (!f) return;
    try {
      setUploadingAvatar(true);
      const res = await uploadUserAvatar(httpClient, AUTH_BASE_URL, f);
      setAvatarPreview(res.url);
      toast({ title: 'Upload successful', description: 'Avatar uploaded', variant: 'default' });
    } catch (err: any) {
      toast({ title: 'Upload failed', description: err?.message ?? 'Unable to upload avatar', variant: 'destructive' });
    } finally {
      setUploadingAvatar(false);
    }
  };

  const doUploadSelectedAvatar = async () => {
    if (!selectedAvatarFile) return;
    await onAvatarFile(selectedAvatarFile);
    setShowUploadDialog(false);
    setSelectedAvatarFile(null);
  };

  const onRemoveAvatar = async () => {
    setAvatarPreview(null);
  };

  if (userQuery.isLoading || projectQuery.isLoading) {
    return (
      <div className="flex items-center justify-center py-20" data-testid="user-loading">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        <span className="ml-2 text-sm text-muted-foreground">Loading…</span>
      </div>
    );
  }

  if (userQuery.isError || projectQuery.isError || !projectQuery.data) {
    return (
      <div className="flex flex-col items-center justify-center py-20 space-y-2" data-testid="user-error">
        <AlertCircle className="h-8 w-8 text-destructive" />
        <p className="text-sm text-destructive font-medium">Failed to load user</p>
        <p className="text-xs text-muted-foreground">{userQuery.error?.message ?? projectQuery.error?.message ?? 'User data is unavailable'}</p>
      </div>
    );
  }

  const roles = rolesQuery.data?.data ?? [];

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
          {userQuery.data?.avatar_url ? (
            <img src={userQuery.data.avatar_url} alt={userQuery.data?.display_name ?? ''} className="h-12 w-12 rounded object-cover border border-border" />
          ) : (
            <div className="h-12 w-12 rounded bg-muted flex items-center justify-center text-sm font-bold text-muted-foreground">
              {(userQuery.data?.display_name || '?').charAt(0).toUpperCase()}
            </div>
          )}
          <div>
            {/* Header shows server value only; local edits do not update header */}
            <h1 className="text-lg font-medium flex items-center gap-3">{userQuery.data?.display_name ?? 'User'}</h1>
            <div className="text-sm text-muted-foreground">
              {userQuery.data?.email}
            </div>
          </div>
        </div>
        <div />
      </div>

      <div className="border rounded p-4 bg-card w-full">
        <h3 className="text-lg font-medium">User Information</h3>
        <div className="text-sm text-muted-foreground">View or edit the user information and manage their role</div>
        <div className="grid grid-cols-12 gap-4 mt-4">
          <div className="col-span-4 flex items-center"><Label>Avatar</Label></div>
          <div className="col-span-8 flex items-center justify-start gap-3">
            <div className="relative">
              {avatarPreview ? (
                <img src={avatarPreview} alt={displayName || userQuery.data?.display_name} className="h-20 w-20 rounded object-cover" />
              ) : (
                <div className="h-20 w-20 rounded bg-muted flex items-center justify-center text-lg font-bold text-muted-foreground">
                  {(displayName || userQuery.data?.display_name || '?').charAt(0).toUpperCase()}
                </div>
              )}
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button size="sm" variant="ghost" className="h-7 w-7 p-0">
                  <MoreVertical className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => { setSelectedAvatarFile(null); setShowUploadDialog(true); }}>
                  <CloudUpload className="h-4 w-4 mr-2" />
                  Upload
                </DropdownMenuItem>
                <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => onRemoveAvatar()} disabled={!avatarPreview && !userQuery.data?.avatar_url}>
                  <Trash2 className="h-4 w-4 mr-2" />
                  Remove
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          <div className="col-span-4 flex items-center"><Label htmlFor="user-name">Name</Label></div>
          <div className="col-span-8">
            <Input
              id="user-name"
              placeholder="Enter user name"
              value={displayName}
              onChange={(e) => {
                handleDisplayNameChange(e.target.value);
                if (!displayNameTouched) setDisplayNameTouched(true);
              }}
              onBlur={() => setDisplayNameTouched(true)}
              className={displayNameTouched && validation.errors.name ? 'border-destructive' : ''}
              maxLength={USER_VALIDATION.name.maxLength}
              aria-invalid={displayNameTouched && !!validation.errors.name}
            />
            {displayNameTouched && validation.errors.name && (
              <div className="text-xs text-destructive mt-1">{validation.errors.name}</div>
            )}
          </div>

          <div className="col-span-4 flex items-center"><Label htmlFor="user-email">Email</Label></div>
          <div className="col-span-8">
            <Input
              id="user-email"
              type="email"
              placeholder="Enter email"
              value={email}
              onChange={(e) => {
                handleEmailChange(e.target.value);
                if (!emailTouched) setEmailTouched(true);
              }}
              onBlur={() => setEmailTouched(true)}
              className={emailTouched && validation.errors.email ? 'border-destructive' : ''}
              maxLength={USER_VALIDATION.email.maxLength}
              aria-invalid={emailTouched && !!validation.errors.email}
            />
            {emailTouched && validation.errors.email && (
              <div className="text-xs text-destructive mt-1">{validation.errors.email}</div>
            )}
          </div>

          <div className="col-span-4 flex items-center"><Label htmlFor="user-password">Password</Label></div>
          <div className="col-span-8">
            <div className="relative">
              <Input
                id="user-password"
                className={`pr-10 ${passwordTouched && validation.errors.password ? 'border-destructive' : ''}`}
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(e) => {
                  handlePasswordChange(e.target.value);
                  if (!passwordTouched) setPasswordTouched(true);
                }}
                onBlur={() => setPasswordTouched(true)}
                placeholder="Leave blank to keep current password"
                aria-invalid={passwordTouched && !!validation.errors.password}
              />
              <button
                aria-label={showPassword ? 'Hide password' : 'Show password'}
                onClick={() => setShowPassword(s => !s)}
                className="absolute right-2 top-1/2 -translate-y-1/2 p-1 rounded text-sm text-muted-foreground hover:bg-muted/5"
                type="button"
              >
                {showPassword ? <Eye className="h-4 w-4" /> : <EyeOff className="h-4 w-4" />}
              </button>
            </div>
            {passwordTouched && validation.errors.password && (
              <div className="text-xs text-destructive mt-1">{validation.errors.password}</div>
            )}
          </div>

          <div className="col-span-4 flex items-center"><Label htmlFor="user-external-id">External User ID</Label></div>
          <div className="col-span-8">
            <Input
              id="user-external-id"
              placeholder="Enter external user ID"
              value={externalUserId}
              onChange={(e) => {
                handleExternalUserIdChange(e.target.value);
                if (!externalUserIdTouched) setExternalUserIdTouched(true);
              }}
              onBlur={() => setExternalUserIdTouched(true)}
              className={externalUserIdTouched && validation.errors.externalUserId ? 'border-destructive' : ''}
              maxLength={USER_VALIDATION.externalUserId.maxLength}
              aria-invalid={externalUserIdTouched && !!validation.errors.externalUserId}
            />
            {externalUserIdTouched && validation.errors.externalUserId && (
              <div className="text-xs text-destructive mt-1">{validation.errors.externalUserId}</div>
            )}
          </div>

          <div className="col-span-4 flex items-center"><Label>Status</Label></div>
          <div className="col-span-8">
            <Select value={isActive ? 'active' : 'inactive'} onValueChange={(v) => setIsActive(v === 'active')}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Select status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="active">Active</SelectItem>
                <SelectItem value="inactive">Inactive</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="col-span-4 flex items-center"><Label>Role</Label></div>
          <div className="col-span-8">
            <Select value={roleId || 'no-role'} onValueChange={(v) => setRoleId(v === 'no-role' ? '' : v)}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Select a role" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="no-role">No Role</SelectItem>
                {roles.map((role: any) => (
                  <SelectItem key={role.id} value={role.id}>{role.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="col-span-12 flex justify-end pt-4">
            <div className="flex items-center gap-2">
              <Button size="sm" variant="destructive" onClick={() => setShowRemoveDialog(true)} disabled={removeLoading}>
                Remove
              </Button>
              <Button
                size="sm"
                onClick={() => {
                  // Mark all fields as touched to show validation errors
                  setDisplayNameTouched(true);
                  setEmailTouched(true);
                  setPasswordTouched(true);
                  setExternalUserIdTouched(true);
                  if (validation.isValid) {
                    const payload: any = {
                      display_name: displayName,
                      email,
                      is_active: isActive,
                      role_id: roleId || undefined,
                      external_user_id: externalUserId || undefined,
                      avatar_url: avatarPreview,
                    };
                    if (password) payload.password = password;
                    updateMutation.mutate(payload);
                  }
                }}
                disabled={updateMutation.status === 'pending' || !validation.isValid || uploadingAvatar}
              >
                {updateMutation.status === 'pending' ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin mr-2" />
                    Saving...
                  </>
                ) : (
                  <>Save</>
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
        title="Remove User"
        description={<span>Are you sure you want to remove <span className="font-medium">{displayName || userQuery.data?.display_name}</span>? This action cannot be undone.</span>}
        confirmLabel={removeLoading ? (<><Loader2 className="h-4 w-4 animate-spin mr-2" />Removing...</>) : 'Remove'}
        destructive
        disabled={removeLoading}
        onConfirm={confirmRemoveUser}
      />
      <ConfirmDialog
        open={showUploadDialog}
        onOpenChange={(open) => { if (!open) { setShowUploadDialog(false); setSelectedAvatarFile(null); } }}
        closeOnConfirm={false}
        title="Upload Avatar"
        description="Select an image file to use as your avatar."
        confirmLabel="Upload"
        onConfirm={doUploadSelectedAvatar}
        disabled={!selectedAvatarFile || uploadingAvatar}
      >
        <div className="py-4">
          {uploadingAvatar ? (
            <div className="flex items-center gap-3 justify-center py-6">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
              <div className="text-sm">Uploading...</div>
            </div>
          ) : (
            <div
              onDragOver={(e) => e.preventDefault()}
              onDrop={(e) => { e.preventDefault(); const f = e.dataTransfer?.files?.[0]; if (f) setSelectedAvatarFile(f); }}
              className="border-2 border-dashed rounded p-6 text-center"
            >
              {selectedAvatarFile ? (
                <div className="space-y-2">
                  <div className="font-medium">{selectedAvatarFile.name}</div>
                  <div className="text-sm text-muted-foreground">{Math.round(selectedAvatarFile.size / 1024)} KB</div>
                </div>
              ) : (
                <div className="text-sm text-muted-foreground">Drag & drop an image here, or click to select</div>
              )}
              <input id="avatar-upload-input" ref={(el) => (inputUploadRef.current = el)} type="file" accept="image/*" className="hidden" onChange={(e) => setSelectedAvatarFile(e.target.files?.[0] ?? null)} />
              <div className="block mt-3">
                <Button size="sm" onClick={() => inputUploadRef.current?.click()}>Select File</Button>
              </div>
              <div className="mt-3 text-xs text-muted-foreground">Accepted formats: JPG, PNG, GIF</div>
            </div>
          )}
        </div>
      </ConfirmDialog>
    </div>
  );
}
