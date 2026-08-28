import React from 'react';
import { Loader2 } from 'lucide-react';
import { Checkbox } from '@/components/ui/checkbox';

interface CheckboxListItem {
  id: string;
  itemLabel: string;
  itemDescription?: string;
  [key: string]: any;
}

interface CheckboxListProps {
  items: CheckboxListItem[];
  selectedIds: string[];
  onToggle: (id: string) => void;
  isFetching?: boolean;
}

export function CheckboxList({ items, selectedIds, onToggle, isFetching }: CheckboxListProps) {
  return (
    <div>
      <div className="border rounded max-h-[30vh] overflow-auto mt-1.5">
        {isFetching ? (
          <div className="flex items-center gap-2 p-3 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span>Loading…</span>
          </div>
        ) : (
          items.map((p: any) => {
            const isSelected = selectedIds.includes(p.id);
            return (
              <div
                key={p.id}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onToggle(p.id); } }}
                onClick={(e) => { if ((e.target as HTMLElement).closest('[data-perm-checkbox]')) return; onToggle(p.id); }}
                className={`flex items-start gap-3 p-2 transition-colors cursor-pointer ${isSelected ? 'bg-primary/10' : 'bg-muted/5 hover:bg-muted/100'}`}
              >
                <div className="pt-1">
                  <input
                    data-perm-checkbox
                    type="checkbox"
                    className={process.env.NODE_ENV === 'test' ? 'h-4 w-4' : 'sr-only'}
                    checked={isSelected}
                    onChange={() => onToggle(p.id)}
                  />
                  <Checkbox
                    checked={isSelected}
                    onCheckedChange={() => onToggle(p.id)}
                  />
                </div>
                    <div className="flex-1">
                      <div className="text-sm font-medium">{p.itemLabel}</div>
                      <div className="text-xs text-muted-foreground mt-1">{p.itemDescription ?? ''}</div>
                    </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}

export default CheckboxList;
