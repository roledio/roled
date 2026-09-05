import logger from '@/lib/logger';
import { buildSigninCallbackRedirect } from '@/lib/redirect';
import { generateCodeChallenge, generateCodeVerifier } from '@/lib/pkce';
import { SessionStorageService } from '@/storage/storageService';
import type { ConfigService } from './configService';
import type { TokenService } from './tokenService';
import type { HttpClient } from './httpClient';

const STATE_KEY = 'console_pkce_state_v1';
const VERIFIER_KEY = 'console_pkce_verifier_v1';

export class AuthService {
  private config: ConfigService;
  private tokens: TokenService;
  private origin: string;

  constructor(config: ConfigService, tokens: TokenService, origin = window.location.origin) {
    this.config = config;
    this.tokens = tokens;
    this.origin = origin.replace(/\/$/, '');
  }

  private session = SessionStorageService;

  async buildSignInRedirect(signup = false): Promise<void> {
    const cfg = this.config.getCachedConfig() ?? (await this.config.loadConfig());
    const clientId = cfg.client_id;
    const codeVerifier = generateCodeVerifier();
    const codeChallenge = await generateCodeChallenge(codeVerifier);
    const state = crypto.randomUUID();

    this.session.set(VERIFIER_KEY, codeVerifier);
    this.session.set(STATE_KEY, state);

    const redirectUri = buildSigninCallbackRedirect(this.origin);

    const params = new URLSearchParams({
      client_id: clientId,
      redirect_uri: redirectUri,
      response_type: 'code',
      code_challenge: codeChallenge,
      code_challenge_method: 'S256',
      state,
    });

    if (signup) {
      params.set('is_signup', 'true');
    }

    const authorizeUrl = `${this.config.getAuthBaseUrl().replace(/\/$/, '')}/authorize?${params.toString()}`;
    window.location.href = authorizeUrl;
  }

  getStoredState(): string | null {
    return this.session.get<string>(STATE_KEY);
  }

  getStoredVerifier(): string | null {
    return this.session.get<string>(VERIFIER_KEY);
  }

  clearStoredPkce() {
    this.session.remove(STATE_KEY);
    this.session.remove(VERIFIER_KEY);
  }

  async exchangeCode(code: string): Promise<void> {
    const verifier = this.getStoredVerifier();
    if (!verifier) throw new Error('Missing code_verifier');
    const cfg = this.config.getCachedConfig() ?? (await this.config.loadConfig());
    const params = new URLSearchParams();
    params.set('grant_type', 'authorization_code');
    params.set('client_id', cfg.client_id);
    params.set('authorization_code', code);
    params.set('redirect_uri', buildSigninCallbackRedirect(this.origin));
    params.set('code_verifier', verifier);

    const res = await fetch(`${this.config.getAuthBaseUrl()}/api/v1/tokens`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: params.toString(),
    });
    if (!res.ok) throw new Error('Token exchange failed');
    const payload = await res.json();
    const data = payload && typeof payload === 'object' && 'data' in payload ? payload.data : payload;
    logger.debug('[AuthService] token exchange payload keys:', Object.keys(payload || {}));
    logger.debug('[AuthService] token data keys:', Object.keys(data || {}));
    this.tokens.setTokens(data as any);
    this.clearStoredPkce();
  }
}

export type TokenInfoUser = {
  id: string;
  display_name: string;
  email: string;
  external_user_id?: string | null;
  avatar_url?: string | null;
};

export type TokenInfoProject = {
  id: string;
  name: string;
  description: string;
  logo_url: string;
};

export type TokenInfoClient = {
  id: string;
  name: string;
};

export type TokenInfoRole = {
  id: string;
  code: string;
  name: string;
  description: string;
};

export type CurrentTokenInfo = {
  id: string;
  issued_at: string;
  expires_at: string;
  project: TokenInfoProject;
  client: TokenInfoClient;
  user: TokenInfoUser;
  role: TokenInfoRole;
  permissions: string[];
};

type CurrentTokenInfoResponse = {
  success: boolean;
  data: CurrentTokenInfo;
  error?: { message: string };
};

export async function fetchCurrentTokenInfo(
  httpClient: HttpClient,
  authBaseUrl: string,
): Promise<CurrentTokenInfo> {
  const url = `${authBaseUrl.replace(/\/$/, '')}/api/v1/tokens/current`;
  try {
    const res = await httpClient.instanceRef.get<CurrentTokenInfoResponse>(url, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success || !res.data?.data) {
      const msg = res.data?.error?.message ?? 'Invalid token response';
      throw new Error(msg);
    }
    return res.data.data;
  } catch (err: any) {
    const msg =
      err?.response?.data?.error?.message ??
      err?.response?.data?.message ??
      err?.message ??
      'Failed to fetch token details';
    throw new Error(msg);
  }
}

export type RevokeTokenPayload = {
  client_id: string;
  refresh_token: string;
};

type RevokeResponse = {
  success: boolean;
  error?: { message: string };
};

export async function revokeCurrentToken(
  httpClient: HttpClient,
  authBaseUrl: string,
  payload: RevokeTokenPayload,
): Promise<void> {
  const url = `${authBaseUrl.replace(/\/$/, '')}/api/v1/tokens/current/revoke`;
  try {
    const res = await httpClient.instanceRef.post<RevokeResponse>(url, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    if (!res.data?.success) {
      const msg = res.data?.error?.message ?? 'Failed to revoke token';
      throw new Error(msg);
    }
  } catch (err: any) {
    const msg =
      err?.response?.data?.error?.message ??
      err?.response?.data?.message ??
      err?.message ??
      'Failed to revoke token';
    throw new Error(msg);
  }
}


