// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"context"
	"testing"
)

// givenSequenceInviteContact builds groups/usecase/Invite with the ports it goes through wrapped by rec
// (rec.Record…), and returns what starts it.
func givenSequenceInviteContact(t *testing.T, rec sequenceRecorder) func(context.Context) error {
	t.Helper()
	w := newWorld()
	g := w.core()
	w.befriend("user_ada", "user_dee")
	uc := NewInvite(rec.RecordRosterGroupsV1(w.groupsPort()), rec.RecordRosterCircleV1(w), rec.RecordRosterInvitationsV1(w.invitationsPort()),
		rec.RecordRosterClockV1(w), rec.RecordRosterPeopleV1(w), rec.RecordRosterNoticesV1(w))
	return func(ctx context.Context) error {
		_, err := uc.Execute(ctx, InviteInput{User: "user_ada", ID: g.ID, Invitee: "user_dee"})
		return err
	}
}
