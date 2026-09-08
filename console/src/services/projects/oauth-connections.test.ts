import type { HttpClient } from '@/services/core/httpClient';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  createOAuthConnectionWithCredentials,
  deleteOAuthConnection,
  fetchOAuthConnections,
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
