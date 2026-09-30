import { useEffect } from 'react';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { useToast } from '@/hooks/use-toast';
import type { HttpClient } from '@/services/core/httpClient';
import { fetchOAuthConnectionByProvider, updateOAuthConnection } from '@/services/projects';

import SocialConnectionState from './social-connection-state';
import { useSocialConnectionContext } from './use-social-connection-context';
import SocialConnectionForm from './social-connection-form';
import { useSocialConnectionForm } from './use-social-connection-form';

interface Props {
  httpClient: HttpClient;
}

export default function SocialConnectionDetails({ httpClient }: Props) {
  const {
    project_id, provider, AUTH_BASE_URL, navigate,
    paramsFrom, providerConfig, projectQuery,
  } = useSocialConnectionContext(httpClient);
  const queryClient = useQueryClient();
  const { toast } = useToast();
  const connectionQuery = useQuery({
    queryKey: ['project', project_id, 'oauth-connection', provider],
    queryFn: () =>
      fetchOAuthConnectionByProvider(httpClient, AUTH_BASE_URL, project_id!, provider!),
    enabled: !!project_id && !!provider,
    retry: 1,
  });

  const project = projectQuery.data;
  const connection = connectionQuery.data;

  const form = useSocialConnectionForm();
  const {
    credentialType, setCredentialType, clientId, setClientId,
    clientSecret, setClientSecret, scopes, setScopes,
    enabled, setEnabled, setClientIdTouched, setClientSecretTouched,
    setScopesTouched, validation,
  } = form;

  // Populate form once connection data arrives
  useEffect(() => {
    if (connection) {
      setCredentialType(connection.credential_type);
      setClientId(connection.client_id ?? '');
      setClientSecret(connection.client_secret ?? '');
      setScopes(connection.scopes ?? []);
      setEnabled(connection.enabled ? 'true' : 'false');
    }
  }, [connection, setCredentialType, setClientId, setClientSecret, setScopes, setEnabled]);

  // Reset touch state when credential type changes (same as create page)
  const handleSelectDefault = () => {
    setCredentialType('default');
    // Don't clear custom field values — they're preserved so the user can
    // switch back to Custom without losing what they typed. The values are
    // simply omitted from the payload when credential_type is 'default'.
    setClientIdTouched(false);
    setClientSecretTouched(false);
    setScopesTouched(false);
  };

  const saveMutation = useMutation({
    mutationFn: () => {
      if (!project_id || !provider) throw new Error('Missing project or provider');
      return updateOAuthConnection(httpClient, AUTH_BASE_URL, project_id, provider, {
        credential_type: credentialType,
        client_id: credentialType === 'custom' ? clientId : undefined,
        client_secret: credentialType === 'custom' ? clientSecret : undefined,
        scopes: credentialType === 'custom' && scopes.length > 0 ? scopes : undefined,
        enabled: enabled === 'true',
      });
    },
    onSuccess: async (data) => {
      // Refresh both the list and this single-connection cache entry
      await queryClient.invalidateQueries({ queryKey: ['project', project_id, 'oauth-connections'] });
      queryClient.setQueryData(
        ['project', project_id, 'oauth-connection', provider],
        data,
      );
      toast({
        title: 'Connection updated',
        description: `${providerConfig?.name ?? provider} social connection has been updated.`,
      });
      navigate(`/projects/${project_id}/details${paramsFrom || '?tab=settings'}`);
    },
    onError: (err: Error) => {
      toast({
        title: 'Save failed',
        description: err?.message ?? 'Unable to update social connection. Please try again.',
        variant: 'destructive',
      });
    },
  });

  const handleSave = () => {
    setClientIdTouched(true);
    setClientSecretTouched(true);
    setScopesTouched(true);
    if (!validation.isValid) return;
    saveMutation.mutate();
  };

  return (
    <SocialConnectionState
      loading={projectQuery.isLoading || connectionQuery.isLoading}
      failed={projectQuery.isError || connectionQuery.isError || !project || !connection}
      error={projectQuery.error?.message ?? connectionQuery.error?.message ?? 'Failed to load connection details'}
      errorTitle="Failed to load connection"
      testId="social-connection-details"
      supported={!!providerConfig}
      provider={provider}
      onBack={() => navigate(`/projects/${project_id}/details${paramsFrom}`)}
    >
      <SocialConnectionForm
        form={form}
        project={project}
        providerConfig={providerConfig}
        mode="edit"
        pending={saveMutation.status === 'pending'}
        onSubmit={handleSave}
        onBack={() => navigate(`/projects/${project_id}/details${paramsFrom}`)}
        onSelectDefault={handleSelectDefault}
      />
    </SocialConnectionState>
  );
}
