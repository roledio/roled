import logger from '@/lib/logger';
import { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import type { AuthService } from '../services/core/authService';
import type { TokenService } from '../services/core/tokenService';

const EXCHANGED_KEY = 'console_pkce_exchanged_v1';
const INFLIGHT_KEY = 'console_pkce_inflight_v1';

type Props = { auth: AuthService; tokenService: TokenService };

export default function SignInCallback({ auth, tokenService }: Props) {
  const navigate = useNavigate();
  const location = useLocation();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    (async () => {
      try {
        // If token exists and is valid, go to projects instead of processing callback 
        const valid = tokenService.isAccessTokenValid();
        const refreshToken = tokenService.getRefreshToken();
        if (valid || refreshToken) {
          logger.debug('[SignInCallback] Already signed in, navigating to /projects');
          navigate('/projects', { replace: true, state: location.state });
          return;
        }

        // Prevent double-processing (React StrictMode mounts twice in dev)
        if (sessionStorage.getItem(EXCHANGED_KEY) === 'true') {
          logger.debug('[SignInCallback] PKCE exchange already completed, navigating to /projects');
          sessionStorage.removeItem(EXCHANGED_KEY);
          navigate('/projects', { replace: true, state: location.state });
          return;
        }

        // If an exchange is already in-flight, wait for it to complete and then navigate
        if (sessionStorage.getItem(INFLIGHT_KEY) === 'true') {
          logger.debug('[SignInCallback] PKCE exchange already in-flight, waiting for completion');
          // another tab/process is exchanging; wait a bit and then navigate
          const wait = () => new Promise((r) => setTimeout(r, 1000));
          await wait();
          if (sessionStorage.getItem(EXCHANGED_KEY) === 'true') {
            logger.debug('[SignInCallback] PKCE exchange completed by another tab/process, navigating to /projects');
            sessionStorage.removeItem(EXCHANGED_KEY);
            sessionStorage.removeItem(INFLIGHT_KEY);
            navigate('/projects', { replace: true, state: location.state });
            return;
          }
        }

        sessionStorage.setItem(INFLIGHT_KEY, 'true');

        const qp = new URLSearchParams(window.location.search);
        const code = qp.get('code');
        const state = qp.get('state');
        const storedState = auth.getStoredState();
        logger.debug('[SignInCallback] Retrieved query params - code:', !!code, 'state:', state);
        logger.debug('[SignInCallback] Retrieved stored PKCE state:', storedState);
        if (!state || state !== storedState) {
          // invalid state - restart auth flow
          logger.debug('[SignInCallback] Invalid state, restarting auth flow');
          sessionStorage.removeItem(INFLIGHT_KEY);
          await auth.buildSignInRedirect();
          return;
        }
        if (!code) throw new Error('Missing authorization code');

        await auth.exchangeCode(code);

        sessionStorage.setItem(EXCHANGED_KEY, 'true');
        sessionStorage.removeItem(INFLIGHT_KEY);

        if (!tokenService.isAccessTokenValid()) {
          throw new Error('Exchange completed but access token not available');
        }

        logger.debug('[SignInCallback] PKCE exchange successful, navigating to /projects');
        navigate('/projects', { replace: true, state: location.state });
      } catch (e: any) {
        sessionStorage.removeItem(INFLIGHT_KEY);
        setError(String(e?.message ?? e));
      }
    })();
  }, [auth, navigate, tokenService]);

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center">
        {error ? <div className="text-red-600">Error: {error}</div> : <div className="text-gray-600">Signing you in…</div>}
      </div>
    </div>
  );
}
