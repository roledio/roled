export function base64UrlEncode(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let binary = '';
  for (let i = 0; i < bytes.byteLength; i++) binary += String.fromCharCode(bytes[i]);
  const base64 = btoa(binary);
  return base64.replace(/=/g, '').replace(/\+/g, '-').replace(/\//g, '_');
}

export function randomString(length = 64): string {
  const array = new Uint8Array(length);
  crypto.getRandomValues(array);
  // map to URL safe chars
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_';
  let out = '';
  for (let i = 0; i < array.length; i++) out += chars[array[i] % chars.length];
  return out;
}
