import { ConfirmDialog } from '@/components/ConfirmDialog';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { useToast } from '@/hooks/use-toast';
import logger from '@/lib/logger';
import type { HttpClient } from '@/services/core/httpClient';
import { createProject, uploadProjectLogo } from '@/services/projects';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { telemetry } from '@/lib/telemetry';
import { ArrowLeft, CloudUpload, Loader2, MoreVertical, Trash2 } from 'lucide-react';
import { useMemo, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

import {
  PROJECT_VALIDATION,
  truncateToMaxLength,
  validateProjectForm,
  type ProjectValidationResult
} from '@/lib/validation';

// Image upload constants
const ACCEPTED_IMAGE_FORMATS = ['image/jpeg', 'image/png', 'image/gif'];
const ACCEPTED_IMAGE_EXTENSIONS = '.jpg, .jpeg, .png, .gif';

interface Props {
  httpClient: HttpClient;
}

export default function NewProject({ httpClient }: Props) {
  const navigate = useNavigate();
  const location = useLocation();
  const queryClient = useQueryClient();
  const { toast } = useToast();

  const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL as string;

  // Form state
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [logoUrl, setLogoUrl] = useState<string | null>(null);
  // Initialize with one empty redirect URI (required field)
  const [redirectUris, setRedirectUris] = useState<Array<{ redirect_uri: string; login_url?: string }>>([
    { redirect_uri: '', login_url: '' },
  ]);

  // Validation state
  const [nameTouched, setNameTouched] = useState(false);
  const [descriptionTouched, setDescriptionTouched] = useState(false);
  const [redirectUrisTouched, setRedirectUrisTouched] = useState(false);

  // Logo upload state
  const [uploadingLogo, setUploadingLogo] = useState(false);
  const [showUploadDialog, setShowUploadDialog] = useState(false);
  const [selectedLogoFile, setSelectedLogoFile] = useState<File | null>(null);

  const stateAny = (location.state || {}) as any;
  const paramsFrom = (stateAny?.from as string) || '';

  const createMutation = useMutation({
    mutationFn: (payload: {
      name: string;
      description?: string | null;
      logo_url?: string | null;
      redirect_uris?: Array<{ redirect_uri: string; login_url?: string }>;
    }) => {
      return createProject(httpClient, AUTH_BASE_URL, payload);
    },
    onSuccess: (data) => {
      telemetry.trackEvent('project_created', { project_id: data.id });
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      toast({ title: 'Project created', description: 'New project has been created', variant: 'default' });
      // Redirect to project details page
      navigate(`/projects/${data.id}/details?tab=project`, { state: { from: paramsFrom } });
    },
    onError: (err: any) => {
      toast({
        title: 'Create project failed',
        description: err?.message ?? 'Unable to create project',
        variant: 'destructive',
      });
    },
  });

  const onLogoFile = async (file?: File) => {
    if (!file) return;
    setUploadingLogo(true);
    try {
      const res = await uploadProjectLogo(httpClient, AUTH_BASE_URL, '', file);
      setLogoUrl(res.logo_url ?? null);
    } catch (err: any) {
      logger.error('Upload failed:', err);
      toast({
        title: 'Upload failed',
        description: err?.message ?? 'Unable to upload logo',
        variant: 'destructive',
      });
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

  const onRemoveLogo = () => {
    setLogoUrl(null);
  };

  const handleDeleteRedirectUri = (index: number) => {
    setRedirectUris((prev) => prev.filter((_, idx) => idx !== index));
  };

  const handleCreate = () => {
    const payload: {
      name: string;
      description?: string | null;
      logo_url?: string | null;
      redirect_uris?: Array<{ redirect_uri: string; login_url?: string }>;
    } = {
      name,
      description: description || null,
    };

    if (logoUrl) {
      payload.logo_url = logoUrl;
    }

    const filteredRedirectUris = redirectUris
      .map((r) => ({
        redirect_uri: (r.redirect_uri || '').trim(),
        login_url: (r.login_url || '').trim(),
      }))
      .filter((r) => r.redirect_uri.length > 0 || r.login_url.length > 0);

    if (filteredRedirectUris.length > 0) {
      payload.redirect_uris = filteredRedirectUris;
    }

    createMutation.mutate(payload);
  };

  const creating = createMutation.status === 'pending';

  // Validation logic using shared utilities
  const validation = useMemo<ProjectValidationResult>(
    () => validateProjectForm(name, description, redirectUris),
    [name, description, redirectUris]
  );

  // Handle name change with validation
  const handleNameChange = (value: string) => {
    setName(truncateToMaxLength(value, PROJECT_VALIDATION.name.maxLength));
  };

  // Handle description change with validation
  const handleDescriptionChange = (value: string) => {
    setDescription(truncateToMaxLength(value, PROJECT_VALIDATION.description.maxLength));
  };

  // Handle file selection with validation
  const handleFileSelect = (file: File | null) => {
    if (!file) {
      setSelectedLogoFile(null);
      return;
    }

    // Validate file type
    if (!ACCEPTED_IMAGE_FORMATS.includes(file.type)) {
      toast({
        title: 'Invalid file format',
        description: `Please upload a valid image file (${ACCEPTED_IMAGE_EXTENSIONS})`,
        variant: 'destructive',
      });
      return;
    }

    setSelectedLogoFile(file);
  };

  return (
    <div className="space-y-6 max-w-4xl">
      <div>
        <Button
          variant="secondary"
          size="sm"
          onClick={() => {
            navigate(`/projects${paramsFrom}`);
          }}
          className="px-3"
        >
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to Projects
        </Button>
      </div>

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          {logoUrl ? (
            <img src={logoUrl} alt={name || 'New Project'} className="h-12 w-12 rounded object-cover" />
          ) : (
            <div className="h-12 w-12 rounded bg-muted flex items-center justify-center text-sm font-bold text-muted-foreground">
              {name?.charAt(0) || 'P'}
            </div>
          )}
          <div>
            <h1 className="text-lg font-medium">Add New Project</h1>
            <div className="text-sm text-muted-foreground">Let's create a new project to integrate with Roled</div>
          </div>
        </div>
      </div>

      <div className="border rounded p-4 bg-card w-full">
        <h3 className="text-lg font-medium">Project Information</h3>
        <div className="text-sm text-muted-foreground">Set the basic information for the new project</div>

        <div className="mt-4">
          <div className="grid grid-cols-12 gap-4 items-center mb-4">
            <div className="col-span-4 flex items-center">
              <Label>Logo</Label>
            </div>
            <div className="col-span-8 flex items-center justify-start gap-3">
              <div className="relative">
                {logoUrl ? (
                  <img src={logoUrl} alt={name || 'Project Logo'} className="h-20 w-20 rounded object-cover" />
                ) : (
                  <div className="h-20 w-20 rounded bg-muted flex items-center justify-center text-lg font-bold text-muted-foreground">
                    {name?.charAt(0) || 'P'}
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
                  <DropdownMenuItem
                    onClick={() => {
                      setSelectedLogoFile(null);
                      setShowUploadDialog(true);
                    }}
                  >
                    <CloudUpload className="h-4 w-4 mr-2" />
                    Upload
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    className="text-destructive focus:text-destructive"
                    onClick={() => onRemoveLogo()}
                    disabled={!logoUrl}
                  >
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
          onOpenChange={(open) => {
            if (!open) {
              setShowUploadDialog(false);
              setSelectedLogoFile(null);
            }
          }}
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
                onDrop={(e) => {
                  e.preventDefault();
                  const f = e.dataTransfer?.files?.[0];
                  if (f) handleFileSelect(f);
                }}
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
                <input
                  id="logo-upload-input"
                  type="file"
                  accept={ACCEPTED_IMAGE_EXTENSIONS}
                  className="hidden"
                  onChange={(e) => handleFileSelect(e.target.files?.[0] ?? null)}
                />
                <div className="block mt-3">
                  <Button size="sm" onClick={() => document.getElementById('logo-upload-input')?.click()}>
                    Select File
                  </Button>
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
            <Label htmlFor="project-name">Name</Label>
          </div>
          <div className="col-span-8">
            <Input
              id="project-name"
              value={name}
              onChange={(e) => {
                handleNameChange(e.target.value);
                if (!nameTouched) setNameTouched(true);
              }}
              onBlur={() => setNameTouched(true)}
              placeholder="Enter project name"
              className={nameTouched && validation.errors.name ? 'border-destructive' : ''}
              aria-invalid={nameTouched && !!validation.errors.name}
              aria-describedby={nameTouched && validation.errors.name ? 'name-error' : undefined}
              maxLength={PROJECT_VALIDATION.name.maxLength}
            />
            {nameTouched && validation.errors.name && (
              <div id="name-error" className="text-xs text-destructive mt-1">
                {validation.errors.name}
              </div>
            )}
          </div>

          <div className="col-span-4 flex items-start pt-1">
            <Label htmlFor="project-description">Description</Label>
          </div>
          <div className="col-span-8">
            <Textarea
              id="project-description"
              rows={3}
              value={description}
              onChange={(e) => {
                handleDescriptionChange(e.target.value);
                if (!descriptionTouched) setDescriptionTouched(true);
              }}
              onBlur={() => setDescriptionTouched(true)}
              placeholder="Enter project description (optional)"
              className={descriptionTouched && validation.errors.description ? 'border-destructive' : ''}
              aria-invalid={descriptionTouched && !!validation.errors.description}
              aria-describedby={descriptionTouched && validation.errors.description ? 'description-error' : undefined}
              maxLength={PROJECT_VALIDATION.description.maxLength}
            />
            {descriptionTouched && validation.errors.description && (
              <div id="description-error" className="text-xs text-destructive mt-1">
                {validation.errors.description}
              </div>
            )}
          </div>

          <div className="col-span-4 flex items-start pt-1">
            <div>
              <Label>Redirect URIs</Label>
              <div className="text-xs text-muted-foreground mt-1">
                Configure allowed redirect URIs and their corresponding login pages. Each redirect URI can have a login
                URL used for initiating authentication.
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
                      className={rowError?.redirect_uri ? 'border-destructive' : ''}
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
                      type="button"
                      variant="outline"
                      className="text-destructive bg-white hover:text-destructive border-destructive/30 hover:bg-destructive/10 shrink-0"
                      onClick={() => handleDeleteRedirectUri(idx)}
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
              <Button type="button" size="sm" onClick={() => setRedirectUris([...redirectUris, { redirect_uri: '', login_url: '' }])}>
                Add Redirect URI
              </Button>
            </div>
          </div>

          <div className="col-span-12 flex justify-end pt-4">
            <div className="flex items-center gap-2">
              <Button
                size="sm"
                onClick={() => {
                  // Mark all fields as touched to show validation errors
                  setNameTouched(true);
                  setDescriptionTouched(true);
                  setRedirectUrisTouched(true);
                  if (validation.isValid) {
                    handleCreate();
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
                  <>Create</>
                )}
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
