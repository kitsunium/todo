import { Menu as MenuIcon, Search } from "lucide-react";
import { useRef, useState, type ReactNode } from "react";
import { IconButton } from "../../components/ui/button";
import { MOD } from "../../components/ui/kbd";
import { Tooltip } from "../../components/ui/tooltip";
import { useT } from "../../i18n";
import { cn } from "../../lib/cn";
import { setUI } from "./store";

/**
 * A page of the app: a slim bar (the title appears in it once the big one
 * scrolls away), then the scrolling column, 760px at most.
 */
export function Page({
  title,
  icon,
  subtitle,
  actions,
  children,
  wide,
}: {
  title: ReactNode;
  icon?: ReactNode;
  subtitle?: ReactNode;
  actions?: ReactNode;
  children: ReactNode;
  wide?: boolean;
}) {
  const t = useT();
  const [scrolled, setScrolled] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  return (
    <div className="flex h-full min-h-0 flex-col">
      <div
        className={cn(
          "relative z-20 flex h-12 shrink-0 items-center gap-2 border-b px-3 transition-colors duration-200 sm:px-4",
          scrolled ? "border-line-soft" : "border-transparent",
        )}
      >
        <IconButton label={t("app.openMenu")} size="sm" className="min-[900px]:hidden" onClick={() => setUI({ drawer: true })}>
          <MenuIcon className="size-[18px]" />
        </IconButton>
        <div
          className={cn(
            "flex min-w-0 items-center gap-2 text-sm font-semibold text-fg transition-[opacity,transform] duration-200",
            scrolled ? "translate-y-0 opacity-100" : "translate-y-1 opacity-0",
          )}
          aria-hidden={!scrolled}
        >
          {icon ? <span className="flex size-4 shrink-0 items-center justify-center text-fg-3 [&>svg]:size-4">{icon}</span> : null}
          <span className="truncate">{title}</span>
        </div>
        <div className="ml-auto flex items-center gap-1">
          <Tooltip content={t("common.search")} keys={[MOD, "K"]}>
            <IconButton label={t("common.search")} size="sm" className="min-[900px]:hidden" onClick={() => setUI({ palette: true, paletteMode: "all" })}>
              <Search className="size-4" />
            </IconButton>
          </Tooltip>
        </div>
      </div>
      <div ref={ref} data-scroller className="min-h-0 flex-1 overflow-y-auto overscroll-contain" onScroll={(e) => setScrolled(e.currentTarget.scrollTop > 56)}>
        <div className={cn("mx-auto w-full px-4 pb-28 sm:px-8", wide ? "max-w-[880px]" : "max-w-[760px]")}>
          <header className="pt-3 pb-6 sm:pt-6">
            <div className="flex items-center gap-3">
              {icon ? <span className="flex shrink-0 items-center justify-center [&>svg]:size-[22px]">{icon}</span> : null}
              <h1 className="min-w-0 truncate text-2xl font-semibold tracking-[-0.025em] text-fg">{title}</h1>
              {actions ? <div className="ml-auto flex shrink-0 items-center gap-1.5">{actions}</div> : null}
            </div>
            {subtitle ? <div className={cn("mt-1 text-sm text-fg-3", icon && "sm:pl-[34px]")}>{subtitle}</div> : null}
          </header>
          {children}
        </div>
      </div>
    </div>
  );
}
