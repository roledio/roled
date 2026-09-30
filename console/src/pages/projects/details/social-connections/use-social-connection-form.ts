import { useMemo, useState } from 'react';
import { validateOAuthConnectionForm, type OAuthConnectionValidationResult } from '@/lib/validation';

export function useSocialConnectionForm() {
  const [credentialType, setCredentialType] = useState<'default' | 'custom'>('default');
  const [clientId, setClientId] = useState('');
  const [clientSecret, setClientSecret] = useState('');
  const [scopes, setScopes] = useState<string[]>([]);
  const [enabled, setEnabled] = useState<'true' | 'false'>('true');

  // Touch tracking — only show errors after the user has interacted with a field
  const [clientIdTouched, setClientIdTouched] = useState(false);
  const [clientSecretTouched, setClientSecretTouched] = useState(false);
  const [scopesTouched, setScopesTouched] = useState(false);

  const validation = useMemo<OAuthConnectionValidationResult>(
    () => validateOAuthConnectionForm(credentialType, clientId, clientSecret, scopes),
    [credentialType, clientId, clientSecret, scopes],
  );

  return {
    credentialType, setCredentialType,
    clientId, setClientId, clientSecret, setClientSecret,
    scopes, setScopes, enabled, setEnabled,
    clientIdTouched, setClientIdTouched,
    clientSecretTouched, setClientSecretTouched,
    scopesTouched, setScopesTouched, validation,
  };
}
