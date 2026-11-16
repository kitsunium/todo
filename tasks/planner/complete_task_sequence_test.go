// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"context"
	"testing"
)

// givenSequenceCompleteTask builds tasks/usecase/MoveTask with the ports it goes through wrapped by rec
// (rec.Record…), and returns what starts it.
func givenSequenceCompleteTask(t *testing.T, rec sequenceRecorder) func(context.Context) error {
	t.Helper()
	uc := NewMoveTask(
		rec.RecordPlannerTaskStoreV1(newMemTasks(open("task_a"))),
		rec.RecordPlannerClockV1(fixedClock(epoch)),
		rec.RecordPlannerJournalV1(&memJournal{}),
	)
	return func(ctx context.Context) error {
		_, err := uc.Execute(ctx, MoveTaskInput{Owner: ada, ID: "task_a", Event: "complete"})
		return err
	}
}
