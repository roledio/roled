export interface IStorage {
  get<T = unknown>(key: string): T | null;
  set<T = unknown>(key: string, value: T): void;
  remove(key: string): void;
}

export class StorageService implements IStorage {
  private storage: Storage;

  constructor(storage: Storage) {
    this.storage = storage;
  }

  get<T = unknown>(key: string): T | null {
    const raw = this.storage.getItem(key);
    if (!raw) return null;
    try {
      return JSON.parse(raw) as T;
    } catch (e) {
      return (raw as unknown) as T;
    }
  }

  set<T = unknown>(key: string, value: T): void {
    const raw = typeof value === 'string' ? (value as unknown as string) : JSON.stringify(value);
    this.storage.setItem(key, raw);
  }

  remove(key: string): void {
    this.storage.removeItem(key);
  }
}

export const LocalStorageService = new StorageService(window.localStorage);
export const SessionStorageService = new StorageService(window.sessionStorage);
