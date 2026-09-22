import { useMemo } from 'react';
import type { BrandingSettings } from '@/services/projects/branding';

import { buildBrandingPreview } from '@/pages/projects/details/branding/branding-preview-document';

export default function BrandingPreview({ settings, projectName }: { settings: BrandingSettings; projectName: string }) {
  const source = useMemo(() => buildBrandingPreview(settings, projectName, import.meta.env.VITE_AUTH_BASE_URL), [settings, projectName]);
  return <div className="overflow-hidden rounded-lg border bg-background">
    <div className="flex items-center gap-2 border-b bg-muted/40 px-4 py-3 text-xs text-muted-foreground">
      <span className="mr-2 flex gap-1" aria-hidden="true"><span className="h-2 w-2 rounded-full bg-muted-foreground/30" /><span className="h-2 w-2 rounded-full bg-muted-foreground/30" /><span className="h-2 w-2 rounded-full bg-muted-foreground/30" /></span>
      {settings.favicon_url && <img src={settings.favicon_url} alt="Favicon preview" className="h-4 w-4 object-contain" />}
      Sign in to {projectName}
    </div>
    <iframe title="Login page preview" sandbox="" srcDoc={source} className="h-[690px] w-full border-0" />
  </div>;
}
