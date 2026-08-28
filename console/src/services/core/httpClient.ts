import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import type { TokenService } from './tokenService';
import type { ConfigService } from './configService';

export class HttpClient {
  private instance: AxiosInstance;
  private tokenService: TokenService;
  private configService: ConfigService;

  constructor(tokenService: TokenService, configService: ConfigService) {
    this.tokenService = tokenService;
    this.configService = configService;
    this.instance = axios.create({ baseURL: undefined });
    this.instance.interceptors.request.use(this.onRequest.bind(this));
    this.instance.interceptors.response.use(this.onResponse.bind(this), this.onResponseError.bind(this));
  }

  private async onRequest(cfg: AxiosRequestConfig) {
    let token = this.tokenService.getAccessToken();
    if (!this.tokenService.isAccessTokenValid()) {
      token = await this.tokenService.refreshAccessToken();
    }
    if (token) cfg.headers = { ...(cfg.headers ?? {}), Authorization: `Bearer ${token}` };
    return cfg;
  }

  private onResponse(res: AxiosResponse) {
    return res;
  }

  private async onResponseError(error: any) {
    const original = error.config as AxiosRequestConfig & { _retry?: boolean };
    if (!original) return Promise.reject(error);
    // attempt refresh only once per request
    if (error.response && error.response.status === 401 && !original._retry) {
      original._retry = true;
      const newToken = await this.tokenService.refreshAccessToken();
      if (!newToken) {
        // refresh failed
        return Promise.reject(error);
      }
      // set header and retry
      original.headers = { ...(original.headers ?? {}), Authorization: `Bearer ${newToken}` };
      return this.instance.request(original);
    }
    return Promise.reject(error);
  }

  get instanceRef() {
    return this.instance;
  }

  get tokenServiceRef() {
    return this.tokenService;
  }
}
