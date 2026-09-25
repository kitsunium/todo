package api

import (
	"cmp"
	"context"
	"slices"
	"time"
	_ "time/tzdata" // the views read the caller's time zone by name, in any image

	"github.com/kitsunium/todo/groups"
	"github.com/kitsunium/todo/identity"
	"github.com/kitsunium/todo/internal/wire"
	"github.com/kitsunium/todo/tasks"
)

// Can says what the caller may do with a task.
type Can struct {
	Edit   bool `json:"edit"`
	Delete bool `json:"delete"`
	Share  bool `json:"share"`
}

// View is a task as a user sees it: the people and the group it names
// resolved, and what they may do with it.
type View struct {
	ID          string             `json:"id"`
	Title       string             `json:"title"`
	Notes       string             `json:"notes"`
	Priority    tasks.Priority     `json:"priority"`
	Status      tasks.Status       `json:"status"`
	Due         *time.Time         `json:"due,omitempty"`
	Owner       identity.UserRef   `json:"owner"`
	Assignee    *identity.UserRef  `json:"assignee,omitempty"`
	Group       *groups.GroupRef   `json:"group,omitempty"`
	SharedWith  []identity.UserRef `json:"sharedWith"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
	CompletedAt *time.Time         `json:"completedAt,omitempty"`
	CompletedBy *identity.UserRef  `json:"completedBy,omitempty"`
	Can         Can                `json:"can"`
}

// resolver knows the users and the groups a set of tasks names.
type resolver struct {
	users  map[string]identity.UserRef
	groups map[string]groups.GroupRef
}

// resolve asks identity for every user the tasks name, and groups for every
// group not already known — in one call each, whatever the number of tasks.
func resolve(ctx context.Context, ts []tasks.Task, known memberships) (resolver, error) {
	var users, unknown []string
	for _, t := range ts {
		users = append(users, t.OwnerID, t.AssigneeID, t.CompletedBy)
		users = append(users, t.SharedWith...)
		if _, ok := known[t.GroupID]; t.GroupID != "" && !ok && !slices.Contains(unknown, t.GroupID) {
			unknown = append(unknown, t.GroupID)
		}
	}
	names, err := identity.Directory(ctx, users...)
	if err != nil {
		return resolver{}, err
	}
	r := resolver{users: names, groups: map[string]groups.GroupRef{}}
	for id, m := range known {
		r.groups[id] = m.Group
	}
	if len(unknown) > 0 {
		found, err := groups.BatchAPI.Call(ctx, groups.IDs{IDs: unknown})
		if err != nil {
			return resolver{}, err
		}
		for _, g := range found.Groups {
			r.groups[g.ID] = g
		}
	}
	return r, nil
}

// user resolves a user ID, nil for none.
func (r resolver) user(id string) *identity.UserRef {
	if id == "" {
		return nil
	}
	u := r.users[id]
	return &u
}

// view is t as seen with access a.
func (r resolver) view(t tasks.Task, a access) View {
	v := View{
		ID: t.ID, Title: t.Title, Notes: t.Notes, Priority: t.Priority, Status: t.Status, Due: t.Due,
		Owner: r.users[t.OwnerID], Assignee: r.user(t.AssigneeID), SharedWith: make([]identity.UserRef, 0, len(t.SharedWith)),
		CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt, CompletedAt: t.CompletedAt, CompletedBy: r.user(t.CompletedBy),
		Can: Can{Edit: a.sees, Delete: a.deletes, Share: a.shares},
	}
	if g, ok := r.groups[t.GroupID]; ok {
		v.Group = &g
	}
	for _, id := range t.SharedWith {
		v.SharedWith = append(v.SharedWith, r.users[id])
	}
	return v
}

// present is one task as the caller sees it.
func present(ctx context.Context, t tasks.Task, a access) (View, error) {
	r, err := resolve(ctx, []tasks.Task{t}, nil)
	if err != nil {
		return View{}, err
	}
	return r.view(t, a), nil
}

// calendar is the caller's day and week, in their time zone.
type calendar struct {
	now       time.Time
	tomorrow  time.Time // the end of today
	weekStart time.Time // Monday, 00:00
}

// calendarAt reads the caller's time zone — an IANA name, UTC by default —
// and places now in it.
func calendarAt(now time.Time, tz string) (calendar, error) {
	loc := time.UTC
	if tz != "" {
		l, err := time.LoadLocation(tz)
		if err != nil || tz == "Local" {
			return calendar{}, wire.Invalid("tz", "timezone", "must be an IANA time zone, like Europe/Paris")
		}
		loc = l
	}
	local := now.In(loc)
	midnight := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	sinceMonday := (int(local.Weekday()) + 6) % 7
	return calendar{now: now, tomorrow: midnight.AddDate(0, 0, 1), weekStart: midnight.AddDate(0, 0, -sinceMonday)}, nil
}

// The views of the task list.
const (
	ViewAll       = ""
	ViewInbox     = "inbox"
	ViewToday     = "today"
	ViewUpcoming  = "upcoming"
	ViewShared    = "shared"
	ViewAssigned  = "assigned"
	ViewCompleted = "completed"
	ViewGroup     = "group"
)

// in reports whether t belongs to a view, for user uid. A view lists what is
// left to do, but three: the whole list and a group's list keep their done
// tasks until they are archived, and completed lists the finished ones —
// done or archived.
func (c calendar) in(view string, t tasks.Task, uid string) bool {
	switch view {
	case ViewInbox:
		return t.OwnerID == uid && t.GroupID == "" && t.Active()
	case ViewToday:
		return t.Active() && t.Due != nil && t.Due.Before(c.tomorrow)
	case ViewUpcoming:
		return t.Active() && t.Due != nil && !t.Due.Before(c.tomorrow)
	case ViewShared:
		return slices.Contains(t.SharedWith, uid) && t.Active()
	case ViewAssigned:
		return t.AssigneeID == uid && t.Active()
	case ViewCompleted:
		return t.Status == tasks.Done || t.Status == tasks.Archived
	}
	return t.Status != tasks.Archived
}

// statusRank orders the states on a list: what is left to do first.
var statusRank = map[tasks.Status]int{tasks.Overdue: 0, tasks.Open: 0, tasks.Done: 1, tasks.Archived: 2}

// priorityRank orders the priorities: urgent first, none last.
func priorityRank(p tasks.Priority) int {
	if p == tasks.NoPriority {
		return 5
	}
	return int(p)
}

// byUrgency orders a list: to do first, then the soonest due, then the most
// urgent, then the newest.
func byUrgency(a, b tasks.Task) int {
	due := func(t tasks.Task) (int, time.Time) {
		if t.Due == nil {
			return 1, time.Time{}
		}
		return 0, *t.Due
	}
	an, ad := due(a)
	bn, bd := due(b)
	return cmp.Or(
		cmp.Compare(statusRank[a.Status], statusRank[b.Status]),
		cmp.Compare(an, bn), ad.Compare(bd),
		cmp.Compare(priorityRank(a.Priority), priorityRank(b.Priority)),
		b.CreatedAt.Compare(a.CreatedAt),
		cmp.Compare(a.ID, b.ID),
	)
}

// byCompletion orders the completed view: the most recently finished first.
func byCompletion(a, b tasks.Task) int {
	at, bt := a.UpdatedAt, b.UpdatedAt
	if a.CompletedAt != nil {
		at = *a.CompletedAt
	}
	if b.CompletedAt != nil {
		bt = *b.CompletedAt
	}
	return cmp.Or(bt.Compare(at), cmp.Compare(a.ID, b.ID))
}
