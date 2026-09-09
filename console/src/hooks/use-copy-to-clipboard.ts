import { useState } from 'react';
import { useToast } from './use-toast';

export function useCopyToClipboard() {
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const { toast } = useToast();

  const handleCopy = async (text: string, label: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedId(label);
      toast({
        title: 'Copied',
        description: `${label} copied to clipboard`,
      });
      setTimeout(() => setCopiedId(null), 1500);
    } catch (err) {
      toast({
        title: 'Copy failed',
        description: `Failed to copy ${label} to clipboard`,
        variant: 'destructive',
      });
    }
  };

  return { copiedId, handleCopy };
}
