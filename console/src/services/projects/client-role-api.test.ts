import { describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '@/services/core/httpClient';
import { createClient, deleteClient, fetchClientById, fetchProjectClients, updateClient } from './clients';
import { createProjectRole, deleteProjectRole, fetchProjectRoles, fetchRoleById, setSignupRole, updateProjectRole } from './roles';

const baseUrl = 'https://auth.example.com/';
const headers = { 'Content-Type': 'application/json' };
const clientPayload = { name: 'Console', permission_ids: ['read-users'] };
const rolePayload = { name: 'Reader', code: 'reader', permission_ids: ['read-users'] };

interface Operation {
  name: string;
  method: 'get' | 'post' | 'put' | 'delete' | 'patch';
  path: string;
  payload?: unknown;
  list?: boolean;
  voidResult?: boolean;
  invalidMessage: string;
  fallbackMessage: string;
  invoke: (client: HttpClient) => Promise<unknown>;
}

const operations: Operation[] = [
  { name: 'list clients', method: 'get', path: '/clients', list: true, invalidMessage: 'Invalid clients response', fallbackMessage: 'Failed to fetch clients', invoke: c => fetchProjectClients(c, baseUrl, 'project') },
  { name: 'get client', method: 'get', path: '/clients/client', invalidMessage: 'Failed to fetch client', fallbackMessage: 'Failed to fetch client', invoke: c => fetchClientById(c, baseUrl, 'project', 'client') },
  { name: 'create client', method: 'post', path: '/clients', payload: clientPayload, invalidMessage: 'Failed to create client', fallbackMessage: 'Failed to create client', invoke: c => createClient(c, baseUrl, 'project', clientPayload) },
  { name: 'update client', method: 'put', path: '/clients/client', payload: { ...clientPayload, is_active: false }, invalidMessage: 'Failed to update client', fallbackMessage: 'Failed to update client', invoke: c => updateClient(c, baseUrl, 'project', 'client', { ...clientPayload, is_active: false }) },
  { name: 'delete client', method: 'delete', path: '/clients/client', voidResult: true, invalidMessage: 'Failed to delete client', fallbackMessage: 'Failed to delete client', invoke: c => deleteClient(c, baseUrl, 'project', 'client') },
  { name: 'list roles', method: 'get', path: '/roles', list: true, invalidMessage: 'Invalid roles response', fallbackMessage: 'Failed to fetch roles', invoke: c => fetchProjectRoles(c, baseUrl, 'project') },
  { name: 'get role', method: 'get', path: '/roles/role', invalidMessage: 'Failed to fetch role', fallbackMessage: 'Failed to fetch role', invoke: c => fetchRoleById(c, baseUrl, 'project', 'role') },
  { name: 'create role', method: 'post', path: '/roles', payload: rolePayload, invalidMessage: 'Failed to create role', fallbackMessage: 'Failed to create role', invoke: c => createProjectRole(c, baseUrl, 'project', rolePayload) },
  { name: 'update role', method: 'put', path: '/roles/role', payload: rolePayload, invalidMessage: 'Failed to update role', fallbackMessage: 'Failed to update role', invoke: c => updateProjectRole(c, baseUrl, 'project', 'role', rolePayload) },
  { name: 'delete role', method: 'delete', path: '/roles/role', voidResult: true, invalidMessage: 'Failed to delete role', fallbackMessage: 'Failed to delete role', invoke: c => deleteProjectRole(c, baseUrl, 'project', 'role') },
  { name: 'set signup role', method: 'patch', path: '/signup-role', payload: { role_id: 'role' }, voidResult: true, invalidMessage: 'Failed to set sign-up role', fallbackMessage: 'Failed to set sign-up role', invoke: c => setSignupRole(c, baseUrl, 'project', 'role') },
];

function makeClient(response: unknown, rejected = false) {
  const request = rejected ? vi.fn().mockRejectedValue(response) : vi.fn().mockResolvedValue(response);
  const client = { instanceRef: { get: request, post: request, put: request, delete: request, patch: request } } as unknown as HttpClient;
  return { client, request };
}

describe.each(operations)('$name', operation => {
  it('uses the API contract and returns the response data', async () => {
    const data = operation.list ? [{ id: 'returned-id', name: 'Returned' }] : { id: 'returned-id', name: 'Returned' };
    const pagination = { page_num: 1, page_size: 1, total_data: 1 };
    // Separate method spies make a wrong HTTP verb fail this contract test.
    const request = vi.fn().mockResolvedValue({ data: { success: true, data, pagination } });
    const methods = { get: vi.fn(), post: vi.fn(), put: vi.fn(), delete: vi.fn(), patch: vi.fn() };
    methods[operation.method] = request;
    const client = { instanceRef: methods } as unknown as HttpClient;
    const result = await operation.invoke(client);
    expect(result).toEqual(operation.voidResult ? undefined : operation.list ? { data, pagination } : data);
    const url = `https://auth.example.com/api/v1/projects/project${operation.path}`;
    if (operation.list) {
      expect(request).toHaveBeenCalledWith(url, {
        headers,
        params: { page_num: 1, page_size: 10, sort_by: '', sort_dir: '', search: undefined, ...(operation.path === '/clients' ? { is_active: null } : {}) },
      });
    } else if (operation.payload) {
      expect(request).toHaveBeenCalledWith(url, operation.payload, { headers });
    } else {
      expect(request).toHaveBeenCalledWith(url, { headers });
    }
    expect(request).toHaveBeenCalledTimes(1);
  });

  it('preserves a rejected API error message', async () => {
    const { client } = makeClient({ data: { success: false, error: { message: 'Permission denied' } } });
    await expect(operation.invoke(client)).rejects.toThrow('Permission denied');
  });

  it.each([{ data: { success: false } }, {}])('rejects unsuccessful or missing response envelopes', async response => {
    const { client } = makeClient(response);
    await expect(operation.invoke(client)).rejects.toThrow(operation.invalidMessage);
  });

  it.each([
    { error: { response: { data: { error: { message: 'Domain error' }, message: 'Outer error' } }, message: 'Transport error' }, message: 'Domain error' },
    { error: { response: { data: { message: 'Outer error' } }, message: 'Transport error' }, message: 'Outer error' },
    { error: new Error('Transport error'), message: 'Transport error' },
  ])('propagates error precedence: $message', async ({ error, message }) => {
    const { client } = makeClient(error, true);
    await expect(operation.invoke(client)).rejects.toThrow(message);
  });

  it('uses the operation fallback for an unknown rejection', async () => {
    const { client } = makeClient(null, true);
    await expect(operation.invoke(client)).rejects.toThrow(operation.fallbackMessage);
  });
});

describe('list filters and empty results', () => {
  it('forwards client pagination, sorting, search, and an explicit inactive filter', async () => {
    const { client, request } = makeClient({ data: { success: true } });
    await expect(fetchProjectClients(client, 'https://auth.example.com', 'project', 3, 20, 'name', 'desc', 'Console', false)).resolves.toEqual({ data: [], pagination: undefined });
    expect(request).toHaveBeenCalledWith('https://auth.example.com/api/v1/projects/project/clients', {
      headers, params: { page_num: 3, page_size: 20, sort_by: 'name', sort_dir: 'desc', search: 'Console', is_active: false },
    });
  });

  it('forwards role filters and supplies an empty list when data is missing', async () => {
    const { client, request } = makeClient({ data: { success: true } });
    await expect(fetchProjectRoles(client, 'https://auth.example.com', 'project', 2, 25, 'name', 'asc', 'Reader')).resolves.toEqual({ data: [], pagination: undefined });
    expect(request).toHaveBeenCalledWith('https://auth.example.com/api/v1/projects/project/roles', {
      headers, params: { page_num: 2, page_size: 25, sort_by: 'name', sort_dir: 'asc', search: 'Reader' },
    });
  });
});
