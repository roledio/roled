import { describe, expect, it } from 'vitest';
import { buildBrandingPreview } from '@/pages/projects/details/branding/branding-preview-document';
import type { BrandingSettings } from '@/services/projects/branding';

const settings: BrandingSettings = {logo_url: null, favicon_url: null, primary_color: '#2563eb', rounding: 'large', enable_shadow: false, enable_border: true};

describe('Branding preview', () => {
  it('uses the auth styles and reflects all appearance settings', () => {
    const html = buildBrandingPreview(settings, 'My Project', 'http://localhost:8082');
    expect(html).toContain('http://localhost:8082/assets/static/roled.css');
    expect(html).toContain('http://localhost:8082/assets/static/branding.css');
    expect(html).toContain('--brand-primary:#2563eb');
    expect(html).toContain('--brand-radius:1rem');
    expect(html).toContain('--brand-shadow:none');
    expect(html).toContain('--brand-border:1px solid var(--roled-primary)');
    const updated = buildBrandingPreview({...settings, rounding:'sharp', enable_shadow:true, enable_border:false}, 'My Project', 'http://localhost:8082');
    expect(updated).toContain('--brand-radius:0');
    expect(updated).toContain('--brand-border:0');
    expect(updated).not.toContain('--brand-shadow:none');
  });
  it('escapes project names and image attributes and rejects CSS injection', () => {
    const html = buildBrandingPreview({...settings, logo_url:'https://example.com/" onerror="alert(1)', primary_color:'red;}body{display:none}'}, '<script>attack()</script>', 'http://localhost:8082');
    const doc = new DOMParser().parseFromString(html, 'text/html');
    expect(doc.querySelector('script')).toBeNull();
    expect(doc.querySelector('img')?.hasAttribute('onerror')).toBe(false);
    expect(doc.querySelector('title')?.textContent).toBe('Sign in to <script>attack()</script>');
    expect(html).toContain('--brand-primary:#ba8d1c');
  });
});
