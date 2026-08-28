import { describe, it, expect, vi, afterEach } from 'vitest';
import { fetchCurrentTokenInfo, revokeCurrentToken } from './authService';
import type { HttpClient } from './httpClient';

const mockHttpClient = (mockResponse: any, isError = false): HttpClient => {
  return {
    instanceRef: {
      get: isError
        ? vi.fn().mockRejectedValue(mockResponse)
        : vi.fn().mockResolvedValue(mockResponse),
      post: isError
        ? vi.fn().mockRejectedValue(mockResponse)
        : vi.fn().mockResolvedValue(mockResponse),
    },
  } as unknown as HttpClient;
};

describe('AuthService API functions', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('fetchCurrentTokenInfo', () => {
    it('successfully fetches current token info', async () => {
      const mockInfo = {
        id: 'token123',
        user: { display_name: 'Admin User', email: 'admin@local.id' },
      };
      const httpClient = mockHttpClient({ data: { success: true, data: mockInfo } });
      const result = await fetchCurrentTokenInfo(httpClient, 'http://localhost:8082');

      expect(httpClient.instanceRef.get).toHaveBeenCalledWith('http://localhost:8082/api/v1/tokens/current', {
        headers: { 'Content-Type': 'application/json' },
      });
      expect(result).toEqual(mockInfo);
    });

    it('throws error when success is false', async () => {
      const httpClient = mockHttpClient({ data: { success: false, error: { message: 'Invalid token' } } });
      await expect(fetchCurrentTokenInfo(httpClient, 'http://localhost:8082')).rejects.toThrow('Invalid token');
    });

    it('throws error when HTTP call fails', async () => {
      const httpClient = mockHttpClient({
        response: { data: { error: { message: 'Unauthorized' } } },
      }, true);
      await expect(fetchCurrentTokenInfo(httpClient, 'http://localhost:8082')).rejects.toThrow('Unauthorized');
    });
  });

  describe('revokeCurrentToken', () => {
    it('successfully revokes current token', async () => {
      const httpClient = mockHttpClient({ data: { success: true } });
      const payload = { client_id: 'cid', refresh_token: 'ref' };
      await revokeCurrentToken(httpClient, 'http://localhost:8082', payload);

      expect(httpClient.instanceRef.post).toHaveBeenCalledWith('http://localhost:8082/api/v1/tokens/current/revoke', payload, {
        headers: { 'Content-Type': 'application/json' },
      });
    });

    it('throws error when success is false', async () => {
      const httpClient = mockHttpClient({ data: { success: false, error: { message: 'Revocation failed' } } });
      const payload = { client_id: 'cid', refresh_token: 'ref' };
      await expect(revokeCurrentToken(httpClient, 'http://localhost:8082', payload)).rejects.toThrow('Revocation failed');
    });
  });
});
