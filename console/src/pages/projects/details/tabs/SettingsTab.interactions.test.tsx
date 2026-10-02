import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '@/services/core/httpClient';
import type { OAuthConnection, ProjectSettings } from '@/services/projects';
import { getProjectTabParams } from '@/lib/paramsStore';
import SettingsTab from './SettingsTab';

const { toast, fetchProjectSettings, fetchOAuthConnections, fetchProjectRoles, updateProjectSettings, deleteOAuthConnection } = vi.hoisted(() => ({
  toast: vi.fn(), fetchProjectSettings: vi.fn(), fetchOAuthConnections: vi.fn(), fetchProjectRoles: vi.fn(), updateProjectSettings: vi.fn(), deleteOAuthConnection: vi.fn(),
}));
vi.mock('@/hooks/use-toast', () => ({ useToast: () => ({ toast }) }));
vi.mock('@/services/projects', () => ({ fetchProjectSettings, fetchOAuthConnections, fetchProjectRoles, updateProjectSettings, deleteOAuthConnection }));
vi.mock('@/services/projects/oauth-connections', () => ({ fetchOAuthConnections }));
vi.mock('@/pages/projects/details/branding/branding-section', () => ({ default: () => <section>Branding summary</section> }));
const httpClient = {} as HttpClient;
const settings: ProjectSettings = { is_signup_enabled: true, default_signup_role_id: 'reader', is_signup_verify_email: true, is_allow_temp_email: true, is_forgot_password_enabled: true };
const google: OAuthConnection = { id: 'connection', project_id: 'p1', provider: 'google', credential_type: 'default', enabled: true, created_at: '', updated_at: '' };
const clients: QueryClient[] = [];
function Destination() { const location = useLocation(); return <output>{location.pathname}</output>; }
function mount() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false, retryDelay: 0, gcTime: Infinity }, mutations: { retry: false } } });
  clients.push(client);
  render(<QueryClientProvider client={client}><MemoryRouter initialEntries={['/projects/p1/details?tab=settings&page=2']}><Routes>
    <Route path="/projects/:project_id/details" element={<SettingsTab httpClient={httpClient} />} />
    <Route path="/projects/:project_id/social-connections/:provider/:action" element={<Destination />} />
  </Routes></MemoryRouter></QueryClientProvider>);
  return client;
}
async function ready() { await screen.findByTestId('settings-save-btn'); }
async function openMenu(button: HTMLElement) { fireEvent.pointerDown(button, { button: 0, ctrlKey: false, pointerType: 'mouse' }); }
async function removeDialog() {
  await openMenu(screen.getByRole('button', { name: '' }));
  fireEvent.click(await screen.findByRole('menuitem', { name: 'Remove', exact: true }));
  return screen.findByRole('alertdialog');
}
beforeEach(() => {
  vi.resetAllMocks(); vi.stubEnv('VITE_AUTH_BASE_URL', 'https://auth.example'); vi.stubGlobal('PointerEvent', MouseEvent);
  sessionStorage.clear();
  HTMLElement.prototype.hasPointerCapture = () => false;
  HTMLElement.prototype.releasePointerCapture = () => {};
  HTMLElement.prototype.scrollIntoView = () => {};
  fetchProjectSettings.mockResolvedValue(settings);
  fetchProjectRoles.mockResolvedValue({ data: [{ id: 'reader', name: 'Reader' }, { id: 'admin', name: 'Admin' }] });
  fetchOAuthConnections.mockResolvedValue({ data: [google] });
  updateProjectSettings.mockImplementation((_http, _base, _id, payload) => Promise.resolve(payload));
  deleteOAuthConnection.mockResolvedValue(undefined);
});
afterEach(() => { cleanup(); clients.splice(0).forEach(c => c.clear()); vi.unstubAllEnvs(); vi.unstubAllGlobals(); });

describe('Settings behavior', () => {
  it('shows loading and then an actionable load error', async () => {
    let reject!: (e: Error) => void;
    fetchProjectSettings.mockReturnValue(new Promise((_r, fail) => { reject = fail; }));
    mount(); expect(screen.getByTestId('settings-loading')).toBeVisible();
    reject(new Error('Access denied'));
    expect(await screen.findByTestId('settings-error')).toHaveTextContent('Access denied');
    expect(screen.queryByTestId('settings-save-btn')).not.toBeInTheDocument();
  });
  it('shows missing settings and loading connections independently', async () => {
    fetchProjectSettings.mockResolvedValue(null); mount();
    expect(await screen.findByTestId('settings-error')).toHaveTextContent('Settings data is unavailable'); cleanup();
    fetchProjectSettings.mockResolvedValue(settings); fetchOAuthConnections.mockReturnValue(new Promise(() => {})); mount();
    await ready(); expect(screen.getByText('Loading connections…')).toBeVisible();
  });
  it('resets dependent signup fields in the saved payload and updates the cache', async () => {
    const client = mount(); await ready();
    fireEvent.click(screen.getByRole('switch', { name: 'Enable sign-up' }));
    expect(screen.getByRole('switch', { name: 'Enable email verification' })).toBeDisabled();
    expect(screen.getByRole('switch', { name: 'Allow temporary emails' })).not.toBeChecked();
    expect(screen.getByRole('combobox')).toBeDisabled();
    fireEvent.click(screen.getByTestId('settings-save-btn'));
    const payload = { ...settings, is_signup_enabled: false, default_signup_role_id: null, is_signup_verify_email: false, is_allow_temp_email: false };
    await waitFor(() => expect(updateProjectSettings).toHaveBeenCalledWith(httpClient, 'https://auth.example', 'p1', payload));
    await waitFor(() => expect(client.getQueryData(['project', 'p1', 'settings'])).toEqual(payload));
    expect(toast).toHaveBeenCalledWith(expect.objectContaining({ title: 'Settings saved' }));
  });
  it('saves selected role and toggled preferences, and supports clearing the role', async () => {
    mount(); await ready();
    fireEvent.keyDown(screen.getByRole('combobox'), { key: 'Enter' });
    fireEvent.click(await screen.findByRole('option', { name: 'Admin', exact: true }));
    for (const name of ['Enable email verification', 'Allow temporary emails', 'Enable forgot-password']) fireEvent.click(screen.getByRole('switch', { name }));
    fireEvent.click(screen.getByTestId('settings-save-btn'));
    await waitFor(() => expect(updateProjectSettings).toHaveBeenCalledWith(httpClient, 'https://auth.example', 'p1', { ...settings, default_signup_role_id: 'admin', is_signup_verify_email: false, is_allow_temp_email: false, is_forgot_password_enabled: false }));
    await waitFor(() => expect(screen.getByTestId('settings-save-btn')).toBeEnabled());
    fireEvent.keyDown(screen.getByRole('combobox'), { key: 'Enter' }); fireEvent.click(await screen.findByRole('option', { name: 'No default role' }));
    fireEvent.click(screen.getByTestId('settings-save-btn'));
    await waitFor(() => expect(updateProjectSettings).toHaveBeenLastCalledWith(httpClient, 'https://auth.example', 'p1', expect.objectContaining({ default_signup_role_id: null })));
  });
  it('disables save during a pending request and preserves settings on failure', async () => {
    let reject!: (e: Error) => void; updateProjectSettings.mockReturnValue(new Promise((_r, fail) => { reject = fail; }));
    const client = mount(); await ready(); fireEvent.click(screen.getByTestId('settings-save-btn'));
    await waitFor(() => expect(screen.getByTestId('settings-save-btn')).toBeDisabled());
    reject(new Error('Permission denied'));
    await waitFor(() => expect(toast).toHaveBeenCalledWith({ title: 'Save failed', description: 'Permission denied', variant: 'destructive' }));
    expect(client.getQueryData(['project', 'p1', 'settings'])).toEqual(settings);
    expect(screen.getByTestId('settings-save-btn')).toBeEnabled();
  });
  it('routes to a new provider and remembers tab parameters', async () => {
    fetchOAuthConnections.mockResolvedValue({ data: [] }); mount(); await ready();
    expect(screen.getByText('No connections found')).toBeVisible();
    await openMenu(screen.getByRole('button', { name: 'Add Connection' })); fireEvent.click(await screen.findByRole('menuitem', { name: 'Google' }));
    expect(await screen.findByText('/projects/p1/social-connections/google/new')).toBeVisible();
    expect(getProjectTabParams('p1', 'settings')).toBe('?tab=settings&page=2');
  });
  it.each(['link', 'menu'])('opens existing connection details through %s', async mode => {
    mount(); await ready(); expect(screen.getByRole('button', { name: 'Add Connection' })).toBeDisabled();
    if (mode === 'link') fireEvent.click(screen.getByRole('button', { name: 'Open Google connection details' }));
    else { await openMenu(screen.getByRole('button', { name: '' })); fireEvent.click(await screen.findByRole('menuitem', { name: 'View Details' })); }
    expect(await screen.findByText('/projects/p1/social-connections/google/details')).toBeVisible();
    expect(getProjectTabParams('p1', 'settings')).toBe('?tab=settings&page=2');
  });
  it('renders custom unknown providers and disabled status', async () => {
    fetchOAuthConnections.mockResolvedValue({ data: [{ ...google, provider: 'github', credential_type: 'custom', enabled: false }] });
    mount(); await ready(); expect(screen.getByText('Custom credentials')).toBeVisible(); expect(screen.getByText('Disabled')).toBeVisible();
    expect(screen.getByRole('button', { name: 'Open github connection details' })).toBeVisible();
  });
  it('cancels removal without calling the API', async () => {
    mount(); await ready(); const dialog = await removeDialog();
    fireEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }));
    expect(deleteOAuthConnection).not.toHaveBeenCalled();
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });
  it('waits for refreshed connections before confirming removal', async () => {
    const client = mount(); await ready(); let finish!: () => void;
    const invalidate = vi.spyOn(client, 'invalidateQueries').mockReturnValue(new Promise<void>(resolve => { finish = resolve; }));
    const dialog = await removeDialog(); fireEvent.click(within(dialog).getByRole('button', { name: 'Remove', exact: true }));
    await waitFor(() => expect(invalidate).toHaveBeenCalledWith({ queryKey: ['project', 'p1', 'oauth-connections'] }));
    expect(deleteOAuthConnection).toHaveBeenCalledWith(httpClient, 'https://auth.example', 'p1', 'google');
    expect(within(dialog).getByRole('button', { name: 'Removing...' })).toBeDisabled();
    expect(toast).not.toHaveBeenCalled(); finish();
    await waitFor(() => expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument());
    expect(toast).toHaveBeenCalledWith(expect.objectContaining({ title: 'Connection removed' }));
  });
  it('retains removal confirmation and connection when the API rejects', async () => {
    deleteOAuthConnection.mockRejectedValue(new Error('Cannot remove')); mount(); await ready();
    const dialog = await removeDialog(); fireEvent.click(within(dialog).getByRole('button', { name: 'Remove', exact: true }));
    await waitFor(() => expect(toast).toHaveBeenCalledWith({ title: 'Remove failed', description: 'Cannot remove', variant: 'destructive' }));
    expect(dialog).toBeVisible(); expect(within(dialog).getByRole('button', { name: 'Remove', exact: true })).toBeEnabled();
  });
});
