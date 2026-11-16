// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:35933b95eb37968572e6d46dc1379b8281e3503fec0e8451a525bac3fb04f4af

package circle

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/kitsunium/todo/contacts/link"
)

// TestAnswerRequestAcceptsARequest is AnswerRequest's scenario "accepts a request" (success).
func TestAnswerRequestAcceptsARequest(t *testing.T) {
	uc, in, then := givenAnswerRequestAcceptsARequest(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("AnswerRequest: %v, want a success", err)
	}
	then(t, out)
}

// TestAnswerRequestRefusesTheRequestOfOthers is AnswerRequest's scenario "refuses the request of others" (refusal).
func TestAnswerRequestRefusesTheRequestOfOthers(t *testing.T) {
	uc, in := givenAnswerRequestRefusesTheRequestOfOthers(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, link.ErrNotFound) {
		t.Fatalf("AnswerRequest: %v, want link.ErrNotFound", err)
	}
}

// TestAnswerRequestRefusesTheRequesterWhoAccepts is AnswerRequest's scenario "refuses the requester who accepts" (refusal).
func TestAnswerRequestRefusesTheRequesterWhoAccepts(t *testing.T) {
	uc, in := givenAnswerRequestRefusesTheRequesterWhoAccepts(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, link.ErrNotAllowed) {
		t.Fatalf("AnswerRequest: %v, want link.ErrNotAllowed", err)
	}
}

// TestAnswerRequestRefusesARequestAlreadyAnswered is AnswerRequest's scenario "refuses a request already answered" (refusal).
func TestAnswerRequestRefusesARequestAlreadyAnswered(t *testing.T) {
	uc, in := givenAnswerRequestRefusesARequestAlreadyAnswered(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, link.ErrNotAllowed) {
		t.Fatalf("AnswerRequest: %v, want link.ErrNotAllowed", err)
	}
}

// TestAreContactsSaysYesForAnAcceptedLinkOnly is AreContacts's scenario "says yes for an accepted link only" (success).
func TestAreContactsSaysYesForAnAcceptedLinkOnly(t *testing.T) {
	uc, in, then := givenAreContactsSaysYesForAnAcceptedLinkOnly(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("AreContacts: %v, want a success", err)
	}
	then(t, out)
}

// TestListContactsListsContactsAndRequestsApart is ListContacts's scenario "lists contacts and requests apart" (success).
func TestListContactsListsContactsAndRequestsApart(t *testing.T) {
	uc, in, then := givenListContactsListsContactsAndRequestsApart(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("ListContacts: %v, want a success", err)
	}
	then(t, out)
}

// TestRemoveContactEndsAContact is RemoveContact's scenario "ends a contact" (success).
func TestRemoveContactEndsAContact(t *testing.T) {
	uc, in, then := givenRemoveContactEndsAContact(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("RemoveContact: %v, want a success", err)
	}
	then(t, out)
}

// TestRemoveContactRefusesAContactOfOthers is RemoveContact's scenario "refuses a contact of others" (refusal).
func TestRemoveContactRefusesAContactOfOthers(t *testing.T) {
	uc, in := givenRemoveContactRefusesAContactOfOthers(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, link.ErrNotFound) {
		t.Fatalf("RemoveContact: %v, want link.ErrNotFound", err)
	}
}

// TestRequestContactAsksAMember is RequestContact's scenario "asks a member" (success).
func TestRequestContactAsksAMember(t *testing.T) {
	uc, in, then := givenRequestContactAsksAMember(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("RequestContact: %v, want a success", err)
	}
	then(t, out)
}

// TestRequestContactAsksAgainAfterARefusal is RequestContact's scenario "asks again after a refusal" (success).
func TestRequestContactAsksAgainAfterARefusal(t *testing.T) {
	uc, in, then := givenRequestContactAsksAgainAfterARefusal(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("RequestContact: %v, want a success", err)
	}
	then(t, out)
}

// TestRequestContactRefusesAnUnknownAddress is RequestContact's scenario "refuses an unknown address" (refusal).
func TestRequestContactRefusesAnUnknownAddress(t *testing.T) {
	uc, in := givenRequestContactRefusesAnUnknownAddress(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, link.ErrUnknownMember) {
		t.Fatalf("RequestContact: %v, want link.ErrUnknownMember", err)
	}
}

// TestRequestContactRefusesOneself is RequestContact's scenario "refuses oneself" (refusal).
func TestRequestContactRefusesOneself(t *testing.T) {
	uc, in := givenRequestContactRefusesOneself(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, link.ErrSelf) {
		t.Fatalf("RequestContact: %v, want link.ErrSelf", err)
	}
}

// TestRequestContactRefusesAPairAlreadyLinked is RequestContact's scenario "refuses a pair already linked" (refusal).
func TestRequestContactRefusesAPairAlreadyLinked(t *testing.T) {
	uc, in := givenRequestContactRefusesAPairAlreadyLinked(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, link.ErrAlready) {
		t.Fatalf("RequestContact: %v, want link.ErrAlready", err)
	}
}

// sequenceRecorder wraps the ports a sequence goes through, and records each
// call in order.
type sequenceRecorder struct {
	calls *[]string
}

// RecordCircleClockV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordCircleClockV1(next ClockV1) ClockV1 {
	return recordedcircleClockV1{next: next, calls: r.calls}
}

type recordedcircleClockV1 struct {
	next  ClockV1
	calls *[]string
}

func (x recordedcircleClockV1) Now(ctx context.Context) (time.Time, error) {
	*x.calls = append(*x.calls, "contacts/port/Clock@v1.Now")
	return x.next.Now(ctx)
}

// RecordCircleLinksV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordCircleLinksV1(next LinksV1) LinksV1 {
	return recordedcircleLinksV1{next: next, calls: r.calls}
}

type recordedcircleLinksV1 struct {
	next  LinksV1
	calls *[]string
}

func (x recordedcircleLinksV1) ByPair(ctx context.Context, pair string) (link.Link, bool, error) {
	*x.calls = append(*x.calls, "contacts/port/Links@v1.ByPair")
	return x.next.ByPair(ctx, pair)
}

func (x recordedcircleLinksV1) Delete(ctx context.Context, id string) error {
	*x.calls = append(*x.calls, "contacts/port/Links@v1.Delete")
	return x.next.Delete(ctx, id)
}

func (x recordedcircleLinksV1) Get(ctx context.Context, id string) (link.Link, bool, error) {
	*x.calls = append(*x.calls, "contacts/port/Links@v1.Get")
	return x.next.Get(ctx, id)
}

func (x recordedcircleLinksV1) NewID(ctx context.Context) (string, error) {
	*x.calls = append(*x.calls, "contacts/port/Links@v1.NewID")
	return x.next.NewID(ctx)
}

func (x recordedcircleLinksV1) Of(ctx context.Context, user string) ([]link.Link, error) {
	*x.calls = append(*x.calls, "contacts/port/Links@v1.Of")
	return x.next.Of(ctx, user)
}

func (x recordedcircleLinksV1) Save(ctx context.Context, l link.Link) error {
	*x.calls = append(*x.calls, "contacts/port/Links@v1.Save")
	return x.next.Save(ctx, l)
}

// RecordCircleNoticesV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordCircleNoticesV1(next NoticesV1) NoticesV1 {
	return recordedcircleNoticesV1{next: next, calls: r.calls}
}

type recordedcircleNoticesV1 struct {
	next  NoticesV1
	calls *[]string
}

func (x recordedcircleNoticesV1) Tell(ctx context.Context, kind string, l link.Link, actor Person) error {
	*x.calls = append(*x.calls, "contacts/port/Notices@v1.Tell")
	return x.next.Tell(ctx, kind, l, actor)
}

// RecordCirclePeopleV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordCirclePeopleV1(next PeopleV1) PeopleV1 {
	return recordedcirclePeopleV1{next: next, calls: r.calls}
}

type recordedcirclePeopleV1 struct {
	next  PeopleV1
	calls *[]string
}

func (x recordedcirclePeopleV1) ByEmail(ctx context.Context, email string) (Person, bool, error) {
	*x.calls = append(*x.calls, "contacts/port/People@v1.ByEmail")
	return x.next.ByEmail(ctx, email)
}

func (x recordedcirclePeopleV1) ByID(ctx context.Context, id string) (Person, bool, error) {
	*x.calls = append(*x.calls, "contacts/port/People@v1.ByID")
	return x.next.ByID(ctx, id)
}

// TestSequenceRequestContact is the order of sequence request-contact: the diagram's links, in the
// diagram's order (D21).
func TestSequenceRequestContact(t *testing.T) {
	var got []string
	run := givenSequenceRequestContact(t, sequenceRecorder{calls: &got})
	if err := run(t.Context()); err != nil {
		t.Fatalf("the sequence failed: %v", err)
	}
	want := []string{"contacts/port/People@v1.ByEmail", "contacts/port/Links@v1.ByPair", "contacts/port/Clock@v1.Now", "contacts/port/Links@v1.NewID", "contacts/port/Links@v1.Save", "contacts/port/People@v1.ByID", "contacts/port/Notices@v1.Tell"}
	if !slices.Equal(got, want) {
		t.Errorf("the calls: %q, want %q", got, want)
	}
}
