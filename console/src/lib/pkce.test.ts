import { describe, it, expect } from 'vitest';
import { generateCodeVerifier, generateCodeChallenge } from './pkce';

describe('PKCE utils', () => {
  it('generates a code verifier of acceptable length', () => {
    const v = generateCodeVerifier();
    expect(typeof v).toBe('string');
    expect(v.length).toBeGreaterThanOrEqual(43);
    expect(v.length).toBeLessThanOrEqual(128);
  });

  it('generates a code challenge from verifier', async () => {
    const v = generateCodeVerifier();
    const c = await generateCodeChallenge(v);
    expect(typeof c).toBe('string');
    // base64url-ish characters
    expect(c).toMatch(/^[A-Za-z0-9_-]+$/);
    expect(c).not.toBe(v);
  });
});
