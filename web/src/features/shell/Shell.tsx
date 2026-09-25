import { matchPath, Navigate, Route, Routes, useLocation, useNavigate, type Location } from "react-router";
import { errorMessage, isUnauthenticated } from "../../api/errors";
import { useMe } from "../../api/queries";
import type { User } from "../../api/types";
import { FoxMark } from "../../components/brand/logo";
import { Button } from "../../components/ui/button";
import { useT } from "../../i18n";
import { ActivityPage } from "../activity/ActivityPage";
import { ContactsPage } from "../contacts/ContactsPage";
import { GroupPage } from "../groups/GroupPage";
import { NotFoundPage } from "../NotFoundPage";
import { useLocaleSync } from "../settings/language";
import { SettingsPage } from "../settings/SettingsPage";
import { TaskPanel } from "../tasks/TaskPanel";
import { TaskViewPage } from "../tasks/TaskViewPage";
import { CommandPalette } from "./CommandPalette";
import { useHotkeys } from "./hotkeys";
import { MobileDrawer } from "./MobileDrawer";
import { VIEWS } from "./nav";
import { NewTaskDialog } from "./NewTaskDialog";
import { ShortcutsDialog } from "./ShortcutsDialog";
import { Sidebar } from "./Sidebar";
import { Splash } from "./session";
import { newTask, setUI } from "./store";

/** /app/*: signed-in only. */
export function AppGate() {
  const t = useT();
  const me = useMe();
  const location = useLocation();
  if (me.isPending) return <Splash />;
  if (me.isError) {
    if (isUnauthenticated(me.error)) {
      return <Navigate to={`/login?next=${encodeURIComponent(location.pathname + location.search)}`} replace />;
    }
    return (
      <div className="flex h-dvh flex-col items-center justify-center gap-4 bg-canvas px-6 text-center">
        <FoxMark className="size-9 opacity-60" />
        <div>
          <p className="text-md font-semibold text-fg">{t("app.cantStart")}</p>
          <p className="mt-1 text-sm text-fg-3">{errorMessage(me.error)}</p>
        </div>
        <Button variant="secondary" onClick={() => void me.refetch()}>
          {t("common.tryAgain")}
        </Button>
      </div>
    );
  }
  return <Shell user={me.data} />;
}

let lastBackground: Location | null = null;

function Shell({ user }: { user: User }) {
  useLocaleSync(user);
  const t = useT();
  const location = useLocation();
  const navigate = useNavigate();
  const taskMatch = matchPath("/app/tasks/:taskId", location.pathname);
  const state = location.state as { background?: Location } | null;
  const background: Location | null = taskMatch ? (state?.background ?? lastBackground) : null;
  if (!taskMatch) lastBackground = location;
  const shown = background ?? (taskMatch ? ({ ...location, pathname: "/app/today", search: "", hash: "" } as Location) : location);

  useHotkeys([
    { key: "k", mod: true, allowInInputs: true, allowInOverlay: true, run: () => setUI({ palette: true, paletteMode: "all" }) },
    { key: "/", run: () => setUI({ palette: true, paletteMode: "tasks" }) },
    { key: "c", run: () => newTask({}) },
    { key: "?", run: () => setUI({ shortcuts: true }) },
    { key: ",", mod: true, run: () => navigate("/app/settings") },
    ...VIEWS.map((v) => ({ key: `g ${v.keys[1]!.toLowerCase()}`, run: () => navigate(v.path) })),
  ]);

  const closePanel = () => {
    if (background) navigate(background, { replace: false });
    else navigate("/app/today");
  };

  return (
    <div className="flex h-dvh overflow-hidden bg-canvas">
      <a
        href="#main"
        className="sr-only z-[100] rounded-md bg-surface px-3 py-2 text-sm shadow-pop focus:not-sr-only focus:fixed focus:top-3 focus:left-3"
      >
        {t("app.skipToContent")}
      </a>
      <aside className="hidden w-[248px] shrink-0 min-[900px]:flex" aria-label={t("app.sidebar")}>
        <Sidebar />
      </aside>
      <MobileDrawer />
      <main id="main" className="flex min-w-0 flex-1 flex-col min-[900px]:py-2 min-[900px]:pr-2">
        <div className="relative flex min-h-0 flex-1 flex-col overflow-hidden bg-sheet min-[900px]:rounded-xl min-[900px]:shadow-sheet">
          {/* Descendant routes of /app/*: their paths are relative to /app. */}
          <Routes location={shown}>
            <Route index element={<Navigate to="/app/today" replace />} />
            {VIEWS.map((v) => (
              <Route key={v.view} path={v.path.slice("/app/".length)} element={<TaskViewPage key={v.view} view={v.view} />} />
            ))}
            <Route path="groups/:groupId" element={<GroupPage />} />
            <Route path="contacts" element={<ContactsPage />} />
            <Route path="activity" element={<ActivityPage />} />
            <Route path="settings" element={<SettingsPage />} />
            <Route path="*" element={<NotFoundPage inApp />} />
          </Routes>
        </div>
      </main>
      {taskMatch?.params.taskId ? (
        <TaskPanel
          id={taskMatch.params.taskId}
          onClose={closePanel}
          onNavigate={(id) => navigate(`/app/tasks/${id}`, { replace: true, state: { background: shown } })}
        />
      ) : null}
      <CommandPalette />
      <NewTaskDialog />
      <ShortcutsDialog />
    </div>
  );
}
