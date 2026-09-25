package api

import (
	"context"
	"slices"
	"time"
	"unicode/utf8"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/activity"
	"github.com/kitsunium/todo/contacts"
	"github.com/kitsunium/todo/internal/wire"
	"github.com/kitsunium/todo/tasks"
)

// The task list's API. Every route acts for the signed-in user, on the tasks
// they see.
//
// fr: L’API de la liste de tâches. Chaque route agit pour l’utilisateur
// connecté, sur les tâches qu’il voit.
var (
	_ = tasks.Service.Endpoint("GET /api/tasks", List, kit.Auth())
	_ = tasks.Service.Endpoint("GET /api/tasks/counts", Counts, kit.Auth())
	_ = tasks.Service.Endpoint("POST /api/tasks", Create, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = tasks.Service.Endpoint("GET /api/tasks/{id}", Get, kit.Auth())
	_ = tasks.Service.Endpoint("PATCH /api/tasks/{id}", Update, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = tasks.Service.Endpoint("POST /api/tasks/{id}/complete", Complete, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = tasks.Service.Endpoint("POST /api/tasks/{id}/reopen", Reopen, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = tasks.Service.Endpoint("POST /api/tasks/{id}/archive", Archive, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = tasks.Service.Endpoint("POST /api/tasks/{id}/restore", Restore, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = tasks.Service.Endpoint("DELETE /api/tasks/{id}", Delete, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = tasks.Service.Endpoint("POST /api/tasks/{id}/share", Share, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = tasks.Service.Endpoint("DELETE /api/tasks/{id}/share/{userId}", Unshare, kit.Auth(), kit.RateLimitPerClient(10, 20))
)

// maxNotes bounds a task's notes, in characters.
const maxNotes = 10000

// ListInput chooses a view of the list.
type ListInput struct {
	// View is inbox, today, upcoming, shared, assigned, completed or group;
	// empty is every task the caller sees, but the archived ones.
	View string `query:"view"`
	// Group narrows the list to a group's tasks; the group view needs it.
	Group string `query:"group"`
	// TZ is the caller's time zone, which says when today ends: an IANA
	// name, UTC by default.
	TZ string `query:"tz"`
}

// ListOutput is a view of the list, the tasks left to do first.
type ListOutput struct {
	Tasks []View `json:"tasks"`
}

// List returns a view of the tasks the caller sees.
//
// fr: List renvoie une vue des tâches que voit l’appelant.
func List(ctx context.Context, in ListInput) (ListOutput, error) {
	uid, err := me(ctx)
	if err != nil {
		return ListOutput{}, err
	}
	switch in.View {
	case ViewAll, ViewInbox, ViewToday, ViewUpcoming, ViewShared, ViewAssigned, ViewCompleted:
	case ViewGroup:
		if in.Group == "" {
			return ListOutput{}, wire.Invalid("group", "required", "is required by the group view")
		}
	default:
		return ListOutput{}, wire.Invalid("view", "oneof", "must be inbox, today, upcoming, shared, assigned, completed or group")
	}
	cal, err := calendarAt(wire.Now(ctx), in.TZ)
	if err != nil {
		return ListOutput{}, err
	}
	all, ms, err := visible(ctx, uid)
	if err != nil {
		return ListOutput{}, err
	}
	if _, member := ms[in.Group]; in.Group != "" && !member {
		return ListOutput{}, kit.NotFound("no such group")
	}
	var picked []tasks.Task
	for _, t := range all {
		if (in.Group == "" || t.GroupID == in.Group) && cal.in(in.View, t, uid) {
			picked = append(picked, t)
		}
	}
	if in.View == ViewCompleted {
		slices.SortFunc(picked, byCompletion)
	} else {
		slices.SortFunc(picked, byUrgency)
	}
	r, err := resolve(ctx, picked, ms)
	if err != nil {
		return ListOutput{}, err
	}
	out := ListOutput{Tasks: make([]View, 0, len(picked))}
	for _, t := range picked {
		out.Tasks = append(out.Tasks, r.view(t, ms.access(t, uid)))
	}
	return out, nil
}

// CountsInput is the caller's time zone.
type CountsInput struct {
	TZ string `query:"tz"`
}

// CountsOutput counts what the sidebar shows: the tasks left to do in each
// view, the overdue ones, the ones finished this week, the open tasks of
// each of the caller's groups, and their unread activity.
type CountsOutput struct {
	Inbox             int            `json:"inbox"`
	Today             int            `json:"today"`
	Upcoming          int            `json:"upcoming"`
	Shared            int            `json:"shared"`
	Assigned          int            `json:"assigned"`
	Overdue           int            `json:"overdue"`
	CompletedThisWeek int            `json:"completedThisWeek"`
	Groups            map[string]int `json:"groups"`
	Unread            int            `json:"unread"`
}

// Counts counts the caller's views, groups and unread activity.
//
// fr: Counts compte les vues de l’appelant, ses groupes et son activité non
// lue.
func Counts(ctx context.Context, in CountsInput) (CountsOutput, error) {
	uid, err := me(ctx)
	if err != nil {
		return CountsOutput{}, err
	}
	cal, err := calendarAt(wire.Now(ctx), in.TZ)
	if err != nil {
		return CountsOutput{}, err
	}
	all, ms, err := visible(ctx, uid)
	if err != nil {
		return CountsOutput{}, err
	}
	out := CountsOutput{Groups: map[string]int{}}
	for id := range ms {
		out.Groups[id] = 0
	}
	counted := []struct {
		view string
		n    *int
	}{
		{ViewInbox, &out.Inbox}, {ViewToday, &out.Today}, {ViewUpcoming, &out.Upcoming},
		{ViewShared, &out.Shared}, {ViewAssigned, &out.Assigned},
	}
	for _, t := range all {
		for _, c := range counted {
			if cal.in(c.view, t, uid) {
				*c.n++
			}
		}
		if t.Status == tasks.Overdue {
			out.Overdue++
		}
		if t.CompletedAt != nil && !t.CompletedAt.Before(cal.weekStart) {
			out.CompletedThisWeek++
		}
		if _, member := ms[t.GroupID]; member && t.Active() {
			out.Groups[t.GroupID]++
		}
	}
	unread, err := activity.UnreadAPI.Call(ctx, activity.UnreadQuery{User: uid})
	if err != nil {
		return CountsOutput{}, err
	}
	out.Unread = unread.Unread
	return out, nil
}

// CreateInput is a new task.
type CreateInput struct {
	Title    string         `json:"title" validate:"required,maxlen=200"`
	Notes    string         `json:"notes" validate:"maxlen=10000"`
	Priority tasks.Priority `json:"priority" validate:"min=0,max=4"`
	// Due is an RFC 3339 time.
	Due *time.Time `json:"due"`
	// GroupID puts the task on one of the caller's groups.
	GroupID string `json:"groupId" validate:"maxlen=64"`
	// AssigneeID is the caller, or a member of the task's group.
	AssigneeID string `json:"assigneeId" validate:"maxlen=64"`
}

// Create adds a task to the caller's list, or to one of their groups'.
//
// fr: Create ajoute une tâche à la liste de l’appelant, ou à celle de l’un de
// ses groupes.
func Create(ctx context.Context, in CreateInput) (View, error) {
	uid, err := me(ctx)
	if err != nil {
		return View{}, err
	}
	title, err := wire.Line("title", in.Title, 200)
	if err != nil {
		return View{}, err
	}
	now := wire.Now(ctx)
	t := tasks.Task{
		ID: kit.NewID("task"), Title: title, Notes: in.Notes, Priority: in.Priority, OwnerID: uid,
		GroupID: in.GroupID, AssigneeID: in.AssigneeID, SharedWith: []string{}, CreatedAt: now, UpdatedAt: now,
	}
	if in.Due != nil {
		due := in.Due.UTC().Truncate(time.Millisecond)
		t.Due = &due
	}
	if t.GroupID != "" {
		role, err := roleIn(ctx, t.GroupID, uid)
		if err != nil {
			return View{}, err
		}
		if role == "" {
			return View{}, wire.Invalid("groupId", "member", "is not a group you belong to")
		}
	}
	if t.AssigneeID != "" {
		ok, err := assignable(ctx, t, t.AssigneeID)
		if err != nil {
			return View{}, err
		}
		if !ok {
			return View{}, wire.Invalid("assigneeId", "audience", "must be you or a member of the task's group")
		}
	}
	t, err = Lifecycle.Start(ctx, t)
	if err != nil {
		return View{}, err
	}
	if t.AssigneeID != "" && t.AssigneeID != uid {
		e := newEvent(ctx, t, tasks.KindAssigned, uid)
		e.UserID = t.AssigneeID
		if err := publish(ctx, t, e); err != nil {
			return View{}, err
		}
	}
	return present(ctx, t, access{sees: true, deletes: true, shares: true})
}

// TaskID addresses a task.
type TaskID struct {
	ID string `path:"id"`
}

// Get returns one task.
//
// fr: Get renvoie une tâche.
func Get(ctx context.Context, in TaskID) (View, error) {
	t, acc, _, err := load(ctx, in.ID)
	if err != nil {
		return View{}, err
	}
	return present(ctx, t, acc)
}

// UpdateInput changes a task: an absent member is unchanged, and an empty
// due, groupId or assigneeId clears it.
type UpdateInput struct {
	ID         string          `path:"id"`
	Title      *string         `json:"title"`
	Notes      *string         `json:"notes"`
	Priority   *tasks.Priority `json:"priority"`
	Due        *string         `json:"due"`
	GroupID    *string         `json:"groupId"`
	AssigneeID *string         `json:"assigneeId"`
}

// edit is a checked change to a task.
type edit struct {
	title, notes, due, group, assignee, priority bool
	next                                         tasks.Task
}

// check validates a change against the task it changes, for the caller.
func (in UpdateInput) check(ctx context.Context, t tasks.Task, uid string) (edit, error) {
	e := edit{next: t}
	if in.Title != nil {
		title, err := wire.Line("title", *in.Title, 200)
		if err != nil {
			return e, err
		}
		e.title, e.next.Title = true, title
	}
	if in.Notes != nil {
		if utf8.RuneCountInString(*in.Notes) > maxNotes {
			return e, wire.Invalid("notes", "maxlen", "must be at most 10000 characters long")
		}
		e.notes, e.next.Notes = true, *in.Notes
	}
	if in.Priority != nil {
		if *in.Priority < tasks.NoPriority || *in.Priority > tasks.Low {
			return e, wire.Invalid("priority", "oneof", "must be 0 (none), 1 (urgent), 2 (high), 3 (medium) or 4 (low)")
		}
		e.priority, e.next.Priority = true, *in.Priority
	}
	if in.Due != nil {
		e.due, e.next.Due = true, nil
		if *in.Due != "" {
			due, err := time.Parse(time.RFC3339Nano, *in.Due)
			if err != nil {
				return e, wire.Invalid("due", "datetime", "must be an RFC 3339 time, or empty to clear it")
			}
			due = due.UTC().Truncate(time.Millisecond)
			e.next.Due = &due
		}
	}
	if in.GroupID != nil && *in.GroupID != t.GroupID {
		if t.OwnerID != uid {
			return e, kit.Forbidden("only the task's owner can move it to another group")
		}
		if *in.GroupID != "" {
			role, err := roleIn(ctx, *in.GroupID, uid)
			if err != nil {
				return e, err
			}
			if role == "" {
				return e, wire.Invalid("groupId", "member", "is not a group you belong to")
			}
		}
		e.group, e.next.GroupID = true, *in.GroupID
	}
	if in.AssigneeID != nil && *in.AssigneeID != t.AssigneeID {
		e.assignee, e.next.AssigneeID = true, *in.AssigneeID
	}
	if e.next.AssigneeID != "" && (e.assignee || e.group) {
		ok, err := assignable(ctx, e.next, e.next.AssigneeID)
		if err != nil {
			return e, err
		}
		switch {
		case !ok && e.assignee:
			return e, wire.Invalid("assigneeId", "audience", "must be the owner, someone the task is shared with, or a member of its group")
		case !ok:
			// The task left the group its assignee does it for.
			e.assignee, e.next.AssigneeID = true, ""
		}
	}
	return e, nil
}

// Update changes a task. Anyone who sees it may; only its owner moves it to
// another group. Moving an overdue task to a later date puts it back on the
// list of things to do.
//
// fr: Update modifie une tâche. Quiconque la voit le peut ; seul son
// propriétaire la déplace vers un autre groupe. Repousser une tâche en retard à
// une date ultérieure la remet parmi les choses à faire.
func Update(ctx context.Context, in UpdateInput) (View, error) {
	t, _, uid, err := load(ctx, in.ID)
	if err != nil {
		return View{}, err
	}
	e, err := in.check(ctx, t, uid)
	if err != nil {
		return View{}, err
	}
	now := wire.Now(ctx)
	updated, err := tasks.Tasks.Update(ctx, t.ID, func(cur *tasks.Task) error {
		if e.title {
			cur.Title = e.next.Title
		}
		if e.notes {
			cur.Notes = e.next.Notes
		}
		if e.priority {
			cur.Priority = e.next.Priority
		}
		if e.due {
			cur.Due = e.next.Due
		}
		if e.group {
			cur.GroupID = e.next.GroupID
		}
		if e.assignee {
			cur.AssigneeID = e.next.AssigneeID
		}
		cur.UpdatedAt = now
		return nil
	})
	if wire.Is(err, kit.CodeNotFound) {
		return View{}, errNoTask()
	}
	if err != nil {
		return View{}, err
	}
	if updated.Status == tasks.Overdue && !pastDue(updated, now) {
		rescheduled, err := Lifecycle.Fire(ctx, updated.ID, "reschedule")
		if err != nil && !wire.Is(err, kit.CodeConflict) {
			return View{}, err
		}
		if err == nil {
			updated = rescheduled
		}
	}
	if err := publish(ctx, updated, newEvent(ctx, updated, tasks.KindUpdated, uid)); err != nil {
		return View{}, err
	}
	if e.assignee && updated.AssigneeID != "" && updated.AssigneeID != uid {
		ev := newEvent(ctx, updated, tasks.KindAssigned, uid)
		ev.UserID = updated.AssigneeID
		if err := publish(ctx, updated, ev); err != nil {
			return View{}, err
		}
	}
	acc, err := accessTo(ctx, updated, uid)
	if err != nil {
		return View{}, err
	}
	return present(ctx, updated, acc)
}

// conflict explains a transition the task's state refuses.
func conflict(err error, what string) error {
	if wire.Is(err, kit.CodeConflict) {
		return kit.Conflict("this task cannot be " + what + " now")
	}
	if wire.Is(err, kit.CodeNotFound) {
		return errNoTask()
	}
	return err
}

// Complete marks a task done.
//
// fr: Complete marque une tâche comme terminée.
func Complete(ctx context.Context, in TaskID) (View, error) {
	t, acc, _, err := load(ctx, in.ID)
	if err != nil {
		return View{}, err
	}
	if t, err = Lifecycle.Fire(ctx, t.ID, "complete"); err != nil {
		return View{}, conflict(err, "completed")
	}
	return present(ctx, t, acc)
}

// Reopen puts a done task back on the list.
//
// fr: Reopen remet une tâche terminée sur la liste.
func Reopen(ctx context.Context, in TaskID) (View, error) {
	t, acc, _, err := load(ctx, in.ID)
	if err != nil {
		return View{}, err
	}
	if t, err = Lifecycle.Fire(ctx, t.ID, "reopen"); err != nil {
		return View{}, conflict(err, "reopened")
	}
	return present(ctx, t, acc)
}

// Archive takes a done task off the lists, before its timer does.
//
// fr: Archive retire une tâche terminée des listes, avant que son timer ne le
// fasse.
func Archive(ctx context.Context, in TaskID) (View, error) {
	t, acc, _, err := load(ctx, in.ID)
	if err != nil {
		return View{}, err
	}
	if t, err = Lifecycle.Fire(ctx, t.ID, "archive"); err != nil {
		return View{}, conflict(err, "archived")
	}
	return present(ctx, t, acc)
}

// Restore puts an archived task back on the list, to do.
//
// fr: Restore remet une tâche archivée sur la liste, à faire.
func Restore(ctx context.Context, in TaskID) (View, error) {
	t, acc, _, err := load(ctx, in.ID)
	if err != nil {
		return View{}, err
	}
	if t, err = Lifecycle.Fire(ctx, t.ID, "restore"); err != nil {
		return View{}, conflict(err, "restored")
	}
	return present(ctx, t, acc)
}

// Delete deletes a task for everyone. Its owner may, and so may an owner or
// an admin of its group.
//
// fr: Delete supprime une tâche pour tout le monde. Son propriétaire le peut,
// tout comme un propriétaire ou un admin de son groupe.
func Delete(ctx context.Context, in TaskID) (kit.Empty, error) {
	t, acc, uid, err := load(ctx, in.ID)
	if err != nil {
		return kit.Empty{}, err
	}
	if !acc.deletes {
		return kit.Empty{}, kit.Forbidden("only the task's owner, or an owner or admin of its group, can delete it")
	}
	e := newEvent(ctx, t, tasks.KindDeleted, uid)
	if e.Recipients, err = recipients(ctx, t, uid); err != nil {
		return kit.Empty{}, err
	}
	if err := tasks.Tasks.Delete(ctx, t.ID); err != nil {
		return kit.Empty{}, conflict(err, "deleted")
	}
	return kit.Empty{}, tasks.Events.Publish(ctx, e)
}

// ShareInput names the contact to share a task with.
type ShareInput struct {
	ID     string `path:"id"`
	UserID string `json:"userId" validate:"required,maxlen=64"`
}

// Share lets one of the owner's contacts see and work on a task. Sharing
// twice with the same user changes nothing.
//
// fr: Share permet à l’un des contacts du propriétaire de voir une tâche et d’y
// travailler. Partager deux fois avec le même utilisateur ne change rien.
func Share(ctx context.Context, in ShareInput) (View, error) {
	t, acc, uid, err := load(ctx, in.ID)
	if err != nil {
		return View{}, err
	}
	switch {
	case !acc.shares:
		return View{}, kit.Forbidden("only the task's owner can share it")
	case in.UserID == uid:
		return View{}, wire.Invalid("userId", "self", "is you: you own this task")
	case slices.Contains(t.SharedWith, in.UserID):
		return present(ctx, t, acc)
	}
	check, err := contacts.CheckAPI.Call(ctx, contacts.CheckInput{A: uid, B: in.UserID})
	if err != nil {
		return View{}, err
	}
	if !check.Contacts {
		return View{}, kit.Forbidden("you can only share a task with your contacts")
	}
	now := wire.Now(ctx)
	updated, err := tasks.Tasks.Update(ctx, t.ID, func(cur *tasks.Task) error {
		if !slices.Contains(cur.SharedWith, in.UserID) {
			cur.SharedWith = append(slices.Clone(cur.SharedWith), in.UserID)
		}
		cur.UpdatedAt = now
		return nil
	})
	if err != nil {
		return View{}, conflict(err, "shared")
	}
	e := newEvent(ctx, updated, tasks.KindShared, uid)
	e.UserID = in.UserID
	if err := publish(ctx, updated, e); err != nil {
		return View{}, err
	}
	return present(ctx, updated, acc)
}

// UnshareInput names a user a task is shared with.
type UnshareInput struct {
	ID     string `path:"id"`
	UserID string `path:"userId"`
}

// Unshare stops sharing a task with a user: the owner may, and so may the
// user, to leave it. A user who no longer sees the task stops being its
// assignee.
//
// fr: Unshare arrête de partager une tâche avec un utilisateur : le
// propriétaire le peut, et l’utilisateur aussi, pour la quitter. Un utilisateur
// qui ne voit plus la tâche n’y est plus assigné.
func Unshare(ctx context.Context, in UnshareInput) (View, error) {
	t, acc, uid, err := load(ctx, in.ID)
	if err != nil {
		return View{}, err
	}
	if !acc.shares && in.UserID != uid {
		return View{}, kit.Forbidden("only the task's owner can stop sharing it")
	}
	if !slices.Contains(t.SharedWith, in.UserID) {
		return present(ctx, t, acc)
	}
	e := newEvent(ctx, t, tasks.KindUnshared, uid)
	e.UserID = in.UserID
	if e.Recipients, err = recipients(ctx, t, uid); err != nil { // everyone who saw it
		return View{}, err
	}
	role, err := roleIn(ctx, t.GroupID, in.UserID)
	if err != nil {
		return View{}, err
	}
	now := wire.Now(ctx)
	updated, err := tasks.Tasks.Update(ctx, t.ID, func(cur *tasks.Task) error {
		cur.SharedWith = slices.DeleteFunc(slices.Clone(cur.SharedWith), func(id string) bool { return id == in.UserID })
		if cur.AssigneeID == in.UserID && role == "" && cur.OwnerID != in.UserID {
			cur.AssigneeID = ""
		}
		cur.UpdatedAt = now
		return nil
	})
	if err != nil {
		return View{}, conflict(err, "unshared")
	}
	e.Status, e.Due = updated.Status, updated.Due
	if err := tasks.Events.Publish(ctx, e); err != nil {
		return View{}, err
	}
	acc, err = accessTo(ctx, updated, uid)
	if err != nil {
		return View{}, err
	}
	return present(ctx, updated, acc)
}
