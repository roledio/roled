import type { IStorage } from '@/storage/storageService';

export type ConsoleConfig = {
  client_id: string;
  [key: string]: unknown;
};

const CONFIG_KEY = 'console_config_v1';

export class ConfigService {
  private storage: IStorage;
  private baseUrl: string;
  private cache: ConsoleConfig | null = null;

  constructor(storage: IStorage, baseUrl: string) {
    this.storage = storage;
    this.baseUrl = baseUrl.replace(/\/$/, '');
    const cached = this.storage.get<ConsoleConfig>(CONFIG_KEY);
    if (cached) this.cache = cached;
  }

  public getAuthBaseUrl(): string {
    return this.baseUrl;
  }

  public async loadConfig(): Promise<ConsoleConfig> {
    if (this.cache) return this.cache;
    const res = await fetch(`${this.baseUrl}/system/console/config`, { method: 'GET' });
    if (!res.ok) throw new Error('Failed to fetch console config');
    const payload = await res.json();
    // payload may be either { client_id } or { success, data: { client_id } }
    let json: ConsoleConfig | null = null;
    if (payload && typeof payload === 'object') {
      if ('client_id' in payload && payload.client_id) {
        json = payload as ConsoleConfig;
      } else if ('data' in payload && payload.data && typeof payload.data === 'object' && 'client_id' in payload.data) {
        json = payload.data as ConsoleConfig;
      }
    }
    if (!json || !json.client_id) {
      throw new Error('console config missing client_id');
    }
    this.cache = json;
    this.storage.set(CONFIG_KEY, json);
    return json;
  }

  public getCachedConfig(): ConsoleConfig | null {
    return this.cache;
  }

  public clear(): void {
    this.cache = null;
    this.storage.remove(CONFIG_KEY);
  }
}
