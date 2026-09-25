// Server state: TanStack Query keys, queries, and the task mutations with
// their optimistic updates (applied to every cached list, rolled back on
// failure, reconciled with the server's answer).
import {
  MutationCache,
  QueryClient,
  useMutation,
  useQuery,
  useQueryClient,
  type QueryKey,
} from "@tanstack/react-query";
import * as ep from "./endpoints";
import type { Counts, Group, NewTask, Task, TaskPatch, TaskView, User, UserRef } from "./types";

export const qk = {
  me: ["me"] as const,
  tasksRoot: ["tasks"] as const,
  tasks: (view?: TaskView, group?: string) => ["tasks", view ?? "all", group ?? ""] as const,
  task: (id: string) => ["task", id] as const,
  counts: ["counts"] as const,
  contacts: ["contacts"] as const,
  groups: ["groups"] as const,
  group: (id: string) => ["group", id] as const,
  invitations: ["invitations"] as const,
  activity: ["activity"] as const,
  sessions: ["sessions"] as const,
};

/**
 * The app's query client. onMutationError hears every failed mutation that
 * did not opt out with meta.silent — one place for the error toast.
 */
export function makeQueryClient(onMutationError?: (err: unknown) => void): QueryClient {
  return new QueryClient({
    mutationCache: new MutationCache({
      onError: (err, _vars, _ctx, mutation) => {
        if (!mutation.meta?.silent) onMutationError?.(err);
      },
    }),
    defaultOptions: {
      queries: {
        staleTime: 20_000,
        gcTime: 5 * 60_000,
        refetchOnWindowFocus: true,
        retry: (count, err) => {
          const status = (err as { status?: number }).status ?? 0;
          if (status >= 400 && status < 500) return false;
          return count < 2;
        },
      },
      mutations: { retry: false },
    },
  });
}

// ── Queries ───────────────────────────────────────────────────────────

export function useMe() {
  return useQuery({ queryKey: qk.me, queryFn: ({ signal }) => ep.auth.me(signal), staleTime: 5 * 60_000 });
}

/** The signed-in user; only call below the auth gate. */
export function useCurrentUser(): User {
  const { data } = useMe();
  if (!data) throw new Error("useCurrentUser outside the auth gate");
  return data;
}

export function useTasks(view?: TaskView, group?: string, enabled = true) {
  return useQuery({
    queryKey: qk.tasks(view, group),
    queryFn: ({ signal }) => ep.tasks.list(view, group, signal),
    enabled,
  });
}

function findCachedTask(qc: QueryClient, id: string): Task | undefined {
  for (const [, list] of qc.getQueriesData<Task[]>({ queryKey: qk.tasksRoot })) {
    const t = list?.find((x) => x.id === id);
    if (t) return t;
  }
  return undefined;
}

export function useTask(id: string | undefined) {
  const qc = useQueryClient();
  return useQuery({
    queryKey: qk.task(id ?? ""),
    queryFn: () => ep.tasks.get(id!),
    enabled: !!id && !id.startsWith("tmp_"),
    initialData: () => (id ? findCachedTask(qc, id) : undefined),
    initialDataUpdatedAt: 0,
  });
}

export function useCounts() {
  return useQuery({ queryKey: qk.counts, queryFn: ep.tasks.counts, refetchInterval: 30_000 });
}

export function useContacts() {
  return useQuery({ queryKey: qk.contacts, queryFn: ep.contacts.list });
}

export function useGroups() {
  return useQuery({ queryKey: qk.groups, queryFn: ep.groups.list });
}

export function useGroup(id: string | undefined) {
  const qc = useQueryClient();
  return useQuery({
    queryKey: qk.group(id ?? ""),
    queryFn: () => ep.groups.get(id!),
    enabled: !!id,
    initialData: () => qc.getQueryData<Group[]>(qk.groups)?.find((g) => g.id === id),
    initialDataUpdatedAt: 0,
  });
}

export function useInvitations() {
  return useQuery({ queryKey: qk.invitations, queryFn: ep.invitations.list, refetchInterval: 60_000 });
}

export function useActivity() {
  return useQuery({ queryKey: qk.activity, queryFn: ep.activity.list });
}

export function useSessions() {
  return useQuery({ queryKey: qk.sessions, queryFn: ep.auth.sessions });
}

/** Everyone the signed-in person can hand a task to: their contacts. */
export function useContactUsers(): UserRef[] {
  const { data } = useContacts();
  return data?.contacts.map((c) => c.user) ?? [];
}

// ── Task mutations ────────────────────────────────────────────────────

type Snapshot = { lists: [QueryKey, Task[] | undefined][]; detail: [QueryKey, Task | undefined][] };

function snapshot(qc: QueryClient): Snapshot {
  return {
    lists: qc.getQueriesData<Task[]>({ queryKey: qk.tasksRoot }),
    detail: qc.getQueriesData<Task>({ queryKey: ["task"] }),
  };
}

function restore(qc: QueryClient, s: Snapshot | undefined) {
  if (!s) return;
  for (const [k, v] of s.lists) qc.setQueryData(k, v);
  for (const [k, v] of s.detail) qc.setQueryData(k, v);
}

/** Applies fn to the task id everywhere it is cached. */
function patchEverywhere(qc: QueryClient, id: string, fn: (t: Task) => Task) {
  qc.setQueriesData<Task[]>({ queryKey: qk.tasksRoot }, (list) => list?.map((t) => (t.id === id ? fn(t) : t)));
  qc.setQueryData<Task>(qk.task(id), (t) => (t ? fn(t) : t));
}

function replaceEverywhere(qc: QueryClient, task: Task, previousId = task.id) {
  qc.setQueriesData<Task[]>({ queryKey: qk.tasksRoot }, (list) =>
    list?.map((t) => (t.id === previousId ? task : t)),
  );
  qc.setQueryData(qk.task(task.id), task);
}

function removeEverywhere(qc: QueryClient, id: string) {
  qc.setQueriesData<Task[]>({ queryKey: qk.tasksRoot }, (list) => list?.filter((t) => t.id !== id));
  qc.removeQueries({ queryKey: qk.task(id), exact: true });
}

function settle(qc: QueryClient) {
  void qc.invalidateQueries({ queryKey: qk.tasksRoot });
  void qc.invalidateQueries({ queryKey: qk.counts });
}

type Lookup = { me?: User; groups?: Group[]; people?: UserRef[] };

/** The optimistic effect of a patch, using what the cache knows about groups and people. */
export function applyPatch(t: Task, patch: TaskPatch, lookup: Lookup, now = new Date()): Task {
  const next: Task = { ...t, updatedAt: now.toISOString() };
  if (patch.title !== undefined) next.title = patch.title;
  if (patch.notes !== undefined) next.notes = patch.notes;
  if (patch.priority !== undefined) next.priority = patch.priority;
  if (patch.due !== undefined) {
    if (patch.due === "") delete next.due;
    else next.due = patch.due;
    if (next.status === "overdue" || next.status === "open") {
      next.status = next.due && Date.parse(next.due) < now.getTime() ? "overdue" : "open";
    }
  }
  if (patch.groupId !== undefined) {
    const g = lookup.groups?.find((x) => x.id === patch.groupId);
    if (patch.groupId === "") delete next.group;
    else if (g) next.group = { id: g.id, name: g.name, color: g.color };
  }
  if (patch.assigneeId !== undefined) {
    if (patch.assigneeId === "") delete next.assignee;
    else {
      const all = [...(lookup.people ?? []), ...(lookup.me ? [lookup.me] : []), t.owner, ...t.sharedWith];
      const u = all.find((x) => x.id === patch.assigneeId);
      if (u) next.assignee = { id: u.id, name: u.name, email: u.email };
    }
  }
  return next;
}

function useLookup(): () => Lookup {
  const qc = useQueryClient();
  return () => ({
    me: qc.getQueryData<User>(qk.me),
    groups: qc.getQueryData<Group[]>(qk.groups),
    people: qc.getQueryData<{ contacts: { user: UserRef }[] }>(qk.contacts)?.contacts.map((c) => c.user),
  });
}

export type CreateVars = { task: NewTask; listKey: QueryKey };

export function useCreateTask() {
  const qc = useQueryClient();
  const lookup = useLookup();
  return useMutation({
    mutationFn: (v: CreateVars) => ep.tasks.create(v.task),
    onMutate: async (v) => {
      await qc.cancelQueries({ queryKey: v.listKey });
      const s = snapshot(qc);
      const l = lookup();
      const me = l.me;
      if (me) {
        const now = new Date().toISOString();
        const tmpId = `tmp_${Math.random().toString(36).slice(2, 10)}`;
        let temp: Task = {
          id: tmpId,
          title: v.task.title,
          notes: v.task.notes ?? "",
          priority: v.task.priority ?? 0,
          status: "open",
          owner: { id: me.id, name: me.name, email: me.email },
          sharedWith: [],
          createdAt: now,
          updatedAt: now,
          can: { edit: true, delete: true, share: true },
        };
        temp = applyPatch(
          temp,
          {
            ...(v.task.due ? { due: v.task.due } : {}),
            ...(v.task.groupId ? { groupId: v.task.groupId } : {}),
            ...(v.task.assigneeId ? { assigneeId: v.task.assigneeId } : {}),
          },
          l,
        );
        qc.setQueryData<Task[]>(v.listKey, (list) => (list ? [...list, temp] : list));
        return { s, tmpId };
      }
      return { s, tmpId: "" };
    },
    onError: (_e, _v, ctx) => restore(qc, ctx?.s),
    onSuccess: (task, v, ctx) => {
      if (ctx?.tmpId) {
        qc.setQueryData<Task[]>(v.listKey, (list) => list?.map((t) => (t.id === ctx.tmpId ? task : t)));
      }
      qc.setQueryData(qk.task(task.id), task);
    },
    onSettled: () => settle(qc),
  });
}

type Action = "complete" | "reopen" | "archive" | "restore";

export function useTaskAction() {
  const qc = useQueryClient();
  const lookup = useLookup();
  return useMutation({
    mutationFn: ({ id, action }: { id: string; action: Action }) => ep.tasks[action](id),
    onMutate: async ({ id, action }) => {
      await qc.cancelQueries({ queryKey: qk.tasksRoot });
      const s = snapshot(qc);
      const me = lookup().me;
      const now = new Date();
      patchEverywhere(qc, id, (t) => {
        const next: Task = { ...t, updatedAt: now.toISOString() };
        if (action === "complete") {
          next.status = "done";
          next.completedAt = now.toISOString();
          if (me) next.completedBy = { id: me.id, name: me.name, email: me.email };
        } else if (action === "archive") next.status = "archived";
        else {
          next.status = t.due && Date.parse(t.due) < now.getTime() ? "overdue" : "open";
          delete next.completedAt;
          delete next.completedBy;
        }
        return next;
      });
      return s;
    },
    onError: (_e, _v, s) => restore(qc, s),
    onSuccess: (task) => replaceEverywhere(qc, task),
    onSettled: () => settle(qc),
  });
}

export function useUpdateTask() {
  const qc = useQueryClient();
  const lookup = useLookup();
  return useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: TaskPatch }) => ep.tasks.update(id, patch),
    onMutate: async ({ id, patch }) => {
      await qc.cancelQueries({ queryKey: qk.tasksRoot });
      await qc.cancelQueries({ queryKey: qk.task(id) });
      const s = snapshot(qc);
      const l = lookup();
      patchEverywhere(qc, id, (t) => applyPatch(t, patch, l));
      return s;
    },
    onError: (_e, _v, s) => restore(qc, s),
    onSuccess: (task) => replaceEverywhere(qc, task),
    onSettled: () => settle(qc),
  });
}

export function useDeleteTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => ep.tasks.remove(id),
    onMutate: async (id) => {
      await qc.cancelQueries({ queryKey: qk.tasksRoot });
      const s = snapshot(qc);
      removeEverywhere(qc, id);
      return s;
    },
    onError: (_e, _v, s) => restore(qc, s),
    onSettled: () => settle(qc),
  });
}

export function useShareTask() {
  const qc = useQueryClient();
  const lookup = useLookup();
  return useMutation({
    mutationFn: ({ id, userId, remove }: { id: string; userId: string; remove?: boolean }) =>
      remove ? ep.tasks.unshare(id, userId) : ep.tasks.share(id, userId),
    onMutate: async ({ id, userId, remove }) => {
      await qc.cancelQueries({ queryKey: qk.tasksRoot });
      const s = snapshot(qc);
      const person = lookup().people?.find((p) => p.id === userId);
      patchEverywhere(qc, id, (t) => ({
        ...t,
        sharedWith: remove
          ? t.sharedWith.filter((u) => u.id !== userId)
          : person && !t.sharedWith.some((u) => u.id === userId)
            ? [...t.sharedWith, person]
            : t.sharedWith,
      }));
      return s;
    },
    onError: (_e, _v, s) => restore(qc, s),
    onSuccess: (task) => {
      if (task) replaceEverywhere(qc, task);
    },
    onSettled: () => settle(qc),
  });
}

// ── Other mutations (no optimistic step: they are rare and confirmable) ─

export function useInvalidate() {
  const qc = useQueryClient();
  return (...keys: QueryKey[]) => Promise.all(keys.map((k) => qc.invalidateQueries({ queryKey: k })));
}

export function setCounts(qc: QueryClient, fn: (c: Counts) => Counts) {
  qc.setQueryData<Counts>(qk.counts, (c) => (c ? fn(c) : c));
}
