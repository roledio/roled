import { cn } from "@/lib/utils";

type StatusStyle = "success" | "info" | "danger" | "warning" | "muted";

interface StatusBadgeProps {
  active?: boolean;
  text?: string;
  label?: string;
  style?: StatusStyle;
  className?: string;
}

const styleClasses: Record<StatusStyle, string> = {
  success: "bg-success/10 text-success",
  info: "bg-info/10 text-info",
  danger: "bg-danger/10 text-danger",
  warning: "bg-warning/10 text-warning",
  muted: "bg-muted text-muted-foreground",
};

export function StatusBadge({ 
  active, 
  text, 
  label, 
  style, 
  className 
}: StatusBadgeProps) {
  // Backward compatibility: if active is provided, use old behavior
  if (active !== undefined) {
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

  // New behavior: use text/label and style
  const displayText = text || label || "Active";
  const badgeStyle = style || "muted";

  return (
    <span
      className={cn(
        "inline-flex items-center px-2 py-0.5 text-xs font-medium rounded",
        styleClasses[badgeStyle],
        className
      )}
    >
      {displayText}
    </span>
  );
}
