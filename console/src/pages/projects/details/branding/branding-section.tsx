import { useLocation, useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useProjectBranding } from '@/hooks/projects/use-project-branding';
import { saveProjectTabParams } from '@/lib/paramsStore';
import type { HttpClient } from '@/services/core/httpClient';
import { brandingError } from '@/services/projects/branding';

export default function BrandingSection({ httpClient, projectId }: { httpClient: HttpClient; projectId: string }) {
  const branding = useProjectBranding(httpClient, projectId);
  const navigate = useNavigate();
  const location = useLocation();
  return <Card>
    <CardHeader className="flex-row items-start justify-between gap-4 space-y-0">
      <div className="space-y-1.5"><CardTitle className="text-lg">Branding</CardTitle>
        <CardDescription>Customize the look of your project's sign-in and account pages.</CardDescription></div>
      <Button variant="outline" disabled={!branding.data} onClick={() => {
        saveProjectTabParams(projectId, 'settings', location.search);
        navigate(`/projects/${projectId}/branding`);
      }}>Configure</Button>
    </CardHeader>
    <CardContent>
      {branding.isPending ? <p className="text-sm text-muted-foreground" role="status">Loading branding…</p> : branding.isError ?
        <div role="alert" className="text-sm text-destructive">{brandingError(branding.error)} <Button variant="link" onClick={() => branding.refetch()}>Try again</Button></div> :
        <div className="flex flex-wrap items-center gap-4 rounded-md border p-4">
          {branding.data.logo_url && <img src={branding.data.logo_url} alt="Current branding logo" className="h-12 w-12 object-contain" />}
          <div className="flex-1 space-y-1"><Badge variant="secondary">{branding.data.is_default ? 'Default branding' : 'Custom branding'}</Badge>
            <p className="text-sm text-muted-foreground">{branding.data.is_default ? 'Inherited from Roled. Configure to create your own branding.' : 'Your saved branding is applied to authentication pages.'}</p>
            <p className="text-xs text-muted-foreground">{branding.data.rounding === 'sharp' ? 'Flat / sharp' : branding.data.rounding} corners · Shadow {branding.data.enable_shadow ? 'on' : 'off'} · Border {branding.data.enable_border ? 'on' : 'off'}</p>
          </div>
          <span className="flex items-center gap-2 font-mono text-sm"><span className="h-6 w-6 rounded border" style={{ backgroundColor: branding.data.primary_color }} aria-hidden="true" />{branding.data.primary_color}</span>
          {branding.data.favicon_url && <img src={branding.data.favicon_url} alt="Current favicon" className="h-6 w-6 object-contain" />}
        </div>}
    </CardContent>
  </Card>;
}
