// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"testing"
	"time"

	"github.com/kitsunium/todo/tasks/task"
)

// givenArchiveDoneArchivesWhatWasDoneLongAgo builds ArchiveDone for "archives what was done long ago": the use case, its input, and
// what its output must be.
func givenArchiveDoneArchivesWhatWasDoneLongAgo(t *testing.T) (ArchiveDone, ArchiveDoneInput, func(*testing.T, ArchiveDoneOutput)) {
	t.Helper()
	now := epoch.Add(48 * time.Hour)
	store := newMemTasks(open("task_open"), done("task_old", epoch), done("task_fresh", now.Add(-time.Hour)))
	journal := &memJournal{}
	return NewArchiveDone(fixedClock(now), store, journal), ArchiveDoneInput{After: 24 * time.Hour}, func(t *testing.T, out ArchiveDoneOutput) {
		t.Helper()
		if out.Archived != 1 {
			t.Errorf("archived %d task(s), want 1", out.Archived)
		}
		for id, want := range map[string]task.LifecycleState{"task_open": task.LifecycleOpen, "task_old": task.LifecycleArchived, "task_fresh": task.LifecycleDone} {
			if got, _ := store.Get(t.Context(), id); got.State != want {
				t.Errorf("%s is %s, want %s", id, got.State, want)
			}
		}
		if len(journal.kinds) != 1 || journal.kinds[0] != "archive" {
			t.Errorf("the journal was told %q, want archive once", journal.kinds)
		}
	}
}
