import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { Toaster as Sonner } from '@/components/ui/sonner';
import { Toaster } from '@/components/ui/toaster';
import { TooltipProvider } from '@/components/ui/tooltip';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React, { useMemo } from 'react';
import { BrowserRouter, Navigate, Outlet, Route, Routes, useLocation, useNavigate } from 'react-router-dom';
import logger from './lib/logger';
import Account from './pages/account/Account';
import Profile from './pages/account/Profile';
import ClientDetails from './pages/projects/details/clients/ClientDetails';
import NewClient from './pages/projects/details/clients/NewClient';
import NewProject from './pages/projects/NewProject';
import NewRole from './pages/projects/details/roles/NewRole';
import NewUser from './pages/projects/details/users/NewUser';
import RoleDetails from './pages/projects/details/roles/RoleDetails';
import UserDetails from './pages/projects/details/users/UserDetails';
import NotFound from './pages/NotFound';
import NewResource from './pages/projects/details/resources/NewResource';
import ResourceDetails from './pages/projects/details/resources/ResourceDetails';
import ProjectDetails from './pages/projects/details/ProjectDetails';
import Projects from './pages/projects/Projects';
import SignIn from './pages/auth/SignIn';
import SignInCallback from './pages/auth/SignInCallback';
import { AuthService } from './services/core/authService';
import { ConfigService } from './services/core/configService';
import { HttpClient } from './services/core/httpClient';
import { TokenService } from './services/core/tokenService';
import { LocalStorageService } from './storage/storageService';

const AUTH_BASE_URL = (import.meta.env.VITE_AUTH_BASE_URL as string);

const queryClient = new QueryClient();

export default function App() {
  const configService = useMemo(() => new ConfigService(LocalStorageService, AUTH_BASE_URL), []);
  const tokenService = useMemo(() => new TokenService(LocalStorageService, configService), [configService]);
  const authService = useMemo(() => new AuthService(configService, tokenService, window.location.origin), [configService, tokenService]);
  const httpClient = useMemo(() => new HttpClient(tokenService, configService), [tokenService, configService]);

  React.useEffect(() => {
    // Preload console config on app start to ensure client_id is available for PKCE redirect
    (async () => {
      try {
        await configService.loadConfig();
      } catch (e) {
        // swallow - SignIn will handle redirect failure if necessary
        // keep console log for diagnostics
        logger.warn('Failed to preload console config', e);
      }
    })();
  }, [configService]);

  return (
    <QueryClientProvider client={queryClient}>
      <TooltipProvider>
        <Toaster />
        <Sonner />
        <BrowserRouter>
          <RouteAuthWatcher tokenService={tokenService} />
          <Routes>
            {/* Public / auth routes (no dashboard layout) */}
            <Route path="/" element={<RootRedirect />} />
            <Route path="/signin" element={<SignIn auth={authService} tokenService={tokenService} />} />
            <Route path="/signup" element={<SignIn auth={authService} tokenService={tokenService} signup />} />
            <Route path="/signin/callback" element={<SignInCallback auth={authService} tokenService={tokenService} />} />

            {/* Protected routes inside DashboardLayout */}
            <Route
              element={
                <RequireAuth tokenService={tokenService}>
                  <DashboardLayout httpClient={httpClient} tokenService={tokenService}>
                    <Outlet />
                  </DashboardLayout>
                </RequireAuth>
              }
            >
              <Route path="/projects" element={<Projects httpClient={httpClient} />} />
              <Route path="/projects/new" element={<NewProject httpClient={httpClient} />} />
              <Route path="/projects/:project_id/details" element={<ProjectDetails httpClient={httpClient} />} />
              <Route path="/projects/:project_id/clients/:client_id/details" element={<ClientDetails httpClient={httpClient} />} />
              <Route path="/projects/:project_id/clients/new" element={<NewClient httpClient={httpClient} />} />
              <Route path="/projects/:project_id/roles/new" element={<NewRole httpClient={httpClient} />} />
              <Route path="/projects/:project_id/roles/:role_id/details" element={<RoleDetails httpClient={httpClient} />} />
              <Route path="/projects/:project_id/users/new" element={<NewUser httpClient={httpClient} />} />
              <Route path="/projects/:project_id/users/:user_id/details" element={<UserDetails httpClient={httpClient} />} />
              <Route path="/projects/:project_id/resources/new" element={<NewResource httpClient={httpClient} />} />
              <Route path="/projects/:project_id/resources/:resource_id/details" element={<ResourceDetails httpClient={httpClient} />} />
              <Route path="/account" element={<Account httpClient={httpClient} tokenService={tokenService} />} />
              <Route path="/profile" element={<Profile httpClient={httpClient} />} />
            </Route>

            {/* Fallback outside layout */}
            <Route path="*" element={<NotFound />} />
          </Routes>
        </BrowserRouter>
      </TooltipProvider>
    </QueryClientProvider>
  );
}

function RootRedirect(): JSX.Element {
  return <Navigate to="/signin" replace />;
}

function RequireAuth({ tokenService, children }: { tokenService: TokenService; children: React.ReactNode }) {
  const location = useLocation();
  const valid = tokenService.isAccessTokenValid();
  const refreshToken = tokenService.getRefreshToken();
  logger.debug('[RequireAuth] Access token valid:', valid);
  logger.debug('[RequireAuth] Access token present:', !!tokenService.getAccessToken());
  logger.debug('[RequireAuth] Refresh token present:', refreshToken);
  if (!valid && !refreshToken) {
    tokenService.clear();
    return <Navigate to="/signin" replace state={location.state} />;
  }
  return <>{children}</>;
}

function RouteAuthWatcher({ tokenService }: { tokenService: TokenService }) {
  const location = useLocation();
  const navigate = useNavigate();

  React.useEffect(() => {
    const path = location.pathname;
    const protectedPrefixes = ['/projects', '/account'];
    const isProtected = protectedPrefixes.some((p) => path === p || path.startsWith(p + '/'));
    if (!isProtected) return;

    const valid = tokenService.isAccessTokenValid();
    const refreshToken = tokenService.getRefreshToken();
    logger.debug('[RouteAuthWatcher] route:', path, 'protected:', isProtected, 'token valid:', valid);
    logger.debug('[RouteAuthWatcher] Access token present:', !!tokenService.getAccessToken());
    logger.debug('[RouteAuthWatcher] Refresh token present:', refreshToken);
    if (!valid && !refreshToken) {
      // Clear any stale tokens and redirect to signin (preserve state)
      tokenService.clear();
      navigate('/signin', { replace: true, state: location.state });
    }
  }, [location.pathname, tokenService, navigate]);

  return null;
}
