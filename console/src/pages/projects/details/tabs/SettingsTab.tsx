import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { useProjectSettings } from '@/hooks/projects';
import { useToast } from '@/hooks/use-toast';
import type { HttpClient } from '@/services/core/httpClient';
import {
    fetchProjectRoles,
    updateProjectSettings,
    type Project,
    type ProjectSettings,
} from '@/services/projects';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { AlertCircle, Loader2 } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';

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

    // ─── Local form state ─────────────────────────────────────────────────────

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

    // ─── Save mutation ────────────────────────────────────────────────────────

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
                    <AlertCircle className="h-8 w-8 text-destructive" />
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
        </div>
    );
}
