import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { HttpClient } from '@/services/core/httpClient';
import { deleteMember, fetchMembers, inviteMember, updateMember } from './index';

const get = vi.fn(), post = vi.fn(), put = vi.fn(), patch = vi.fn(), remove = vi.fn();
const http = { instanceRef: { get, post, put, patch, delete: remove } } as unknown as HttpClient;
const base = 'https://auth.example/';
const headers = { 'Content-Type': 'application/json' };
const member = { id: 'member', display_name: 'Alice', email: 'alice@example.com', is_admin: false, is_active: true, is_verified: true, created_at: '', updated_at: '' };
beforeEach(() => vi.resetAllMocks());

const operations = [
  { name: 'list members', method: get, run: () => fetchMembers(http, base), data: [member], expected: { data: [member], pagination: undefined }, args: ['https://auth.example/api/v1/members', { params: {}, headers }], invalid: 'Invalid members response', fallback: 'Failed to fetch members' },
  { name: 'invite member', method: post, run: () => inviteMember(http, base, member.email), data: member, expected: member, args: ['https://auth.example/api/v1/members', { email: member.email }, { headers }], invalid: 'Failed to invite member', fallback: 'Failed to invite member' },
  { name: 'delete member', method: remove, run: () => deleteMember(http, base, 'a/b ?'), data: undefined, expected: undefined, args: ['https://auth.example/api/v1/members/a%2Fb%20%3F', { headers }], invalid: 'Failed to delete member', fallback: 'Failed to delete member' },
  { name: 'update member', method: patch, run: () => updateMember(http, base, 'a/b ?', { is_admin: false, account_id: 'account' }), data: member, expected: member, args: ['https://auth.example/api/v1/members/a%2Fb%20%3F', { is_admin: false, account_id: 'account' }, { headers }], invalid: 'Failed to update member', fallback: 'Failed to update member' },
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

describe('member payload edge cases', () => {
  it('preserves false filters and pagination returned by the API', async () => {
    const params = { page_num: 2, page_size: 5, is_admin: false, is_active: false, is_verified: false, search: ' Alice ' };
    const pagination = { page_num: 2, page_size: 5, total_data: 8 };
    get.mockResolvedValue({ data: { success: true, data: [member], pagination } });
    expect(await fetchMembers(http, base, params)).toEqual({ data: [member], pagination });
    expect(get).toHaveBeenCalledWith('https://auth.example/api/v1/members', { params, headers });
  });
  it.each([undefined, null])('defaults missing member lists to empty: %s', async data => {
    get.mockResolvedValue({ data: { success: true, data } });
    expect(await fetchMembers(http, base)).toEqual({ data: [], pagination: undefined });
  });
  it.each([undefined, '', 'https://app.example/activate'])('includes only a nonempty invitation redirect: %s', async redirect => {
    post.mockResolvedValue({ data: { success: true, data: member } });
    await inviteMember(http, base, member.email, redirect);
    const payload = redirect ? { email: member.email, redirect_uri: redirect } : { email: member.email };
    expect(post).toHaveBeenCalledWith('https://auth.example/api/v1/members', payload, { headers });
  });
});
