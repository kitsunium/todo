import { Toaster as Sonner, toast } from "sonner";
import { useTheme } from "../../lib/theme";

export { toast };

export function Toaster() {
  const { resolved } = useTheme();
  return (
    <Sonner
      theme={resolved}
      position="bottom-center"
      offset={20}
      gap={8}
      visibleToasts={3}
      duration={4500}
      toastOptions={{
        unstyled: true,
        classNames: {
          toast:
            "group flex w-[356px] max-w-[calc(100vw-32px)] items-center gap-2.5 rounded-xl bg-[#1c1917] px-3.5 py-2.5 text-sm text-[#fafaf9] shadow-pop dark:bg-raised dark:text-fg",
          title: "min-w-0 flex-1 font-medium leading-5",
          description: "text-xs text-white/60 dark:text-fg-3",
          icon: "flex size-4 shrink-0 items-center justify-center",
          actionButton:
            "ml-auto shrink-0 rounded-md px-2 py-1 text-sm font-semibold text-[#ffb68a] hover:bg-white/10 dark:text-accent-ink dark:hover:bg-hover",
          cancelButton: "shrink-0 rounded-md px-2 py-1 text-sm text-white/60 hover:bg-white/10",
          error: "",
          success: "",
        },
      }}
    />
  );
}
