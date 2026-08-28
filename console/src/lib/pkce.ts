import { base64UrlEncode, randomString } from './crypto';

export function generateCodeVerifier(length = 96): string {
  return randomString(length);
}

export async function generateCodeChallenge(codeVerifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(codeVerifier);
  const digest = await crypto.subtle.digest('SHA-256', data);
  return base64UrlEncode(digest);
}
