import { Dialog, DialogBody, DialogContent, DialogHeader } from "../../components/ui/dialog";
import { Kbd, MOD } from "../../components/ui/kbd";
import { useT, type Key } from "../../i18n";
import { VIEWS } from "./nav";
import { setUI, useUI } from "./store";

const GROUPS: { title: Key; rows: [Key, string[]][] }[] = [
  {
    title: "shortcuts.anywhere",
    rows: [
      ["common.searchAndCommands", [MOD, "K"]],
      ["shortcuts.searchTasks", ["/"]],
      ["sidebar.newTask", ["C"]],
      ["user.shortcuts", ["?"]],
    ],
  },
  {
    title: "shortcuts.goTo",
    rows: VIEWS.map((v) => [v.label, v.keys]),
  },
  {
    title: "shortcuts.inList",
    rows: [
      ["shortcuts.nextPrevious", ["J", "K"]],
      ["shortcuts.complete", ["X"]],
      ["shortcuts.open", ["E"]],
      ["shortcuts.priority", ["1", "2", "3", "4", "0"]],
      ["shortcuts.letGo", ["Esc"]],
    ],
  },
];

export function ShortcutsDialog() {
  const t = useT();
  const open = useUI((s) => s.shortcuts);
  return (
    <Dialog open={open} onOpenChange={(o) => setUI({ shortcuts: o })}>
      <DialogContent
        size="lg"
        tabIndex={-1}
        onOpenAutoFocus={(e) => {
          e.preventDefault();
          (e.currentTarget as HTMLElement | null)?.focus({ preventScroll: true });
        }}
      >
        <DialogHeader title={t("user.shortcuts")} description={t("shortcuts.description")} />
        <DialogBody className="grid gap-6 pb-6 sm:grid-cols-3">
          {GROUPS.map((g) => (
            <section key={g.title}>
              <h3 className="mb-2 text-2xs font-semibold tracking-[0.06em] text-fg-4 uppercase">{t(g.title)}</h3>
              <ul className="flex flex-col gap-2">
                {g.rows.map(([label, keys]) => (
                  <li key={label} className="flex items-center justify-between gap-3 text-sm text-fg-2">
                    <span>{t(label)}</span>
                    <span className="flex shrink-0 gap-1">
                      {keys.map((k) => (
                        <Kbd key={k}>{k}</Kbd>
                      ))}
                    </span>
                  </li>
                ))}
              </ul>
            </section>
          ))}
        </DialogBody>
      </DialogContent>
    </Dialog>
  );
}
