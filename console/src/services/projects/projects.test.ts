import { describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '@/services/core/httpClient';
import { createProject, deleteProject, fetchProjectById, fetchProjects, updateProject, uploadProjectLogo, uploadUserAvatar } from './projects';

const baseUrl = 'https://auth.example.com/';
const payload = { name: 'Example', description: 'Description', logo_url: null, redirect_uris: [{ redirect_uri: 'https://app.example.com/callback' }] };
const image = new File(['test-image'], 'logo.png', { type: 'image/png' });

describe.each([
  { name: 'list', method: 'get', path: '/api/v1/projects', invalid: 'Invalid projects response', fallback: 'Failed to fetch projects', invoke: (c: HttpClient) => fetchProjects(c, baseUrl) },
  { name: 'details', method: 'get', path: '/api/v1/projects/project', invalid: 'Invalid project response', fallback: 'Failed to fetch project', invoke: (c: HttpClient) => fetchProjectById(c, baseUrl, 'project') },
  { name: 'create', method: 'post', path: '/api/v1/projects', invalid: 'Failed to create project', fallback: 'Failed to create project', invoke: (c: HttpClient) => createProject(c, baseUrl, payload) },
  { name: 'update', method: 'put', path: '/api/v1/projects/project', invalid: 'Failed to update project', fallback: 'Failed to update project', invoke: (c: HttpClient) => updateProject(c, baseUrl, 'project', payload) },
  { name: 'delete', method: 'post', path: '/api/v1/projects/project/delete', invalid: 'Failed to delete project', fallback: 'Failed to delete project', invoke: (c: HttpClient) => deleteProject(c, baseUrl, 'project', { name: 'Example' }) },
  { name: 'logo', method: 'post', path: '/api/v1/uploads', invalid: 'Failed to upload logo', fallback: 'Failed to upload logo', invoke: (c: HttpClient) => uploadProjectLogo(c, baseUrl, 'project', image) },
  { name: 'avatar', method: 'post', path: '/api/v1/uploads', invalid: 'Failed to upload avatar', fallback: 'Failed to upload avatar', invoke: (c: HttpClient) => uploadUserAvatar(c, baseUrl, image) },
])('project API $name', operation => {
  function setup(response: unknown, rejected = false) {
    const request = rejected ? vi.fn().mockRejectedValue(response) : vi.fn().mockResolvedValue(response);
    const methods = { get: vi.fn(), post: vi.fn(), put: vi.fn() };
    methods[operation.method as keyof typeof methods] = request;
    return { client: { instanceRef: methods } as unknown as HttpClient, request };
  }

  it('preserves payload and response using the documented endpoint', async () => {
    const isUpload = operation.name === 'logo' || operation.name === 'avatar';
    const data = isUpload ? { url: 'https://uploads.example.com/image.png' } : operation.name === 'list' ? [{ id: 'project', ...payload }] : { id: 'project', ...payload };
    const pagination = { page_num: 1, page_size: 1, total_data: 1 };
    const { client, request } = setup({ data: { success: true, data, pagination } });
    const got = await operation.invoke(client);
    if (operation.name === 'delete') expect(got).toBeUndefined();
    else if (operation.name === 'logo') expect(got).toEqual({ logo_url: 'https://uploads.example.com/image.png' });
    else if (operation.name === 'list') expect(got).toEqual({ data, pagination });
    else expect(got).toEqual(data);
    expect(request).toHaveBeenCalledTimes(1);
    const url = `https://auth.example.com${operation.path}`;
    const jsonOptions = { headers: { 'Content-Type': 'application/json' } };
    if (isUpload) {
      expect(request).toHaveBeenCalledWith(url, expect.any(FormData), { headers: { 'Content-Type': 'multipart/form-data' } });
      const form = request.mock.calls[0][1] as FormData;
      expect(form.get('file')).toBe(image);
      expect(form.get('type')).toBe(operation.name === 'logo' ? 'project-logo' : 'user-avatar');
    } else if (operation.name === 'list') expect(request).toHaveBeenCalledWith(url, { ...jsonOptions, params: {} });
    else if (operation.name === 'details') expect(request).toHaveBeenCalledWith(url, jsonOptions);
    else expect(request).toHaveBeenCalledWith(url, operation.name === 'delete' ? { name: 'Example' } : payload, jsonOptions);
  });

  it.each([
    { response: { data: { success: false, error: { message: 'Forbidden' } } }, message: 'Forbidden' },
    { response: { data: { success: false } }, message: undefined },
    { response: {}, message: undefined },
  ])('rejects unsuccessful and absent response envelopes', async ({ response, message }) => {
    const { client } = setup(response);
    await expect(operation.invoke(client)).rejects.toThrow(message ?? operation.invalid);
  });

  it.each([
    { error: { response: { data: { error: { message: 'Domain error' }, message: 'Outer error' } }, message: 'Transport error' }, message: 'Domain error' },
    { error: { response: { data: { message: 'Outer error' } }, message: 'Transport error' }, message: 'Outer error' },
    { error: new Error('Transport error'), message: 'Transport error' },
    { error: null, message: undefined },
  ])('preserves backend error precedence and fallback', async ({ error, message }) => {
    const { client } = setup(error, true);
    await expect(operation.invoke(client)).rejects.toThrow(message ?? operation.fallback);
  });
});

it('forwards project filters and defaults missing result data to an empty list', async () => {
  const get = vi.fn().mockResolvedValue({ data: { success: true } });
  const client = { instanceRef: { get } } as unknown as HttpClient;
  const params = { page_num: 3, page_size: 20, search: 'Example', is_active: false, sort_by: 'name', sort_dir: 'desc' };
  await expect(fetchProjects(client, 'https://auth.example.com', params)).resolves.toEqual({ data: [], pagination: undefined });
  expect(get).toHaveBeenCalledWith('https://auth.example.com/api/v1/projects', { params, headers: { 'Content-Type': 'application/json' } });
});
