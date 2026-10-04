import { act, render, screen, waitFor, cleanup } from '@testing-library/react';
import { beforeEach, afterEach, describe, expect, it, vi } from 'vitest';
import SignInCallback from './SignInCallback';
import type { ComponentProps } from 'react';
const navigate = vi.hoisted(() => vi.fn());
vi.mock('react-router-dom', () => ({ useNavigate: () => navigate, useLocation: () => ({ state: { from: '/projects' } }) }));
vi.mock('@/lib/logger', () => ({ default: { debug: vi.fn() } }));
const exchanged = 'console_pkce_exchanged_v1';
const inflight = 'console_pkce_inflight_v1';
function setup() {
 const auth = { getStoredState: vi.fn(() => 'valid'), buildSignInRedirect: vi.fn().mockResolvedValue(undefined), exchangeCode: vi.fn().mockResolvedValue(undefined) };
 const tokenService = { isAccessTokenValid: vi.fn().mockReturnValueOnce(false).mockReturnValue(true), getRefreshToken: vi.fn(() => null) };
 return { auth, tokenService, mount: () => render(<SignInCallback {...({ auth, tokenService } as unknown as ComponentProps<typeof SignInCallback>)} />) };
}
beforeEach(() => { vi.clearAllMocks(); sessionStorage.clear(); window.history.replaceState({}, '', '/callback?code=authorization&state=valid'); });
afterEach(() => { cleanup(); vi.useRealTimers(); });
describe('SignInCallback', () => {
 it('exchanges the code once, clears inflight and preserves navigation state', async () => {
  const f = setup(); f.mount(); await waitFor(() => expect(navigate).toHaveBeenCalledWith('/projects', { replace: true, state: { from: '/projects' } }));
  expect(f.auth.exchangeCode).toHaveBeenCalledExactlyOnceWith('authorization'); expect(sessionStorage.getItem(exchanged)).toBe('true'); expect(sessionStorage.getItem(inflight)).toBeNull();
 });
 it.each(['access token', 'refresh token', 'completed exchange'])('skips exchanging for %s', async mode => {
  const f = setup(); if (mode === 'access token') f.tokenService.isAccessTokenValid.mockReset().mockReturnValue(true);
  if (mode === 'refresh token') f.tokenService.getRefreshToken.mockReturnValue('refresh');
  if (mode === 'completed exchange') sessionStorage.setItem(exchanged, 'true');
  f.mount(); await waitFor(() => expect(navigate).toHaveBeenCalled()); expect(f.auth.exchangeCode).not.toHaveBeenCalled();
  if (mode === 'completed exchange') expect(sessionStorage.getItem(exchanged)).toBeNull();
 });
 it.each(['', '?state=foreign&code=code'])('restarts authorization for invalid state %s', async query => {
  window.history.replaceState({}, '', '/callback' + query); const f = setup(); f.mount();
  await waitFor(() => expect(f.auth.buildSignInRedirect).toHaveBeenCalledOnce()); expect(f.auth.exchangeCode).not.toHaveBeenCalled(); expect(navigate).not.toHaveBeenCalled(); expect(sessionStorage.getItem(inflight)).toBeNull();
 });
 it.each(['missing code', 'exchange failure', 'token missing', 'string failure', 'redirect failure'])('renders %s and releases inflight flag', async mode => {
  const f = setup(); let message = 'Offline';
  if (mode === 'missing code') { window.history.replaceState({}, '', '/callback?state=valid'); message = 'Missing authorization code'; }
  if (mode === 'exchange failure') f.auth.exchangeCode.mockRejectedValue(new Error(message));
  if (mode === 'string failure') f.auth.exchangeCode.mockRejectedValue(message);
  if (mode === 'token missing') { f.tokenService.isAccessTokenValid.mockReset().mockReturnValue(false); message = 'Exchange completed but access token not available'; }
  if (mode === 'redirect failure') { window.history.replaceState({}, '', '/callback'); f.auth.buildSignInRedirect.mockRejectedValue(new Error(message)); }
  f.mount(); expect(await screen.findByText('Error: ' + message)).toBeInTheDocument(); expect(navigate).not.toHaveBeenCalled(); expect(sessionStorage.getItem(inflight)).toBeNull();
 });
 it.each([true, false])('waits for another exchange, completed=%s', async completed => {
  vi.useFakeTimers(); sessionStorage.setItem(inflight, 'true'); const f = setup(); f.mount();
  expect(screen.getByText('Signing you in…')).toBeInTheDocument(); expect(f.auth.exchangeCode).not.toHaveBeenCalled();
  if (completed) sessionStorage.setItem(exchanged, 'true');
  await act(async () => { await vi.advanceTimersByTimeAsync(1000); });
  expect(navigate).toHaveBeenCalled(); expect(f.auth.exchangeCode).toHaveBeenCalledTimes(completed ? 0 : 1); expect(sessionStorage.getItem(inflight)).toBeNull();
 });
});
