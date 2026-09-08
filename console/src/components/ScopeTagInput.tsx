import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { X } from 'lucide-react';
import { useRef, useState } from 'react';

interface Props {
  scopes: string[];
  onChange: (scopes: string[]) => void;
  disabled?: boolean;
  hasError?: boolean;
  id?: string;
  'aria-describedby'?: string;
}

/**
 * Tag-style input for OAuth scopes.
 * Typing a comma (or pressing Enter) converts the current text into a removable tag.
 */
export function ScopeTagInput({
  scopes,
  onChange,
  disabled = false,
  hasError = false,
  id,
  'aria-describedby': ariaDescribedBy,
}: Props) {
  const [inputValue, setInputValue] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);

  const addScope = (raw: string) => {
    const trimmed = raw.trim().replace(/,+$/, '').trim();
    if (!trimmed) return;
    // Avoid duplicates
    if (!scopes.includes(trimmed)) {
      onChange([...scopes, trimmed]);
    }
    setInputValue('');
  };

  const removeScope = (index: number) => {
    onChange(scopes.filter((_, i) => i !== index));
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      addScope(inputValue);
    } else if (e.key === 'Backspace' && !inputValue && scopes.length > 0) {
      // Remove last tag when backspacing on empty input
      onChange(scopes.slice(0, -1));
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    // Commit the current text as a tag when the user types a comma
    if (val.includes(',')) {
      const [before] = val.split(',');
      addScope(before);
    } else {
      setInputValue(val);
    }
  };

  return (
    <div
      className={[
        'flex flex-wrap items-center gap-1.5 min-h-9 rounded-md border bg-background px-3 py-2 text-sm',
        'focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-0',
        disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-text',
        hasError ? 'border-destructive' : 'border-input',
      ].join(' ')}
      onClick={() => !disabled && inputRef.current?.focus()}
      aria-label="Scopes"
    >
      {scopes.map((scope, idx) => (
        <Badge
          key={idx}
          variant="secondary"
          className="flex items-center gap-1 px-2 py-0.5 text-xs font-normal"
        >
          {scope}
          {!disabled && (
            <button
              type="button"
              aria-label={`Remove scope ${scope}`}
              onClick={(e) => {
                e.stopPropagation();
                removeScope(idx);
              }}
              className="ml-0.5 rounded-full hover:bg-muted-foreground/20 transition-colors p-0.5"
            >
              <X className="h-3 w-3" />
            </button>
          )}
        </Badge>
      ))}

      <input
        ref={inputRef}
        id={id}
        type="text"
        value={inputValue}
        onChange={handleChange}
        onKeyDown={handleKeyDown}
        disabled={disabled}
        placeholder={scopes.length === 0 ? 'e.g. email, profile, openid' : ''}
        aria-describedby={ariaDescribedBy}
        className="flex-1 min-w-[120px] bg-transparent outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed text-sm"
        data-testid="scope-tag-input"
      />
    </div>
  );
}
