import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate, useParams } from 'react-router-dom';
import { ArrowLeft, Loader2, Pipette } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useToast } from '@/hooks/use-toast';
import { brandingQueryKey, useProjectBranding } from '@/hooks/projects/use-project-branding';
import { getProjectTabParams } from '@/lib/paramsStore';
import type { HttpClient } from '@/services/core/httpClient';
import { fetchProjectById, type Project } from '@/services/projects';
import { brandingError, updateProjectBranding, uploadBrandingAsset, type BrandingSettings, type ProjectBranding } from '@/services/projects/branding';
import BrandingPreview from '@/pages/projects/details/branding/branding-preview';

type EyeDropperConstructor = new () => { open: () => Promise<{ sRGBHex: string }> };

function BrandingEditor({ httpClient, project, branding }: { httpClient: HttpClient; project: Project; branding: ProjectBranding }) {
  const [settings, setSettings] = useState<BrandingSettings>(() => ({
    logo_url: branding.is_default ? project.logo_url ?? null : branding.logo_url,
    favicon_url: branding.is_default ? null : branding.favicon_url,
    primary_color: branding.primary_color,
    rounding: branding.rounding,
    enable_shadow: branding.is_default ? true : branding.enable_shadow,
    enable_border: branding.is_default ? false : branding.enable_border,
  }));
  const [uploading, setUploading] = useState<'logo_url' | 'favicon_url' | null>(null);
  const [picking, setPicking] = useState(false);
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { toast } = useToast();
  const baseUrl = import.meta.env.VITE_AUTH_BASE_URL;
  const EyeDropper = (window as Window & { EyeDropper?: EyeDropperConstructor }).EyeDropper;
  const back = () => {
    const params = new URLSearchParams(getProjectTabParams(project.id, 'settings'));
    params.set('tab', 'settings');
    navigate(`/projects/${project.id}/details?${params}`);
  };
  const save = useMutation({
    mutationFn: () => updateProjectBranding(httpClient, baseUrl, project.id, settings),
    onSuccess: data => {
      queryClient.setQueryData(brandingQueryKey(project.id), data);
      setSettings(data);
      toast({ title: 'Branding saved', description: 'Your authentication pages now use this branding.' });
    },
    onError: error => toast({ title: 'Save failed', description: brandingError(error), variant: 'destructive' }),
  });
  const busy = Boolean(uploading) || save.isPending;
  const set = <K extends keyof BrandingSettings>(key: K, value: BrandingSettings[K]) => setSettings(previous => ({ ...previous, [key]: value }));
  const upload = async (field: 'logo_url' | 'favicon_url', file?: File) => {
    if (!file) return;
    if (file.size > 2 * 1024 * 1024) {
      toast({ title: 'Image too large', description: 'Choose an image smaller than 2 MB.', variant: 'destructive' });
      return;
    }
    setUploading(field);
    try { set(field, await uploadBrandingAsset(httpClient, baseUrl, file, field === 'logo_url' ? 'project-logo' : 'favicon')); }
    catch (error) { toast({ title: 'Upload failed', description: brandingError(error), variant: 'destructive' }); }
    finally { setUploading(null); }
  };
  const pickColor = async () => {
    if (!EyeDropper) return;
    setPicking(true);
    try { set('primary_color', (await new EyeDropper().open()).sRGBHex); }
    catch (error) {
      if (!(error instanceof DOMException && error.name === 'AbortError')) toast({ title: 'Color picker unavailable', description: brandingError(error), variant: 'destructive' });
    } finally { setPicking(false); }
  };
  return <div className="space-y-6">
    <div className="flex flex-wrap items-center justify-between gap-4">
      <div className="flex items-center gap-3"><Button variant="ghost" size="icon" aria-label="Back to project settings" onClick={back}><ArrowLeft className="h-5 w-5" /></Button>
        <div><h1 className="text-2xl font-semibold">Branding</h1><p className="text-sm text-muted-foreground">{project.name} · Customize your authentication pages</p></div></div>
      <Button onClick={() => save.mutate()} disabled={busy || picking || !/^#[0-9a-f]{6}$/i.test(settings.primary_color)}>
        {save.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}Save branding</Button>
    </div>
    <div className="grid items-start gap-6 lg:grid-cols-[minmax(300px,380px)_minmax(0,1fr)]">
      <Card><CardHeader><CardTitle className="text-lg">Appearance</CardTitle><CardDescription>Changes appear in the preview. Save to apply them to your project.</CardDescription></CardHeader>
        <CardContent><fieldset disabled={busy} className="space-y-6">
          {(['logo_url', 'favicon_url'] as const).map(field => <div key={field} className="space-y-2">
            <Label htmlFor={`branding-${field}`}>{field === 'logo_url' ? 'Logo' : 'Favicon'}</Label>
            <div className="flex items-center gap-3">
              <div className="flex h-14 w-14 shrink-0 items-center justify-center rounded-md border bg-muted/30">
                {uploading === field ? <Loader2 className="h-5 w-5 animate-spin" aria-label="Uploading image" /> : settings[field] ? <img src={settings[field]} alt={field === 'logo_url' ? 'Logo preview' : 'Favicon preview'} className="max-h-10 max-w-10 object-contain" /> : <span className="text-xs text-muted-foreground">None</span>}
              </div>
              <Input id={`branding-${field}`} type="file" accept={field === 'favicon_url' ? "image/png,image/jpeg,image/gif,image/x-icon,image/vnd.microsoft.icon,.ico" : "image/png,image/jpeg,image/gif"} aria-label={`Upload ${field === 'logo_url' ? 'logo' : 'favicon'}`} onChange={event => { void upload(field, event.target.files?.[0]); event.target.value = ''; }} className="text-xs" />
            </div>
            <p className="text-xs text-muted-foreground">{field === 'favicon_url' ? 'ICO, PNG, JPG or GIF.' : 'PNG, JPG or GIF.'} Maximum 2 MB.</p>
            <Button variant="link" size="sm" className="h-auto p-0" onClick={() => set(field, field === 'logo_url' ? project.logo_url ?? null : null)}>{field === 'logo_url' ? 'Use project logo' : 'Remove favicon'}</Button>
          </div>)}
          <div className="space-y-2"><Label htmlFor="brand-color">Brand / primary color</Label>
            <div className="flex gap-2"><Input type="color" aria-label="Choose primary color" value={/^#[0-9a-f]{6}$/i.test(settings.primary_color) ? settings.primary_color : '#ba8d1c'} onChange={event => set('primary_color', event.target.value)} className="w-12 shrink-0 p-1" />
              <Input id="brand-color" value={settings.primary_color} maxLength={7} onChange={event => set('primary_color', event.target.value)} aria-invalid={!/^#[0-9a-f]{6}$/i.test(settings.primary_color)} className="font-mono" />
              <Button variant="outline" size="icon" aria-label="Pick color from screen" title={EyeDropper ? 'Pick color from screen' : 'Screen color picker requires a supported browser in a secure context'} disabled={!EyeDropper || picking} onClick={pickColor}><Pipette className="h-4 w-4" /></Button>
            </div><p className="text-xs text-muted-foreground">Used for buttons, links, and borders.</p>
            {!EyeDropper && <p className="text-xs text-muted-foreground">Screen eyedropper is unavailable in this browser. Use the color picker or enter a hex color.</p>}
          </div>
          <div className="space-y-2"><Label htmlFor="brand-rounding">Rounding corner</Label>
            <Select value={settings.rounding} onValueChange={(value: BrandingSettings['rounding']) => set('rounding', value)} disabled={busy}><SelectTrigger id="brand-rounding"><SelectValue /></SelectTrigger><SelectContent>
              <SelectItem value="sharp">Flat / sharp</SelectItem><SelectItem value="small">Small</SelectItem><SelectItem value="medium">Medium</SelectItem><SelectItem value="large">Large</SelectItem>
            </SelectContent></Select>
          </div>
          <div className="flex items-center justify-between"><Label htmlFor="brand-shadow">Enable shadow</Label><Switch id="brand-shadow" checked={settings.enable_shadow} onCheckedChange={value => set('enable_shadow', value)} disabled={busy} /></div>
          <div className="flex items-center justify-between"><Label htmlFor="brand-border">Enable border</Label><Switch id="brand-border" checked={settings.enable_border} onCheckedChange={value => set('enable_border', value)} disabled={busy} /></div>
        </fieldset></CardContent>
      </Card>
      <div className="space-y-3"><div className="flex items-center justify-between"><h2 className="font-medium">Login page preview</h2><span className="text-xs text-muted-foreground">Updates as you edit</span></div>
        <BrandingPreview settings={settings} projectName={project.name} />
        <p className="text-xs text-muted-foreground">Preview shows sample fields. Available sign-in options follow your project settings.</p>
      </div>
    </div>
  </div>;
}

export default function Branding({ httpClient }: { httpClient: HttpClient }) {
  const { project_id = '' } = useParams<{ project_id: string }>();
  const branding = useProjectBranding(httpClient, project_id);
  const project = useQuery({ queryKey: ['project', project_id], queryFn: () => fetchProjectById(httpClient, import.meta.env.VITE_AUTH_BASE_URL, project_id), enabled: Boolean(project_id) });
  if (branding.isError || project.isError) return <div role="alert" className="space-y-3"><p>{brandingError(branding.error ?? project.error)}</p><Button variant="outline" onClick={() => { void branding.refetch(); void project.refetch(); }}>Try again</Button></div>;
  if (!branding.data || !project.data) return <div role="status" className="flex items-center gap-2"><Loader2 className="h-5 w-5 animate-spin" />Loading branding…</div>;
  return <BrandingEditor key={project_id} httpClient={httpClient} branding={branding.data} project={project.data} />;
}
