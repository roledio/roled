import { ScopeTagInput } from '@/components/ScopeTagInput';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useToast } from '@/hooks/use-toast';
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard';
import { getProjectTabParams } from '@/lib/paramsStore';
import {
  validateOAuthConnectionForm,
  type OAuthConnectionValidationResult,
} from '@/lib/validation';
import type { HttpClient } from '@/services/core/httpClient';
import {
  createOAuthConnectionWithCredentials,
  fetchProjectById,
} from '@/services/projects';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { AlertCircle, ArrowLeft, Copy, Check, Eye, EyeOff, Loader2 } from 'lucide-react';
import { useMemo, useState } from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';
import { cn } from '@/lib/utils';
// ─── Provider config ──────────────────────────────────────────────────────────

const PROVIDER_CONFIG: Record<
  string,
  { name: string; description: string; logo: React.ReactNode }
> = {
  google: {
    name: 'Google',
    description: 'Allow users to sign in with their Google account.',
    logo: (
      <svg
        className="h-8 w-8"
        viewBox="0 0 24 24"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        aria-hidden="true"
      >
        <path
          d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
          fill="#4285F4"
        />
        <path
          d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
          fill="#34A853"
        />
        <path
          d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
          fill="#FBBC05"
        />
        <path
          d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
          fill="#EA4335"
        />
      </svg>
    ),
  },
};

// ─── Credential option card ───────────────────────────────────────────────────

interface CredentialCardProps {
  selected: boolean;
  onSelect: () => void;
  title: string;
  description: string;
  disabled?: boolean;
}

function CredentialCard({
  selected,
  onSelect,
  title,
  description,
  disabled = false,
}: CredentialCardProps) {
  return (
    <button
      type="button"
      role="radio"
      aria-checked={selected}
      disabled={disabled}
      onClick={onSelect}
      className={cn(
        'flex flex-col items-start gap-1 rounded-lg border p-4 text-left transition-colors w-full',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
        selected
          ? 'border-primary bg-primary/5 ring-1 ring-primary'
          : 'border-border hover:border-muted-foreground/40 hover:bg-muted/30',
        disabled && 'opacity-50 cursor-not-allowed',
      )}
    >
      <span className="text-sm font-medium">{title}</span>
      <span className="text-xs text-muted-foreground leading-snug">{description}</span>
    </button>
  );
}

// ─── Props ────────────────────────────────────────────────────────────────────

interface Props {
  httpClient: HttpClient;
}

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function NewSocialConnection({ httpClient }: Props) {
  const { project_id, provider } = useParams<{
    project_id: string;
    provider: string;
  }>();
  const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL as string;
  const navigate = useNavigate();
  const location = useLocation();
  const queryClient = useQueryClient();
  const { toast } = useToast();

  const stateAny = (location.state || {}) as any;
  const paramsFrom =
    getProjectTabParams(project_id ?? '', 'settings') || stateAny?.from || '';

  const providerConfig = provider ? PROVIDER_CONFIG[provider] : undefined;

  // ─── Project data ───────────────────────────────────────────────────────────
  const projectQuery = useQuery({
    queryKey: ['project', project_id],
    queryFn: () => fetchProjectById(httpClient, AUTH_BASE_URL, project_id!),
    enabled: !!project_id,
    retry: 1,
  });

  const project = projectQuery.data;

  // ─── Form state ─────────────────────────────────────────────────────────────
  const [credentialType, setCredentialType] = useState<'default' | 'custom'>('default');
  const [clientId, setClientId] = useState('');
  const [clientSecret, setClientSecret] = useState('');
  const [showClientSecret, setShowClientSecret] = useState(false);
  const [scopes, setScopes] = useState<string[]>([]);
  const [enabled, setEnabled] = useState<'true' | 'false'>('true');

  // Touch tracking — only show errors after the user has interacted with a field
  const [clientIdTouched, setClientIdTouched] = useState(false);
  const [clientSecretTouched, setClientSecretTouched] = useState(false);
  const [scopesTouched, setScopesTouched] = useState(false);

  const { copiedId, handleCopy } = useCopyToClipboard();

  // ─── Validation ─────────────────────────────────────────────────────────────
  const validation = useMemo<OAuthConnectionValidationResult>(
    () => validateOAuthConnectionForm(credentialType, clientId, clientSecret, scopes),
    [credentialType, clientId, clientSecret, scopes],
  );

  // ─── Mutation ───────────────────────────────────────────────────────────────
  const createMutation = useMutation({
    mutationFn: () => {
      if (!project_id || !provider) throw new Error('Missing project or provider');
      return createOAuthConnectionWithCredentials(
        httpClient,
        AUTH_BASE_URL,
        project_id,
        provider,
        {
          credential_type: credentialType,
          client_id: credentialType === 'custom' ? clientId : undefined,
          client_secret: credentialType === 'custom' ? clientSecret : undefined,
          scopes: scopes.length > 0 ? scopes : undefined,
          enabled: enabled === 'true',
        },
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['project', project_id, 'oauth-connections'],
      });
      toast({
        title: 'Connection created',
        description: `${providerConfig?.name ?? provider} social connection has been added.`,
      });
      // Navigate back to settings tab, restoring any saved tab params (search, page, etc.)
      // paramsFrom already contains the serialised search string from the settings tab
      // (e.g. "?tab=settings&..."), so this lands the user exactly where they left off.
      navigate(`/projects/${project_id}/details${paramsFrom || '?tab=settings'}`);
    },
    onError: (err: any) => {
      toast({
        title: 'Create failed',
        description:
          err?.message ?? 'Unable to create social connection. Please try again.',
        variant: 'destructive',
      });
    },
  });

  const handleCreate = () => {
    // Touch all fields to surface any errors before submitting
    setClientIdTouched(true);
    setClientSecretTouched(true);
    setScopesTouched(true);

    if (!validation.isValid) return;
    createMutation.mutate();
  };

  // ─── Loading / error states ─────────────────────────────────────────────────
  if (projectQuery.isLoading) {
    return (
      <div
        className="flex items-center justify-center py-20"
        data-testid="new-social-connection-loading"
      >
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        <span className="ml-2 text-sm text-muted-foreground">Loading…</span>
      </div>
    );
  }

  if (projectQuery.isError || !project) {
    return (
      <div
        className="flex flex-col items-center justify-center py-20 space-y-2"
        data-testid="new-social-connection-error"
      >
        <AlertCircle className="h-8 w-8 text-destructive" />
        <p className="text-sm text-destructive font-medium">Failed to load project</p>
        <p className="text-xs text-muted-foreground">
          {projectQuery.error?.message ?? 'Project data is unavailable'}
        </p>
      </div>
    );
  }

  // Unknown provider guard
  if (!providerConfig) {
    return (
      <div
        className="flex flex-col items-center justify-center py-20 space-y-2"
        data-testid="new-social-connection-unknown-provider"
      >
        <AlertCircle className="h-8 w-8 text-destructive" />
        <p className="text-sm text-destructive font-medium">Unknown provider</p>
        <p className="text-xs text-muted-foreground">
          The provider &quot;{provider}&quot; is not supported.
        </p>
        <Button
          size="sm"
          variant="secondary"
          onClick={() => navigate(`/projects/${project_id}/details${paramsFrom}`)}
        >
          Go back
        </Button>
      </div>
    );
  }

  const creating = createMutation.status === 'pending';

  // ─── Render ─────────────────────────────────────────────────────────────────
  return (
    <div className="space-y-6 max-w-3xl">
      {/* Back button */}
      <div>
        <Button
          variant="secondary"
          size="sm"
          onClick={() => navigate(`/projects/${project_id}/details${paramsFrom}`)}
          className="px-3"
        >
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back to {project.name}
        </Button>
      </div>

      {/* Page header */}
      <div className="flex items-center gap-4">
        {project.logo_url ? (
          <img
            src={project.logo_url}
            alt={project.name}
            className="h-12 w-12 rounded object-cover"
          />
        ) : (
          <div className="h-12 w-12 rounded bg-muted flex items-center justify-center text-sm font-bold text-muted-foreground">
            {project.name.charAt(0)}
          </div>
        )}
        <div>
          <h1 className="text-lg font-medium">
            Add {providerConfig.name} Connection
          </h1>
          <div className="text-sm text-muted-foreground">
            {providerConfig.description}
          </div>
        </div>
      </div>

      {/* Form card */}
      <div className="border rounded p-4 bg-card w-full">
        <div className="flex items-center gap-3 mb-1">
          <div className="flex-shrink-0">{providerConfig.logo}</div>
          <div>
            <h3 className="text-lg font-medium">{providerConfig.name}</h3>
            <p className="text-sm text-muted-foreground">Social connection settings</p>
          </div>
        </div>

        <div className="mt-6 grid grid-cols-12 gap-y-6 gap-x-4">
          {/* ── Credentials ──────────────────────────────────────────────── */}
          <div className="col-span-4 flex flex-col justify-start pt-1">
            <Label>Credentials</Label>
            <p className="text-xs text-muted-foreground mt-1">
              Choose how to authenticate with {providerConfig.name}.
            </p>
          </div>
          <div
            className="col-span-8 grid grid-cols-2 gap-3"
            role="radiogroup"
            aria-label="Credential type"
          >
            <CredentialCard
              selected={credentialType === 'default'}
              onSelect={() => {
                setCredentialType('default');
                // Clear custom fields when switching back to default
                setClientId('');
                setClientSecret('');
                setScopes([]);
                setClientIdTouched(false);
                setClientSecretTouched(false);
                setScopesTouched(false);
              }}
              title="Default"
              description="Use Roled's shared credentials. No configuration required."
            />
            <CredentialCard
              selected={credentialType === 'custom'}
              onSelect={() => setCredentialType('custom')}
              title="Custom"
              description="Use your own OAuth credentials for full control."
            />
          </div>

          {/* ── Custom credential fields (shown only when custom) ──────── */}
          {credentialType === 'custom' && (
            <>
              {/* Client ID */}
              <div className="col-span-4 flex flex-col justify-center">
                <Label htmlFor="oauth-client-id">Client ID</Label>
                <p className="text-xs text-muted-foreground mt-1">
                  OAuth 2.0 Client ID from {providerConfig.name}.
                </p>
              </div>
              <div className="col-span-8">
                <Input
                  id="oauth-client-id"
                  placeholder={`Enter ${providerConfig.name} Client ID`}
                  value={clientId}
                  onChange={(e) => {
                    setClientId(e.target.value);
                    if (!clientIdTouched) setClientIdTouched(true);
                  }}
                  onBlur={() => setClientIdTouched(true)}
                  className={cn(
                    clientIdTouched && validation.errors.clientId && 'border-destructive',
                  )}
                  aria-invalid={clientIdTouched && !!validation.errors.clientId}
                  aria-describedby={
                    clientIdTouched && validation.errors.clientId
                      ? 'oauth-client-id-error'
                      : undefined
                  }
                  disabled={creating}
                />
                {clientIdTouched && validation.errors.clientId && (
                  <p
                    id="oauth-client-id-error"
                    className="text-xs text-destructive mt-1"
                  >
                    {validation.errors.clientId}
                  </p>
                )}
              </div>

              {/* Client Secret */}
              <div className="col-span-4 flex flex-col justify-center">
                <Label htmlFor="oauth-client-secret">Client Secret</Label>
                <p className="text-xs text-muted-foreground mt-1">
                  OAuth 2.0 Client Secret from {providerConfig.name}.
                </p>
              </div>
              <div className="col-span-8">
                <div className="relative">
                  <Input
                    id="oauth-client-secret"
                    type={showClientSecret ? 'text' : 'password'}
                    placeholder={`Enter ${providerConfig.name} Client Secret`}
                    value={clientSecret}
                    onChange={(e) => {
                      setClientSecret(e.target.value);
                      if (!clientSecretTouched) setClientSecretTouched(true);
                    }}
                    onBlur={() => setClientSecretTouched(true)}
                    className={cn(
                      'pr-10',
                      clientSecretTouched &&
                        validation.errors.clientSecret &&
                        'border-destructive',
                    )}
                    aria-invalid={
                      clientSecretTouched && !!validation.errors.clientSecret
                    }
                    aria-describedby={
                      clientSecretTouched && validation.errors.clientSecret
                        ? 'oauth-client-secret-error'
                        : undefined
                    }
                    disabled={creating}
                  />
                  <button
                    type="button"
                    aria-label={showClientSecret ? 'Hide client secret' : 'Show client secret'}
                    onClick={() => setShowClientSecret((s) => !s)}
                    className="absolute right-2 top-1/2 -translate-y-1/2 p-1 rounded text-muted-foreground hover:bg-muted/5 transition-colors"
                  >
                    {showClientSecret ? (
                      <Eye className="h-4 w-4" />
                    ) : (
                      <EyeOff className="h-4 w-4" />
                    )}
                  </button>
                </div>
                {clientSecretTouched && validation.errors.clientSecret && (
                  <p
                    id="oauth-client-secret-error"
                    className="text-xs text-destructive mt-1"
                  >
                    {validation.errors.clientSecret}
                  </p>
                )}
              </div>

              {/* Scopes */}
              <div className="col-span-4 flex flex-col justify-start pt-1">
                <Label htmlFor="oauth-scopes">Scopes</Label>
                <p className="text-xs text-muted-foreground mt-1">
                  Permissions requested from {providerConfig.name}. Type a scope and press{' '}
                  <kbd className="rounded border px-1 py-0.5 text-[10px] font-mono bg-muted">
                    ,
                  </kbd>{' '}
                  or{' '}
                  <kbd className="rounded border px-1 py-0.5 text-[10px] font-mono bg-muted">
                    Enter
                  </kbd>{' '}
                  to add it.
                </p>
              </div>
              <div className="col-span-8">
                <ScopeTagInput
                  id="oauth-scopes"
                  scopes={scopes}
                  onChange={(next) => {
                    setScopes(next);
                    if (!scopesTouched) setScopesTouched(true);
                  }}
                  hasError={scopesTouched && !!validation.errors.scopes}
                  disabled={creating}
                  aria-describedby={
                    scopesTouched && validation.errors.scopes
                      ? 'oauth-scopes-error'
                      : undefined
                  }
                />
                {scopesTouched && validation.errors.scopes && (
                  <p id="oauth-scopes-error" className="text-xs text-destructive mt-1">
                    {validation.errors.scopes}
                  </p>
                )}
              </div>
            </>
          )}

          {/* ── Redirect URI (always shown for custom) ──────────────────── */}
          {credentialType === 'custom' && (
            <>
              <div className="col-span-4 flex flex-col justify-center">
                <Label htmlFor="oauth-redirect-uri">Redirect URI</Label>
                <p className="text-xs text-muted-foreground mt-1">
                  Add this URL to your {providerConfig.name} OAuth application settings.
                </p>
              </div>
              <div className="col-span-8">
                <div className="relative">
                  <Input
                    id="oauth-redirect-uri"
                    value={`${AUTH_BASE_URL}/oauth/google/callback`}
                    readOnly
                    className="select-text pr-10"
                    onFocus={(e) => e.target.select()}
                  />
                  <button
                    type="button"
                    aria-label="Copy redirect URI"
                    onClick={() => handleCopy(`${AUTH_BASE_URL}/oauth/google/callback`, 'Redirect URI')}
                    className="absolute right-2 top-1/2 -translate-y-1/2 p-1 rounded text-sm text-muted-foreground hover:bg-muted/5 transition-colors"
                  >
                    {copiedId === 'Redirect URI' ? (
                      <Check className="h-4 w-4 text-green-500" />
                    ) : (
                      <Copy className="h-4 w-4" />
                    )}
                  </button>
                </div>
              </div>
            </>
          )}

          {/* ── Status ──────────────────────────────────────────────────── */}
          <div className="col-span-4 flex flex-col justify-center">
            <Label htmlFor="oauth-status">Status</Label>
            <p className="text-xs text-muted-foreground mt-1">
              Enable or disable this connection on the login page.
            </p>
          </div>
          <div className="col-span-8">
            <Select
              value={enabled}
              onValueChange={(v) => setEnabled(v as 'true' | 'false')}
              disabled={creating}
            >
              <SelectTrigger id="oauth-status" className="w-[160px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="true">Enabled</SelectItem>
                <SelectItem value="false">Disabled</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* ── Submit ──────────────────────────────────────────────────── */}
          <div className="col-span-12 flex justify-end pt-2">
            <Button
              size="sm"
              onClick={handleCreate}
              disabled={creating}
              data-testid="create-social-connection-btn"
            >
              {creating ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                  Creating…
                </>
              ) : (
                'Create'
              )}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
