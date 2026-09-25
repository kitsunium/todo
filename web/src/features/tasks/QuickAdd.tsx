import { endOfDay } from "date-fns";
import { CalendarDays, CornerDownLeft, Hash, Plus, Signal, X } from "lucide-react";
import { useEffect, useMemo, useRef, useState, type KeyboardEvent, type ReactNode } from "react";
import { useNavigate } from "react-router";
import { useContacts, useCreateTask, useCurrentUser, useGroups, useShareTask, useUpdateTask, qk } from "../../api/queries";
import type { Group, NewTask, Priority, TaskView } from "../../api/types";
import { Avatar } from "../../components/ui/avatar";
import { Button } from "../../components/ui/button";
import { Chip } from "../../components/ui/chip";
import { GroupDot } from "../../components/ui/group-color";
import { Kbd } from "../../components/ui/kbd";
import { Menu, MenuContent, MenuTrigger } from "../../components/ui/menu";
import { Popover, PopoverContent, PopoverTrigger } from "../../components/ui/popover";
import { PRIORITY_LABEL, PriorityIcon } from "../../components/ui/priority";
import { toast } from "../../components/ui/toast";
import { Tooltip } from "../../components/ui/tooltip";
import { tr, useT, type T } from "../../i18n";
import { cn } from "../../lib/cn";
import { formatDue } from "../../lib/dates";
import { useNow } from "../../lib/now";
import { mentionAt, mentionText, normalize, parseQuickAdd, suggest, type Candidate, type TokenKind } from "../../lib/quickadd";
import { registerQuickAdd } from "../shell/store";
import { DuePicker } from "./DuePicker";
import { GroupItems, PriorityItems } from "./menus";

type Override = { due?: string; priority?: Priority; groupId?: string };

const TOKEN_TONE: Record<TokenKind, string> = {
  due: "bg-accent-soft",
  priority: "bg-[color-mix(in_srgb,var(--high)_16%,transparent)]",
  group: "bg-[color-mix(in_srgb,var(--done)_16%,transparent)]",
  contact: "bg-[color-mix(in_srgb,#3e63dd_16%,transparent)]",
};

/** Where a new task lands, said the way the sidebar says it. */
function destination(task: NewTask, groups: Group[], t: T): string {
  if (task.groupId) return groups.find((g) => g.id === task.groupId)?.name ?? t("quickadd.itsGroup");
  if (task.due && Date.parse(task.due) > endOfDay(new Date()).getTime()) return t("view.upcoming");
  if (task.due) return t("view.today");
  return t("view.inbox");
}

function belongs(view: TaskView | undefined, task: NewTask, meId: string, groupId?: string): boolean {
  const eod = endOfDay(new Date()).getTime();
  switch (view) {
    case "inbox":
      return !task.groupId;
    case "today":
      return !!task.due && Date.parse(task.due) <= eod;
    case "upcoming":
      return !!task.due && Date.parse(task.due) > eod;
    case "assigned":
      return task.assigneeId === meId;
    case "group":
      return task.groupId === groupId;
    default:
      return false;
  }
}

export function QuickAdd({
  view,
  groupId,
  defaults = {},
  variant = "inline",
  autoFocus = false,
  initialText = "",
  onDone,
}: {
  initialText?: string;
  view?: TaskView;
  groupId?: string;
  defaults?: Partial<NewTask>;
  variant?: "inline" | "dialog";
  autoFocus?: boolean;
  onDone?: () => void;
}) {
  const me = useCurrentUser();
  const t = useT();
  const now = useNow();
  const navigate = useNavigate();
  const { data: groups = [] } = useGroups();
  const { data: contactsData } = useContacts();
  const create = useCreateTask();
  const share = useShareTask();
  const update = useUpdateTask();
  const [text, setText] = useState(initialText);
  const [focused, setFocused] = useState(autoFocus);
  const [ignore, setIgnore] = useState<Set<TokenKind>>(new Set());
  const [override, setOverride] = useState<Override>({});
  const [caret, setCaret] = useState(0);
  const [pick, setPick] = useState(0);
  const [shake, setShake] = useState(false);
  const input = useRef<HTMLInputElement>(null);
  const mirror = useRef<HTMLDivElement>(null);
  const root = useRef<HTMLDivElement>(null);

  const groupCands: Candidate[] = useMemo(() => groups.map((g) => ({ id: g.id, name: g.name })), [groups]);
  const contactCands: Candidate[] = useMemo(() => {
    const list = (contactsData?.contacts ?? []).map((c) => ({ id: c.user.id, name: c.user.name, email: c.user.email }));
    return [{ id: me.id, name: me.name, email: me.email }, ...list];
  }, [contactsData, me]);

  const parsed = useMemo(
    () => parseQuickAdd(text, now, { groups: groupCands, contacts: contactCands, ignore }),
    [text, now, groupCands, contactCands, ignore],
  );

  useEffect(() => {
    if (variant !== "inline") return;
    return registerQuickAdd(() => {
      input.current?.focus();
      input.current?.scrollIntoView({ block: "nearest" });
    });
  }, [variant]);

  const mention = focused ? mentionAt(text, caret) : null;
  const suggestions = mention
    ? suggest(mention.query, mention.kind === "group" ? groupCands : contactCands.filter((c) => c.id !== me.id || mention.query), mention.kind === "contact")
    : [];
  // Hide once the mention names exactly the one suggestion: nothing left to choose.
  const exact =
    !!mention &&
    suggestions.length === 1 &&
    normalize(mention.query) !== "" &&
    [suggestions[0]!.name, suggestions[0]!.name.split(" ")[0] ?? "", suggestions[0]!.email?.split("@")[0] ?? ""].some(
      (k) => normalize(k) === normalize(mention.query),
    );
  const showSuggestions = !!mention && suggestions.length > 0 && !exact;

  function syncScroll() {
    if (mirror.current && input.current) mirror.current.scrollLeft = input.current.scrollLeft;
    setCaret(input.current?.selectionStart ?? text.length);
  }

  function complete(c: Candidate) {
    if (!mention) return;
    const insert = mentionText(mention.kind, c);
    const next = text.slice(0, mention.start) + insert + text.slice(caret).replace(/^\S*\s?/, "");
    setText(next);
    const pos = mention.start + insert.length;
    requestAnimationFrame(() => {
      input.current?.setSelectionRange(pos, pos);
      setCaret(pos);
    });
    setPick(0);
  }

  const effective = {
    priority: parsed.priority ?? override.priority ?? defaults.priority,
    due: parsed.due ? parsed.due.date.toISOString() : (override.due ?? defaults.due),
    groupId: parsed.group?.id ?? override.groupId ?? defaults.groupId,
    contact: parsed.contact,
  };

  function reset() {
    setText("");
    setIgnore(new Set());
    setOverride({});
    setPick(0);
  }

  async function submit() {
    const title = parsed.title.trim();
    if (!title) {
      setShake(true);
      setTimeout(() => setShake(false), 400);
      return;
    }
    const task: NewTask = { title };
    if (effective.priority) task.priority = effective.priority;
    if (effective.due) task.due = effective.due;
    if (effective.groupId) task.groupId = effective.groupId;
    let shareWith: string | null = null;
    const c = effective.contact;
    if (c) {
      const group = groups.find((g) => g.id === task.groupId);
      if (c.id === me.id || group?.members.some((m) => m.user.id === c.id)) task.assigneeId = c.id;
      else shareWith = c.id;
    } else if (defaults.assigneeId) task.assigneeId = defaults.assigneeId;

    const here = belongs(view, task, me.id, groupId);
    reset();
    if (variant === "dialog") onDone?.();
    try {
      const created = await create.mutateAsync({ task, listKey: here ? qk.tasks(view, groupId) : ["tasks", "__none__"] });
      if (shareWith && c) {
        await share.mutateAsync({ id: created.id, userId: shareWith });
        await update.mutateAsync({ id: created.id, patch: { assigneeId: shareWith } });
      }
      if (!here) {
        const say = tr();
        toast(say("quickadd.addedTo", { place: destination(task, groups, say) }), {
          action: { label: say("common.open"), onClick: () => navigate(`/app/tasks/${created.id}`) },
        });
      }
    } catch {
      /* the mutation cache already said what went wrong */
    }
  }

  function onKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (showSuggestions) {
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setPick((p) => (p + 1) % suggestions.length);
        return;
      }
      if (e.key === "ArrowUp") {
        e.preventDefault();
        setPick((p) => (p - 1 + suggestions.length) % suggestions.length);
        return;
      }
      if (e.key === "Enter" || e.key === "Tab") {
        e.preventDefault();
        complete(suggestions[Math.min(pick, suggestions.length - 1)]!);
        return;
      }
    }
    if (e.key === "Enter") {
      e.preventDefault();
      void submit();
    } else if (e.key === "Escape") {
      e.preventDefault();
      e.stopPropagation();
      if (text) reset();
      else if (variant === "dialog") onDone?.();
      else input.current?.blur();
    }
  }

  // The highlighted copy of the text, behind the transparent-background input.
  const segments: ReactNode[] = [];
  let at = 0;
  for (const tok of parsed.tokens) {
    if (tok.start > at) segments.push(text.slice(at, tok.start));
    segments.push(
      <mark key={tok.start} className={cn("rounded-[4px] text-transparent", tok.resolved ? TOKEN_TONE[tok.kind] : "bg-warn-soft")} style={{ boxShadow: "0 0 0 2px transparent" }}>
        {text.slice(tok.start, tok.end)}
      </mark>,
    );
    at = tok.end;
  }
  segments.push(text.slice(at));

  const expanded = variant === "dialog" || focused || text.length > 0;
  const dismiss = (k: TokenKind) => {
    setIgnore((s) => new Set(s).add(k));
    setOverride((o) => {
      const n = { ...o };
      if (k === "due") delete n.due;
      if (k === "priority") delete n.priority;
      if (k === "group") delete n.groupId;
      return n;
    });
  };

  const chips: ReactNode[] = [];
  if (effective.due) {
    const d = formatDue(effective.due, now, t.locale);
    chips.push(
      <TokenChip key="due" onRemove={() => dismiss("due")} label={t("quickadd.removeDate")} tone={d.overdue ? "danger" : "accent"}>
        <CalendarDays className="size-3" /> {d.label}
      </TokenChip>,
    );
  }
  if (effective.priority) {
    chips.push(
      <TokenChip key="priority" onRemove={() => dismiss("priority")} label={t("quickadd.removePriority")}>
        <PriorityIcon priority={effective.priority} className="size-3" /> {t(PRIORITY_LABEL[effective.priority])}
      </TokenChip>,
    );
  }
  const g = groups.find((x) => x.id === effective.groupId);
  if (g && !(view === "group" && g.id === groupId)) {
    chips.push(
      <TokenChip key="group" onRemove={() => dismiss("group")} label={t("quickadd.removeGroup")}>
        <GroupDot color={g.color} className="size-[7px]" /> {g.name}
      </TokenChip>,
    );
  } else if (parsed.groupQuery) {
    chips.push(
      <Chip key="group?" tone="warn">
        <Hash className="size-3" /> {t("quickadd.noGroup", { name: parsed.groupQuery })}
      </Chip>,
    );
  }
  if (effective.contact) {
    const u = effective.contact;
    chips.push(
      <TokenChip key="contact" onRemove={() => dismiss("contact")} label={t("quickadd.removeAssignee")}>
        <Avatar user={{ id: u.id, name: u.name, email: u.email ?? "" }} size="xs" className="-ml-0.5" />
        {u.id === me.id ? t("quickadd.assignMe") : t("quickadd.assignTo", { name: u.name.split(" ")[0] ?? u.name })}
      </TokenChip>,
    );
  } else if (parsed.contactQuery) {
    chips.push(
      <Chip key="contact?" tone="warn">
        {t("quickadd.noContact", { name: parsed.contactQuery })}
      </Chip>,
    );
  }

  return (
    <div
      ref={root}
      className={cn(
        "relative",
        variant === "inline" && "mb-5",
        shake && "animate-[shake_0.35s_ease-in-out]",
      )}
      onBlur={(e) => {
        if (!root.current?.contains(e.relatedTarget as Node | null) && !(e.relatedTarget as HTMLElement | null)?.closest("[data-radix-popper-content-wrapper]")) {
          setFocused(false);
        }
      }}
    >
      <div
        className={cn(
          "rounded-xl transition-[background-color,box-shadow] duration-150",
          expanded ? "bg-surface shadow-[0_0_0_1px_var(--line),0_4px_16px_-6px_rgb(0_0_0/0.08)]" : "hover:bg-hover",
          variant === "dialog" && "shadow-none",
        )}
      >
        <div className="flex items-center gap-3 px-2.5 py-2">
          <span
            className={cn(
              "flex size-[18px] shrink-0 items-center justify-center rounded-full transition-colors",
              expanded ? "bg-accent text-white" : "border-[1.5px] border-dashed border-fg-4 text-fg-4",
            )}
            aria-hidden="true"
          >
            <Plus className="size-3" strokeWidth={expanded ? 3 : 2.5} />
          </span>
          <div className="relative min-w-0 flex-1">
            <div
              ref={mirror}
              aria-hidden="true"
              className="pointer-events-none absolute inset-0 overflow-hidden text-[14px] leading-6 whitespace-pre text-transparent"
            >
              {segments}
            </div>
            <input
              ref={input}
              value={text}
              autoFocus={autoFocus}
              onChange={(e) => {
                setText(e.target.value);
                setCaret(e.target.selectionStart ?? e.target.value.length);
                setPick(0);
                if (!e.target.value) setIgnore(new Set());
              }}
              onFocus={() => setFocused(true)}
              onKeyDown={onKeyDown}
              onKeyUp={syncScroll}
              onClick={syncScroll}
              onSelect={syncScroll}
              onScroll={syncScroll}
              placeholder={expanded ? t("quickadd.example") : t("quickadd.placeholder")}
              aria-label={t("quickadd.label")}
              aria-autocomplete="list"
              aria-expanded={showSuggestions}
              enterKeyHint="done"
              maxLength={200}
              className="relative block h-6 w-full bg-transparent text-[14px] leading-6 text-fg caret-accent outline-none placeholder:text-fg-4"
            />
          </div>
          {!expanded ? <Kbd className="max-[899px]:hidden">C</Kbd> : null}
        </div>

        {expanded ? (
          <div className="flex flex-wrap items-center gap-1.5 px-2.5 pb-2.5 pl-[42px] animate-fade-in">
            {chips}
            <Toolbar
              override={override}
              setOverride={setOverride}
              hasDue={!!effective.due}
              hasPriority={!!effective.priority}
              hasGroup={!!g}
              inGroup={view === "group"}
              clearIgnore={(k) => setIgnore((s) => {
                const n = new Set(s);
                n.delete(k);
                return n;
              })}
            />
            <div className="ml-auto flex items-center gap-1.5">
              {variant === "inline" ? (
                <Button
                  size="sm"
                  variant="ghost"
                  onMouseDown={(e) => e.preventDefault()}
                  onClick={() => {
                    reset();
                    input.current?.blur();
                    setFocused(false);
                  }}
                >
                  {t("common.cancel")}
                </Button>
              ) : null}
              <Button size="sm" variant="primary" disabled={!parsed.title.trim()} onMouseDown={(e) => e.preventDefault()} onClick={() => void submit()}>
                {t("quickadd.add")}
                <CornerDownLeft className="size-3.5 opacity-60" />
              </Button>
            </div>
          </div>
        ) : null}
      </div>

      {showSuggestions ? (
        <div
          role="listbox"
          aria-label={mention.kind === "group" ? t("quickadd.groups") : t("quickadd.people")}
          className="absolute top-[calc(100%+6px)] left-[34px] z-30 w-[260px] overflow-hidden rounded-lg bg-raised p-1 shadow-pop animate-pop-in"
        >
          {suggestions.map((c, i) => (
            <button
              key={c.id}
              type="button"
              role="option"
              aria-selected={i === pick}
              onMouseDown={(e) => e.preventDefault()}
              onClick={() => complete(c)}
              onMouseEnter={() => setPick(i)}
              className={cn("flex h-8 w-full items-center gap-2 rounded-md px-2 text-left text-sm", i === pick && "bg-hover")}
            >
              {mention.kind === "group" ? (
                <GroupDot color={groups.find((x) => x.id === c.id)?.color ?? "slate"} />
              ) : (
                <Avatar user={{ id: c.id, name: c.name, email: c.email ?? "" }} size="xs" />
              )}
              <span className="min-w-0 flex-1 truncate">{c.id === me.id ? t("common.nameYou", { name: c.name }) : c.name}</span>
              {i === pick ? <Kbd>↵</Kbd> : null}
            </button>
          ))}
        </div>
      ) : null}
    </div>
  );
}

function TokenChip({ children, onRemove, label, tone = "neutral" }: { children: ReactNode; onRemove: () => void; label: string; tone?: "neutral" | "accent" | "danger" }) {
  return (
    <Chip tone={tone} className="animate-pop-in gap-1 pr-0.5">
      {children}
      <button
        type="button"
        aria-label={label}
        onMouseDown={(e) => e.preventDefault()}
        onClick={onRemove}
        className="ml-0.5 flex size-4 items-center justify-center rounded-[4px] opacity-60 hover:bg-active hover:opacity-100"
      >
        <X className="size-3" />
      </button>
    </Chip>
  );
}

function Toolbar({
  override,
  setOverride,
  hasDue,
  hasPriority,
  hasGroup,
  inGroup,
  clearIgnore,
}: {
  override: Override;
  setOverride: (fn: (o: Override) => Override) => void;
  hasDue: boolean;
  hasPriority: boolean;
  hasGroup: boolean;
  inGroup: boolean;
  clearIgnore: (k: TokenKind) => void;
}) {
  const t = useT();
  const [dueOpen, setDueOpen] = useState(false);
  const btn =
    "inline-flex h-6 items-center gap-1 rounded-md px-1.5 text-xs text-fg-3 shadow-[inset_0_0_0_1px_var(--line)] transition-colors hover:bg-hover hover:text-fg aria-expanded:bg-hover aria-expanded:text-fg";
  return (
    <>
      {!hasDue ? (
        <Popover open={dueOpen} onOpenChange={setDueOpen}>
          <Tooltip content={t("quickadd.dateTip")}>
            <PopoverTrigger asChild>
              <button type="button" className={btn} onMouseDown={(e) => e.preventDefault()}>
                <CalendarDays className="size-3.5" /> {t("quickadd.date")}
              </button>
            </PopoverTrigger>
          </Tooltip>
          <PopoverContent className="p-0" onOpenAutoFocus={(e) => e.preventDefault()}>
            <DuePicker
              value={override.due}
              onChange={(due) => {
                clearIgnore("due");
                setOverride((o) => {
                  const n = { ...o };
                  if (due) n.due = due;
                  else delete n.due;
                  return n;
                });
              }}
              onPicked={() => setDueOpen(false)}
            />
          </PopoverContent>
        </Popover>
      ) : null}
      {!hasPriority ? (
        <Menu>
          <Tooltip content={t("quickadd.priorityTip")}>
            <MenuTrigger asChild>
              <button type="button" className={btn} onMouseDown={(e) => e.preventDefault()}>
                <Signal className="size-3.5" /> {t("quickadd.priority")}
              </button>
            </MenuTrigger>
          </Tooltip>
          <MenuContent onCloseAutoFocus={(e) => e.preventDefault()}>
            <PriorityItems value={override.priority ?? 0} onPick={(p) => (clearIgnore("priority"), setOverride((o) => ({ ...o, priority: p })))} />
          </MenuContent>
        </Menu>
      ) : null}
      {!hasGroup && !inGroup ? (
        <Menu>
          <Tooltip content={t("quickadd.groupTip")}>
            <MenuTrigger asChild>
              <button type="button" className={btn} onMouseDown={(e) => e.preventDefault()}>
                <Hash className="size-3.5" /> {t("quickadd.group")}
              </button>
            </MenuTrigger>
          </Tooltip>
          <MenuContent onCloseAutoFocus={(e) => e.preventDefault()}>
            <GroupItems
              value={override.groupId}
              onPick={(id) => {
                clearIgnore("group");
                setOverride((o) => {
                  const n = { ...o };
                  if (id) n.groupId = id;
                  else delete n.groupId;
                  return n;
                });
              }}
            />
          </MenuContent>
        </Menu>
      ) : null}
    </>
  );
}
