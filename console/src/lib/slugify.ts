// Reusable slugify utility for generating URL/code-friendly identifiers.
// Keeps behavior deterministic and safe for production: lowercased, ASCII-only,
// collapses separators to single dash, trims leading/trailing dashes.
export function slugify(input: string): string {
  if (!input) return '';
  // Normalize and remove diacritics
  let s = input.normalize('NFKD').replace(/\p{M}/gu, '');
  s = s.toLowerCase();
  // Replace any non-alphanumeric characters with dashes
  s = s.replace(/[^a-z0-9]+/g, '-');
  // Collapse multiple dashes
  s = s.replace(/-+/g, '-');
  // Trim leading/trailing dashes
  s = s.replace(/^-|-$/g, '');
  return s;
}

export default slugify;
