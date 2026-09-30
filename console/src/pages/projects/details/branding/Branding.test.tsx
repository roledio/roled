import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '@/services/core/httpClient';
import type { ProjectBranding } from '@/services/projects/branding';
import { brandingQueryKey } from '@/hooks/projects/use-project-branding';
import { getProjectTabParams } from '@/lib/paramsStore';
import Branding from './Branding';

const { toast } = vi.hoisted(() => ({ toast: vi.fn() }));
vi.mock('@/hooks/use-toast', () => ({ useToast: () => ({ toast }) }));
vi.mock('@/lib/paramsStore', () => ({ getProjectTabParams: vi.fn() }));

const stored: ProjectBranding = {
  project_id: 'project', source_project_id: 'project', is_default: false,
  logo_url: 'https://images.example/custom.png', favicon_url: 'https://images.example/icon.png',
  primary_color: '#123456', rounding: 'small', enable_shadow: false, enable_border: true,
};
const project = { id: 'project', name: 'Acme', logo_url: 'https://images.example/project.png' };
let branding: ProjectBranding;
const get = vi.fn(), put = vi.fn(), post = vi.fn();
const httpClient = { instanceRef: { get, put, post } } as unknown as HttpClient;
const clients: QueryClient[] = [];
const success = (data: unknown) => ({ data: { success: true, data } });

function Destination() { const location = useLocation(); return <output>{location.pathname}{location.search}</output>; }
function mount() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: Infinity }, mutations: { retry: false } } });
  clients.push(client);
  render(<QueryClientProvider client={client}><MemoryRouter initialEntries={['/projects/project/branding']}><Routes>
    <Route path="/projects/:project_id/branding" element={<Branding httpClient={httpClient} />} />
    <Route path="/projects/:project_id/details" element={<Destination />} />
  </Routes></MemoryRouter></QueryClientProvider>);
  return client;
}
const ready = () => screen.findByRole('button', { name: 'Save branding' });
const preview = () => screen.getByTitle('Login page preview').getAttribute('srcdoc');
const upload = (name: 'logo' | 'favicon', file?: File) => fireEvent.change(screen.getByLabelText(`Upload ${name}`), { target: { files: file ? [file] : [] } });

beforeEach(() => {
  vi.resetAllMocks();
  vi.stubEnv('VITE_AUTH_BASE_URL', 'https://auth.example/');
  branding = { ...stored };
  get.mockImplementation((url: string) => Promise.resolve(success(url.endsWith('/branding') ? branding : project)));
  put.mockImplementation((_url: string, settings: ProjectBranding) => Promise.resolve(success({ ...branding, ...settings })));
  post.mockResolvedValue(success({ url: 'https://images.example/upload.png' }));
  vi.mocked(getProjectTabParams).mockReturnValue('?tab=branding&page=2');
  HTMLElement.prototype.hasPointerCapture = () => false;
  HTMLElement.prototype.scrollIntoView = () => {};
});
afterEach(() => { cleanup(); clients.splice(0).forEach(c => c.clear()); vi.unstubAllGlobals(); vi.unstubAllEnvs(); });

describe('Branding editor', () => {
  it('loads both resources before showing the editor', async () => {
    let resolve!: (value: unknown) => void;
    get.mockImplementation((url: string) => url.endsWith('/branding') ? new Promise(r => { resolve = r; }) : Promise.resolve(success(project)));
    mount();
    expect(screen.getByRole('status')).toHaveTextContent('Loading branding');
    resolve(success(branding));
    await ready();
    expect(get).toHaveBeenCalledWith('https://auth.example/api/v1/projects/project/branding');
    expect(preview()).toContain('Acme');
    expect(preview()).toContain('--brand-primary:#123456');
    expect(screen.getByRole('button', { name: 'Pick color from screen' })).toBeDisabled();
  });

  it.each(['project', 'branding'])('shows %s load errors and supports retry', async resource => {
    let fail = true;
    get.mockImplementation((url: string) => {
      if (fail && url.endsWith(resource === 'branding' ? '/branding' : '/project')) return Promise.reject(new Error('Access denied'));
      return Promise.resolve(success(url.endsWith('/branding') ? branding : project));
    });
    mount();
    expect(await screen.findByRole('alert')).toHaveTextContent('Access denied');
    fail = false;
    fireEvent.click(screen.getByText('Try again'));
    await ready();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('initializes inherited branding with the project logo and local appearance defaults', async () => {
    branding = { ...stored, is_default: true, source_project_id: 'system' };
    mount(); await ready();
    expect(screen.getByAltText('Logo preview')).toHaveAttribute('src', project.logo_url);
    expect(screen.queryByAltText('Favicon preview')).not.toBeInTheDocument();
    expect(screen.getByRole('switch', { name: 'Enable shadow' })).toHaveAttribute('aria-checked', 'true');
    expect(screen.getByRole('switch', { name: 'Enable border' })).toHaveAttribute('aria-checked', 'false');
  });

  it('edits appearance with live preview then persists settings and query cache', async () => {
    const client = mount(); await ready();
    fireEvent.change(screen.getByLabelText('Brand / primary color'), { target: { value: '#abcdef' } });
    fireEvent.click(screen.getByRole('switch', { name: 'Enable shadow' }));
    fireEvent.click(screen.getByRole('switch', { name: 'Enable border' }));
    fireEvent.click(screen.getByRole('combobox'));
    fireEvent.click(await screen.findByRole('option', { name: 'Large' }));
    expect(preview()).toContain('--brand-primary:#abcdef');
    expect(preview()).toContain('--brand-radius:1rem');
    expect(put).not.toHaveBeenCalled();
    fireEvent.click(await ready());
    await waitFor(() => expect(toast).toHaveBeenCalledWith(expect.objectContaining({ title: 'Branding saved' })));
    const expected = { logo_url: stored.logo_url, favicon_url: stored.favicon_url, primary_color: '#abcdef', rounding: 'large', enable_shadow: true, enable_border: false };
    expect(put).toHaveBeenCalledWith('https://auth.example/api/v1/projects/project/branding', expected);
    expect(client.getQueryData(brandingQueryKey('project'))).toEqual({ ...stored, ...expected });
  });

  it('blocks invalid colors and accepts a color-input change', async () => {
    mount(); await ready();
    fireEvent.change(screen.getByLabelText('Brand / primary color'), { target: { value: 'invalid' } });
    expect(await ready()).toBeDisabled();
    expect(screen.getByLabelText('Brand / primary color')).toHaveAttribute('aria-invalid', 'true');
    fireEvent.change(screen.getByLabelText('Choose primary color'), { target: { value: '#556677' } });
    expect(await ready()).not.toBeDisabled();
    expect(preview()).toContain('--brand-primary:#556677');
  });

  it('disables editing during save and preserves the draft after a failed save', async () => {
    let reject!: (reason: Error) => void;
    put.mockReturnValue(new Promise((_, r) => { reject = r; }));
    mount(); fireEvent.click(await ready());
    await waitFor(() => expect(screen.getByLabelText('Brand / primary color')).toBeDisabled());
    expect(await ready()).toBeDisabled();
    reject(new Error('Permission denied'));
    await waitFor(() => expect(toast).toHaveBeenCalledWith(expect.objectContaining({ title: 'Save failed', description: 'Permission denied' })));
    expect(await ready()).not.toBeDisabled();
    expect(screen.getByLabelText('Brand / primary color')).toHaveValue('#123456');
  });

  it.each(['logo', 'favicon'] as const)('uploads %s with the correct multipart type and updates preview', async name => {
    mount(); await ready();
    const file = new File(['image'], 'image.png', { type: 'image/png' });
    upload(name, file);
    await waitFor(() => {
      if (name === 'logo') expect(preview()).toContain('https://images.example/upload.png');
      else screen.getAllByAltText('Favicon preview').forEach(image => expect(image).toHaveAttribute('src', 'https://images.example/upload.png'));
    });
    expect(post).toHaveBeenCalledWith('https://auth.example/api/v1/uploads', expect.any(FormData), { headers: { 'Content-Type': 'multipart/form-data' } });
    const form = post.mock.calls[0][1] as FormData;
    expect(form.get('file')).toBe(file);
    expect(form.get('type')).toBe(name === 'logo' ? 'project-logo' : 'favicon');
    expect(put).not.toHaveBeenCalled();
  });

  it('ignores an empty selection and rejects oversized uploads without an API call', async () => {
    mount(); await ready();
    upload('logo');
    upload('favicon', new File([new Uint8Array(2 * 1024 * 1024 + 1)], 'large.png'));
    expect(post).not.toHaveBeenCalled();
    expect(toast).toHaveBeenCalledWith(expect.objectContaining({ title: 'Image too large' }));
  });

  it('blocks saves during upload and restores controls on upload failure', async () => {
    let reject!: (reason: Error) => void;
    post.mockReturnValue(new Promise((_, r) => { reject = r; }));
    mount(); await ready();
    upload('logo', new File(['image'], 'logo.png'));
    expect(await ready()).toBeDisabled();
    expect(screen.getByLabelText('Uploading image')).toBeInTheDocument();
    reject(new Error('Upload rejected'));
    await waitFor(() => expect(toast).toHaveBeenCalledWith(expect.objectContaining({ title: 'Upload failed', description: 'Upload rejected' })));
    expect(await ready()).not.toBeDisabled();
    expect(screen.getByAltText('Logo preview')).toHaveAttribute('src', stored.logo_url);
  });

  it('resets logo, removes favicon, and returns to the saved settings parameters', async () => {
    mount(); await ready();
    fireEvent.click(screen.getByText('Use project logo'));
    fireEvent.click(screen.getByText('Remove favicon'));
    expect(screen.getByAltText('Logo preview')).toHaveAttribute('src', project.logo_url);
    expect(screen.queryByAltText('Favicon preview')).not.toBeInTheDocument();
    fireEvent.click(screen.getByLabelText('Back to project settings'));
    expect(screen.getByRole('status')).toHaveTextContent('/projects/project/details?tab=settings&page=2');
  });

  it.each(['success', 'cancel', 'error'])('handles screen color picking: %s', async outcome => {
    let resolve!: (value: { sRGBHex: string }) => void;
    let reject!: (reason: Error) => void;
    const open = vi.fn().mockReturnValue(new Promise((a, b) => { resolve = a; reject = b; }));
    vi.stubGlobal('EyeDropper', class { open = open; });
    mount(); await ready();
    fireEvent.click(screen.getByLabelText('Pick color from screen'));
    expect(await ready()).toBeDisabled();
    if (outcome === 'success') resolve({ sRGBHex: '#998877' });
    else reject(outcome === 'cancel' ? new DOMException('Cancelled', 'AbortError') : new Error('Unavailable'));
    await waitFor(() => expect(screen.getByLabelText('Pick color from screen')).not.toBeDisabled());
    if (outcome === 'success') expect(screen.getByLabelText('Brand / primary color')).toHaveValue('#998877');
    if (outcome === 'cancel') expect(toast).not.toHaveBeenCalled();
    if (outcome === 'error') expect(toast).toHaveBeenCalledWith(expect.objectContaining({ title: 'Color picker unavailable', description: 'Unavailable' }));
  });
});
