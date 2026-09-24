import { Dialog, DialogBody, DialogContent, DialogHeader } from "../../components/ui/dialog";
import { Kbd, MOD } from "../../components/ui/kbd";
import { setUI, useUI } from "./store";

const GROUPS: { title: string; rows: [string, string[]][] }[] = [
  {
    title: "Anywhere",
    rows: [
      ["Search and commands", [MOD, "K"]],
      ["Search tasks", ["/"]],
      ["New task", ["C"]],
      ["Keyboard shortcuts", ["?"]],
      ["Switch theme", ["⇧", "T"]],
    ],
  },
  {
    title: "Go to",
    rows: [
      ["Inbox", ["G", "I"]],
      ["Today", ["G", "T"]],
      ["Upcoming", ["G", "U"]],
      ["Shared with me", ["G", "S"]],
      ["Assigned to me", ["G", "A"]],
      ["Completed", ["G", "C"]],
    ],
  },
  {
    title: "In a list",
    rows: [
      ["Next / previous task", ["J", "K"]],
      ["Complete", ["X"]],
      ["Open", ["E"]],
      ["Priority", ["1", "2", "3", "4", "0"]],
      ["Let go", ["Esc"]],
    ],
  },
];

export function ShortcutsDialog() {
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
        <DialogHeader title="Keyboard shortcuts" description="Everything is a key away." />
        <DialogBody className="grid gap-6 pb-6 sm:grid-cols-3">
          {GROUPS.map((g) => (
            <section key={g.title}>
              <h3 className="mb-2 text-2xs font-semibold tracking-[0.06em] text-fg-4 uppercase">{g.title}</h3>
              <ul className="flex flex-col gap-2">
                {g.rows.map(([label, keys]) => (
                  <li key={label} className="flex items-center justify-between gap-3 text-sm text-fg-2">
                    <span>{label}</span>
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
