// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:ad3708a6c0120e4ecb328f05fe017711e138b2d9494fff3cf0c4cd1c6596edae

// Package roster — Creating and running a group, inviting a contact into it,
// answering an invitation, and what the other components ask of one's groups.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package roster

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/groups/team"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 9, 2, 255), "NOT_IMPLEMENTED", "not implemented yet", "roster: a hand-written body is still a skeleton")

// Belonging — A group a member belongs to, and their role in it.
type Belonging struct {
	Group GroupRef  `json:"group"`
	Role  team.Role `json:"role" app:"public"`
}

// GroupRef — How a group is shown elsewhere.
type GroupRef struct {
	ID    string `json:"id" app:"public"`
	Name  string `json:"name" app:"public"`
	Color string `json:"color" app:"public"`
}

// GroupView — A group as one of its members sees it, the members ranked.
type GroupView struct {
	ID    string `json:"id" app:"public"`
	Name  string `json:"name" app:"public"`
	Color string `json:"color" app:"public"`
	// Role is the one of the member who looks.
	Role      team.Role    `json:"role" app:"public"`
	Members   []MemberView `json:"members"`
	CreatedAt time.Time    `json:"createdAt" app:"public"`
	UpdatedAt time.Time    `json:"updatedAt" app:"public"`
}

// InvitationView — An invitation into a group, as the inviter and the
// invitee read it.
type InvitationView struct {
	ID      string   `json:"id" app:"public"`
	Group   GroupRef `json:"group"`
	Inviter Person   `json:"inviter"`
	// Invitee is who is invited: the inviter's answer to an invitation holds it.
	Invitee   Person             `json:"invitee,omitzero"`
	State     team.InvitingState `json:"state" app:"public"`
	CreatedAt time.Time          `json:"createdAt" app:"public"`
}

// MemberView — A member as the group's page shows them.
type MemberView struct {
	User     Person    `json:"user"`
	Role     team.Role `json:"role" app:"public"`
	JoinedAt time.Time `json:"joinedAt" app:"public"`
}

// PendingInvitation — An invitation a group is waiting on, as its owner and
// admins read it, without any address.
type PendingInvitation struct {
	ID string `json:"id" app:"public"`
	// User is who is invited.
	User      Person    `json:"user"`
	Inviter   Person    `json:"inviter"`
	CreatedAt time.Time `json:"createdAt" app:"public"`
}

// Person — A member as a group shows them.
type Person struct {
	ID   string `json:"id" app:"public"`
	Name string `json:"name" app:"personal"`
}

// CircleV1 knows who is a contact of whom.
type CircleV1 interface {
	AreContacts(ctx context.Context, one string, other string) (bool, error)
}

// ClockV1 tells the time.
type ClockV1 interface {
	Now(ctx context.Context) (time.Time, error)
}

// GroupsV1 keeps the groups.
type GroupsV1 interface {
	// Change applies a change to the group of an id as it is kept at that instant,
	// alone among the writers of the store, and keeps the result; an error of the
	// change keeps nothing, and a group that is not there is ErrNotFound.
	Change(ctx context.Context, id string, change func(*team.Group) error) (team.Group, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (team.Group, bool, error)
	NewID(ctx context.Context) (string, error)
	Of(ctx context.Context, user string) ([]team.Group, error)
	Save(ctx context.Context, g team.Group) error
}

// InvitationsV1 keeps the invitations.
type InvitationsV1 interface {
	// Change applies a change to the invitation of an id as it is kept at that
	// instant, alone among the writers of the store, and keeps the result; an
	// error of the change keeps nothing, and an invitation that is not there is
	// ErrNoInvitation.
	Change(ctx context.Context, id string, change func(*team.Invitation) error) (team.Invitation, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (team.Invitation, bool, error)
	NewID(ctx context.Context) (string, error)
	Of(ctx context.Context, user string) ([]team.Invitation, error)
	OfGroup(ctx context.Context, group string) ([]team.Invitation, error)
	Save(ctx context.Context, inv team.Invitation) error
}

// NoticesV1 is told what happened in a group, and to whom, for whoever tells
// the members.
type NoticesV1 interface {
	Tell(ctx context.Context, kind string, g team.Group, actor Person, to []string) error
}

// PeopleV1 is the directory of the members.
type PeopleV1 interface {
	ByID(ctx context.Context, id string) (Person, bool, error)
}

// AnswerInvitation moves a pending invitation by one event, accept or decline,
// for its invitee; accepting joins the group as a member.
type AnswerInvitation interface {
	Execute(ctx context.Context, in AnswerInvitationInput) (AnswerInvitationOutput, error)
}

// AnswerInvitationInput is AnswerInvitation's input: a field added breaks no caller.
type AnswerInvitationInput struct {
	User  string
	ID    string
	Event string
}

// AnswerInvitationOutput is AnswerInvitation's output.
type AnswerInvitationOutput struct {
	Group GroupView
}

// AnswerInvitationHandler runs AnswerInvitation: its body is the hand-written execute.
type AnswerInvitationHandler struct {
	clock       ClockV1
	invitations InvitationsV1
	groups      GroupsV1
	people      PeopleV1
	notices     NoticesV1
}

// NewAnswerInvitation builds AnswerInvitation on what it calls.
func NewAnswerInvitation(clock ClockV1, invitations InvitationsV1, groups GroupsV1, people PeopleV1, notices NoticesV1) *AnswerInvitationHandler {
	return &AnswerInvitationHandler{clock: clock, invitations: invitations, groups: groups, people: people, notices: notices}
}

// Execute runs AnswerInvitation.
func (h *AnswerInvitationHandler) Execute(ctx context.Context, in AnswerInvitationInput) (AnswerInvitationOutput, error) {
	return h.execute(ctx, &in)
}

var _ AnswerInvitation = (*AnswerInvitationHandler)(nil)

// CreateGroup starts a group; whoever creates it is its owner.
type CreateGroup interface {
	Execute(ctx context.Context, in CreateGroupInput) (CreateGroupOutput, error)
}

// CreateGroupInput is CreateGroup's input: a field added breaks no caller.
type CreateGroupInput struct {
	User  string
	Name  string
	Color string
}

// CreateGroupOutput is CreateGroup's output.
type CreateGroupOutput struct {
	Group GroupView
}

// CreateGroupHandler runs CreateGroup: its body is the hand-written execute.
type CreateGroupHandler struct {
	clock  ClockV1
	groups GroupsV1
	people PeopleV1
}

// NewCreateGroup builds CreateGroup on what it calls.
func NewCreateGroup(clock ClockV1, groups GroupsV1, people PeopleV1) *CreateGroupHandler {
	return &CreateGroupHandler{clock: clock, groups: groups, people: people}
}

// Execute runs CreateGroup.
func (h *CreateGroupHandler) Execute(ctx context.Context, in CreateGroupInput) (CreateGroupOutput, error) {
	return h.execute(ctx, &in)
}

var _ CreateGroup = (*CreateGroupHandler)(nil)

// DeleteGroup deletes a group and its invitations; only its owner may. Its
// members are told.
type DeleteGroup interface {
	Execute(ctx context.Context, in DeleteGroupInput) (DeleteGroupOutput, error)
}

// DeleteGroupInput is DeleteGroup's input: a field added breaks no caller.
type DeleteGroupInput struct {
	User string
	ID   string
}

// DeleteGroupOutput is DeleteGroup's output.
type DeleteGroupOutput struct {
}

// DeleteGroupHandler runs DeleteGroup: its body is the hand-written execute.
type DeleteGroupHandler struct {
	groups      GroupsV1
	invitations InvitationsV1
	people      PeopleV1
	notices     NoticesV1
}

// NewDeleteGroup builds DeleteGroup on what it calls.
func NewDeleteGroup(groups GroupsV1, invitations InvitationsV1, people PeopleV1, notices NoticesV1) *DeleteGroupHandler {
	return &DeleteGroupHandler{groups: groups, invitations: invitations, people: people, notices: notices}
}

// Execute runs DeleteGroup.
func (h *DeleteGroupHandler) Execute(ctx context.Context, in DeleteGroupInput) (DeleteGroupOutput, error) {
	return h.execute(ctx, &in)
}

var _ DeleteGroup = (*DeleteGroupHandler)(nil)

// EditGroup renames a group or changes its color, each only when it is sent;
// its owner and admins may.
type EditGroup interface {
	Execute(ctx context.Context, in EditGroupInput) (EditGroupOutput, error)
}

// EditGroupInput is EditGroup's input: a field added breaks no caller.
type EditGroupInput struct {
	User  string
	ID    string
	Name  *string
	Color *string
}

// EditGroupOutput is EditGroup's output.
type EditGroupOutput struct {
	Group GroupView
}

// EditGroupHandler runs EditGroup: its body is the hand-written execute.
type EditGroupHandler struct {
	clock  ClockV1
	groups GroupsV1
	people PeopleV1
}

// NewEditGroup builds EditGroup on what it calls.
func NewEditGroup(clock ClockV1, groups GroupsV1, people PeopleV1) *EditGroupHandler {
	return &EditGroupHandler{clock: clock, groups: groups, people: people}
}

// Execute runs EditGroup.
func (h *EditGroupHandler) Execute(ctx context.Context, in EditGroupInput) (EditGroupOutput, error) {
	return h.execute(ctx, &in)
}

var _ EditGroup = (*EditGroupHandler)(nil)

// GetGroup shows one of a member's groups; a group they are not in does not
// exist for them.
type GetGroup interface {
	Execute(ctx context.Context, in GetGroupInput) (GetGroupOutput, error)
}

// GetGroupInput is GetGroup's input: a field added breaks no caller.
type GetGroupInput struct {
	User string
	ID   string
}

// GetGroupOutput is GetGroup's output.
type GetGroupOutput struct {
	Group GroupView
}

// GetGroupHandler runs GetGroup: its body is the hand-written execute.
type GetGroupHandler struct {
	groups GroupsV1
	people PeopleV1
}

// NewGetGroup builds GetGroup on what it calls.
func NewGetGroup(groups GroupsV1, people PeopleV1) *GetGroupHandler {
	return &GetGroupHandler{groups: groups, people: people}
}

// Execute runs GetGroup.
func (h *GetGroupHandler) Execute(ctx context.Context, in GetGroupInput) (GetGroupOutput, error) {
	return h.execute(ctx, &in)
}

var _ GetGroup = (*GetGroupHandler)(nil)

// Invite asks one of the caller's contacts to join a group; its owner and
// admins may. The invitee is told.
type Invite interface {
	Execute(ctx context.Context, in InviteInput) (InviteOutput, error)
}

// InviteInput is Invite's input: a field added breaks no caller.
type InviteInput struct {
	User    string
	ID      string
	Invitee string
}

// InviteOutput is Invite's output.
type InviteOutput struct {
	Invitation InvitationView
}

// InviteHandler runs Invite: its body is the hand-written execute.
type InviteHandler struct {
	groups      GroupsV1
	circle      CircleV1
	invitations InvitationsV1
	clock       ClockV1
	people      PeopleV1
	notices     NoticesV1
}

// NewInvite builds Invite on what it calls.
func NewInvite(groups GroupsV1, circle CircleV1, invitations InvitationsV1, clock ClockV1, people PeopleV1, notices NoticesV1) *InviteHandler {
	return &InviteHandler{groups: groups, circle: circle, invitations: invitations, clock: clock, people: people, notices: notices}
}

// Execute runs Invite.
func (h *InviteHandler) Execute(ctx context.Context, in InviteInput) (InviteOutput, error) {
	return h.execute(ctx, &in)
}

var _ Invite = (*InviteHandler)(nil)

// ListGroupInvitations lists the invitations a group is waiting on, the newest
// first; its owner and admins may ask, as they may invite.
type ListGroupInvitations interface {
	Execute(ctx context.Context, in ListGroupInvitationsInput) (ListGroupInvitationsOutput, error)
}

// ListGroupInvitationsInput is ListGroupInvitations's input: a field added breaks no caller.
type ListGroupInvitationsInput struct {
	User string
	ID   string
}

// ListGroupInvitationsOutput is ListGroupInvitations's output.
type ListGroupInvitationsOutput struct {
	Invitations []PendingInvitation
}

// ListGroupInvitationsHandler runs ListGroupInvitations: its body is the hand-written execute.
type ListGroupInvitationsHandler struct {
	groups      GroupsV1
	invitations InvitationsV1
	people      PeopleV1
}

// NewListGroupInvitations builds ListGroupInvitations on what it calls.
func NewListGroupInvitations(groups GroupsV1, invitations InvitationsV1, people PeopleV1) *ListGroupInvitationsHandler {
	return &ListGroupInvitationsHandler{groups: groups, invitations: invitations, people: people}
}

// Execute runs ListGroupInvitations.
func (h *ListGroupInvitationsHandler) Execute(ctx context.Context, in ListGroupInvitationsInput) (ListGroupInvitationsOutput, error) {
	return h.execute(ctx, &in)
}

var _ ListGroupInvitations = (*ListGroupInvitationsHandler)(nil)

// ListGroups lists a member's groups by name.
type ListGroups interface {
	Execute(ctx context.Context, in ListGroupsInput) (ListGroupsOutput, error)
}

// ListGroupsInput is ListGroups's input: a field added breaks no caller.
type ListGroupsInput struct {
	User string
}

// ListGroupsOutput is ListGroups's output.
type ListGroupsOutput struct {
	Groups []GroupView
}

// ListGroupsHandler runs ListGroups: its body is the hand-written execute.
type ListGroupsHandler struct {
	groups GroupsV1
	people PeopleV1
}

// NewListGroups builds ListGroups on what it calls.
func NewListGroups(groups GroupsV1, people PeopleV1) *ListGroupsHandler {
	return &ListGroupsHandler{groups: groups, people: people}
}

// Execute runs ListGroups.
func (h *ListGroupsHandler) Execute(ctx context.Context, in ListGroupsInput) (ListGroupsOutput, error) {
	return h.execute(ctx, &in)
}

var _ ListGroups = (*ListGroupsHandler)(nil)

// ListInvitations lists the invitations waiting for a member's answer, the
// newest first.
type ListInvitations interface {
	Execute(ctx context.Context, in ListInvitationsInput) (ListInvitationsOutput, error)
}

// ListInvitationsInput is ListInvitations's input: a field added breaks no caller.
type ListInvitationsInput struct {
	User string
}

// ListInvitationsOutput is ListInvitations's output.
type ListInvitationsOutput struct {
	Invitations []InvitationView
}

// ListInvitationsHandler runs ListInvitations: its body is the hand-written execute.
type ListInvitationsHandler struct {
	invitations InvitationsV1
	groups      GroupsV1
	people      PeopleV1
}

// NewListInvitations builds ListInvitations on what it calls.
func NewListInvitations(invitations InvitationsV1, groups GroupsV1, people PeopleV1) *ListInvitationsHandler {
	return &ListInvitationsHandler{invitations: invitations, groups: groups, people: people}
}

// Execute runs ListInvitations.
func (h *ListInvitationsHandler) Execute(ctx context.Context, in ListInvitationsInput) (ListInvitationsOutput, error) {
	return h.execute(ctx, &in)
}

var _ ListInvitations = (*ListInvitationsHandler)(nil)

// MembersOf lists the members of a group, whoever they are, empty for a group
// that does not exist; the components that tell a group's members ask it.
type MembersOf interface {
	Execute(ctx context.Context, in MembersOfInput) (MembersOfOutput, error)
}

// MembersOfInput is MembersOf's input: a field added breaks no caller.
type MembersOfInput struct {
	Group string
}

// MembersOfOutput is MembersOf's output.
type MembersOfOutput struct {
	Users []string
}

// MembersOfHandler runs MembersOf: its body is the hand-written execute.
type MembersOfHandler struct {
	groups GroupsV1
}

// NewMembersOf builds MembersOf on what it calls.
func NewMembersOf(groups GroupsV1) *MembersOfHandler {
	return &MembersOfHandler{groups: groups}
}

// Execute runs MembersOf.
func (h *MembersOfHandler) Execute(ctx context.Context, in MembersOfInput) (MembersOfOutput, error) {
	return h.execute(ctx, &in)
}

var _ MembersOf = (*MembersOfHandler)(nil)

// Memberships lists the groups a member belongs to, with their role in each;
// the other components ask it.
type Memberships interface {
	Execute(ctx context.Context, in MembershipsInput) (MembershipsOutput, error)
}

// MembershipsInput is Memberships's input: a field added breaks no caller.
type MembershipsInput struct {
	User string
}

// MembershipsOutput is Memberships's output.
type MembershipsOutput struct {
	Memberships []Belonging
}

// MembershipsHandler runs Memberships: its body is the hand-written execute.
type MembershipsHandler struct {
	groups GroupsV1
}

// NewMemberships builds Memberships on what it calls.
func NewMemberships(groups GroupsV1) *MembershipsHandler {
	return &MembershipsHandler{groups: groups}
}

// Execute runs Memberships.
func (h *MembershipsHandler) Execute(ctx context.Context, in MembershipsInput) (MembershipsOutput, error) {
	return h.execute(ctx, &in)
}

var _ Memberships = (*MembershipsHandler)(nil)

// RemoveMember takes a member out of a group. Anyone but the owner may leave;
// the owner removes anyone, an admin removes members.
type RemoveMember interface {
	Execute(ctx context.Context, in RemoveMemberInput) (RemoveMemberOutput, error)
}

// RemoveMemberInput is RemoveMember's input: a field added breaks no caller.
type RemoveMemberInput struct {
	User   string
	ID     string
	Member string
}

// RemoveMemberOutput is RemoveMember's output.
type RemoveMemberOutput struct {
}

// RemoveMemberHandler runs RemoveMember: its body is the hand-written execute.
type RemoveMemberHandler struct {
	clock   ClockV1
	groups  GroupsV1
	people  PeopleV1
	notices NoticesV1
}

// NewRemoveMember builds RemoveMember on what it calls.
func NewRemoveMember(clock ClockV1, groups GroupsV1, people PeopleV1, notices NoticesV1) *RemoveMemberHandler {
	return &RemoveMemberHandler{clock: clock, groups: groups, people: people, notices: notices}
}

// Execute runs RemoveMember.
func (h *RemoveMemberHandler) Execute(ctx context.Context, in RemoveMemberInput) (RemoveMemberOutput, error) {
	return h.execute(ctx, &in)
}

var _ RemoveMember = (*RemoveMemberHandler)(nil)

// RoleOf is the role of a member in a group, empty for a stranger or for a
// group that does not exist; the other components ask it.
type RoleOf interface {
	Execute(ctx context.Context, in RoleOfInput) (RoleOfOutput, error)
}

// RoleOfInput is RoleOf's input: a field added breaks no caller.
type RoleOfInput struct {
	Group string
	User  string
}

// RoleOfOutput is RoleOf's output.
type RoleOfOutput struct {
	Role team.Role
}

// RoleOfHandler runs RoleOf: its body is the hand-written execute.
type RoleOfHandler struct {
	groups GroupsV1
}

// NewRoleOf builds RoleOf on what it calls.
func NewRoleOf(groups GroupsV1) *RoleOfHandler {
	return &RoleOfHandler{groups: groups}
}

// Execute runs RoleOf.
func (h *RoleOfHandler) Execute(ctx context.Context, in RoleOfInput) (RoleOfOutput, error) {
	return h.execute(ctx, &in)
}

var _ RoleOf = (*RoleOfHandler)(nil)

// SetRole makes a member an admin, or an admin a member; only the owner may,
// and never of themselves.
type SetRole interface {
	Execute(ctx context.Context, in SetRoleInput) (SetRoleOutput, error)
}

// SetRoleInput is SetRole's input: a field added breaks no caller.
type SetRoleInput struct {
	User   string
	ID     string
	Member string
	Role   string
}

// SetRoleOutput is SetRole's output.
type SetRoleOutput struct {
	Group GroupView
}

// SetRoleHandler runs SetRole: its body is the hand-written execute.
type SetRoleHandler struct {
	clock   ClockV1
	groups  GroupsV1
	people  PeopleV1
	notices NoticesV1
}

// NewSetRole builds SetRole on what it calls.
func NewSetRole(clock ClockV1, groups GroupsV1, people PeopleV1, notices NoticesV1) *SetRoleHandler {
	return &SetRoleHandler{clock: clock, groups: groups, people: people, notices: notices}
}

// Execute runs SetRole.
func (h *SetRoleHandler) Execute(ctx context.Context, in SetRoleInput) (SetRoleOutput, error) {
	return h.execute(ctx, &in)
}

var _ SetRole = (*SetRoleHandler)(nil)

// WithdrawInvitation takes back an invitation a group is waiting on; its owner
// and admins may. The invitation is answered once, whoever the invitee or the
// manager is, so a withdrawal and an acceptance at the same instant leave one
// winner.
type WithdrawInvitation interface {
	Execute(ctx context.Context, in WithdrawInvitationInput) (WithdrawInvitationOutput, error)
}

// WithdrawInvitationInput is WithdrawInvitation's input: a field added breaks no caller.
type WithdrawInvitationInput struct {
	User         string
	ID           string
	InvitationID string
}

// WithdrawInvitationOutput is WithdrawInvitation's output.
type WithdrawInvitationOutput struct {
}

// WithdrawInvitationHandler runs WithdrawInvitation: its body is the hand-written execute.
type WithdrawInvitationHandler struct {
	groups      GroupsV1
	invitations InvitationsV1
}

// NewWithdrawInvitation builds WithdrawInvitation on what it calls.
func NewWithdrawInvitation(groups GroupsV1, invitations InvitationsV1) *WithdrawInvitationHandler {
	return &WithdrawInvitationHandler{groups: groups, invitations: invitations}
}

// Execute runs WithdrawInvitation.
func (h *WithdrawInvitationHandler) Execute(ctx context.Context, in WithdrawInvitationInput) (WithdrawInvitationOutput, error) {
	return h.execute(ctx, &in)
}

var _ WithdrawInvitation = (*WithdrawInvitationHandler)(nil)
