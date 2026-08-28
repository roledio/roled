import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";

interface ConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: React.ReactNode;
  confirmLabel?: React.ReactNode;
  destructive?: boolean;
  disabled?: boolean;
  onConfirm: () => void;
  /**
   * When false, do not use the AlertDialog action which auto-closes the dialog.
   * Defaults to true to preserve existing behavior.
   */
  closeOnConfirm?: boolean;
  children?: React.ReactNode;
  contentClassName?: string;
}

export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmLabel = "Confirm",
  destructive = false,
  disabled = false,
  onConfirm,
  closeOnConfirm = true,
  children,
  contentClassName = '',
}: ConfirmDialogProps) {
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription>{description}</AlertDialogDescription>
        </AlertDialogHeader>
        <div className={`${contentClassName}`}>{children}</div>
        <AlertDialogFooter className="bottom-0 bg-background">
          <AlertDialogCancel className="h-9 rounded-md px-3">Cancel</AlertDialogCancel>
          {closeOnConfirm ? (
            <AlertDialogAction
              onClick={onConfirm}
              disabled={disabled}
              className={destructive ? "bg-destructive text-destructive-foreground hover:bg-destructive/90 h-9 rounded-md px-3" : "h-9 rounded-md px-3"}
            >
              {confirmLabel}
            </AlertDialogAction>
          ) : (
            <Button size="sm" onClick={onConfirm} disabled={disabled} className={destructive ? "bg-destructive text-destructive-foreground hover:bg-destructive/90 h-9 rounded-md px-3" : "h-9 rounded-md px-3"}>
              {confirmLabel}
            </Button>
          )}
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
