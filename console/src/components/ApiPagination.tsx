import React from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { Pagination, PaginationContent, PaginationItem, PaginationLink, PaginationEllipsis } from '@/components/ui/pagination';
import { cn } from '@/lib/utils';

type ApiPaginationProps = {
  pagination?: { page_num: number; page_size: number; total_data: number } | null;
  pageSize?: number;
  onPageChange?: (page: number) => void;
  className?: string;
};

export default function ApiPagination({ pagination, pageSize = 10, onPageChange, className }: ApiPaginationProps) {
  if (!pagination) return null;

  const total = pagination.total_data ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / (pageSize || pagination.page_size || 1)));
  const current = Math.min(Math.max(1, pagination.page_num || 1), totalPages);

  const handlePage = (p: number) => { if (onPageChange) onPageChange(p); };

  const pages: number[] = [];
  if (totalPages <= 7) {
    for (let i = 1; i <= totalPages; i++) pages.push(i);
  } else {
    const set = new Set<number>([1, totalPages, current]);
    if (current - 1 > 1) set.add(current - 1);
    if (current + 1 < totalPages) set.add(current + 1);
    Array.from(set).filter(p => p >= 1 && p <= totalPages).sort((a,b)=>a-b).forEach(p => pages.push(p));
  }

  return (
    <div className={cn('flex items-center', className)}>
      {total > 0 && (
        <Pagination>
          <PaginationContent>
            <PaginationItem>
              <PaginationLink href="#" aria-label="Previous page" onClick={(e) => { e.preventDefault(); handlePage(Math.max(1, current - 1)); }}>
                <ChevronLeft className="h-4 w-4" />
              </PaginationLink>
            </PaginationItem>
            {pages.map((p, idx) => (
              <PaginationItem key={p}>
                {idx > 0 && p - pages[idx - 1] > 1 ? <PaginationEllipsis /> : null}
                <PaginationLink href="#" isActive={p === current} onClick={(e) => { e.preventDefault(); handlePage(p); }}>
                  {p}
                </PaginationLink>
              </PaginationItem>
            ))}
            <PaginationItem>
              <PaginationLink href="#" aria-label="Next page" onClick={(e) => { e.preventDefault(); handlePage(Math.min(totalPages, current + 1)); }}>
                <ChevronRight className="h-4 w-4" />
              </PaginationLink>
            </PaginationItem>
          </PaginationContent>
        </Pagination>
      )}
    </div>
  );
}
