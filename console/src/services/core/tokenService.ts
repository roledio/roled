import type { IStorage } from '@/storage/storageService';
import type { ConfigService } from './configService';
import { decodeJwtPayload } from '@/lib/jwt';
import logger from '@/lib/logger';

type TokenResponse = {
  access_token: string;
  refresh_token?: string;
  expires_in?: number;
  [key: string]: unknown;
};

const REFRESH_KEY = 'console_refresh_token_v1';

export class TokenService {
  private storage: IStorage;
  private config: ConfigService;
  private isRefreshing = false;
  private waiters: Array<(token: string | null) => void> = [];
  private accessToken: string | null = null;
  private accessTokenPayload: any | null = null;
  private currentTokenInfo: any | null = null;
  private currentMember: any | undefined = undefined;

  constructor(storage: IStorage, config: ConfigService) {
    this.storage = storage;
    this.config = config;
  }

  getAccessToken(): string | null {
    return this.accessToken;
  }

  getCurrentTokenInfo(): any | null {
    return this.currentTokenInfo;
  }

  setCurrentTokenInfo(info: any | null) {
    this.currentTokenInfo = info;
  }

  getCurrentMember(): any | null {
    return this.currentMember;
  }

  setCurrentMember(member: any | null) {
    this.currentMember = member;
  }

  isAccessTokenValid(): boolean {
    logger.debug('[TokenService] Checking access token validity. Access token present:', this.accessTokenPayload);
    if (this.accessTokenPayload) {
      const exp = this.accessTokenPayload?.exp;
      if (!exp) return false; // if we have a payload but no exp, consider it invalid
      return Date.now() < exp * 1000 - 5000; // consider token expired 5s before
    }
    return false;
  }

  getRefreshToken(): string | null {
    return this.storage.get<string>(REFRESH_KEY);
  }

  setTokens(tokens: TokenResponse) {
    // support wrapped payloads { success: true, data: { access_token, ... } }
    const t: any = (tokens as any)?.data ?? tokens;
    logger.debug('[TokenService] setTokens keys:', Object.keys(t || {}));
    if (t.refresh_token) this.storage.set(REFRESH_KEY, t.refresh_token);
    if (t.access_token) {
      this.accessToken = t.access_token;
      this.accessTokenPayload = decodeJwtPayload(t.access_token);
    }
  }

  clear() {
    this.storage.remove(REFRESH_KEY);
    this.accessToken = null;
    this.accessTokenPayload = null;
    this.currentTokenInfo = null;
    this.currentMember = undefined;
    this.config.clear();
    try {
      window.localStorage.clear();
      window.sessionStorage.clear();
    } catch (e) {
      logger.error('Failed to clear storage:', e);
    }
  }

  async refreshAccessToken(): Promise<string | null> {
    // If already refreshing, return a promise that resolves when refresh completes
    if (this.isRefreshing) {
      return new Promise((resolve) => this.waiters.push(resolve));
    }
    const refreshToken = this.getRefreshToken();
    if (!refreshToken) return null;
    this.isRefreshing = true;
    try {
      const cfg = this.config;
      const url = `${cfg.getAuthBaseUrl()}/api/v1/tokens`;
      const params = new URLSearchParams();
      params.set('grant_type', 'refresh_token');
      params.set('refresh_token', refreshToken);
      const cached = cfg.getCachedConfig();
      if (!cached) throw new Error('missing client config');
      params.set('client_id', cached.client_id);

      const res = await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: params.toString(),
      });
      if (!res.ok) {
        this.clear();
        this.waiters.forEach((w) => w(null));
        this.waiters = [];
        return null;
      }
      const payload = await res.json();
      const data = (payload && typeof payload === 'object' && 'data' in payload) ? payload.data : payload;
      this.setTokens(data as TokenResponse);
      const newAccess = (data && data.access_token) || null;
      this.waiters.forEach((w) => w(newAccess));
      this.waiters = [];
      return newAccess;
    } finally {
      this.isRefreshing = false;
    }
  }
}
