import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '@/services/core/httpClient';
import { deleteAccount, fetchCurrentAccount, updateAccount } from './index';

const get = vi.fn(), post = vi.fn(), put = vi.fn(), patch = vi.fn(), remove = vi.fn();
const http = { instanceRef: { get, post, put, patch, delete: remove } } as unknown as HttpClient;
const base = 'https://auth.example/';
const headers = { 'Content-Type': 'application/json' };
const account = { id: 'account', name: 'Acme', description: '', is_active: true, created_at: '', updated_at: '' };
beforeEach(() => vi.resetAllMocks());

const operations = [
  { name: 'current account', method: get, run: () => fetchCurrentAccount(http, base), data: account, expected: account, args: ['https://auth.example/api/v1/accounts/current'], invalid: 'Invalid account response', fallback: 'Failed to fetch account' },
  { name: 'update account', method: put, run: () => updateAccount(http, base, 'account', { name: 'Acme', description: '' }), data: account, expected: account, args: ['https://auth.example/api/v1/accounts/account', { name: 'Acme', description: '' }, { headers }], invalid: 'Invalid account response', fallback: 'Failed to update account' },
  { name: 'delete account', method: post, run: () => deleteAccount(http, base, 'account', { password: 'confirmation' }), data: undefined, expected: undefined, args: ['https://auth.example/api/v1/accounts/account/delete', { password: 'confirmation' }, { headers }], invalid: 'Failed to delete account', fallback: 'Failed to delete account' },
];
describe.each(operations)('$name', operation => {
  it('sends the expected request and returns the successful response', async () => {
    operation.method.mockResolvedValue({ data: { success: true, data: operation.data } });
    expect(await operation.run()).toEqual(operation.expected);
    expect(operation.method).toHaveBeenCalledExactlyOnceWith(...operation.args);
  });
  it('preserves the server rejection message', async () => {
    operation.method.mockResolvedValue({ data: { success: false, error: { message: 'Access denied' } } });
    await expect(operation.run()).rejects.toThrow('Access denied');
  });
  it.each([{}, { data: {} }, { data: { success: false } }])('rejects malformed or failed success envelopes: %j', async response => {
    operation.method.mockResolvedValue(response);
    await expect(operation.run()).rejects.toThrow(operation.invalid);
  });
  it.each([
    { error: { response: { data: { error: { message: 'Nested message' }, message: 'Outer message' } }, message: 'Transport' }, expected: 'Nested message' },
    { error: { response: { data: { message: 'Outer message' } }, message: 'Transport' }, expected: 'Outer message' },
    { error: new Error('Transport'), expected: 'Transport' },
  ])('uses error precedence: $expected', async ({ error, expected }) => {
    operation.method.mockRejectedValue(error);
    await expect(operation.run()).rejects.toThrow(expected);
  });
  it('provides a fallback for an unstructured rejection', async () => {
    operation.method.mockRejectedValue(null);
    await expect(operation.run()).rejects.toThrow(operation.fallback);
  });
});

describe('account payload edge cases', () => {
  it.each(['get', 'update'])('rejects absent account data on successful %s response', async op => {
    get.mockResolvedValue({ data: { success: true } }); put.mockResolvedValue({ data: { success: true } });
    await expect(op === 'get' ? fetchCurrentAccount(http, base) : updateAccount(http, base, 'account', {})).rejects.toThrow('Invalid account response');
  });
});
