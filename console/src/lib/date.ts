import { parseISO, isValid } from 'date-fns';

export function formatDate(iso?: string | null) {
  if (!iso) return '';
  const d = typeof iso === 'string' ? parseISO(iso) : new Date(iso as any);
  if (!isValid(d)) return '';
  try {
    // Use user's locale and timezone with medium date and short time styles
    return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(d);
  } catch (err) {
    // Fallback to toLocaleString if Intl options aren't supported in the environment
    return d.toLocaleString();
  }
}

export default formatDate;
