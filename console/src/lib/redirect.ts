export function buildSigninCallbackRedirect(origin?: string) {
  const o = (origin ?? (typeof window !== 'undefined' ? window.location.origin : '')) as string;
  return o.replace(/\/$/, '') + '/signin/callback';
}
