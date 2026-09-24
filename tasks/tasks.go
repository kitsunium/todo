// Package tasks is the model of the task list, and all other services see of
// it: what a task is, the store that keeps it, the events that announce what
// happens to it, and the questions other services may ask about it.
//
// The lifecycle and the HTTP API live in tasks/api, on this same service.
// They check groups, contacts and activity; groups and activity in turn ask
// this package about tasks — so the part they import depends on nothing,
// and Go's imports stay a tree while the services call each other both ways.
package tasks

import (
	"slices"
	"time"

	"github.com/kitsunium/platform/kit"
)

// Service owns the task list.
var Service = kit.NewService("tasks", "The task list: every task, its lifecycle, who sees it and what happens to it.\n\nfr: La liste de tâches : chaque tâche, son cycle de vie, qui la voit et ce qui lui arrive.")

// Status is where a task stands in its lifecycle.
type Status string

// The states of a task.
const (
	Open     Status = "open"
	Overdue  Status = "overdue"
	Done     Status = "done"
	Archived Status = "archived"
)

// Priority orders tasks the way Linear does: 1 is the most urgent, 0 none.
type Priority int

// The priorities.
const (
	NoPriority Priority = 0
	Urgent     Priority = 1
	High       Priority = 2
	Medium     Priority = 3
	Low        Priority = 4
)

// Task is one thing to do.
type Task struct {
	ID       string     `json:"id"`
	Title    string     `json:"title"`
	Notes    string     `json:"notes"`
	Priority Priority   `json:"priority"`
	Status   Status     `json:"status"`
	Due      *time.Time `json:"due,omitempty"`
	// OwnerID created the task. The owner shares it, moves it between
	// groups, and deletes it.
	OwnerID string `json:"ownerId"`
	// AssigneeID is who does it: the owner, a user it is shared with, or a
	// member of its group.
	AssigneeID string `json:"assigneeId,omitempty"`
	// GroupID puts the task on a group's list: every member sees it.
	GroupID string `json:"groupId,omitempty"`
	// SharedWith are the contacts of the owner who see it too.
	SharedWith  []string   `json:"sharedWith"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CompletedBy string     `json:"completedBy,omitempty"`
}

// Active reports whether the task is still to do: open or overdue.
func (t Task) Active() bool { return t.Status == Open || t.Status == Overdue }

// Audience is who a task concerns in person: its owner, its assignee and the
// users it is shared with — not the members of its group.
func (t Task) Audience() []string {
	out := []string{t.OwnerID}
	for _, id := range append([]string{t.AssigneeID}, t.SharedWith...) {
		if id != "" && !slices.Contains(out, id) {
			out = append(out, id)
		}
	}
	return out
}

// Tasks keeps every task, indexed by the users and the group it concerns:
// what a user sees is found through the indexes, never by reading them all.
//
// fr: Tasks garde chaque tâche, indexée par les utilisateurs et le groupe
// qu’elle concerne : ce que voit un utilisateur se trouve par les index, jamais
// en les lisant toutes.
var Tasks = Service.Store("tasks", func(t Task) string { return t.ID },
	kit.Index("owner", func(t Task) []string { return []string{t.OwnerID} }),
	kit.Index("shared", func(t Task) []string { return t.SharedWith }),
	kit.Index("group", func(t Task) []string { return []string{t.GroupID} }),
	kit.Index("assignee", func(t Task) []string { return []string{t.AssigneeID} }))

// Kind says what happened to a task.
type Kind string

// The kinds of Event.
const (
	KindCreated   Kind = "created"
	KindUpdated   Kind = "updated"
	KindCompleted Kind = "completed"
	KindReopened  Kind = "reopened"
	KindArchived  Kind = "archived"
	KindRestored  Kind = "restored"
	KindOverdue   Kind = "overdue"
	KindShared    Kind = "shared"
	KindUnshared  Kind = "unshared"
	KindAssigned  Kind = "assigned"
	KindDeleted   Kind = "deleted"
)

// Event is what happened to a task, and whom to tell.
type Event struct {
	// ID identifies the event: a redelivery carries the same one.
	ID     string `json:"id"`
	TaskID string `json:"taskId"`
	Title  string `json:"title"`
	Kind   Kind   `json:"kind"`
	// ActorID did it; empty when the product itself did — the overdue
	// guard, the auto-archive timer.
	ActorID string `json:"actorId,omitempty"`
	// UserID is whom a shared, unshared or assigned event is about.
	UserID string `json:"userId,omitempty"`
	// GroupID is the task's group.
	GroupID string `json:"groupId,omitempty"`
	// Recipients are everyone who sees the task, the actor excepted — for
	// an unshared or deleted task, everyone who saw it.
	Recipients []string  `json:"recipients"`
	At         time.Time `json:"at"`
	// From and To are the states of a transition.
	From Status `json:"from,omitempty"`
	To   Status `json:"to,omitempty"`
	// Status and Due are the task's, after the event.
	Status   Status     `json:"status"`
	Due      *time.Time `json:"due,omitempty"`
	Priority Priority   `json:"priority"`
}

// Events announces everything that happens to a task: activity writes the
// feeds, notify mails the people concerned and tracks due dates.
//
// fr: Events annonce tout ce qui arrive à une tâche : activity écrit les fils
// d’activité, notify écrit aux personnes concernées et suit les échéances.
var Events = Service.Topic[Event]("events")
