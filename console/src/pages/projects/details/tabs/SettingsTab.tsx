import { ConfirmDialog } from '@/components/ConfirmDialog';
import { StatusBadge } from '@/components/StatusBadge';
import { Button } from '@/components/ui/button';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Label } from '@/components/ui/label';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { useProjectSettings, useOAuthConnections } from '@/hooks/projects';
import { useToast } from '@/hooks/use-toast';
import type { HttpClient } from '@/services/core/httpClient';
import {
    fetchProjectRoles,
    updateProjectSettings,
    createOAuthConnection,
    deleteOAuthConnection,
    type Project,
    type ProjectSettings,
    type OAuthConnection,
} from '@/services/projects';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Eye, Loader2, MoreVertical, Trash2 } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';

// Provider configuration for OAuth connections
const PROVIDER_CONFIG: Record<
  string,
  { name: string; logo: React.ReactNode }
> = {
  google: {
    name: 'Google',
    logo: (
      <svg
        className="h-5 w-5"
        viewBox="0 0 24 24"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
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

interface Props {
    httpClient: HttpClient;
    project?: Project | null;
}

/** A placeholder value used in the role Select when no role is chosen. */
const NO_ROLE_VALUE = '__none__';

export default function SettingsTab({ httpClient, project }: Props) {
    const { project_id } = useParams<{ project_id: string }>();
    const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL as string;

    const queryClient = useQueryClient();
    const { toast } = useToast();

    // ─── Remote data ──────────────────────────────────────────────────────────

    const { settings, isLoading: settingsLoading, error: settingsError } = useProjectSettings({
        httpClient,
        baseUrl: AUTH_BASE_URL,
        projectId: project_id,
    });

    // Load all roles so we can populate the default signup role dropdown.
    // Fetch up to 100 roles (unlikely a project has more than that for this
    // selector; if needed this can be converted to a search-as-you-type later).
    const rolesQuery = useQuery({
        queryKey: ['project', project_id, 'roles', 'all'],
        queryFn: () =>
            fetchProjectRoles(httpClient, AUTH_BASE_URL, project_id!, 1, 100, 'name', 'asc'),
        enabled: Boolean(project_id),
        retry: 1,
    });

    const roles = rolesQuery.data?.data ?? [];

    // ─── OAuth Connections ────────────────────────────────────────────────

    const [removeConnection, setRemoveConnection] = useState<OAuthConnection | null>(null);

    const { connections: oauthConnections, isLoading: oauthLoading } = useOAuthConnections({
        httpClient,
        baseUrl: AUTH_BASE_URL,
        projectId: project_id,
    });

    const createConnectionMutation = useMutation({
        mutationFn: (provider: string) => {
            if (!project_id) throw new Error('Project not loaded');
            return createOAuthConnection(httpClient, AUTH_BASE_URL, project_id, {
                provider,
                credential_type: 'default',
            });
        },
        onSuccess: (data) => {
            queryClient.invalidateQueries({ queryKey: ['project', project_id, 'oauth-connections'] });
            toast({
                title: 'Connection added',
                description: `${PROVIDER_CONFIG[data.provider]?.name || data.provider} connection added successfully`,
            });
        },
        onError: (err: any) => {
            toast({
                title: 'Add failed',
                description: err?.message ?? 'Unable to add oauth connection',
                variant: 'destructive',
            });
        },
    });

    const deleteConnectionMutation = useMutation({
        mutationFn: (provider: string) => {
            if (!project_id) throw new Error('Project not loaded');
            return deleteOAuthConnection(httpClient, AUTH_BASE_URL, project_id, provider);
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['project', project_id, 'oauth-connections'] });
            toast({
                title: 'Connection removed',
                description: 'OAuth connection has been removed',
            });
            setRemoveConnection(null);
        },
        onError: (err: any) => {
            toast({
                title: 'Remove failed',
                description: err?.message ?? 'Unable to remove oauth connection',
                variant: 'destructive',
            });
        },
    });

    // Get available providers (not already added)
    const addedProviders = oauthConnections.map((c) => c.provider);
    const availableProviders = Object.keys(PROVIDER_CONFIG).filter(
        (p) => !addedProviders.includes(p)
    );

    // ─── Local form state ─────────────────────────────────────────────────

    const [isSignupEnabled, setIsSignupEnabled] = useState(false);
    const [defaultSignupRoleId, setDefaultSignupRoleId] = useState<string | null>(null);
    const [isSignupVerifyEmail, setIsSignupVerifyEmail] = useState(false);
    const [isForgotPasswordEnabled, setIsForgotPasswordEnabled] = useState(false);
    const [isAllowTempEmail, setIsAllowTempEmail] = useState(false);

    // Sync local state from fetched settings whenever they arrive or change.
    useEffect(() => {
        if (settings) {
            setIsSignupEnabled(settings.is_signup_enabled);
            setDefaultSignupRoleId(settings.default_signup_role_id);
            setIsSignupVerifyEmail(settings.is_signup_verify_email);
            setIsForgotPasswordEnabled(settings.is_forgot_password_enabled);
            setIsAllowTempEmail(settings.is_allow_temp_email);
        }
    }, [settings]);

    // ─── Save mutation ────────────────────────────────────────────────────

    const saveMutation = useMutation<ProjectSettings, Error, ProjectSettings>({
        mutationFn: (payload: ProjectSettings) => {
            if (!project_id) throw new Error('Project not loaded');
            return updateProjectSettings(httpClient, AUTH_BASE_URL, project_id, payload);
        },
        onSuccess: (data) => {
            // Update the cached settings so the form stays in sync.
            queryClient.setQueryData(['project', project_id, 'settings'], data);
            toast({ title: 'Settings saved', description: 'Project settings updated successfully' });
        },
        onError: (err: any) => {
            toast({
                title: 'Save failed',
                description: err?.message ?? 'Unable to save project settings',
                variant: 'destructive',
            });
        },
    });

    const handleSave = () => {
        // When sign-up is disabled the dependent fields must be reset per API contract.
        const payload: ProjectSettings = {
            is_signup_enabled: isSignupEnabled,
            default_signup_role_id: isSignupEnabled ? defaultSignupRoleId : null,
            is_signup_verify_email: isSignupEnabled ? isSignupVerifyEmail : false,
            is_forgot_password_enabled: isForgotPasswordEnabled,
            is_allow_temp_email: isSignupEnabled ? isAllowTempEmail : false,
        };
        saveMutation.mutate(payload);
    };

    // ─── Loading / error states ───────────────────────────────────────────────

    if (settingsLoading) {
        return (
            <div className="flex items-center" data-testid="settings-loading">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                <span className="ml-2 text-sm text-muted-foreground">Loading…</span>
            </div>
        );
    }

    if (settingsError || !settings) {
        return (
            <section className="border rounded p-4 bg-card w-full">
                <div
                    className="flex flex-col items-center justify-center py-20 space-y-2"
                    data-testid="settings-error"
                >
                    <p className="text-sm text-destructive font-medium">Failed to load settings</p>
                    <p className="text-xs text-muted-foreground">
                        {settingsError?.message ?? 'Settings data is unavailable'}
                    </p>
                </div>
            </section>
        );
    }

    // ─── Render ───────────────────────────────────────────────────────────────

    return (
        <div className="space-y-4">
            <section className="border rounded p-4 bg-card w-full">
                <h3 className="text-lg font-medium">Project Settings</h3>
                <div className="text-sm text-muted-foreground">
                    Configure the advanced settings for this project
                </div>

                <div className="mt-6 grid grid-cols-12 gap-y-6 gap-x-4">

                    {/* ── Enable Sign up ───────────────────────────────────── */}
                    <div className="col-span-4 flex flex-col justify-center">
                        <Label htmlFor="setting-signup-enabled">Enable sign-up</Label>
                        <p className="text-xs text-muted-foreground mt-1">
                            Allow users to self-register to this project.
                        </p>
                    </div>
                    <div className="col-span-8 flex items-center">
                        <Switch
                            id="setting-signup-enabled"
                            checked={isSignupEnabled}
                            onCheckedChange={setIsSignupEnabled}
                        />
                    </div>

                    {/* ── Default Sign up Role ─────────────────────────────── */}
                    <div className="col-span-4 flex flex-col justify-center">
                        <Label
                            htmlFor="setting-default-role"
                            className={!isSignupEnabled ? 'opacity-50' : ''}
                        >
                            Default sign-up role
                        </Label>
                        <p
                            className={`text-xs text-muted-foreground mt-1 ${!isSignupEnabled ? 'opacity-50' : ''}`}
                        >
                            The role users will be assigned automatically when they self-register to
                            this project.
                        </p>
                    </div>
                    <div className="col-span-8 flex items-center">
                        <Select
                            disabled={!isSignupEnabled}
                            value={defaultSignupRoleId ?? NO_ROLE_VALUE}
                            onValueChange={(v) =>
                                setDefaultSignupRoleId(v === NO_ROLE_VALUE ? null : v)
                            }
                        >
                            <SelectTrigger
                                id="setting-default-role"
                                className="w-[260px]"
                                aria-disabled={!isSignupEnabled}
                            >
                                <SelectValue placeholder="Select a role…" />
                            </SelectTrigger>
                            <SelectContent>
                                {/* Allow clearing the selection */}
                                <SelectItem value={NO_ROLE_VALUE}>
                                    <span className="text-muted-foreground">No default role</span>
                                </SelectItem>
                                {rolesQuery.isLoading ? (
                                    <SelectItem value="__loading__" disabled>
                                        <span className="flex items-center gap-2">
                                            <Loader2 className="h-3 w-3 animate-spin" />
                                            Loading roles…
                                        </span>
                                    </SelectItem>
                                ) : (
                                    roles.map((role: any) => (
                                        <SelectItem key={role.id} value={role.id}>
                                            {role.name}
                                        </SelectItem>
                                    ))
                                )}
                            </SelectContent>
                        </Select>
                    </div>

                    {/* ── Sign up verifies email ───────────────────────────── */}
                    <div className="col-span-4 flex flex-col justify-center">
                        <Label
                            htmlFor="setting-verify-email"
                            className={!isSignupEnabled ? 'opacity-50' : ''}
                        >
                            Enable email verification
                        </Label>
                        <p
                            className={`text-xs text-muted-foreground mt-1 ${!isSignupEnabled ? 'opacity-50' : ''}`}
                        >
                            Users will receive an email to verify their email address.
                        </p>
                    </div>
                    <div className="col-span-8 flex items-center">
                        <Switch
                            id="setting-verify-email"
                            disabled={!isSignupEnabled}
                            checked={isSignupEnabled ? isSignupVerifyEmail : false}
                            onCheckedChange={setIsSignupVerifyEmail}
                        />
                    </div>

                    {/* ── Allow temporary emails ───────────────────────────── */}
                    <div className="col-span-4 flex flex-col justify-center">
                        <Label
                            htmlFor="setting-allow-temp-email"
                            className={!isSignupEnabled ? 'opacity-50' : ''}
                        >
                            Allow temporary emails
                        </Label>
                        <p
                            className={`text-xs text-muted-foreground mt-1 ${!isSignupEnabled ? 'opacity-50' : ''}`}
                        >
                            Allow users to sign up using temporary email addresses.
                        </p>
                    </div>
                    <div className="col-span-8 flex items-center">
                        <Switch
                            id="setting-allow-temp-email"
                            disabled={!isSignupEnabled}
                            checked={isSignupEnabled ? isAllowTempEmail : false}
                            onCheckedChange={setIsAllowTempEmail}
                        />
                    </div>

                    {/* ── Enable Forgot password ───────────────────────────── */}
                    <div className="col-span-4 flex flex-col justify-center">
                        <Label htmlFor="setting-forgot-password">Enable forgot-password</Label>
                        <p className="text-xs text-muted-foreground mt-1">
                            Allow users to reset their password using the forgot-password flow.
                        </p>
                    </div>
                    <div className="col-span-8 flex items-center">
                        <Switch
                            id="setting-forgot-password"
                            checked={isForgotPasswordEnabled}
                            onCheckedChange={setIsForgotPasswordEnabled}
                        />
                    </div>

                    {/* ── Save button ──────────────────────────────────────── */}
                    <div className="col-span-12 flex justify-end pt-2">
                        <Button
                            size="sm"
                            onClick={handleSave}
                            disabled={saveMutation.status === 'pending'}
                            data-testid="settings-save-btn"
                        >
                            {saveMutation.status === 'pending' ? (
                                <>
                                    <Loader2 className="h-4 w-4 animate-spin mr-2" />
                                    Saving…
                                </>
                            ) : (
                                'Save'
                            )}
                        </Button>
                    </div>
                </div>
            </section>

            {/* Social Connections Section */}
            <section className="border rounded p-4 bg-card w-full">
                <h3 className="text-lg font-medium">Social Connections</h3>
                <div className="text-sm text-muted-foreground mb-4">
                    Configure the social login providers available on your project's login page.
                </div>

                <div className="flex items-center justify-between mb-3">
                    <div className="text-sm text-muted-foreground">
                        Showing {oauthConnections.length} connection{oauthConnections.length !== 1 ? 's' : ''}
                    </div>

                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button size="sm" disabled={availableProviders.length === 0}>
                                Add Connection
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                            {availableProviders.length === 0 ? (
                                <DropdownMenuItem disabled>
                                    All providers configured
                                </DropdownMenuItem>
                            ) : (
                                availableProviders.map((provider) => (
                                    <DropdownMenuItem
                                        key={provider}
                                        onClick={() => createConnectionMutation.mutate(provider)}
                                        disabled={createConnectionMutation.isPending}
                                    >
                                        {PROVIDER_CONFIG[provider]?.name || provider}
                                    </DropdownMenuItem>
                                ))
                            )}
                        </DropdownMenuContent>
                    </DropdownMenu>
                </div>

                <div className="border rounded bg-card shadow-sm overflow-hidden">
                    <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                            <thead>
                                <tr className="border-b bg-muted/40">
                                    <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                                        Provider
                                    </th>
                                    <th className="text-left font-medium text-muted-foreground px-4 py-2.5">
                                        Status
                                    </th>
                                    <th className="w-10 px-4 py-2.5"></th>
                                </tr>
                            </thead>
                            <tbody>
                                {oauthLoading ? (
                                    <tr>
                                        <td colSpan={3} className="px-4 py-3">
                                            <div className="flex items-center gap-2">
                                                <Loader2 className="h-4 w-4 animate-spin" />
                                                <span>Loading connections…</span>
                                            </div>
                                        </td>
                                    </tr>
                                ) : oauthConnections.length === 0 ? (
                                    <tr>
                                        <td colSpan={3} className="px-4 py-3 text-sm text-muted-foreground">No connections found</td>
                                    </tr>
                                ) : (
                                    oauthConnections.map((connection) => {
                                        const config = PROVIDER_CONFIG[connection.provider];
                                        const credentialLabel =
                                            connection.credential_type === 'default'
                                                ? 'Default credential'
                                                : 'Custom credentials';

                                        return (
                                            <tr
                                                key={connection.id}
                                                className="border-b last:border-0 hover:bg-muted/20 transition-colors"
                                            >
                                                {/* Provider Column with Logo and Badge */}
                                                <td className="px-4 py-3 align-middle">
                                                    <div className="flex items-center gap-3">
                                                        {config?.logo ? (
                                                            <div className="flex-shrink-0 rounded bg-muted p-1.5 flex items-center justify-center">
                                                                {config.logo}
                                                            </div>
                                                        ) : (
                                                            <div className="flex-shrink-0 h-8 w-8 rounded bg-muted flex items-center justify-center text-xs font-bold text-muted-foreground">
                                                                {connection.provider.charAt(0).toUpperCase()}
                                                            </div>
                                                        )}
                                                        <div className="flex-1">
                                                            <div className="text-sm font-medium">
                                                                {config?.name || connection.provider}
                                                            </div>
                                                            <div className="text-xs text-muted-foreground mt-0.5">
                                                                {credentialLabel}
                                                            </div>
                                                        </div>
                                                    </div>
                                                </td>

                                                {/* Status Column */}
                                                <td className="px-4 py-3 align-middle">
                                                    <StatusBadge active={connection.enabled} />
                                                </td>

                                                {/* Actions Column */}
                                                <td className="px-4 py-3 align-middle">
                                                    <div className="flex items-center justify-end">
                                                        <DropdownMenu>
                                                            <DropdownMenuTrigger asChild>
                                                                <Button size="sm" variant="ghost" className="h-7 w-7 p-0">
                                                                    <MoreVertical className="h-4 w-4" />
                                                                </Button>
                                                            </DropdownMenuTrigger>
                                                            <DropdownMenuContent align="end">
                                                                <DropdownMenuItem>
                                                                    <Eye className="h-4 w-4 mr-2" />
                                                                    View Details
                                                                </DropdownMenuItem>
                                                                <DropdownMenuItem
                                                                    className="text-destructive focus:text-destructive"
                                                                    onClick={() => setRemoveConnection(connection)}
                                                                >
                                                                    <Trash2 className="h-4 w-4 mr-2" />
                                                                    Remove
                                                                </DropdownMenuItem>
                                                            </DropdownMenuContent>
                                                        </DropdownMenu>
                                                    </div>
                                                </td>
                                            </tr>
                                        );
                                    })
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
            </section>

            {/* Delete Connection Confirmation Dialog */}
            <ConfirmDialog
                open={!!removeConnection}
                onOpenChange={(open) => {
                    if (!open) setRemoveConnection(null);
                }}
                closeOnConfirm={false}
                title="Remove Social Connection"
                description={
                    <span>
                        Are you sure you want to remove{' '}
                        <span className="font-medium">
                            {removeConnection ? PROVIDER_CONFIG[removeConnection.provider]?.name : ''}{' '}
                        </span>
                        connection? This action cannot be undone.
                    </span>
                }
                confirmLabel={deleteConnectionMutation.isPending ? (<><Loader2 className="h-4 w-4 animate-spin mr-2" />Removing...</>) : 'Remove'}
                destructive
                disabled={deleteConnectionMutation.isPending}
                onConfirm={() => {
                    if (removeConnection) {
                        deleteConnectionMutation.mutate(removeConnection.provider);
                    }
                }}
            />
        </div>
    );
}
