import { describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '@/services/core/httpClient';
import { fetchProjectSettings, updateProjectSettings, type ProjectSettings } from './settings';

const disabled: ProjectSettings = {
  is_signup_enabled: false,
  default_signup_role_id: null,
  is_signup_verify_email: false,
  is_forgot_password_enabled: true,
  is_allow_temp_email: false,
};

describe.each([
  { name: 'fetch', method: 'get' as const, invalid: 'Invalid project settings response', fallback: 'Failed to fetch project settings', invoke: (c: HttpClient) => fetchProjectSettings(c, 'https://auth.example.com/', 'project') },
  { name: 'update', method: 'put' as const, invalid: 'Failed to update project settings', fallback: 'Failed to update project settings', invoke: (c: HttpClient) => updateProjectSettings(c, 'https://auth.example.com/', 'project', disabled) },
])('$name project settings', operation => {
  function setup(response: unknown, reject = false) {
    const request = reject ? vi.fn().mockRejectedValue(response) : vi.fn().mockResolvedValue(response);
    const methods = { get: vi.fn(), put: vi.fn() };
    methods[operation.method] = request;
    return { client: { instanceRef: methods } as unknown as HttpClient, request };
  }

  it('preserves boolean and null fields and uses the correct method and URL', async () => {
    const { client, request } = setup({ data: { success: true, data: disabled } });
    await expect(operation.invoke(client)).resolves.toEqual(disabled);
    const url = 'https://auth.example.com/api/v1/projects/project/settings';
    const options = { headers: { 'Content-Type': 'application/json' } };
    if (operation.method === 'put') expect(request).toHaveBeenCalledWith(url, disabled, options);
    else expect(request).toHaveBeenCalledWith(url, options);
    expect(request).toHaveBeenCalledTimes(1);
  });

  it.each([
    { response: { data: { success: false, error: { message: 'Forbidden' } } }, message: 'Forbidden' },
    { response: { data: { success: false } }, message: undefined },
    { response: {}, message: undefined },
  ])('rejects invalid response envelopes', async ({ response, message }) => {
    const { client } = setup(response);
    await expect(operation.invoke(client)).rejects.toThrow(message ?? operation.invalid);
  });

  it.each([
    { error: { response: { data: { error: { message: 'Domain error' }, message: 'Outer error' } }, message: 'Transport error' }, message: 'Domain error' },
    { error: { response: { data: { message: 'Outer error' } }, message: 'Transport error' }, message: 'Outer error' },
    { error: new Error('Transport error'), message: 'Transport error' },
    { error: null, message: undefined },
  ])('propagates errors in priority order', async ({ error, message }) => {
    const { client } = setup(error, true);
    await expect(operation.invoke(client)).rejects.toThrow(message ?? operation.fallback);
  });
});
