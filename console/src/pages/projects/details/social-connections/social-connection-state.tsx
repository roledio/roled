import type { ReactNode } from 'react';
import { AlertCircle, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface Props {
  loading: boolean;
  failed: boolean;
  error: string;
  errorTitle: string;
  supported: boolean;
  provider: string;
  testId: string;
  onBack: () => void;
  children: ReactNode;
}

export default function SocialConnectionState({ loading, failed, error, errorTitle, supported, provider, testId, onBack, children }: Props) {
  if (loading) return (
    <div className="flex items-center justify-center py-20" data-testid={`${testId}-loading`}>
      <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      <span className="ml-2 text-sm text-muted-foreground">Loading…</span>
    </div>
  );
  if (failed || !supported) return (
    <div className="flex flex-col items-center justify-center py-20 space-y-2" data-testid={`${testId}-${failed ? 'error' : 'unknown-provider'}`}>
      <AlertCircle className="h-8 w-8 text-destructive" />
      <p className="text-sm text-destructive font-medium">{failed ? errorTitle : 'Unknown provider'}</p>
      <p className="text-xs text-muted-foreground">{failed ? error : <>The provider &quot;{provider}&quot; is not supported.</>}</p>
      {!failed && <Button size="sm" variant="secondary" onClick={onBack}>Go back</Button>}
    </div>
  );
  return children;
}
