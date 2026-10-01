import { describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '@/services/core/httpClient';
import { fetchCurrentUser, updateCurrentUser } from '@/services/users';
import { createProjectUser, deleteProjectUser, fetchProjectUsers, fetchUserById, inviteProjectUser, requestProjectUserPasswordReset, resendProjectUserVerificationEmail, updateProjectUser } from './users';

const base = 'https://auth.example/';
const body = { display_name: 'Member', email: 'member@example.com', password: 'test-password', avatar_url: null };
const invitation = { email: 'invite@example.com', role_id: 'reader', redirect_uri: 'https://app.example/callback' };
const reset = { redirect_uri: 'https://app.example/callback' };
const options = { headers: { 'Content-Type': 'application/json' } };
const operations = [
  { name: 'list', method: 'get', path: '/projects/tenant/users', invalid: 'Invalid users response', fallback: 'Failed to fetch users', body: undefined, invoke: (c: HttpClient) => fetchProjectUsers(c, base, 'tenant') },
  { name: 'details', method: 'get', path: '/projects/tenant/users/user', invalid: 'Failed to fetch user', fallback: 'Failed to fetch user', body: undefined, invoke: (c: HttpClient) => fetchUserById(c, base, 'tenant', 'user') },
  { name: 'create', method: 'post', path: '/projects/tenant/users', invalid: 'Failed to create user', fallback: 'Failed to create user', body, invoke: (c: HttpClient) => createProjectUser(c, base, 'tenant', body) },
  { name: 'update', method: 'put', path: '/projects/tenant/users/user', invalid: 'Failed to update user', fallback: 'Failed to update user', body, invoke: (c: HttpClient) => updateProjectUser(c, base, 'tenant', 'user', body) },
  { name: 'delete', method: 'delete', path: '/projects/tenant/users/user', invalid: 'Failed to delete user', fallback: 'Failed to delete user', body: undefined, invoke: (c: HttpClient) => deleteProjectUser(c, base, 'tenant', 'user') },
  { name: 'verify', method: 'post', path: '/projects/tenant/users/user/verification-email', invalid: 'Failed to send verification email', fallback: 'Failed to send verification email', body: reset, invoke: (c: HttpClient) => resendProjectUserVerificationEmail(c, base, 'tenant', 'user', reset) },
  { name: 'reset', method: 'post', path: '/projects/tenant/users/user/password-reset', invalid: 'Failed to request password reset', fallback: 'Failed to request password reset', body: reset, invoke: (c: HttpClient) => requestProjectUserPasswordReset(c, base, 'tenant', 'user', reset) },
  { name: 'invite', method: 'post', path: '/projects/tenant/users/invitation', invalid: 'Failed to invite user', fallback: 'Failed to invite user', body: invitation, invoke: (c: HttpClient) => inviteProjectUser(c, base, 'tenant', invitation) },
  { name: 'current', method: 'get', path: '/users/current', invalid: 'Invalid user response', fallback: 'Failed to fetch current user', body: undefined, invoke: (c: HttpClient) => fetchCurrentUser(c, base) },
  { name: 'update current', method: 'put', path: '/users/current', invalid: 'Failed to update current user', fallback: 'Failed to update current user', body, invoke: (c: HttpClient) => updateCurrentUser(c, base, body) },
];

function setup(method: string, response: unknown, rejected = false) {
  const request = rejected ? vi.fn().mockRejectedValue(response) : vi.fn().mockResolvedValue(response);
  return { request, client: { instanceRef: { [method]: request } } as unknown as HttpClient };
}

describe.each(operations)('user API $name', operation => {
  it('uses the tenant endpoint, preserves the payload, and returns the API data', async () => {
    const data = operation.name === 'list' ? [{ id: 'user', ...body }] : { id: 'user', ...body };
    const pagination = { page_num: 1, page_size: 10, total_data: 1 };
    const { client, request } = setup(operation.method, { data: { success: true, data, pagination } });
    const result = await operation.invoke(client);
    if (['delete', 'verify', 'reset'].includes(operation.name)) expect(result).toBeUndefined();
    else if (operation.name === 'list') expect(result).toEqual({ data, pagination });
    else expect(result).toEqual(data);
    const url = `https://auth.example/api/v1${operation.path}`;
    if (operation.name === 'list') expect(request).toHaveBeenCalledWith(url, { ...options, params: { page_num: 1, page_size: 10, sort_by: '', sort_dir: '', search: undefined } });
    else if (operation.body) expect(request).toHaveBeenCalledWith(url, operation.body, options);
    else expect(request).toHaveBeenCalledWith(url, options);
    expect(request).toHaveBeenCalledTimes(1);
  });

  it.each([
    { response: { data: { success: false, error: { message: 'Forbidden' } } }, message: 'Forbidden' },
    { response: { data: { success: false } }, message: undefined },
    { response: {}, message: undefined },
  ])('rejects unsuccessful or absent envelopes', async ({ response, message }) => {
    const { client } = setup(operation.method, response);
    await expect(operation.invoke(client)).rejects.toThrow(message ?? operation.invalid);
  });

  it.each([
    { error: { response: { data: { error: { message: 'Domain error' }, message: 'Outer error' } }, message: 'Transport error' }, message: 'Domain error' },
    { error: { response: { data: { message: 'Outer error' } }, message: 'Transport error' }, message: 'Outer error' },
    { error: new Error('Offline'), message: 'Offline' },
    { error: {}, message: undefined },
  ])('propagates errors with the documented precedence', async ({ error, message }) => {
    const { client } = setup(operation.method, error, true);
    await expect(operation.invoke(client)).rejects.toThrow(message ?? operation.fallback);
  });
});

it.each([true, false])('preserves the active filter %s and optional role/search/sort pagination', async active => {
  const { client, request } = setup('get', { data: { success: true } });
  expect(await fetchProjectUsers(client, base, 'tenant', 3, 25, 'email', 'desc', 'member', active, 'reader')).toEqual({ data: [], pagination: undefined });
  expect(request).toHaveBeenCalledWith('https://auth.example/api/v1/projects/tenant/users', { ...options, params: { page_num: 3, page_size: 25, sort_by: 'email', sort_dir: 'desc', search: 'member', is_active: active, role_id: 'reader' } });
});

it.each([resendProjectUserVerificationEmail, requestProjectUserPasswordReset])('sends an empty object when no redirect override is supplied', async invoke => {
  const { client, request } = setup('post', { data: { success: true } });
  await invoke(client, base, 'tenant', 'user');
  expect(request.mock.calls[0][1]).toEqual({});
});

it.each([fetchCurrentUser, (c: HttpClient, url: string) => updateCurrentUser(c, url, body)])('rejects a missing current-user payload even with success=true', async invoke => {
  const request = vi.fn().mockResolvedValue({ data: { success: true } });
  const client = { instanceRef: { get: request, put: request } } as unknown as HttpClient;
  await expect(invoke(client, base)).rejects.toThrow();
});
