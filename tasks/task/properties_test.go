// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package task

import (
	"errors"
	"math/rand/v2"
	"testing"
	"time"
)

// propertyTaskOverdueNeverWhenDone holds tasks/Task.Overdue: whatever its
// dates, a task that is not open is never overdue, and an open one is
// exactly when it has a due date behind now.
func propertyTaskOverdueNeverWhenDone(t *testing.T, seed uint64) {
	t.Helper()
	r := rand.New(rand.NewPCG(seed, seed))
	states := []LifecycleState{LifecycleOpen, LifecycleDone, LifecycleArchived}
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for range 2000 {
		now := base.Add(time.Duration(r.Int64N(int64(1000 * time.Hour))))
		tk := Task{State: states[r.IntN(len(states))]}
		if r.IntN(4) > 0 {
			tk.Due = base.Add(time.Duration(r.Int64N(int64(1000 * time.Hour))))
		}
		got := tk.Overdue(now)
		if tk.State != LifecycleOpen && got {
			t.Fatalf("a %s task due %s is overdue at %s", tk.State, tk.Due, now)
		}
		if want := tk.State == LifecycleOpen && !tk.Due.IsZero() && tk.Due.Before(now); got != want {
			t.Fatalf("an open task due %s at %s: overdue %v, want %v", tk.Due, now, got, want)
		}
	}
}

// propertyParsePriorityKnownOrRefused holds tasks/function/ParsePriority: a
// text is read as one of the three priorities, or refused with
// ErrUnknownPriority and no priority.
func propertyParsePriorityKnownOrRefused(t *testing.T, seed uint64) {
	t.Helper()
	r := rand.New(rand.NewPCG(seed, seed))
	words := []string{"", "low", "normal", "high", "LOW", " High ", "urgent", "0", "lo w", "normal\n", "haut"}
	letters := []rune("lowhignrmaLOWH 1é")
	for range 2000 {
		s := words[r.IntN(len(words))]
		if r.IntN(3) == 0 {
			b := make([]rune, r.IntN(8))
			for i := range b {
				b[i] = letters[r.IntN(len(letters))]
			}
			s = string(b)
		}
		p, err := ParsePriority(s)
		switch {
		case err == nil:
			if p != PriorityLow && p != PriorityNormal && p != PriorityHigh {
				t.Fatalf("ParsePriority(%q) = %q, no priority of the design", s, p)
			}
		case !errors.Is(err, ErrUnknownPriority):
			t.Fatalf("ParsePriority(%q): %v, want ErrUnknownPriority", s, err)
		case p != "":
			t.Fatalf("ParsePriority(%q) refuses and still answers %q", s, p)
		}
	}
}

// propertyTaskSeenByOwnerSharedAssigneeOrMember holds tasks/Task.SeenBy: a
// member sees a task exactly when they own it, it is shared with them, it is
// assigned to them, or it is on a group's list and they are in that group;
// nobody is never one of them.
func propertyTaskSeenByOwnerSharedAssigneeOrMember(t *testing.T, seed uint64) {
	t.Helper()
	r := rand.New(rand.NewPCG(seed, seed))
	who := func() string { return "user_" + string(rune('a'+r.IntN(6))) }
	for range 2000 {
		tk := Task{OwnerID: who()}
		for range r.IntN(4) {
			tk.SharedWith = append(tk.SharedWith, who())
		}
		if r.IntN(2) == 0 {
			tk.AssigneeID = who()
		}
		if r.IntN(2) == 0 {
			tk.GroupID = "group_core"
		}
		u, inGroup := who(), r.IntN(2) == 0
		want := u == tk.OwnerID || u == tk.AssigneeID || inGroup && tk.GroupID != ""
		for _, s := range tk.SharedWith {
			want = want || s == u
		}
		if got := tk.SeenBy(u, inGroup); got != want {
			t.Fatalf("%s (in the group %v) sees the task of %s shared with %q, assigned to %q, on %q: %v, want %v", u, inGroup, tk.OwnerID, tk.SharedWith, tk.AssigneeID, tk.GroupID, got, want)
		}
		if tk.SeenBy("", true) {
			t.Fatalf("nobody sees the task of %s", tk.OwnerID)
		}
	}
}
