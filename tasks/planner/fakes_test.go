// The ports the use cases' tests run on: tasks in memory, a clock that
// stands still, a journal that remembers.

package planner

import (
	"context"
	"strconv"
	"time"

	"github.com/kitsunium/todo/tasks/task"
)

// ada owns every task a test makes; bob and cy own none.
const (
	ada = "user_ada"
	bob = "user_bob"
	cy  = "user_cy"
)

// epoch is the instant every test starts at.
var epoch = time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC)

// memTasks is TaskStoreV1 in memory.
type memTasks struct {
	byID map[string]task.Task
	next int
}

func newMemTasks(ts ...task.Task) *memTasks {
	m := &memTasks{byID: map[string]task.Task{}}
	for _, t := range ts {
		m.byID[t.ID] = t
	}
	return m
}

func (m *memTasks) NewID(context.Context) (string, error) {
	m.next++
	return "task_" + strconv.Itoa(m.next), nil
}

func (m *memTasks) Save(_ context.Context, t task.Task) error {
	m.byID[t.ID] = t
	return nil
}

func (m *memTasks) Get(_ context.Context, id string) (task.Task, error) {
	t, ok := m.byID[id]
	if !ok {
		return task.Task{}, task.ErrNotFound
	}
	return t, nil
}

func (m *memTasks) Delete(_ context.Context, id string) error {
	delete(m.byID, id)
	return nil
}

// contacts is CircleV1 over the pairs it holds, "one|other".
type contacts map[string]bool

func (c contacts) AreContacts(_ context.Context, one, other string) (bool, error) {
	return c[one+"|"+other] || c[other+"|"+one], nil
}

// shared is an open task of Ada's shared with Bob.
func shared(id string) task.Task {
	t := open(id)
	t.SharedWith = []string{bob}
	return t
}

func (m *memTasks) List(context.Context) ([]task.Task, error) {
	out := make([]task.Task, 0, len(m.byID))
	for _, t := range m.byID {
		out = append(out, t)
	}
	return out, nil
}

// fixedClock is ClockV1 at one instant.
type fixedClock time.Time

func (c fixedClock) Now(context.Context) (time.Time, error) { return time.Time(c), nil }

// memJournal is JournalV1 that remembers what it was told.
type memJournal struct {
	kinds []string
}

func (j *memJournal) Record(_ context.Context, kind string, _ task.Task) error {
	j.kinds = append(j.kinds, kind)
	return nil
}

// open, done and archived are tasks in each state, for a test to adjust.
func open(id string) task.Task {
	return task.Task{ID: id, OwnerID: ada, Title: "task " + id, Priority: task.PriorityNormal, State: task.LifecycleOpen, CreatedAt: epoch, UpdatedAt: epoch}
}

func done(id string, at time.Time) task.Task {
	t := open(id)
	t.State, t.DoneAt, t.UpdatedAt = task.LifecycleDone, at, at
	return t
}
