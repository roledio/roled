import type { HttpClient } from '@/services/core/httpClient';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  createOAuthConnectionWithCredentials,
  deleteOAuthConnection,
  fetchOAuthConnections,
  fetchOAuthConnectionByProvider,
  updateOAuthConnection,
  type OAuthConnection,
} from './oauth-connections';

function createMockHttpClient(): HttpClient {
  return {
    instanceRef: {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
    },
  } as unknown as HttpClient;
}

const BASE_URL = 'http://localhost:8082';
const PROJECT_ID = 'proj_123';

const mockOAuthConnection: OAuthConnection = {
  id: 'oauth_1',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
  project_id: PROJECT_ID,
  provider: 'google',
  credential_type: 'default',
  enabled: true,
};

describe('OAuthConnections Service', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('fetchOAuthConnections', () => {
    it('fetches oauth connections successfully', async () => {
      const httpClient = createMockHttpClient();
      const mockResponse = {
        data: {
          success: true,
          data: [mockOAuthConnection],
          pagination: { page_num: 1, page_size: 10, total_data: 1 },
        },
      };
      vi.mocked(httpClient.instanceRef.get).mockResolvedValueOnce(mockResponse as any);

      const result = await fetchOAuthConnections(httpClient, BASE_URL, PROJECT_ID);

      expect(result.data).toEqual([mockOAuthConnection]);
      expect(result.pagination).toEqual({
        page_num: 1,
        page_size: 10,
        total_data: 1,
      });
      expect(httpClient.instanceRef.get).toHaveBeenCalledWith(
        `${BASE_URL}/api/v1/projects/${PROJECT_ID}/oauth-connections`,
        expect.objectContaining({
          headers: { 'Content-Type': 'application/json' },
        })
      );
    });

    it('returns empty array when no connections', async () => {
      const httpClient = createMockHttpClient();
      const mockResponse = {
        data: {
          success: true,
          data: [],
        },
      };
      vi.mocked(httpClient.instanceRef.get).mockResolvedValueOnce(mockResponse as any);

      const result = await fetchOAuthConnections(httpClient, BASE_URL, PROJECT_ID);

      expect(result.data).toEqual([]);
    });

    it('throws error on unsuccessful response', async () => {
      const httpClient = createMockHttpClient();
      const mockResponse = {
        data: {
          success: false,
          error: { message: 'Failed to fetch' },
        },
      };
      vi.mocked(httpClient.instanceRef.get).mockResolvedValueOnce(mockResponse as any);

      await expect(fetchOAuthConnections(httpClient, BASE_URL, PROJECT_ID)).rejects.toThrow(
        'Failed to fetch'
      );
    });

    it('throws error with fallback message on network error', async () => {
      const httpClient = createMockHttpClient();
      vi.mocked(httpClient.instanceRef.get).mockRejectedValueOnce(
        new Error('Network error')
      );

      await expect(fetchOAuthConnections(httpClient, BASE_URL, PROJECT_ID)).rejects.toThrow(
        'Network error'
      );
    });

    it('handles baseUrl with trailing slash', async () => {
      const httpClient = createMockHttpClient();
      const mockResponse = { data: { success: true, data: [] } };
      vi.mocked(httpClient.instanceRef.get).mockResolvedValueOnce(mockResponse as any);

      await fetchOAuthConnections(httpClient, `${BASE_URL}/`, PROJECT_ID);

      expect(httpClient.instanceRef.get).toHaveBeenCalledWith(
        `${BASE_URL}/api/v1/projects/${PROJECT_ID}/oauth-connections`,
        expect.any(Object)
      );
    });
  });

  describe('createOAuthConnectionWithCredentials', () => {
    it('creates oauth connection successfully', async () => {
      const httpClient = createMockHttpClient();
      const mockResponse = {
        data: {
          success: true,
          data: mockOAuthConnection,
        },
      };
      vi.mocked(httpClient.instanceRef.post).mockResolvedValueOnce(mockResponse as any);

      const result = await createOAuthConnectionWithCredentials(httpClient, BASE_URL, PROJECT_ID, 'google', {
        credential_type: 'default',
      });

      expect(result).toEqual(mockOAuthConnection);
      expect(httpClient.instanceRef.post).toHaveBeenCalledWith(
        `${BASE_URL}/api/v1/projects/${PROJECT_ID}/oauth-connections/google`,
        { credential_type: 'default' },
        expect.objectContaining({
          headers: { 'Content-Type': 'application/json' },
        })
      );
    });

    it('throws error on unsuccessful response', async () => {
      const httpClient = createMockHttpClient();
      const mockResponse = {
        data: {
          success: false,
          error: { message: 'Provider already exists' },
        },
      };
      vi.mocked(httpClient.instanceRef.post).mockResolvedValueOnce(mockResponse as any);

      await expect(
        createOAuthConnectionWithCredentials(httpClient, BASE_URL, PROJECT_ID, 'google', {
          credential_type: 'default',
        })
      ).rejects.toThrow('Provider already exists');
    });

    it('throws error on network failure', async () => {
      const httpClient = createMockHttpClient();
      vi.mocked(httpClient.instanceRef.post).mockRejectedValueOnce(
        new Error('Network error')
      );

      await expect(
        createOAuthConnectionWithCredentials(httpClient, BASE_URL, PROJECT_ID, 'google', {
          credential_type: 'default',
        })
      ).rejects.toThrow('Network error');
    });
  });

  describe('deleteOAuthConnection', () => {
    it('deletes oauth connection successfully', async () => {
      const httpClient = createMockHttpClient();
      const mockResponse = {
        data: {
          success: true,
        },
      };
      vi.mocked(httpClient.instanceRef.delete).mockResolvedValueOnce(mockResponse as any);

      await deleteOAuthConnection(httpClient, BASE_URL, PROJECT_ID, 'oauth_1');

      expect(httpClient.instanceRef.delete).toHaveBeenCalledWith(
        `${BASE_URL}/api/v1/projects/${PROJECT_ID}/oauth-connections/oauth_1`,
        expect.objectContaining({
          headers: { 'Content-Type': 'application/json' },
        })
      );
    });

    it('throws error on unsuccessful response', async () => {
      const httpClient = createMockHttpClient();
      const mockResponse = {
        data: {
          success: false,
          error: { message: 'Connection not found' },
        },
      };
      vi.mocked(httpClient.instanceRef.delete).mockResolvedValueOnce(mockResponse as any);

      await expect(
        deleteOAuthConnection(httpClient, BASE_URL, PROJECT_ID, 'oauth_1')
      ).rejects.toThrow('Connection not found');
    });

    it('throws error on network failure', async () => {
      const httpClient = createMockHttpClient();
      vi.mocked(httpClient.instanceRef.delete).mockRejectedValueOnce(
        new Error('Network error')
      );

      await expect(
        deleteOAuthConnection(httpClient, BASE_URL, PROJECT_ID, 'oauth_1')
      ).rejects.toThrow('Network error');
    });

    it('handles baseUrl with trailing slash', async () => {
      const httpClient = createMockHttpClient();
      const mockResponse = { data: { success: true } };
      vi.mocked(httpClient.instanceRef.delete).mockResolvedValueOnce(mockResponse as any);

      await deleteOAuthConnection(httpClient, `${BASE_URL}/`, PROJECT_ID, 'oauth_1');

      expect(httpClient.instanceRef.delete).toHaveBeenCalledWith(
        `${BASE_URL}/api/v1/projects/${PROJECT_ID}/oauth-connections/oauth_1`,
        expect.any(Object)
      );
    });
  });
});

 describe("OAuth endpoint contracts", () => {
const payload = { credential_type: 'custom' as const, client_id: 'client', client_secret: 'secret', scopes: ['email'], enabled: false };
const connection = { id: 'connection', provider: 'google', ...payload };
const base = 'https://auth.example/';
const root = 'https://auth.example/api/v1/projects/project/oauth-connections';
const headers = { headers: { 'Content-Type': 'application/json' } };
const endpoints = [
 { name: 'list', method: 'get', url: root, call: (http: HttpClient) => fetchOAuthConnections(http, base, 'project'), invalid: 'Invalid oauth connections response', fallback: 'Failed to fetch oauth connections', data: [connection] },
 { name: 'details', method: 'get', url: root + '/google', call: (http: HttpClient) => fetchOAuthConnectionByProvider(http, base, 'project', 'google'), invalid: 'Invalid oauth connection response', fallback: 'Failed to fetch oauth connection', data: connection },
 { name: 'create', method: 'post', url: root + '/google', call: (http: HttpClient) => createOAuthConnectionWithCredentials(http, base, 'project', 'google', payload), invalid: 'Failed to create oauth connection with credentials', fallback: 'Failed to create oauth connection with credentials', data: connection },
 { name: 'update', method: 'put', url: root + '/google', call: (http: HttpClient) => updateOAuthConnection(http, base, 'project', 'google', payload), invalid: 'Failed to update oauth connection', fallback: 'Failed to update oauth connection', data: connection },
 { name: 'delete', method: 'delete', url: root + '/google', call: (http: HttpClient) => deleteOAuthConnection(http, base, 'project', 'google'), invalid: 'Failed to delete oauth connection', fallback: 'Failed to delete oauth connection', data: undefined },
];
for (const endpoint of endpoints) describe(endpoint.name, () => {
 function setup() { const request = vi.fn(); return { request, http: { instanceRef: { [endpoint.method]: request } } as unknown as HttpClient }; }
 it('uses the scoped endpoint and preserves data, pagination and credential payload', async () => {
  const { request, http } = setup(); const pagination = { page_num: 1, page_size: 5, total_data: 1 }; request.mockResolvedValue({ data: { success: true, data: endpoint.data, pagination } });
  expect(await endpoint.call(http)).toEqual(endpoint.name === 'list' ? { data: endpoint.data, pagination } : endpoint.data);
  expect(request).toHaveBeenCalledExactlyOnceWith(...(['post', 'put'].includes(endpoint.method) ? [endpoint.url, payload, headers] : [endpoint.url, headers]));
 });
 it.each([undefined, {}, { success: false }, { success: false, error: { message: 'Denied' } }])('rejects malformed or failed API response %j', async data => {
  const { request, http } = setup(); request.mockResolvedValue({ data }); await expect(endpoint.call(http)).rejects.toThrow(data?.error?.message ?? endpoint.invalid);
 });
 it.each([
  [{ response: { data: { error: { message: 'nested' }, message: 'outer' } }, message: 'transport' }, 'nested'],
  [{ response: { data: { message: 'outer' } }, message: 'transport' }, 'outer'],
  [new Error('transport'), 'transport'],
  [{}, null], [null, null],
 ] as const)('propagates the highest priority error %j', async (failure, message) => {
  const { request, http } = setup(); request.mockRejectedValue(failure); await expect(endpoint.call(http)).rejects.toThrow(message ?? endpoint.fallback);
 });
});
it('normalizes a successful list with missing data to an empty list', async () => {
 const get = vi.fn().mockResolvedValue({ data: { success: true } });
 expect(await fetchOAuthConnections({ instanceRef: { get } } as unknown as HttpClient, base, 'project')).toEqual({ data: [], pagination: undefined });
});

});
