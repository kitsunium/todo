package tasks

import (
	"context"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/internal/wire"
)

// The questions other services may ask about tasks, in-process. A service's
// data is its own: stats, groups and notify ask here rather than read the
// store.
//
// fr: Les questions que les autres services peuvent poser sur les tâches, dans
// le processus. Les données d’un service lui appartiennent : stats, groups et
// notify demandent ici plutôt que de lire le store.
var (
	// CensusAPI counts every task, for stats.
	//
	// fr: CensusAPI compte toutes les tâches, pour stats.
	CensusAPI = Service.Endpoint("GET /internal/tasks/census", Census, kit.Private())

	// OpenByGroupAPI counts the tasks left to do on groups' lists, for
	// groups.
	//
	// fr: OpenByGroupAPI compte les tâches qui restent à faire sur les listes
	// des groupes, pour groups.
	OpenByGroupAPI = Service.Endpoint("POST /internal/tasks/open-by-group", OpenByGroup, kit.Private())

	// AudienceAPI says whom a task concerns right now, for notify's
	// reminders.
	//
	// fr: AudienceAPI dit qui une tâche concerne en ce moment, pour les rappels
	// de notify.
	AudienceAPI = Service.Endpoint("GET /internal/tasks/audience", Audience, kit.Private())
)

// priorityNames names the priorities in the census.
var priorityNames = map[Priority]string{NoPriority: "none", Urgent: "urgent", High: "high", Medium: "medium", Low: "low"}

// CensusOutput counts the tasks.
type CensusOutput struct {
	// Counts has every status, zeros included.
	Counts map[Status]int `json:"counts"`
	Total  int            `json:"total"`
	// OpenByPriority counts the tasks left to do per priority.
	OpenByPriority map[string]int `json:"openByPriority"`
	// Shared counts the tasks shared with at least one user; Grouped those
	// on a group's list.
	Shared  int `json:"shared"`
	Grouped int `json:"grouped"`
}

// Census counts every task by status and the open ones by priority.
//
// fr: Census compte toutes les tâches par statut, et les tâches ouvertes par
// priorité.
func Census(ctx context.Context, _ kit.Empty) (CensusOutput, error) {
	all, err := Tasks.List(ctx)
	if err != nil {
		return CensusOutput{}, err
	}
	out := CensusOutput{
		Counts:         map[Status]int{Open: 0, Overdue: 0, Done: 0, Archived: 0},
		OpenByPriority: map[string]int{"none": 0, "urgent": 0, "high": 0, "medium": 0, "low": 0},
		Total:          len(all),
	}
	for _, t := range all {
		out.Counts[t.Status]++
		if t.Active() {
			out.OpenByPriority[priorityNames[t.Priority]]++
		}
		if len(t.SharedWith) > 0 {
			out.Shared++
		}
		if t.GroupID != "" {
			out.Grouped++
		}
	}
	return out, nil
}

// GroupIDs are groups to count.
type GroupIDs struct {
	IDs []string `json:"ids"`
}

// OpenByGroupOutput counts, per group asked, the tasks left to do.
type OpenByGroupOutput struct {
	Counts map[string]int `json:"counts"`
}

// OpenByGroup counts the open and overdue tasks on each group's list.
//
// fr: OpenByGroup compte les tâches ouvertes et en retard sur la liste de
// chaque groupe.
func OpenByGroup(ctx context.Context, in GroupIDs) (OpenByGroupOutput, error) {
	out := OpenByGroupOutput{Counts: map[string]int{}}
	for _, id := range in.IDs {
		if _, done := out.Counts[id]; done || id == "" {
			continue
		}
		listed, err := Tasks.Find(ctx, "group", id)
		if err != nil {
			return OpenByGroupOutput{}, err
		}
		n := 0
		for _, t := range listed {
			if t.Active() {
				n++
			}
		}
		out.Counts[id] = n
	}
	return out, nil
}

// AudienceInput names a task.
type AudienceInput struct {
	Task string `query:"task"`
}

// Brief is a task as a notification shows it.
type Brief struct {
	ID       string     `json:"id"`
	Title    string     `json:"title"`
	Status   Status     `json:"status"`
	Priority Priority   `json:"priority"`
	Due      *time.Time `json:"due,omitempty"`
}

// AudienceOutput is the task, and the users it concerns in person: owner,
// assignee, and the users it is shared with. Task is nil for a task that no
// longer exists.
type AudienceOutput struct {
	Task  *Brief   `json:"task,omitempty"`
	Users []string `json:"users"`
}

// Audience says, now, whom a task concerns — a reminder is for the people a
// task concerns when it fires, not when it was scheduled.
//
// fr: Audience dit, maintenant, qui une tâche concerne — un rappel s’adresse
// aux personnes que la tâche concerne quand il part, pas quand il a été
// programmé.
func Audience(ctx context.Context, in AudienceInput) (AudienceOutput, error) {
	t, err := Tasks.Get(ctx, in.Task)
	if wire.Is(err, kit.CodeNotFound) {
		return AudienceOutput{Users: []string{}}, nil
	}
	if err != nil {
		return AudienceOutput{}, err
	}
	return AudienceOutput{
		Task:  &Brief{ID: t.ID, Title: t.Title, Status: t.Status, Priority: t.Priority, Due: t.Due},
		Users: t.Audience(),
	}, nil
}
