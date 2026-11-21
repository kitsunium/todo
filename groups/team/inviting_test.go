package team

import (
	"errors"
	"testing"
)

// An invitation is pending until it is accepted or declined; once answered
// or expired it moves no more.
func TestInvitingMovesAPendingInvitationOnly(t *testing.T) {
	w := NewInviting()
	for event, want := range map[InvitingEvent]InvitingState{InvitingOnAccept: InvitingAccepted, InvitingOnDecline: InvitingDeclined} {
		got, err := w.Fire(t.Context(), &Invitation{}, InvitingInitial, event)
		if err != nil || got != want {
			t.Errorf("%s from pending: %q (%v), want %q", event, got, err, want)
		}
	}
	for _, from := range []InvitingState{InvitingAccepted, InvitingDeclined, InvitingExpired} {
		for _, event := range []InvitingEvent{InvitingOnAccept, InvitingOnDecline} {
			if got, err := w.Fire(t.Context(), &Invitation{}, from, event); !errors.Is(err, ErrInvitingNoTransition) || got != from {
				t.Errorf("%s from %s: %q (%v), want to stay", event, from, got, err)
			}
		}
	}
}
