import { useQueryClient } from "@tanstack/react-query";
import { useEffect, useRef } from "react";
import { useLocation, useNavigate } from "react-router";
import { onUnauthorized } from "../../api/client";
import { FoxMark } from "../../components/brand/logo";
import { toast } from "../../components/ui/toast";
import { tr, useT } from "../../i18n";

/** A calm splash while the session is checked. */
export function Splash() {
  const t = useT();
  return (
    <div className="flex h-dvh items-center justify-center bg-canvas" role="status" aria-label={t("common.loading")}>
      <FoxMark className="size-9 animate-pulse" />
    </div>
  );
}

/** Any 401 while signed in: back to /login, then back here after. */
export function useSessionGuard() {
  const navigate = useNavigate();
  const location = useLocation();
  const qc = useQueryClient();
  const here = useRef(location);
  here.current = location;
  useEffect(
    () =>
      onUnauthorized(() => {
        const l = here.current;
        if (!l.pathname.startsWith("/app")) return;
        qc.clear();
        toast(tr()("app.sessionEnded"), { id: "session" });
        navigate(`/login?next=${encodeURIComponent(l.pathname + l.search)}`, { replace: true });
      }),
    [navigate, qc],
  );
}

