import { useEffect, useState, useCallback } from 'react';
import type { ConfigService } from '../services/core/configService';
import type { AuthService } from '../services/core/authService';
import type { TokenService } from '../services/core/tokenService';

export type AuthDependencies = {
  config: ConfigService;
  auth: AuthService;
  tokens: TokenService;
};

export function useAuth(deps: AuthDependencies) {
  const { tokens, config, auth } = deps;
  const [isAuthenticated, setAuthenticated] = useState<boolean>(!!tokens.getAccessToken());

  useEffect(() => {
    setAuthenticated(!!tokens.getAccessToken());
  }, [tokens]);

  const signIn = useCallback(() => {
    return auth.buildSignInRedirect();
  }, [auth]);

  const signOut = useCallback(() => {
    tokens.clear();
    setAuthenticated(false);
    window.location.href = '/signin';
  }, [tokens]);

  const ensureConfigLoaded = useCallback(async () => {
    const cached = config.getCachedConfig();
    if (!cached) await config.loadConfig();
  }, [config]);

  return {
    isAuthenticated,
    signIn,
    signOut,
    ensureConfigLoaded,
  } as const;
}
