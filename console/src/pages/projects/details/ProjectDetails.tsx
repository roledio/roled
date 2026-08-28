import { StatusBadge } from '@/components/StatusBadge';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useProject } from '@/hooks/projects';
import { getProjectsParams, getProjectTabParams, saveProjectTabParams } from '@/lib/paramsStore';
import type { HttpClient } from '@/services/core/httpClient';
import { AlertCircle, ArrowLeft, Loader2 } from 'lucide-react';
import React, { Suspense, useEffect } from 'react';
import { useLocation, useNavigate, useParams, useSearchParams } from 'react-router-dom';

import ProjectTab from './tabs/ProjectTab';
const ResourcesTab = React.lazy(() => import('./tabs/ResourcesTab'));
const RolesTab = React.lazy(() => import('./tabs/RolesTab'));
const UsersTab = React.lazy(() => import('./tabs/UsersTab'));
const SettingsTab = React.lazy(() => import('./tabs/SettingsTab'));

interface Props { httpClient: HttpClient }

export default function ProjectDetails({ httpClient }: Props) {
    const navigate = useNavigate();
    const { project_id } = useParams<{ project_id: string }>();
    const AUTH_BASE_URL = import.meta.env.VITE_AUTH_BASE_URL as string;

    const location = useLocation();
    const stateAny = (location.state || {}) as any;
    const paramsFrom = getProjectsParams() || (stateAny?.from as string || '');
    const tabParamsState = stateAny?.tabParams ?? {} as Record<string, string>;

    const { project, isLoading: projectLoading, error: projectError } = useProject({ httpClient, baseUrl: AUTH_BASE_URL, projectId: project_id });

    // If project is present, preload the ProjectTab bundle so its content
    // becomes available quickly (helps tests and UX when default tab is open).
    useEffect(() => {
        if (project) {
            void import('./tabs/ProjectTab');
        }
    }, [project]);

    const [searchParams, setSearchParams] = useSearchParams();
    const tab = searchParams.get('tab') ?? 'project';

    const onSetTab = (value: string) => {
        const currentTab = searchParams.get('tab') ?? 'project';
        // Save current tab's params (excluding the tab key)
        const current = new URLSearchParams(searchParams.toString());
        current.delete('tab');
        const currentSerialized = current.toString();
        if (project_id) saveProjectTabParams(project_id, currentTab, currentSerialized);

        // Restore params for the target tab from session storage (or empty)
        const restoredSerialized = project_id ? getProjectTabParams(project_id, value) : '';
        const restored = new URLSearchParams(restoredSerialized ?? '');
        restored.set('tab', value);

        setSearchParams(restored, { replace: true });
    };

    // Ensure tab param present in URL (default to project)
    useEffect(() => {
        if (!searchParams.get('tab')) {
            const next = new URLSearchParams(searchParams.toString());
            next.set('tab', 'project');
            setSearchParams(next, { replace: true, state: location.state });
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    // Header should reflect server-side values only; do not update header
    // optimistically while the form is edited. Use project values directly.
    const effectiveLogo = project?.logo_url;
    const effectiveName = project?.name;
    const effectiveIsActive = project?.is_active;

    // Loading state (centered)
    if (projectLoading) {
        return (
            <div className="flex items-center justify-center py-20" data-testid="project-loading">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                <span className="ml-2 text-sm text-muted-foreground">Loading…</span>
            </div>
        );
    }

    // Error state (centered like Account page)
    if (projectError || !project) {
        return (
            <div className="flex flex-col items-center justify-center py-20 space-y-2" data-testid="project-error">
                <AlertCircle className="h-8 w-8 text-destructive" />
                <p className="text-sm text-destructive font-medium">Failed to load project</p>
                <p className="text-xs text-muted-foreground">{projectError?.message ?? 'Project data is unavailable'}</p>
            </div>
        );
    }

    const backToProjects = () => {
        navigate(`/projects${paramsFrom}`);
    };

    return (
        <div className="space-y-6 max-w-4xl">
            <div>
                <Button size="sm" variant="secondary" onClick={backToProjects} className="px-3">
                    <ArrowLeft className="h-4 w-4 mr-2" />
                    Back to Projects
                </Button>
            </div>

            <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                    {effectiveLogo ? (
                        <img src={effectiveLogo} alt={effectiveName ?? project?.name} className="h-12 w-12 rounded object-cover" />
                    ) : (
                        <div className="h-12 w-12 rounded bg-muted flex items-center justify-center text-sm font-bold text-muted-foreground">{(effectiveName ?? project?.name)?.charAt(0)}</div>
                    )}
                    <div>
                        <h2 className="text-lg font-medium flex items-center gap-3">
                            {effectiveName}
                        </h2>
                        <span className=""><StatusBadge active={!!effectiveIsActive} /></span>
                    </div>
                </div>
                <div />
            </div>

            <Tabs value={tab} onValueChange={onSetTab}>
                <TabsList>
                    <TabsTrigger value="project">Project</TabsTrigger>
                    <TabsTrigger value="resources">Resources</TabsTrigger>
                    <TabsTrigger value="roles">Roles</TabsTrigger>
                    <TabsTrigger value="users">Users</TabsTrigger>
                    <TabsTrigger value="settings">Settings</TabsTrigger>
                </TabsList>

                <TabsContent value="project">
                    <Suspense>
                        <ProjectTab httpClient={httpClient} project={project} />
                    </Suspense>
                </TabsContent>

                <TabsContent value="resources">
                    <Suspense>
                        <ResourcesTab httpClient={httpClient} project={project} />
                    </Suspense>
                </TabsContent>

                <TabsContent value="roles">
                    <Suspense>
                        <RolesTab httpClient={httpClient} project={project} />
                    </Suspense>
                </TabsContent>

                <TabsContent value="users">
                    <Suspense>
                        <UsersTab httpClient={httpClient} project={project} />
                    </Suspense>
                </TabsContent>

                <TabsContent value="settings">
                    <Suspense>
                        <SettingsTab httpClient={httpClient} project={project} />
                    </Suspense>
                </TabsContent>
            </Tabs>
        </div>
    );
}
