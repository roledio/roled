import { useEffect, useMemo, useRef, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@/components/ui/dropdown-menu';
import { useToast } from '@/hooks/use-toast';
import { useCurrentUser } from '@/hooks/users';
import type { CurrentTokenAndMemberInfo } from '@/hooks/use-current-token-info';
import type { CurrentTokenInfo } from '@/services/core/authService';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ConfirmDialog } from '@/components/ConfirmDialog';
import { MoreVertical, CloudUpload, Trash2, Loader2, Eye, EyeOff } from 'lucide-react';
import type { HttpClient } from '@/services/core/httpClient';
import { updateCurrentUser, type CurrentUser, type UpdateCurrentUserPayload } from '@/services/users';
import { uploadUserAvatar } from '@/services/projects/projects';
import {
  USER_VALIDATION,
  truncateToMaxLength,
  validateProfileForm,
  type ProfileValidationResult,
} from '@/lib/validation';

interface Props { httpClient: HttpClient }

export default function Profile({ httpClient }: Props) {
  const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL as string;
  const { user, isLoading, error } = useCurrentUser({ httpClient, baseUrl: AUTH_BASE_URL });
  const queryClient = useQueryClient();
  const { toast } = useToast();

  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [avatarPreview, setAvatarPreview] = useState<string | null>(null);
  const [showUploadDialog, setShowUploadDialog] = useState(false);
  const [selectedAvatarFile, setSelectedAvatarFile] = useState<File | null>(null);
  const inputUploadRef = useRef<HTMLInputElement | null>(null);
  const [uploadingAvatar, setUploadingAvatar] = useState(false);

  // Validation touched state
  const [displayNameTouched, setDisplayNameTouched] = useState(false);
  const [emailTouched, setEmailTouched] = useState(false);
  const [passwordTouched, setPasswordTouched] = useState(false);

  useEffect(() => {
    if (user) {
      setDisplayName(user.display_name ?? '');
      setEmail(user.email ?? '');
      setPassword('');
      setAvatarPreview(user.avatar_url ?? null);
      // Reset touched state when data loads
      setDisplayNameTouched(false);
      setEmailTouched(false);
      setPasswordTouched(false);
    }
  }, [user]);

  // Validation logic using shared utilities
  const validation = useMemo<ProfileValidationResult>(
    () => validateProfileForm(displayName, email, password),
    [displayName, email, password]
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

  const mutation = useMutation<CurrentUser, Error, UpdateCurrentUserPayload>({
    mutationFn: (payload: UpdateCurrentUserPayload) => {
      if (!user) return Promise.reject(new Error('user missing'));
      return updateCurrentUser(httpClient, AUTH_BASE_URL, payload);
    },
    onSuccess: (updatedUser) => {
      queryClient.setQueryData<CurrentUser>(['user', 'current'], updatedUser);
      queryClient.setQueryData<CurrentTokenAndMemberInfo | null>(['currentTokenAndMemberInfo'], (current) => {
        if (!current) return current;
        return {
          ...current,
          tokenInfo: {
            ...current.tokenInfo,
            user: {
              ...current.tokenInfo.user,
              display_name: updatedUser.display_name,
              email: updatedUser.email,
              avatar_url: updatedUser.avatar_url ?? null,
            },
          },
        };
      });
      queryClient.setQueryData<CurrentTokenInfo | null>(['currentTokenInfo'], (current) => {
        if (!current) return current;
        return {
          ...current,
          user: {
            ...current.user,
            display_name: updatedUser.display_name,
            email: updatedUser.email,
            avatar_url: updatedUser.avatar_url ?? null,
          },
        };
      });

      const currentTokenInfo = httpClient.tokenServiceRef.getCurrentTokenInfo() as CurrentTokenInfo | null;
      if (currentTokenInfo) {
        httpClient.tokenServiceRef.setCurrentTokenInfo({
          ...currentTokenInfo,
          user: {
            ...currentTokenInfo.user,
            display_name: updatedUser.display_name,
            email: updatedUser.email,
            avatar_url: updatedUser.avatar_url ?? null,
          },
        });
      }

      toast({ title: 'Saved', description: 'Profile updated', variant: 'default' });
    },
    onError: (err: Error) => {
      toast({ title: 'Save failed', description: err.message ?? 'Unable to update profile', variant: 'destructive' });
    },
  });

  const handleSave = () => {
    if (!validation.isValid) return;
    const payload: UpdateCurrentUserPayload = {
      display_name: displayName,
      email,
    };
    if (password) payload.password = password;
    if (avatarPreview) payload.avatar_url = avatarPreview;
    else if (avatarPreview === null && user.avatar_url) payload.avatar_url = null;
    mutation.mutate(payload);
  };

  if (isLoading) return <div className="flex items-center justify-center py-20">Loading…</div>;
  if (error || !user) return <div className="flex flex-col items-center justify-center py-20">Failed to load profile</div>;

  const avatarLetter = (user.display_name || '?').charAt(0).toUpperCase();
  const headerDisplayName = user.display_name;
  const headerEmail = user.email;

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
    // set preview to null to indicate removal
    setAvatarPreview(null);
  };

  return (
    <div className="max-w-4xl space-y-6">
      <div className="flex items-center gap-4">
        {user.avatar_url ? (
          <img src={user.avatar_url} alt={headerDisplayName} className="h-12 w-12 rounded object-cover border border-border" />
        ) : (
          <div className="h-12 w-12 rounded bg-muted flex items-center justify-center text-sm font-bold text-muted-foreground">{avatarLetter}</div>
        )}
        <div>
          <h2 className="text-lg font-semibold">{headerDisplayName}</h2>
          <div className="text-sm text-muted-foreground">{headerEmail}</div>
        </div>
      </div>

      <div className="border rounded p-4 bg-card w-full">
        <h3 className="text-lg font-medium">Profile</h3>
        <div className="text-sm text-muted-foreground mb-4">View or edit your profile information</div>

        <div className="grid grid-cols-12 gap-4 mt-4">
          <div className="col-span-4 flex items-center"><Label>Avatar</Label></div>
          <div className="col-span-8 flex items-center justify-start gap-3">
            <div className="relative">
              {avatarPreview ? (
                <img src={avatarPreview} alt={displayName || user.display_name} className="h-20 w-20 rounded object-cover" />
              ) : (
                <div className="h-20 w-20 rounded bg-muted flex items-center justify-center text-lg font-bold text-muted-foreground">{avatarLetter}</div>
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
                <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => onRemoveAvatar()} disabled={!avatarPreview && !user.avatar_url}>
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
              placeholder="Enter your name"
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
              placeholder="Enter your email"
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
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(e) => {
                  handlePasswordChange(e.target.value);
                  if (!passwordTouched) setPasswordTouched(true);
                }}
                onBlur={() => setPasswordTouched(true)}
                placeholder="Leave empty to keep current password"
                className={`pr-10 ${passwordTouched && validation.errors.password ? 'border-destructive' : ''}`}
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

          <div className="col-span-12 flex justify-end pt-4">
            <div className="flex items-center gap-2">
              <Button size="sm" onClick={handleSave} disabled={!validation.isValid || mutation.status === 'pending' || uploadingAvatar}>
                {mutation.status === 'pending' ? (
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
