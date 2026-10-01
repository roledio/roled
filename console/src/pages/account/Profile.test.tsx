import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '@/services/core/httpClient';
import type { CurrentUser } from '@/services/users';
import Profile from './Profile';

const { toast, useCurrentUser, updateCurrentUser, uploadUserAvatar } = vi.hoisted(() => ({
  toast: vi.fn(), useCurrentUser: vi.fn(), updateCurrentUser: vi.fn(), uploadUserAvatar: vi.fn(),
}));
vi.mock('@/hooks/use-toast', () => ({ useToast: () => ({ toast }) }));
vi.mock('@/hooks/users', () => ({ useCurrentUser }));
vi.mock('@/services/users', () => ({ updateCurrentUser }));
vi.mock('@/services/projects/projects', () => ({ uploadUserAvatar }));
const user: CurrentUser = { id: 'user', display_name: 'Alice', email: 'alice@example.com', avatar_url: 'https://images.example/old.png', created_at: '', updated_at: '', is_active: true, is_email_verified: true };
const token = { id: 'token', user: { id: 'user', display_name: 'Alice', email: user.email, avatar_url: user.avatar_url }, scopes: ['read'] };
const getCurrentTokenInfo = vi.fn(), setCurrentTokenInfo = vi.fn();
const httpClient = { tokenServiceRef: { getCurrentTokenInfo, setCurrentTokenInfo } } as unknown as HttpClient;
const clients: QueryClient[] = [];
function mount(seed = true) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  clients.push(client);
  if (seed) {
    client.setQueryData(['currentTokenInfo'], token);
    client.setQueryData(['currentTokenAndMemberInfo'], { tokenInfo: token, member: { id: 'member' } });
  }
  render(<QueryClientProvider client={client}><Profile httpClient={httpClient} /></QueryClientProvider>);
  return client;
}
async function menu(name: string) {
  // The avatar menu is the only unnamed button in this view.
  fireEvent.pointerDown(screen.getByRole('button', { name: '' }), { button: 0, ctrlKey: false, pointerType: 'mouse' });
  fireEvent.click(await screen.findByRole('menuitem', { name }));
}
async function uploadDialog(file?: File) {
  await menu('Upload');
  const dialog = await screen.findByRole('alertdialog');
  if (file) fireEvent.change(dialog.querySelector('input[type=file]')!, { target: { files: [file] } });
  return dialog;
}
beforeEach(() => {
  vi.resetAllMocks();
  vi.stubGlobal("PointerEvent", MouseEvent);
  vi.stubEnv('VITE_AUTH_BASE_URL', 'https://auth.example');
  useCurrentUser.mockReturnValue({ user, isLoading: false, error: null });
  getCurrentTokenInfo.mockReturnValue(token);
  updateCurrentUser.mockImplementation((_http, _url, payload) => Promise.resolve({ ...user, ...payload }));
  uploadUserAvatar.mockResolvedValue({ url: 'https://images.example/new.png' });
  HTMLElement.prototype.hasPointerCapture = () => false;
  HTMLElement.prototype.scrollIntoView = () => {};
});
afterEach(() => { cleanup(); clients.splice(0).forEach(c => c.clear()); vi.unstubAllEnvs(); vi.unstubAllGlobals(); });

describe('Profile', () => {
  it('shows loading then handles unavailable profiles', () => {
    useCurrentUser.mockReturnValue({ isLoading: true }); mount();
    expect(screen.getByText('Loading…')).toBeVisible(); cleanup();
    useCurrentUser.mockReturnValue({ error: new Error('offline') }); mount();
    expect(screen.getByText('Failed to load profile')).toBeVisible();
    expect(screen.queryByRole('button', { name: 'Save' })).not.toBeInTheDocument();
  });
  it('saves profile and synchronizes all identity caches without losing token/member data', async () => {
    const client = mount();
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Alicia' } });
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'new@example.com' } });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'Secret123!' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    await waitFor(() => expect(toast).toHaveBeenCalledWith(expect.objectContaining({ title: 'Saved' })));
    const identity = { ...token.user, display_name: 'Alicia', email: 'new@example.com' };
    expect(updateCurrentUser).toHaveBeenCalledWith(httpClient, 'https://auth.example', { display_name: 'Alicia', email: 'new@example.com', password: 'Secret123!', avatar_url: user.avatar_url });
    expect(client.getQueryData(['user', 'current'])).toMatchObject(identity);
    expect(client.getQueryData(['currentTokenInfo'])).toEqual({ ...token, user: identity });
    expect(client.getQueryData(['currentTokenAndMemberInfo'])).toEqual({ tokenInfo: { ...token, user: identity }, member: { id: 'member' } });
    expect(setCurrentTokenInfo).toHaveBeenCalledWith({ ...token, user: identity });
  });
  it('supports absent token caches and omits unchanged password and absent avatar', async () => {
    useCurrentUser.mockReturnValue({ user: { ...user, avatar_url: null }, isLoading: false });
    getCurrentTokenInfo.mockReturnValue(null);
    const client = mount(false);
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    await waitFor(() => expect(toast).toHaveBeenCalledWith(expect.objectContaining({ title: 'Saved' })));
    expect(updateCurrentUser).toHaveBeenCalledWith(httpClient, 'https://auth.example', { display_name: 'Alice', email: user.email });
    expect(client.getQueryData(['currentTokenInfo'])).toBeUndefined();
    expect(client.getQueryData(['currentTokenAndMemberInfo'])).toBeUndefined();
    expect(setCurrentTokenInfo).not.toHaveBeenCalled();
  });
  it.each([['Name', ''], ['Email', 'invalid'], ['Password', 'x']])('blocks invalid %s and exposes validation after editing', (label, value) => {
    mount(); const input = screen.getByLabelText(label);
    fireEvent.change(input, { target: { value } }); fireEvent.blur(input);
    expect(input).toHaveAttribute('aria-invalid', 'true');
    expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled();
    expect(updateCurrentUser).not.toHaveBeenCalled();
  });
  it('toggles password visibility without changing its value', () => {
    mount(); fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'Secret123!' } });
    fireEvent.click(screen.getByRole('button', { name: 'Show password' }));
    expect(screen.getByLabelText('Password')).toHaveAttribute('type', 'text');
    fireEvent.click(screen.getByRole('button', { name: 'Hide password' }));
    expect(screen.getByLabelText('Password')).toHaveAttribute('type', 'password');
    expect(screen.getByLabelText('Password')).toHaveValue('Secret123!');
  });
  it('keeps save disabled while pending and reports backend errors without updating caches', async () => {
    let reject!: (reason: Error) => void;
    updateCurrentUser.mockReturnValue(new Promise((_resolve, r) => { reject = r; }));
    const client = mount(); fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    expect(await screen.findByRole('button', { name: 'Saving...' })).toBeDisabled();
    reject(new Error('Email already exists'));
    await waitFor(() => expect(toast).toHaveBeenCalledWith({ title: 'Save failed', description: 'Email already exists', variant: 'destructive' }));
    expect(client.getQueryData(['currentTokenInfo'])).toEqual(token);
    expect(setCurrentTokenInfo).not.toHaveBeenCalled();
    expect(screen.getByRole('button', { name: 'Save' })).toBeEnabled();
  });
  it('removes avatar only when the profile is saved', async () => {
    mount(); await menu('Remove');
    expect(updateCurrentUser).not.toHaveBeenCalled();
    expect(screen.getAllByRole('img')).toHaveLength(1);
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    await waitFor(() => expect(updateCurrentUser).toHaveBeenCalledWith(httpClient, 'https://auth.example', { display_name: 'Alice', email: user.email, avatar_url: null }));
  });
  it('uploads a selected image and saves the resulting URL after upload completes', async () => {
    let resolve!: (value: { url: string }) => void;
    uploadUserAvatar.mockReturnValue(new Promise(r => { resolve = r; }));
    mount(); const file = new File(['image'], 'avatar.png', { type: 'image/png' });
    const dialog = await uploadDialog(file);
    expect(within(dialog).getByText('avatar.png')).toBeVisible();
    fireEvent.click(within(dialog).getByRole('button', { name: 'Upload', exact: true }));
    expect(await within(dialog).findByText('Uploading...')).toBeVisible();
    expect(uploadUserAvatar).toHaveBeenCalledWith(httpClient, 'https://auth.example', file);
    resolve({ url: 'https://images.example/new.png' });
    await waitFor(() => expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument());
    expect(screen.getAllByRole('img').some(i => i.getAttribute('src') === 'https://images.example/new.png')).toBe(true);
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));
    await waitFor(() => expect(updateCurrentUser).toHaveBeenCalledWith(httpClient, 'https://auth.example', expect.objectContaining({ avatar_url: 'https://images.example/new.png' })));
  });
  it('reports failed upload and preserves the previous avatar', async () => {
    uploadUserAvatar.mockRejectedValue(new Error('Image too large'));
    mount(); const dialog = await uploadDialog(new File(['image'], 'avatar.png'));
    fireEvent.click(within(dialog).getByRole('button', { name: 'Upload', exact: true }));
    await waitFor(() => expect(toast).toHaveBeenCalledWith({ title: 'Upload failed', description: 'Image too large', variant: 'destructive' }));
    expect(screen.getAllByRole('img').every(i => i.getAttribute('src') === user.avatar_url)).toBe(true);
    expect(updateCurrentUser).not.toHaveBeenCalled();
  });
  it('accepts drag and drop and clears the selected file when cancelled', async () => {
    mount(); let dialog = await uploadDialog();
    expect(within(dialog).getByRole('button', { name: 'Upload', exact: true })).toBeDisabled();
    const drop = within(dialog).getByText('Drag & drop an image here, or click to select').parentElement!;
    fireEvent.dragOver(drop); fireEvent.drop(drop, { dataTransfer: { files: [new File(['image'], 'dropped.png')] } });
    expect(within(dialog).getByText('dropped.png')).toBeVisible();
    fireEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }));
    dialog = await uploadDialog();
    expect(within(dialog).queryByText('dropped.png')).not.toBeInTheDocument();
    expect(within(dialog).getByRole('button', { name: 'Upload', exact: true })).toBeDisabled();
    expect(uploadUserAvatar).not.toHaveBeenCalled();
  });
});
