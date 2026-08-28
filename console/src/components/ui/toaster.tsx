import { useToast } from "@/hooks/use-toast";
import { Toast, ToastClose, ToastDescription, ToastProvider, ToastTitle, ToastViewport } from "@/components/ui/toast";

export function Toaster({ position = 'top-right' as any }: { position?: string }) {
  const { toasts } = useToast();

  return (
    <ToastProvider>
      {toasts.map(function ({ id, title, description, action, ...props }) {
        // choose entry animation based on configured position (top vs bottom)
        const entryClass = (position || '').startsWith('top') ? 'data-[state=open]:slide-in-from-top-full' : 'data-[state=open]:slide-in-from-bottom-full';
        return (
          <Toast key={id} {...props} className={entryClass} data-position={position}>
            <div className="grid gap-1">
              {title && <ToastTitle>{title}</ToastTitle>}
              {description && <ToastDescription>{description}</ToastDescription>}
            </div>
            {action}
            <ToastClose />
          </Toast>
        );
      })}
      <ToastViewport position={position as any} />
    </ToastProvider>
  );
}
