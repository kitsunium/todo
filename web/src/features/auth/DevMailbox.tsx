import { useQuery } from "@tanstack/react-query";
import { ArrowUpRight } from "lucide-react";
import { lazy, Suspense } from "react";
import { studioAvailable } from "../../api/endpoints";
import { useT } from "../../i18n";

// In mock mode the mailbox is the mock's own; the import disappears from a
// production build with the branch.
const MockMailbox = import.meta.env.MODE === "mock" ? lazy(() => import("../../mock/MockMailbox")) : null;

/**
 * Development only: a way to the mail that was just sent. With a running kit
 * app in dev, the Studio's mailbox (shown only when /_kit/api/graph answers);
 * in mock mode, the mock's.
 */
export function DevMailbox({ email }: { email?: string }) {
  const t = useT();
  const studio = useQuery({ queryKey: ["studio"], queryFn: studioAvailable, staleTime: Infinity, retry: false, enabled: !MockMailbox });
  if (MockMailbox) {
    return (
      <Suspense fallback={null}>
        <MockMailbox {...(email ? { email } : {})} />
      </Suspense>
    );
  }
  if (!studio.data) return null;
  return (
    <a
      href="/_kit/#/mail"
      target="_blank"
      rel="noopener"
      className="mt-6 flex items-center justify-between rounded-lg border border-dashed border-line-strong px-3.5 py-3 text-sm text-fg-2 transition-colors hover:border-accent hover:text-fg focus-visible:outline-2 focus-visible:outline-ring"
    >
      <span>
        <span className="mr-2 rounded bg-accent-soft px-1.5 py-0.5 text-2xs font-semibold tracking-wide text-accent-ink uppercase">Dev</span>
        {t("dev.mailbox")}
      </span>
      <ArrowUpRight className="size-4" aria-hidden="true" />
    </a>
  );
}
