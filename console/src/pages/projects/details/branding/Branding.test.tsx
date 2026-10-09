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

// Render dropdown items inline so tests can click them without Radix portal/pointer-event machinery.
vi.mock('@/components/ui/dropdown-menu', () => ({
  DropdownMenu: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DropdownMenuTrigger: ({ children, asChild }: { children: React.ReactNode; asChild?: boolean }) => asChild ? children : <button>{children}</button>,
  DropdownMenuContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DropdownMenuItem: ({ children, onClick }: { children: React.ReactNode; onClick?: () => void }) => <button onClick={onClick}>{children}</button>,
}));

const stored: ProjectBranding = {
  project_id: 'project', source_project_id: 'project', is_default: false,
  logo_url: 'https://images.example/custom.png', favicon_url: 'https://images.example/icon.png',
  primary_color: '#123456', rounding: 'small', enable_shadow: false, enable_border: true,
};
const project = { id: 'project', name: 'Acme', logo_url: 'https://images.example/project.png' };
let branding: ProjectBranding;
const get = vi.fn(), put = vi.fn(), post = vi.fn();
const httpClient = { instanceRef: { get, put, post }, configServiceRef: { loadConfig: async () => ({ client_id: 'console-client', project_id: 'system-project' }) } } as unknown as HttpClient;
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

// The Save button has no accessible name suffix — just "Save".
const ready = () => screen.findByRole('button', { name: 'Save' });
const preview = () => screen.getByTitle('Login page preview').getAttribute('srcdoc');
// The file input is hidden but still has an aria-label, so fireEvent.change works directly.
const upload = (name: 'logo' | 'favicon', file?: File) => fireEvent.change(screen.getByLabelText(`Upload ${name}`), { target: { files: file ? [file] : [] } });
// With the mocked dropdown, items are rendered inline as buttons — click directly by text.
const clickDropdownItem = (itemName: string) => fireEvent.click(screen.getByText(itemName));

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
    // When is_default, logo comes from project.logo_url not branding.logo_url.
    expect(screen.getByAltText('Logo preview')).toHaveAttribute('src', project.logo_url);
    // Favicon is always taken from branding.favicon_url regardless of is_default.
    expect(screen.getAllByAltText('Favicon preview')[0]).toHaveAttribute('src', stored.favicon_url);
    // Default overrides: shadow on, border off.
    expect(screen.getByRole('switch', { name: 'Enable shadow' })).toHaveAttribute('aria-checked', 'true');
    expect(screen.getByRole('switch', { name: 'Enable border' })).toHaveAttribute('aria-checked', 'false');
  });

  it('edits appearance with live preview then persists settings and query cache', async () => {
    const client = mount(); await ready();
    fireEvent.change(screen.getByLabelText('Primary color'), { target: { value: '#abcdef' } });
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
    fireEvent.change(screen.getByLabelText('Primary color'), { target: { value: 'invalid' } });
    expect(await ready()).toBeDisabled();
    expect(screen.getByLabelText('Primary color')).toHaveAttribute('aria-invalid', 'true');
    fireEvent.change(screen.getByLabelText('Choose primary color'), { target: { value: '#556677' } });
    expect(await ready()).not.toBeDisabled();
    expect(preview()).toContain('--brand-primary:#556677');
  });

  it('disables editing during save and preserves the draft after a failed save', async () => {
    let reject!: (reason: Error) => void;
    put.mockReturnValue(new Promise((_, r) => { reject = r; }));
    mount(); fireEvent.click(await ready());
    await waitFor(() => expect(screen.getByLabelText('Primary color')).toBeDisabled());
    expect(await ready()).toBeDisabled();
    reject(new Error('Permission denied'));
    await waitFor(() => expect(toast).toHaveBeenCalledWith(expect.objectContaining({ title: 'Save failed', description: 'Permission denied' })));
    expect(await ready()).not.toBeDisabled();
    expect(screen.getByLabelText('Primary color')).toHaveValue('#123456');
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

  it('resets logo via dropdown and restores the branding logo_url value', async () => {
    mount(); await ready();
    // First change the logo to something else via upload to confirm reset works.
    const file = new File(['img'], 'new.png', { type: 'image/png' });
    upload('logo', file);
    await waitFor(() => expect(preview()).toContain('https://images.example/upload.png'));
    // Click "Use project logo" from the inline-rendered dropdown.
    clickDropdownItem('Use project logo');
    // Should reset to branding.logo_url (the stored value from API, not the uploaded one).
    expect(screen.getByAltText('Logo preview')).toHaveAttribute('src', stored.logo_url);
  });

  it('resets favicon via dropdown and restores the branding favicon_url value', async () => {
    mount(); await ready();
    // Change favicon via upload first.
    const file = new File(['img'], 'new.png', { type: 'image/png' });
    upload('favicon', file);
    await waitFor(() => screen.getAllByAltText('Favicon preview').forEach(img => expect(img).toHaveAttribute('src', 'https://images.example/upload.png')));
    // Click "Use project favicon" from the inline-rendered dropdown.
    clickDropdownItem('Use project favicon');
    // Should reset to branding.favicon_url — check the thumbnail (first match in the card).
    expect(screen.getAllByAltText('Favicon preview')[0]).toHaveAttribute('src', stored.favicon_url);
  });

  it('navigates back to project settings with restored tab params', async () => {
    mount(); await ready();
    fireEvent.click(screen.getByRole('button', { name: /Back to Acme/i }));
    expect(screen.getByRole('status')).toHaveTextContent('/projects/project/details?tab=settings&page=2');
  });

});
