// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:f78b90bd39d042efd423450d5f8ea59f0f6ddf7d6ea3d6874b6a7bc17cd1eabd

package roster

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/kitsunium/todo/groups/team"
)

// TestAnswerInvitationAcceptsAnInvitation is AnswerInvitation's scenario "accepts an invitation" (success).
func TestAnswerInvitationAcceptsAnInvitation(t *testing.T) {
	uc, in, then := givenAnswerInvitationAcceptsAnInvitation(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("AnswerInvitation: %v, want a success", err)
	}
	then(t, out)
}

// TestAnswerInvitationDeclinesAnInvitation is AnswerInvitation's scenario "declines an invitation" (success).
func TestAnswerInvitationDeclinesAnInvitation(t *testing.T) {
	uc, in, then := givenAnswerInvitationDeclinesAnInvitation(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("AnswerInvitation: %v, want a success", err)
	}
	then(t, out)
}

// TestAnswerInvitationRefusesTheInvitationOfOthers is AnswerInvitation's scenario "refuses the invitation of others" (refusal).
func TestAnswerInvitationRefusesTheInvitationOfOthers(t *testing.T) {
	uc, in := givenAnswerInvitationRefusesTheInvitationOfOthers(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrNoInvitation) {
		t.Fatalf("AnswerInvitation: %v, want team.ErrNoInvitation", err)
	}
}

// TestAnswerInvitationRefusesAnInvitationAlreadyAnswered is AnswerInvitation's scenario "refuses an invitation already answered" (refusal).
func TestAnswerInvitationRefusesAnInvitationAlreadyAnswered(t *testing.T) {
	uc, in := givenAnswerInvitationRefusesAnInvitationAlreadyAnswered(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrNotPending) {
		t.Fatalf("AnswerInvitation: %v, want team.ErrNotPending", err)
	}
}

// TestAnswerInvitationRefusesAnInvitationIntoADeletedGroup is AnswerInvitation's scenario "refuses an invitation into a deleted group" (refusal).
func TestAnswerInvitationRefusesAnInvitationIntoADeletedGroup(t *testing.T) {
	uc, in := givenAnswerInvitationRefusesAnInvitationIntoADeletedGroup(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrNotFound) {
		t.Fatalf("AnswerInvitation: %v, want team.ErrNotFound", err)
	}
}

// TestCreateGroupCreatesAGroup is CreateGroup's scenario "creates a group" (success).
func TestCreateGroupCreatesAGroup(t *testing.T) {
	uc, in, then := givenCreateGroupCreatesAGroup(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("CreateGroup: %v, want a success", err)
	}
	then(t, out)
}

// TestCreateGroupPicksAColorFromTheName is CreateGroup's scenario "picks a color from the name" (success).
func TestCreateGroupPicksAColorFromTheName(t *testing.T) {
	uc, in, then := givenCreateGroupPicksAColorFromTheName(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("CreateGroup: %v, want a success", err)
	}
	then(t, out)
}

// TestCreateGroupRefusesABlankName is CreateGroup's scenario "refuses a blank name" (refusal).
func TestCreateGroupRefusesABlankName(t *testing.T) {
	uc, in := givenCreateGroupRefusesABlankName(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrBadName) {
		t.Fatalf("CreateGroup: %v, want team.ErrBadName", err)
	}
}

// TestCreateGroupRefusesAnUnknownColor is CreateGroup's scenario "refuses an unknown color" (refusal).
func TestCreateGroupRefusesAnUnknownColor(t *testing.T) {
	uc, in := givenCreateGroupRefusesAnUnknownColor(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrBadColor) {
		t.Fatalf("CreateGroup: %v, want team.ErrBadColor", err)
	}
}

// TestDeleteGroupDeletesAGroup is DeleteGroup's scenario "deletes a group" (success).
func TestDeleteGroupDeletesAGroup(t *testing.T) {
	uc, in, then := givenDeleteGroupDeletesAGroup(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("DeleteGroup: %v, want a success", err)
	}
	then(t, out)
}

// TestDeleteGroupRefusesTheGroupOfOthers is DeleteGroup's scenario "refuses the group of others" (refusal).
func TestDeleteGroupRefusesTheGroupOfOthers(t *testing.T) {
	uc, in := givenDeleteGroupRefusesTheGroupOfOthers(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrNotFound) {
		t.Fatalf("DeleteGroup: %v, want team.ErrNotFound", err)
	}
}

// TestDeleteGroupRefusesAnAdmin is DeleteGroup's scenario "refuses an admin" (refusal).
func TestDeleteGroupRefusesAnAdmin(t *testing.T) {
	uc, in := givenDeleteGroupRefusesAnAdmin(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrNotOwner) {
		t.Fatalf("DeleteGroup: %v, want team.ErrNotOwner", err)
	}
}

// TestEditGroupEditsAGroup is EditGroup's scenario "edits a group" (success).
func TestEditGroupEditsAGroup(t *testing.T) {
	uc, in, then := givenEditGroupEditsAGroup(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("EditGroup: %v, want a success", err)
	}
	then(t, out)
}

// TestEditGroupRefusesTheGroupOfOthers is EditGroup's scenario "refuses the group of others" (refusal).
func TestEditGroupRefusesTheGroupOfOthers(t *testing.T) {
	uc, in := givenEditGroupRefusesTheGroupOfOthers(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrNotFound) {
		t.Fatalf("EditGroup: %v, want team.ErrNotFound", err)
	}
}

// TestEditGroupRefusesAPlainMember is EditGroup's scenario "refuses a plain member" (refusal).
func TestEditGroupRefusesAPlainMember(t *testing.T) {
	uc, in := givenEditGroupRefusesAPlainMember(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrNotManager) {
		t.Fatalf("EditGroup: %v, want team.ErrNotManager", err)
	}
}

// TestEditGroupRefusesABlankColor is EditGroup's scenario "refuses a blank color" (refusal).
func TestEditGroupRefusesABlankColor(t *testing.T) {
	uc, in := givenEditGroupRefusesABlankColor(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrBadColor) {
		t.Fatalf("EditGroup: %v, want team.ErrBadColor", err)
	}
}

// TestGetGroupShowsAGroupWithItsMembersRanked is GetGroup's scenario "shows a group with its members ranked" (success).
func TestGetGroupShowsAGroupWithItsMembersRanked(t *testing.T) {
	uc, in, then := givenGetGroupShowsAGroupWithItsMembersRanked(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("GetGroup: %v, want a success", err)
	}
	then(t, out)
}

// TestGetGroupRefusesTheGroupOfOthers is GetGroup's scenario "refuses the group of others" (refusal).
func TestGetGroupRefusesTheGroupOfOthers(t *testing.T) {
	uc, in := givenGetGroupRefusesTheGroupOfOthers(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrNotFound) {
		t.Fatalf("GetGroup: %v, want team.ErrNotFound", err)
	}
}

// TestInviteInvitesAContact is Invite's scenario "invites a contact" (success).
func TestInviteInvitesAContact(t *testing.T) {
	uc, in, then := givenInviteInvitesAContact(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("Invite: %v, want a success", err)
	}
	then(t, out)
}

// TestInviteRefusesTheGroupOfOthers is Invite's scenario "refuses the group of others" (refusal).
func TestInviteRefusesTheGroupOfOthers(t *testing.T) {
	uc, in := givenInviteRefusesTheGroupOfOthers(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrNotFound) {
		t.Fatalf("Invite: %v, want team.ErrNotFound", err)
	}
}

// TestInviteRefusesAPlainMember is Invite's scenario "refuses a plain member" (refusal).
func TestInviteRefusesAPlainMember(t *testing.T) {
	uc, in := givenInviteRefusesAPlainMember(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrNotManager) {
		t.Fatalf("Invite: %v, want team.ErrNotManager", err)
	}
}

// TestInviteRefusesSomeoneAlreadyIn is Invite's scenario "refuses someone already in" (refusal).
func TestInviteRefusesSomeoneAlreadyIn(t *testing.T) {
	uc, in := givenInviteRefusesSomeoneAlreadyIn(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrAlreadyMember) {
		t.Fatalf("Invite: %v, want team.ErrAlreadyMember", err)
	}
}

// TestInviteRefusesWhoIsNoContact is Invite's scenario "refuses who is no contact" (refusal).
func TestInviteRefusesWhoIsNoContact(t *testing.T) {
	uc, in := givenInviteRefusesWhoIsNoContact(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrNotContact) {
		t.Fatalf("Invite: %v, want team.ErrNotContact", err)
	}
}

// TestInviteRefusesAnInvitationAlreadyWaiting is Invite's scenario "refuses an invitation already waiting" (refusal).
func TestInviteRefusesAnInvitationAlreadyWaiting(t *testing.T) {
	uc, in := givenInviteRefusesAnInvitationAlreadyWaiting(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrAlreadyInvited) {
		t.Fatalf("Invite: %v, want team.ErrAlreadyInvited", err)
	}
}

// TestListGroupsListsTheGroupsOfAMemberByName is ListGroups's scenario "lists the groups of a member by name" (success).
func TestListGroupsListsTheGroupsOfAMemberByName(t *testing.T) {
	uc, in, then := givenListGroupsListsTheGroupsOfAMemberByName(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("ListGroups: %v, want a success", err)
	}
	then(t, out)
}

// TestListInvitationsListsThePendingInvitationsNewestFirst is ListInvitations's scenario "lists the pending invitations newest first" (success).
func TestListInvitationsListsThePendingInvitationsNewestFirst(t *testing.T) {
	uc, in, then := givenListInvitationsListsThePendingInvitationsNewestFirst(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("ListInvitations: %v, want a success", err)
	}
	then(t, out)
}

// TestMembersOfListsTheMembersOfAGroup is MembersOf's scenario "lists the members of a group" (success).
func TestMembersOfListsTheMembersOfAGroup(t *testing.T) {
	uc, in, then := givenMembersOfListsTheMembersOfAGroup(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("MembersOf: %v, want a success", err)
	}
	then(t, out)
}

// TestMembershipsListsTheGroupsAMemberBelongsTo is Memberships's scenario "lists the groups a member belongs to" (success).
func TestMembershipsListsTheGroupsAMemberBelongsTo(t *testing.T) {
	uc, in, then := givenMembershipsListsTheGroupsAMemberBelongsTo(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("Memberships: %v, want a success", err)
	}
	then(t, out)
}

// TestRemoveMemberLeavesAGroup is RemoveMember's scenario "leaves a group" (success).
func TestRemoveMemberLeavesAGroup(t *testing.T) {
	uc, in, then := givenRemoveMemberLeavesAGroup(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("RemoveMember: %v, want a success", err)
	}
	then(t, out)
}

// TestRemoveMemberRemovesAMember is RemoveMember's scenario "removes a member" (success).
func TestRemoveMemberRemovesAMember(t *testing.T) {
	uc, in, then := givenRemoveMemberRemovesAMember(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("RemoveMember: %v, want a success", err)
	}
	then(t, out)
}

// TestRemoveMemberRefusesAnUnknownMember is RemoveMember's scenario "refuses an unknown member" (refusal).
func TestRemoveMemberRefusesAnUnknownMember(t *testing.T) {
	uc, in := givenRemoveMemberRefusesAnUnknownMember(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrNoMember) {
		t.Fatalf("RemoveMember: %v, want team.ErrNoMember", err)
	}
}

// TestRemoveMemberRefusesTheOwnerWhoLeaves is RemoveMember's scenario "refuses the owner who leaves" (refusal).
func TestRemoveMemberRefusesTheOwnerWhoLeaves(t *testing.T) {
	uc, in := givenRemoveMemberRefusesTheOwnerWhoLeaves(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrOwnerStays) {
		t.Fatalf("RemoveMember: %v, want team.ErrOwnerStays", err)
	}
}

// TestRemoveMemberRefusesAPlainMemberWhoRemoves is RemoveMember's scenario "refuses a plain member who removes" (refusal).
func TestRemoveMemberRefusesAPlainMemberWhoRemoves(t *testing.T) {
	uc, in := givenRemoveMemberRefusesAPlainMemberWhoRemoves(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrCannotRemove) {
		t.Fatalf("RemoveMember: %v, want team.ErrCannotRemove", err)
	}
}

// TestRemoveMemberRefusesAnAdminWhoRemovesAnAdmin is RemoveMember's scenario "refuses an admin who removes an admin" (refusal).
func TestRemoveMemberRefusesAnAdminWhoRemovesAnAdmin(t *testing.T) {
	uc, in := givenRemoveMemberRefusesAnAdminWhoRemovesAnAdmin(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrCannotRemove) {
		t.Fatalf("RemoveMember: %v, want team.ErrCannotRemove", err)
	}
}

// TestRemoveMemberRefusesTheRemovalOfTheOwner is RemoveMember's scenario "refuses the removal of the owner" (refusal).
func TestRemoveMemberRefusesTheRemovalOfTheOwner(t *testing.T) {
	uc, in := givenRemoveMemberRefusesTheRemovalOfTheOwner(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrCannotRemove) {
		t.Fatalf("RemoveMember: %v, want team.ErrCannotRemove", err)
	}
}

// TestRemoveMemberRefusesTheGroupOfOthers is RemoveMember's scenario "refuses the group of others" (refusal).
func TestRemoveMemberRefusesTheGroupOfOthers(t *testing.T) {
	uc, in := givenRemoveMemberRefusesTheGroupOfOthers(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrNotFound) {
		t.Fatalf("RemoveMember: %v, want team.ErrNotFound", err)
	}
}

// TestRoleOfTellsTheRoleOfAMemberAndNothingOfAStranger is RoleOf's scenario "tells the role of a member and nothing of a stranger" (success).
func TestRoleOfTellsTheRoleOfAMemberAndNothingOfAStranger(t *testing.T) {
	uc, in, then := givenRoleOfTellsTheRoleOfAMemberAndNothingOfAStranger(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("RoleOf: %v, want a success", err)
	}
	then(t, out)
}

// TestSetRoleMakesAMemberAnAdmin is SetRole's scenario "makes a member an admin" (success).
func TestSetRoleMakesAMemberAnAdmin(t *testing.T) {
	uc, in, then := givenSetRoleMakesAMemberAnAdmin(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("SetRole: %v, want a success", err)
	}
	then(t, out)
}

// TestSetRoleRefusesAnUnknownRole is SetRole's scenario "refuses an unknown role" (refusal).
func TestSetRoleRefusesAnUnknownRole(t *testing.T) {
	uc, in := givenSetRoleRefusesAnUnknownRole(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrBadRole) {
		t.Fatalf("SetRole: %v, want team.ErrBadRole", err)
	}
}

// TestSetRoleRefusesTheGroupOfOthers is SetRole's scenario "refuses the group of others" (refusal).
func TestSetRoleRefusesTheGroupOfOthers(t *testing.T) {
	uc, in := givenSetRoleRefusesTheGroupOfOthers(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrNotFound) {
		t.Fatalf("SetRole: %v, want team.ErrNotFound", err)
	}
}

// TestSetRoleRefusesAnAdmin is SetRole's scenario "refuses an admin" (refusal).
func TestSetRoleRefusesAnAdmin(t *testing.T) {
	uc, in := givenSetRoleRefusesAnAdmin(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrNotOwner) {
		t.Fatalf("SetRole: %v, want team.ErrNotOwner", err)
	}
}

// TestSetRoleRefusesAnUnknownMember is SetRole's scenario "refuses an unknown member" (refusal).
func TestSetRoleRefusesAnUnknownMember(t *testing.T) {
	uc, in := givenSetRoleRefusesAnUnknownMember(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrNoMember) {
		t.Fatalf("SetRole: %v, want team.ErrNoMember", err)
	}
}

// TestSetRoleRefusesToChangeTheOwner is SetRole's scenario "refuses to change the owner" (refusal).
func TestSetRoleRefusesToChangeTheOwner(t *testing.T) {
	uc, in := givenSetRoleRefusesToChangeTheOwner(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, team.ErrOwnerRole) {
		t.Fatalf("SetRole: %v, want team.ErrOwnerRole", err)
	}
}

// sequenceRecorder wraps the ports a sequence goes through, and records each
// call in order.
type sequenceRecorder struct {
	calls *[]string
}

// RecordRosterCircleV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordRosterCircleV1(next CircleV1) CircleV1 {
	return recordedrosterCircleV1{next: next, calls: r.calls}
}

type recordedrosterCircleV1 struct {
	next  CircleV1
	calls *[]string
}

func (x recordedrosterCircleV1) AreContacts(ctx context.Context, one string, other string) (bool, error) {
	*x.calls = append(*x.calls, "groups/port/Circle@v1.AreContacts")
	return x.next.AreContacts(ctx, one, other)
}

// RecordRosterClockV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordRosterClockV1(next ClockV1) ClockV1 {
	return recordedrosterClockV1{next: next, calls: r.calls}
}

type recordedrosterClockV1 struct {
	next  ClockV1
	calls *[]string
}

func (x recordedrosterClockV1) Now(ctx context.Context) (time.Time, error) {
	*x.calls = append(*x.calls, "groups/port/Clock@v1.Now")
	return x.next.Now(ctx)
}

// RecordRosterGroupsV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordRosterGroupsV1(next GroupsV1) GroupsV1 {
	return recordedrosterGroupsV1{next: next, calls: r.calls}
}

type recordedrosterGroupsV1 struct {
	next  GroupsV1
	calls *[]string
}

func (x recordedrosterGroupsV1) Delete(ctx context.Context, id string) error {
	*x.calls = append(*x.calls, "groups/port/Groups@v1.Delete")
	return x.next.Delete(ctx, id)
}

func (x recordedrosterGroupsV1) Get(ctx context.Context, id string) (team.Group, bool, error) {
	*x.calls = append(*x.calls, "groups/port/Groups@v1.Get")
	return x.next.Get(ctx, id)
}

func (x recordedrosterGroupsV1) NewID(ctx context.Context) (string, error) {
	*x.calls = append(*x.calls, "groups/port/Groups@v1.NewID")
	return x.next.NewID(ctx)
}

func (x recordedrosterGroupsV1) Of(ctx context.Context, user string) ([]team.Group, error) {
	*x.calls = append(*x.calls, "groups/port/Groups@v1.Of")
	return x.next.Of(ctx, user)
}

func (x recordedrosterGroupsV1) Save(ctx context.Context, g team.Group) error {
	*x.calls = append(*x.calls, "groups/port/Groups@v1.Save")
	return x.next.Save(ctx, g)
}

// RecordRosterInvitationsV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordRosterInvitationsV1(next InvitationsV1) InvitationsV1 {
	return recordedrosterInvitationsV1{next: next, calls: r.calls}
}

type recordedrosterInvitationsV1 struct {
	next  InvitationsV1
	calls *[]string
}

func (x recordedrosterInvitationsV1) Delete(ctx context.Context, id string) error {
	*x.calls = append(*x.calls, "groups/port/Invitations@v1.Delete")
	return x.next.Delete(ctx, id)
}

func (x recordedrosterInvitationsV1) Get(ctx context.Context, id string) (team.Invitation, bool, error) {
	*x.calls = append(*x.calls, "groups/port/Invitations@v1.Get")
	return x.next.Get(ctx, id)
}

func (x recordedrosterInvitationsV1) NewID(ctx context.Context) (string, error) {
	*x.calls = append(*x.calls, "groups/port/Invitations@v1.NewID")
	return x.next.NewID(ctx)
}

func (x recordedrosterInvitationsV1) Of(ctx context.Context, user string) ([]team.Invitation, error) {
	*x.calls = append(*x.calls, "groups/port/Invitations@v1.Of")
	return x.next.Of(ctx, user)
}

func (x recordedrosterInvitationsV1) OfGroup(ctx context.Context, group string) ([]team.Invitation, error) {
	*x.calls = append(*x.calls, "groups/port/Invitations@v1.OfGroup")
	return x.next.OfGroup(ctx, group)
}

func (x recordedrosterInvitationsV1) Save(ctx context.Context, inv team.Invitation) error {
	*x.calls = append(*x.calls, "groups/port/Invitations@v1.Save")
	return x.next.Save(ctx, inv)
}

// RecordRosterNoticesV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordRosterNoticesV1(next NoticesV1) NoticesV1 {
	return recordedrosterNoticesV1{next: next, calls: r.calls}
}

type recordedrosterNoticesV1 struct {
	next  NoticesV1
	calls *[]string
}

func (x recordedrosterNoticesV1) Tell(ctx context.Context, kind string, g team.Group, actor Person, to []string) error {
	*x.calls = append(*x.calls, "groups/port/Notices@v1.Tell")
	return x.next.Tell(ctx, kind, g, actor, to)
}

// RecordRosterPeopleV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordRosterPeopleV1(next PeopleV1) PeopleV1 {
	return recordedrosterPeopleV1{next: next, calls: r.calls}
}

type recordedrosterPeopleV1 struct {
	next  PeopleV1
	calls *[]string
}

func (x recordedrosterPeopleV1) ByID(ctx context.Context, id string) (Person, bool, error) {
	*x.calls = append(*x.calls, "groups/port/People@v1.ByID")
	return x.next.ByID(ctx, id)
}

// TestSequenceInviteContact is the order of sequence invite-contact: the diagram's links, in the
// diagram's order (D21).
func TestSequenceInviteContact(t *testing.T) {
	var got []string
	run := givenSequenceInviteContact(t, sequenceRecorder{calls: &got})
	if err := run(t.Context()); err != nil {
		t.Fatalf("the sequence failed: %v", err)
	}
	want := []string{"groups/port/Groups@v1.Get", "groups/port/Circle@v1.AreContacts", "groups/port/Invitations@v1.OfGroup", "groups/port/Clock@v1.Now", "groups/port/Invitations@v1.NewID", "groups/port/Invitations@v1.Save", "groups/port/People@v1.ByID", "groups/port/People@v1.ByID", "groups/port/Notices@v1.Tell"}
	if !slices.Equal(got, want) {
		t.Errorf("the calls: %q, want %q", got, want)
	}
}
