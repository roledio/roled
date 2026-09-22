import type { BrandingSettings } from '@/services/projects/branding';

function escapeHTML(value: string): string {
  return value.replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c]!);
}

export function buildBrandingPreview(settings: BrandingSettings, name: string, authBaseUrl: string): string {
  const base = escapeHTML(authBaseUrl.replace(/\/$/, ''));
  const color = /^#[0-9a-f]{6}$/i.test(settings.primary_color) ? settings.primary_color : '#ba8d1c';
  const radius = { sharp: '0', small: '0.25rem', medium: '0.5rem', large: '1rem' }[settings.rounding] ?? '0.25rem';
  const shadow = settings.enable_shadow ? '0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)' : 'none';
  return `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Sign in to ${escapeHTML(name)}</title>
    <link rel="stylesheet" href="${base}/assets/static/roled.css">
    <link rel="stylesheet" href="${base}/assets/static/branding.css?v=1">
    <style>:root{--brand-primary:${color};--brand-radius:${radius};--brand-shadow:${shadow};--brand-border:${settings.enable_border ? '1px solid var(--roled-primary)' : '0'}}
    .roled-auth-page{min-height:100vh;padding:2.5rem 1.25rem} .roled-auth-container{width:100%}</style>
    </head><body><main class="roled-auth-page"><div class="roled-auth-container"><div class="roled-auth-card">
    <header class="roled-auth-header">
    ${settings.logo_url ? `<img class="roled-auth-logo" src="${escapeHTML(settings.logo_url)}" alt="${escapeHTML(name)} logo">` : ''}
    <h1 class="roled-auth-title">Sign in</h1><p class="roled-auth-subtitle">Sign in to <strong>${escapeHTML(name)}</strong></p></header>
    <div class="roled-auth-body"><form>
    <div class="roled-form-group"><label class="roled-label" for="preview-email">Email</label><input class="roled-input" id="preview-email" type="email" placeholder="Enter your email" autocomplete="off"></div>
    <div class="roled-form-group"><label class="roled-label" for="preview-password">Password</label><input class="roled-input" id="preview-password" type="password" placeholder="Enter your password" autocomplete="off"></div>
    <div class="roled-form-group" style="text-align:right"><a href="#">Forgot password?</a></div>
    <div class="roled-form-group"><button type="button" class="roled-btn roled-btn-primary roled-btn-block">Sign in</button></div>
    </form></div><footer class="roled-auth-footer"><p class="roled-auth-toggle-text">Don't have an account? <a href="#">Sign up</a></p></footer>
    </div><div class="roled-powered-by">Powered by <strong>Roled</strong></div></div></main></body></html>`;
}

