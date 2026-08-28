import { cn } from "@/lib/utils";

interface StatusBadgeProps {
  active: boolean;
  className?: string;
}

export function StatusBadge({ active, className }: StatusBadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center px-2 py-0.5 text-xs font-medium rounded",
        active
          ? "bg-success/10 text-success"
          : "bg-muted text-muted-foreground",
        className
      )}
    >
      {active ? "Active" : "Inactive"}
    </span>
  );
}
