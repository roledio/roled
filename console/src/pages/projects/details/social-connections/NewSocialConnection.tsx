import { useMutation, useQueryClient } from '@tanstack/react-query';

import { useToast } from '@/hooks/use-toast';
import type { HttpClient } from '@/services/core/httpClient';
import { createOAuthConnectionWithCredentials } from '@/services/projects';

import SocialConnectionState from './social-connection-state';
import { useSocialConnectionContext } from './use-social-connection-context';
import SocialConnectionForm from './social-connection-form';
import { useSocialConnectionForm } from './use-social-connection-form';

interface Props {
  httpClient: HttpClient;
}

export default function NewSocialConnection({ httpClient }: Props) {
  const {
    project_id, provider, AUTH_BASE_URL, navigate,
    paramsFrom, providerConfig, projectQuery,
  } = useSocialConnectionContext(httpClient);
  const queryClient = useQueryClient();
  const { toast } = useToast();
  const project = projectQuery.data;

  const form = useSocialConnectionForm();
  const {
    credentialType, setCredentialType, clientId, setClientId,
    clientSecret, setClientSecret, scopes, setScopes,
    enabled, setClientIdTouched, setClientSecretTouched, setScopesTouched,
    validation,
  } = form;

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
    onSuccess: async () => {
      await queryClient.invalidateQueries({
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
    onError: (err: Error) => {
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

  return (
    <SocialConnectionState
      loading={projectQuery.isLoading}
      failed={projectQuery.isError || !project}
      error={projectQuery.error?.message ?? 'Project data is unavailable'}
      errorTitle="Failed to load project"
      testId="new-social-connection"
      supported={!!providerConfig}
      provider={provider}
      onBack={() => navigate(`/projects/${project_id}/details${paramsFrom}`)}
    >
      <SocialConnectionForm
        form={form}
        project={project}
        providerConfig={providerConfig}
        mode="create"
        pending={createMutation.status === 'pending'}
        onSubmit={handleCreate}
        onBack={() => navigate(`/projects/${project_id}/details${paramsFrom}`)}
        onSelectDefault={() => {
          setCredentialType('default');
          setClientId('');
          setClientSecret('');
          setScopes([]);
          setClientIdTouched(false);
          setClientSecretTouched(false);
          setScopesTouched(false);
        }}
      />
    </SocialConnectionState>
  );
}
