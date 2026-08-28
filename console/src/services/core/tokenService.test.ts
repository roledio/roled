import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { mock } from 'vitest-mock-extended';
import { TokenService } from './tokenService';
import { IStorage } from '@/storage/storageService';
import { ConfigService } from './configService';

describe('TokenService', () => {
  let storageMock: ReturnType<typeof mock<IStorage>>;
  let configServiceMock: ReturnType<typeof mock<ConfigService>>;
  let svc: TokenService;

  beforeEach(() => {
    storageMock = mock<IStorage>();
    configServiceMock = mock<ConfigService>();
    configServiceMock.getAuthBaseUrl.mockReturnValue('http://auth.local');
    configServiceMock.getCachedConfig.mockReturnValue({ client_id: 'cid' });
    svc = new TokenService(storageMock, configServiceMock);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('setTokens persists access/refresh tokens and expiry', () => {
    // create a minimal fake JWT with future exp so decodeJwtPayload returns a payload
    const createJwt = (payload: object) => {
      const base64 = (v: string) => Buffer.from(v).toString('base64').replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
      const header = base64(JSON.stringify({ alg: 'none' }));
      const body = base64(JSON.stringify(payload));
      return `${header}.${body}.`; // no signature
    };

    const token = createJwt({ exp: Math.floor(Date.now() / 1000) + 30 });
    svc.setTokens({ access_token: token, refresh_token: 'r', expires_in: 60 });

    // TokenService currently persists the refresh token to storage; access token
    // is kept in-memory via getAccessToken(). Assert refresh persisted and access stored in memory.
    expect(storageMock.set).toHaveBeenCalledWith('console_refresh_token_v1', 'r');
    expect(svc.getAccessToken()).toBe(token);
    expect(svc.isAccessTokenValid()).toBe(true);
  });

  it('refreshAccessToken calls backend and stores new tokens', async () => {
    const createJwt = (payload: object) => {
      const base64 = (v: string) => Buffer.from(v).toString('base64').replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
      const header = base64(JSON.stringify({ alg: 'none' }));
      const body = base64(JSON.stringify(payload));
      return `${header}.${body}.`;
    };

    const newToken = createJwt({ exp: Math.floor(Date.now() / 1000) + 60 });
    const mocked = {
      ok: true,
      json: async () => ({ data: { access_token: newToken, refresh_token: 'newR', expires_in: 60 } }),
    } as any;
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(mocked));

    // set an existing refresh token so refreshAccessToken will run
    storageMock.get.mockReturnValue('oldR');
    const newAccess = await svc.refreshAccessToken();
    expect(newAccess).toBe(newToken);
    expect(storageMock.set).toHaveBeenCalledWith('console_refresh_token_v1', 'newR');
    expect(svc.getAccessToken()).toBe(newToken);
  });

  it('stores and clears token info and current member info in memory', () => {
    const mockTokenInfo = { id: 'tok-123', user: { email: 'user@example.com' } };
    const mockMember = { id: 'mem-123', email: 'user@example.com', is_admin: true };

    expect(svc.getCurrentTokenInfo()).toBeNull();
    expect(svc.getCurrentMember()).toBeUndefined();

    svc.setCurrentTokenInfo(mockTokenInfo);
    svc.setCurrentMember(mockMember);

    expect(svc.getCurrentTokenInfo()).toEqual(mockTokenInfo);
    expect(svc.getCurrentMember()).toEqual(mockMember);

    svc.clear();

    expect(svc.getCurrentTokenInfo()).toBeNull();
    expect(svc.getCurrentMember()).toBeUndefined();
  });
});
