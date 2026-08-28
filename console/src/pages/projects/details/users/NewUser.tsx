import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { ConfirmDialog } from '@/components/ConfirmDialog';
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
import { createProjectUser, fetchProjectById, fetchProjectRoles, uploadUserAvatar } from '@/services/projects';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { telemetry } from '@/lib/telemetry';
import { AlertCircle, ArrowLeft, Eye, EyeOff, Loader2, MoreVertical, CloudUpload, Trash2 } from 'lucide-react';
import { useMemo, useRef, useState } from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';

interface Props { httpClient: HttpClient }

export default function NewUser({ httpClient }: Props) {
  const { project_id } = useParams<{ project_id: string }>();
  const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL as string;
  const navigate = useNavigate();
  const location = useLocation();
  const queryClient = useQueryClient();
  const { toast } = useToast();

  const stateAny = (location.state || {}) as any;
  const paramsFrom = getProjectTabParams(project_id ?? '', 'users') || stateAny?.from || '';

  const projectQuery = useQuery({
    queryKey: ['project', project_id],
    queryFn: () => fetchProjectById(httpClient, AUTH_BASE_URL, project_id!),
    enabled: !!project_id,
  });

  const project = projectQuery.data;
  const projectLoading = projectQuery.isLoading;

  const rolesQuery = useQuery({
    queryKey: ['project', project?.id, 'roles'],
    queryFn: () => fetchProjectRoles(httpClient, AUTH_BASE_URL, project!.id, 1, 100, 'name', 'asc'),
    enabled: !!project?.id,
  });

  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [externalUserId, setExternalUserId] = useState('');
  const [roleId, setRoleId] = useState<string>('');

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

  // Validation logic using shared utilities
  const validation = useMemo<UserValidationResult>(
    () => validateUserForm(displayName, email, password, externalUserId, false),
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

  const createMutation = useMutation({
    mutationFn: (payload: { display_name: string; email?: string; password?: string; external_user_id?: string; role_id?: string; avatar_url?: string | null }) => {
      return createProjectUser(httpClient, AUTH_BASE_URL, project!.id, payload);
    },
    onSuccess: () => {
      telemetry.trackEvent('user_created', {
        project_id: project?.id,
        has_role: Boolean(roleId),
        has_external_user_id: Boolean(externalUserId),
      });
      queryClient.invalidateQueries({ queryKey: ['project', project?.id, 'users'] });
      toast({ title: 'User created', description: 'New user has been added', variant: 'default' });
      navigate(`/projects/${project?.id}/details${paramsFrom}`, { state: { from: paramsFrom } });
    },
    onError: (err: any) => {
      toast({ title: 'Create user failed', description: err?.message ?? 'Unable to create user', variant: 'destructive' });
    },
  });

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

  if (projectLoading) {
    return (
      <div className="flex items-center justify-center py-20" data-testid="newuser-loading">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        <span className="ml-2 text-sm text-muted-foreground">Loading…</span>
      </div>
    );
  }

  if (projectQuery.isError || !project) {
    return (
      <div className="flex flex-col items-center justify-center py-20 space-y-2" data-testid="newuser-error">
        <AlertCircle className="h-8 w-8 text-destructive" />
        <p className="text-sm text-destructive font-medium">Failed to load project</p>
        <p className="text-xs text-muted-foreground">{projectQuery.error?.message ?? 'Project data is unavailable'}</p>
      </div>
    );
  }

  const roles = rolesQuery.data?.data ?? [];
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
            <h1 className="text-lg font-medium">Add New User</h1>
            <div className="text-sm text-muted-foreground">Create a new user for {project?.name}</div>
          </div>
        </div>
      </div>

      <div className="border rounded p-4 bg-card w-full">
        <h3 className="text-lg font-medium">User Information</h3>
        <div className="text-sm text-muted-foreground">Set the basic information for the new user and assign a role</div>
        <div className="grid grid-cols-12 gap-4 mt-4">
          <div className="col-span-4 flex items-center"><Label>Avatar</Label></div>
          <div className="col-span-8 flex items-center justify-start gap-3">
            <div className="relative">
              {avatarPreview ? (
                <img src={avatarPreview} alt={displayName} className="h-20 w-20 rounded object-cover" />
              ) : (
                <div className="h-20 w-20 rounded bg-muted flex items-center justify-center text-lg font-bold text-muted-foreground">
                  {(displayName || 'U').charAt(0).toUpperCase()}
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
                <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => onRemoveAvatar()} disabled={!avatarPreview}>
                  <Trash2 className="h-4 w-4 mr-2" />
                  Remove
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          <div className="col-span-4 flex items-center">
            <Label htmlFor="user-name">Name</Label>
          </div>
          <div className="col-span-8">
            <Input
              id="user-name"
              placeholder="Enter user name"
              className={`${displayNameTouched && validation.errors.name ? 'border-destructive' : ''}`}
              value={displayName}
              onChange={(e) => {
                handleDisplayNameChange(e.target.value);
                if (!displayNameTouched) setDisplayNameTouched(true);
              }}
              onBlur={() => setDisplayNameTouched(true)}
              maxLength={USER_VALIDATION.name.maxLength}
              aria-invalid={displayNameTouched && !!validation.errors.name}
            />
            {displayNameTouched && validation.errors.name && (
              <div className="text-xs text-destructive mt-1">{validation.errors.name}</div>
            )}
          </div>

          <div className="col-span-4 flex items-center">
            <Label htmlFor="user-email">Email</Label>
          </div>
          <div className="col-span-8">
            <Input
              id="user-email"
              type="email"
              placeholder="Enter email"
              className={`${emailTouched && validation.errors.email ? 'border-destructive' : ''}`}
              value={email}
              onChange={(e) => {
                handleEmailChange(e.target.value);
                if (!emailTouched) setEmailTouched(true);
              }}
              onBlur={() => setEmailTouched(true)}
              maxLength={USER_VALIDATION.email.maxLength}
              aria-invalid={emailTouched && !!validation.errors.email}
            />
            {emailTouched && validation.errors.email && (
              <div className="text-xs text-destructive mt-1">{validation.errors.email}</div>
            )}
          </div>

          <div className="col-span-4 flex items-center">
            <Label htmlFor="user-password">Password</Label>
          </div>
          <div className="col-span-8">
            <div className="relative">
              <Input
                id="user-password"
                className={`pr-10 ${passwordTouched && validation.errors.password ? 'border-destructive' : ''}`}
                type={showPassword ? 'text' : 'password'}
                value={password}
                placeholder="Enter password"
                onChange={(e) => {
                  handlePasswordChange(e.target.value);
                  if (!passwordTouched) setPasswordTouched(true);
                }}
                onBlur={() => setPasswordTouched(true)}
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

          <div className="col-span-4 flex items-center">
            <Label htmlFor="user-external-id">External User ID</Label>
          </div>
          <div className="col-span-8">
            <Input
              id="user-external-id"
              placeholder="Enter external user ID"
              className={`${externalUserIdTouched && validation.errors.externalUserId ? 'border-destructive' : ''}`}
              value={externalUserId}
              onChange={(e) => {
                handleExternalUserIdChange(e.target.value);
                if (!externalUserIdTouched) setExternalUserIdTouched(true);
              }}
              onBlur={() => setExternalUserIdTouched(true)}
              maxLength={USER_VALIDATION.externalUserId.maxLength}
              aria-invalid={externalUserIdTouched && !!validation.errors.externalUserId}
            />
            {externalUserIdTouched && validation.errors.externalUserId && (
              <div className="text-xs text-destructive mt-1">{validation.errors.externalUserId}</div>
            )}
          </div>

          <div className="col-span-4 flex items-center">
            <Label>Role</Label>
          </div>
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
                      email: email || undefined,
                      password: password || undefined,
                      external_user_id: externalUserId || undefined,
                      role_id: roleId || undefined,
                      avatar_url: avatarPreview || undefined,
                    };
                    createMutation.mutate(payload);
                  }
                }}
                disabled={creating || !validation.isValid || uploadingAvatar}
              >
                {creating ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin mr-2" />
                    Creating...
                  </>
                ) : (
                  <>Create</>
                )}
              </Button>
            </div>
          </div>
        </div>
      </div>
      <ConfirmDialog
        open={showUploadDialog}
        onOpenChange={(open) => { if (!open) { setShowUploadDialog(false); setSelectedAvatarFile(null); } }}
        closeOnConfirm={false}
        title="Upload Avatar"
        description="Select an image file to use as the user's avatar."
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
