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
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard';
import { cn } from '@/lib/utils';
import { ArrowLeft, Copy, Check, Eye, EyeOff, Loader2 } from 'lucide-react';
import { useState } from 'react';
import type { useSocialConnectionForm } from './use-social-connection-form';
import type { PROVIDER_CONFIG } from './social-providers';

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

interface Props {
  form: ReturnType<typeof useSocialConnectionForm>;
  project: { name: string; logo_url?: string };
  providerConfig: (typeof PROVIDER_CONFIG)[string];
  mode: 'create' | 'edit';
  pending: boolean;
  onSubmit: () => void;
  onBack: () => void;
  onSelectDefault: () => void;
}

export default function SocialConnectionForm({ form, project, providerConfig, mode, pending, onSubmit, onBack, onSelectDefault }: Props) {
  const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL as string;
  const { copiedId, handleCopy } = useCopyToClipboard();
  const [showClientSecret, setShowClientSecret] = useState(false);
  const { credentialType, setCredentialType, clientId, setClientId, clientSecret, setClientSecret, scopes, setScopes, enabled, setEnabled, clientIdTouched, setClientIdTouched, clientSecretTouched, setClientSecretTouched, scopesTouched, setScopesTouched, validation } = form;
  return (
    <div className="space-y-6 max-w-3xl">
      {/* Back button */}
      <div>
        <Button
          variant="secondary"
          size="sm"
          onClick={() => onBack()}
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
            {mode === 'create' ? 'Add ' : ''}{providerConfig.name} Connection
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
              onSelect={onSelectDefault}
              disabled={pending}
              title="Default"
              description="Use Roled's shared credentials. No configuration required."
            />
            <CredentialCard
              selected={credentialType === 'custom'}
              onSelect={() => setCredentialType('custom')}
              disabled={pending}
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
                  disabled={pending}
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
                    disabled={pending}
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
                  disabled={pending}
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
              disabled={pending}
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
              onClick={onSubmit}
              disabled={pending}
              data-testid={`${mode === 'create' ? 'create' : 'save'}-social-connection-btn`}
            >
              {pending ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                  {mode === 'create' ? 'Creating…' : 'Saving…'}
                </>
              ) : (
                mode === 'create' ? 'Create' : 'Save'
              )}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
