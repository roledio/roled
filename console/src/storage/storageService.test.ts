import { mock } from 'vitest-mock-extended';
import { describe, it, expect, beforeEach } from 'vitest';
import { StorageService } from './storageService';

// Example test suite for StorageService using vitest-mock-extended

describe('StorageService', () => {
  let storageMock: ReturnType<typeof mock<Storage>>;
  let service: StorageService;

  beforeEach(() => {
    storageMock = mock<Storage>();
    service = new StorageService(storageMock);
  });

  it('should set and get a value', () => {
    // Simulate setItem and getItem
    storageMock.setItem.mockImplementation((key, value) => {
      storageMock.getItem.mockReturnValueOnce(value);
    });
    service.set('foo', 'bar');
    expect(service.get('foo')).toBe('bar');
    expect(storageMock.setItem).toHaveBeenCalledWith('foo', 'bar');
    expect(storageMock.getItem).toHaveBeenCalledWith('foo');
  });

  it('should remove a value', () => {
    service.remove('foo');
    expect(storageMock.removeItem).toHaveBeenCalledWith('foo');
  });
});
