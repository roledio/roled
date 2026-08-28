import logger from '@/lib/logger';
import { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import type { AuthService } from '../services/core/authService';
import type { TokenService } from '../services/core/tokenService';

type Props = { auth: AuthService; tokenService: TokenService; signup?: boolean };

export default function SignIn({ auth, tokenService, signup = false }: Props) {
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    (async () => {
      try {
        const valid = tokenService.isAccessTokenValid();
        const refreshToken = tokenService.getRefreshToken();
        if (valid || refreshToken) {
          logger.debug('[SignIn] Already signed in, navigating to /projects');
          navigate('/projects', { replace: true, state: location.state });
          return;
        }

        logger.debug('[SignIn] No valid access token or refresh token, starting sign-in redirect');
        await auth.buildSignInRedirect(signup);
      } catch (e) {
        logger.error('Failed to start sign-in', e);
      }
    })();
  }, [auth, tokenService, navigate]);

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center">Redirecting to authentication page…</div>
    </div>
  );
}
