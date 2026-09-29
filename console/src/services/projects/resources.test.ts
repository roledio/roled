import { describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '@/services/core/httpClient';
import { createProjectResource, deleteProjectResource, fetchProjectResourceById, fetchProjectResources, updateProjectResource } from './resources';

const baseUrl = 'https://auth.example.com/';
const headers = { 'Content-Type': 'application/json' };
const resourcePayload = { name: 'Console', code: 'console', permissions: [{ name: 'Read', code: 'read' }] };

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
  { name: 'list resources', method: 'get', path: '/resources', list: true, invalidMessage: 'Failed to fetch resources', fallbackMessage: 'Failed to fetch resources', invoke: c => fetchProjectResources(c, baseUrl, 'project') },
  { name: 'get resource', method: 'get', path: '/resources/resource', invalidMessage: 'Failed to fetch resource', fallbackMessage: 'Failed to fetch resource', invoke: c => fetchProjectResourceById(c, baseUrl, 'project', 'resource') },
  { name: 'create resource', method: 'post', path: '/resources', payload: resourcePayload, invalidMessage: 'Failed to create resource', fallbackMessage: 'Failed to create resource', invoke: c => createProjectResource(c, baseUrl, 'project', resourcePayload) },
  { name: 'update resource', method: 'put', path: '/resources/resource', payload: { ...resourcePayload, description: 'Updated resource' }, invalidMessage: 'Failed to update resource', fallbackMessage: 'Failed to update resource', invoke: c => updateProjectResource(c, baseUrl, 'project', 'resource', { ...resourcePayload, description: 'Updated resource' }) },
  { name: 'delete resource', method: 'delete', path: '/resources/resource', voidResult: true, invalidMessage: 'Failed to delete resource', fallbackMessage: 'Failed to delete resource', invoke: c => deleteProjectResource(c, baseUrl, 'project', 'resource') },

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
        params: { page_num: 1, page_size: 10, sort_by: '', sort_dir: '', search: undefined, ...(operation.path === '/resources' ? { is_default: null } : {}) },
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
  it('forwards resource pagination, sorting, search, and an explicit non-default filter', async () => {
    const { client, request } = makeClient({ data: { success: true } });
    await expect(fetchProjectResources(client, 'https://auth.example.com', 'project', 3, 20, 'name', 'desc', 'Console', false)).resolves.toEqual({ data: [], pagination: undefined });
    expect(request).toHaveBeenCalledWith('https://auth.example.com/api/v1/projects/project/resources', {
      headers, params: { page_num: 3, page_size: 20, sort_by: 'name', sort_dir: 'desc', search: 'Console', is_default: false },
    });
  });

});
