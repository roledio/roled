import { AxiosError, type AxiosResponse } from 'axios';
import { describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '@/services/core/httpClient';
import { brandingError, fetchProjectBranding, updateProjectBranding, uploadBrandingAsset, type BrandingSettings } from './branding';

const settings: BrandingSettings = { logo_url: null, favicon_url: null, primary_color: '#123456', rounding: 'sharp', enable_shadow: false, enable_border: false };
const file = new File(['image'], 'icon.ico', { type: 'image/x-icon' });
describe.each([
  { name: 'fetch', method: 'get', fallback: 'Unable to load branding', invoke: (c: HttpClient) => fetchProjectBranding(c, 'https://auth.example/', 'tenant') },
  { name: 'update', method: 'put', fallback: 'Unable to save branding', invoke: (c: HttpClient) => updateProjectBranding(c, 'https://auth.example/', 'tenant', settings) },
  { name: 'upload', method: 'post', fallback: 'Unable to upload image', invoke: (c: HttpClient) => uploadBrandingAsset(c, 'https://auth.example/', file, 'favicon') },
])('branding $name errors', operation => {
  it.each([undefined, 'Not authorized'])('rejects unsuccessful envelopes with message %s', async message => {
    const request = vi.fn().mockResolvedValue({ data: { success: false, error: message ? { message } : undefined } });
    const client = { instanceRef: { [operation.method]: request } } as unknown as HttpClient;
    await expect(operation.invoke(client)).rejects.toThrow(message ?? operation.fallback);
    expect(request).toHaveBeenCalledTimes(1);
  });
  it('preserves transport errors for the caller', async () => {
    const error = new Error('Offline');
    const client = { instanceRef: { [operation.method]: vi.fn().mockRejectedValue(error) } } as unknown as HttpClient;
    await expect(operation.invoke(client)).rejects.toBe(error);
  });
});

it('selects a server error message before the Axios fallback', () => {
  const error = new AxiosError('Transport error');
  error.response = { data: { error: { message: 'Forbidden' } } } as AxiosResponse;
  expect(brandingError(error)).toBe('Forbidden');
  error.response = { data: {} } as AxiosResponse;
  expect(brandingError(error)).toBe('Transport error');
  error.response = undefined;
  expect(brandingError(error)).toBe('Transport error');
  expect(brandingError(new Error('Invalid image'))).toBe('Invalid image');
  expect(brandingError(null)).toBe('Unable to update branding');
});
