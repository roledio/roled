import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { mock } from 'vitest-mock-extended';
import type { HttpClient } from '@/services/core/httpClient';
import * as api from '@/services/projects';
import { getProjectTabParams } from '@/lib/paramsStore';
import NewSocialConnection from './NewSocialConnection';
import SocialConnectionDetails from './SocialConnectionDetails';

vi.mock('@/services/projects', () => ({
  fetchProjectById: vi.fn(), fetchOAuthConnectionByProvider: vi.fn(),
  createOAuthConnectionWithCredentials: vi.fn(), updateOAuthConnection: vi.fn(),
}));
vi.mock('@/lib/paramsStore', () => ({ getProjectTabParams: vi.fn() }));
const { toast } = vi.hoisted(() => ({ toast: vi.fn() }));
vi.mock('@/hooks/use-toast', () => ({ useToast: () => ({ toast }) }));

const httpClient = mock<HttpClient>();
const project = { ...mock<Awaited<ReturnType<typeof api.fetchProjectById>>>(), id: 'project-1', name: 'Acme', logo_url: undefined };
const connection = {
  id: 'connection-1', project_id: 'project-1', provider: 'google',
  credential_type: 'custom' as const, client_id: 'existing-client', client_secret: 'existing-secret',
  scopes: ['email'], enabled: true, created_at: '', updated_at: '',
};
const clients: QueryClient[] = [];

function Destination() {
  const location = useLocation();
  return <output data-testid="destination">{location.pathname}{location.search}</output>;
}

function mount(mode: 'create' | 'edit', provider = 'google', from?: string) {
  const client = new QueryClient({ defaultOptions: { queries: { retryDelay: 0, gcTime: Infinity }, mutations: { retry: false } } });
  clients.push(client);
  const invalidate = vi.spyOn(client, 'invalidateQueries');
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={[{ pathname: `/projects/project-1/connections/${provider}`, state: { from } }]}>
        <Routes>
          <Route path="/projects/:project_id/connections/:provider" element={mode === 'create' ? <NewSocialConnection httpClient={httpClient} /> : <SocialConnectionDetails httpClient={httpClient} />} />
          <Route path="/projects/:project_id/details" element={<Destination />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
  return { client, invalidate };
}

function change(label: string, value: string) {
  fireEvent.change(screen.getByLabelText(label), { target: { value } });
}

beforeEach(() => {
  vi.resetAllMocks();
  vi.stubEnv('VITE_AUTH_BASE_URL', 'https://auth.example');
  vi.mocked(api.fetchProjectById).mockResolvedValue(project);
  vi.mocked(api.fetchOAuthConnectionByProvider).mockResolvedValue(connection);
  vi.mocked(api.createOAuthConnectionWithCredentials).mockResolvedValue(connection);
  vi.mocked(api.updateOAuthConnection).mockResolvedValue(connection);
  vi.mocked(getProjectTabParams).mockReturnValue('');
  Object.defineProperty(navigator, 'clipboard', { configurable: true, value: { writeText: vi.fn().mockResolvedValue(undefined) } });
  HTMLElement.prototype.hasPointerCapture = () => false;
  HTMLElement.prototype.scrollIntoView = () => {};
});

afterEach(() => {
  cleanup();
  clients.splice(0).forEach(client => client.clear());
  vi.unstubAllEnvs();
});

describe.each(['create', 'edit'] as const)('%s social connection', mode => {
  const action = () => screen.getByTestId(`${mode === 'create' ? 'create' : 'save'}-social-connection-btn`);
  const mutation = () => mode === 'create' ? api.createOAuthConnectionWithCredentials : api.updateOAuthConnection;
  const ready = async () => screen.findByRole('radio', { name: /Custom/ });

  it('shows loading while project data is pending', () => {
    vi.mocked(api.fetchProjectById).mockReturnValue(new Promise(() => {}));
    mount(mode);
    expect(screen.getByText('Loading…')).toBeInTheDocument();
    expect(screen.queryByRole('radiogroup')).not.toBeInTheDocument();
  });

  it('reports project load failures without displaying an editable form', async () => {
    vi.mocked(api.fetchProjectById).mockRejectedValue(new Error('Project access denied'));
    mount(mode);
    expect(await screen.findByText('Project access denied')).toBeInTheDocument();
    expect(screen.queryByRole('radiogroup')).not.toBeInTheDocument();
  });

  it('rejects unknown providers and restores back navigation', async () => {
    mount(mode, 'unsupported', '?tab=settings&page=2');
    expect(await screen.findByText('Unknown provider')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Go back'));
    expect(await screen.findByTestId('destination')).toHaveTextContent('/projects/project-1/details?tab=settings&page=2');
    expect(mutation()).not.toHaveBeenCalled();
  });

  it('validates custom credentials before sending a mutation', async () => {
    mount(mode);
    fireEvent.click(await ready());
    change('Client ID', ' ');
    change('Client Secret', '');
    fireEvent.blur(screen.getByLabelText('Client ID'));
    fireEvent.blur(screen.getByLabelText('Client Secret'));
    fireEvent.click(action());
    expect(screen.getByText('Client ID is required for custom credentials')).toBeInTheDocument();
    expect(screen.getByText('Client Secret is required for custom credentials')).toBeInTheDocument();
    expect(screen.getByLabelText('Client ID')).toHaveAttribute('aria-invalid', 'true');
    expect(mutation()).not.toHaveBeenCalled();
  });

  it('submits custom credentials, updates cache, and restores saved settings parameters', async () => {
    vi.mocked(getProjectTabParams).mockReturnValue('?tab=settings&search=google');
    const { client, invalidate } = mount(mode, 'google', '?tab=settings&page=3');
    fireEvent.click(await ready());
    change('Client ID', 'new-client');
    change('Client Secret', 'new-secret');
    fireEvent.change(screen.getByTestId('scope-tag-input'), { target: { value: 'openid,' } });
    fireEvent.click(action());
    await waitFor(() => expect(mutation()).toHaveBeenCalledWith(httpClient, 'https://auth.example', 'project-1', 'google', {
      credential_type: 'custom', client_id: 'new-client', client_secret: 'new-secret', enabled: true,
      scopes: mode === 'create' ? ['openid'] : ['email', 'openid'],
    }));
    expect(await screen.findByTestId('destination')).toHaveTextContent('?tab=settings&search=google');
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ['project', 'project-1', 'oauth-connections'] });
    expect(toast).toHaveBeenCalledWith(expect.objectContaining({ title: mode === 'create' ? 'Connection created' : 'Connection updated' }));
    if (mode === 'edit') expect(client.getQueryData(['project', 'project-1', 'oauth-connection', 'google'])).toEqual(connection);
  });

  it('omits custom credentials when using defaults and handles disabled status', async () => {
    mount(mode);
    await ready();
    fireEvent.click(screen.getByRole('radio', { name: /Default/ }));
    fireEvent.click(screen.getByRole('combobox'));
    fireEvent.click(await screen.findByRole('option', { name: 'Disabled' }));
    fireEvent.click(action());
    await waitFor(() => expect(mutation()).toHaveBeenCalledWith(httpClient, 'https://auth.example', 'project-1', 'google', {
      credential_type: 'default', client_id: undefined, client_secret: undefined, scopes: undefined, enabled: false,
    }));
    expect(await screen.findByTestId('destination')).toHaveTextContent('?tab=settings');
  });

  it('disables controls while saving and displays the API error without navigation', async () => {
    let reject!: (error: Error) => void;
    vi.mocked(mutation()).mockReturnValue(new Promise((_, rejectPromise) => { reject = rejectPromise; }));
    mount(mode);
    fireEvent.click(await ready());
    change('Client ID', 'valid-client');
    change('Client Secret', 'valid-secret');
    fireEvent.click(action());
    await waitFor(() => expect(action()).toBeDisabled());
    expect(screen.getByRole('radio', { name: /Default/ })).toBeDisabled();
    expect(screen.getByLabelText('Client ID')).toBeDisabled();
    expect(screen.getByTestId('scope-tag-input')).toBeDisabled();
    reject(new Error('Permission denied'));
    await waitFor(() => expect(toast).toHaveBeenCalledWith(expect.objectContaining({ description: 'Permission denied', variant: 'destructive' })));
    expect(action()).not.toBeDisabled();
    expect(screen.getByLabelText('Client ID')).toHaveValue('valid-client');
    expect(screen.queryByTestId('destination')).not.toBeInTheDocument();
  });

  it('toggles secret visibility and copies the callback URL', async () => {
    mount(mode);
    fireEvent.click(await ready());
    const secret = screen.getByLabelText('Client Secret');
    expect(secret).toHaveAttribute('type', 'password');
    fireEvent.click(screen.getByRole('button', { name: 'Show client secret' }));
    expect(secret).toHaveAttribute('type', 'text');
    fireEvent.click(screen.getByRole('button', { name: 'Hide client secret' }));
    expect(secret).toHaveAttribute('type', 'password');
    const redirect = screen.getByLabelText('Redirect URI');
    fireEvent.focus(redirect);
    expect(redirect).toHaveValue('https://auth.example/oauth/google/callback');
    fireEvent.click(screen.getByRole('button', { name: 'Copy redirect URI' }));
    await waitFor(() => expect(toast).toHaveBeenCalledWith(expect.objectContaining({ title: 'Copied' })));
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('https://auth.example/oauth/google/callback');
  });

  it('preserves edit values but clears create values when switching back to defaults', async () => {
    mount(mode);
    fireEvent.click(await ready());
    change('Client ID', 'draft-client');
    change('Client Secret', 'draft-secret');
    fireEvent.click(screen.getByRole('radio', { name: /Default/ }));
    expect(screen.queryByLabelText('Client ID')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('radio', { name: /Custom/ }));
    expect(screen.getByLabelText('Client ID')).toHaveValue(mode === 'edit' ? 'draft-client' : '');
    expect(screen.getByLabelText('Client Secret')).toHaveValue(mode === 'edit' ? 'draft-secret' : '');
    expect(screen.queryByText('Client ID is required for custom credentials')).not.toBeInTheDocument();
  });

  it('deduplicates and removes scope tags, and prevents submission with oversized scopes', async () => {
    mount(mode);
    fireEvent.click(await ready());
    change('Client ID', 'client');
    change('Client Secret', 'secret');
    const input = screen.getByTestId('scope-tag-input');
    const add = (value: string) => {
      fireEvent.change(input, { target: { value } });
      fireEvent.keyDown(input, { key: 'Enter' });
    };
    add('profile');
    add('profile');
    expect(screen.getAllByRole('button', { name: 'Remove scope profile' })).toHaveLength(1);
    fireEvent.click(screen.getByRole('button', { name: 'Remove scope profile' }));
    expect(screen.queryByRole('button', { name: 'Remove scope profile' })).not.toBeInTheDocument();
    add('temporary');
    fireEvent.keyDown(input, { key: 'Backspace' });
    expect(screen.queryByRole('button', { name: 'Remove scope temporary' })).not.toBeInTheDocument();
    add(' ');
    add('x'.repeat(300));
    fireEvent.click(action());
    expect(screen.getByText(/Each scope must be/)).toBeInTheDocument();
    expect(mutation()).not.toHaveBeenCalled();
  });

  it('shows a project logo and returns to the saved settings tab', async () => {
    vi.mocked(api.fetchProjectById).mockResolvedValue({ ...project, logo_url: 'https://example.com/logo.png' });
    mount(mode, 'google', '?tab=settings&page=4');
    await ready();
    expect(screen.getByRole('img', { name: 'Acme' })).toHaveAttribute('src', 'https://example.com/logo.png');
    fireEvent.click(screen.getByRole('button', { name: 'Back to Acme' }));
    expect(await screen.findByTestId('destination')).toHaveTextContent('?tab=settings&page=4');
  });
});

it('reports a connection lookup failure on the edit page', async () => {
  vi.mocked(api.fetchOAuthConnectionByProvider).mockRejectedValue(new Error('Connection missing'));
  mount('edit');
  expect(await screen.findByText('Connection missing')).toBeInTheDocument();
  expect(screen.queryByRole('radiogroup')).not.toBeInTheDocument();
});

it('populates disabled default connections with absent custom fields', async () => {
  vi.mocked(api.fetchOAuthConnectionByProvider).mockResolvedValue({ ...connection, credential_type: 'default', client_id: undefined, client_secret: undefined, scopes: undefined, enabled: false });
  mount('edit');
  expect(await screen.findByRole('radio', { name: /Default/ })).toHaveAttribute('aria-checked', 'true');
  expect(screen.getByRole('combobox')).toHaveTextContent('Disabled');
  fireEvent.click(screen.getByRole('radio', { name: /Custom/ }));
  expect(screen.getByLabelText('Client ID')).toHaveValue('');
});
