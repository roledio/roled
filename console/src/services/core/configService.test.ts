import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { mock } from 'vitest-mock-extended';
import { ConfigService } from './configService';
import { IStorage } from '../storage/storageService';

describe('ConfigService', () => {
  let storageMock: ReturnType<typeof mock<IStorage>>;
  beforeEach(() => {
    storageMock = mock<IStorage>();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('loadConfig accepts wrapped payloads and caches result', async () => {
    const response = { data: { client_id: 'abc123' } };
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true, json: async () => response }));

    // Mock storage set/get
    storageMock.set.mockImplementation((k, v) => {
      storageMock.get.mockReturnValueOnce(v);
    });

    const cfg = new ConfigService(storageMock, 'http://localhost:8082');
    const loaded = await cfg.loadConfig();
    expect(loaded.client_id).toBe('abc123');
    expect(cfg.getCachedConfig()?.client_id).toBe('abc123');
    expect(storageMock.set).toHaveBeenCalledWith('console_config_v1', response.data);
    expect(storageMock.get).toHaveBeenCalledWith('console_config_v1');
  });
});
