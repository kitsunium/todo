// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:d9e6eb32619ad64ff259ee580b4b28ea598fe9a15bc77181628dbbd887d57f92

// Package groupsapi — The HTTP API of a member's groups and invitations, and
// the questions the rest of the product asks of them.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package groupsapi

import (
	"context"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/groups/roster"
	"github.com/kitsunium/todo/groups/team"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 9, 4, 255), "NOT_IMPLEMENTED", "not implemented yet", "groupsapi: a hand-written body is still a skeleton")

// Belongings — The groups a member belongs to.
type Belongings struct {
	Memberships []roster.Belonging `json:"memberships"`
}

// ByGroup — The group a route names.
type ByGroup struct {
	ID string `path:"id"`
}

// ByInvitation — The invitation a route names.
type ByInvitation struct {
	ID string `path:"id"`
}

// ByMember — A member of a group a route names.
type ByMember struct {
	ID     string `path:"id"`
	UserID string `path:"userId"`
}

// CreateRequest — The group to start; no color picks one from the name.
type CreateRequest struct {
	Name  string `json:"name" app:"public"`
	Color string `json:"color" app:"public"`
}

// EditRequest — What a member changes of a group, each field only when it is
// sent.
type EditRequest struct {
	ID    string  `path:"id"`
	Name  *string `json:"name" app:"public"`
	Color *string `json:"color" app:"public"`
}

// GroupList — A member's groups.
type GroupList struct {
	Groups []roster.GroupView `json:"groups"`
}

// InvitationList — The invitations waiting for a member.
type InvitationList struct {
	Invitations []roster.InvitationView `json:"invitations"`
}

// InviteRequest — The contact a group invites.
type InviteRequest struct {
	ID     string `path:"id"`
	UserID string `json:"userId" app:"public"`
}

// Members — The members of a group.
type Members struct {
	Users []string `json:"users" app:"public"`
}

// NoInput — A request, or an answer, that says nothing.
type NoInput struct {
}

// RoleAnswer — The role of a member in a group; empty when they are not in
// it.
type RoleAnswer struct {
	Role team.Role `json:"role" app:"public"`
}

// RoleAsked — A group and a member.
type RoleAsked struct {
	Group string `json:"group" app:"public"`
	User  string `json:"user" app:"public"`
}

// RoleRequest — The role a member of a group is given.
type RoleRequest struct {
	ID     string `path:"id"`
	UserID string `path:"userId"`
	Role   string `json:"role" app:"public"`
}

// WhichGroup — A group whose members are asked.
type WhichGroup struct {
	Group string `json:"group" app:"public"`
}

// WhoseGroups — A member whose groups are asked.
type WhoseGroups struct {
	User string `json:"user" app:"public"`
}

// Service is groupsapi's framework service: what it declares, drawn and run.
var Service = app.NewService("groupsapi", "The HTTP API of a member's groups and invitations, and the questions the rest of the product asks of them.")

var _ = Service.Endpoint("GET /api/groups", List, app.Auth())

// List serves GET /api/groups: its body is the hand-written list.
func List(ctx context.Context, in NoInput) (GroupList, error) {
	return list(ctx, in)
}

var _ = Service.Endpoint("POST /api/groups", Create, app.Auth(), app.RateLimitPerClient(10, 20))

// Create serves POST /api/groups: its body is the hand-written create.
func Create(ctx context.Context, in CreateRequest) (roster.GroupView, error) {
	return create(ctx, in)
}

var _ = Service.Endpoint("GET /api/groups/{id}", Get, app.Auth())

// Get serves GET /api/groups/{id}: its body is the hand-written get.
func Get(ctx context.Context, in ByGroup) (roster.GroupView, error) {
	return get(ctx, in)
}

var _ = Service.Endpoint("PATCH /api/groups/{id}", Edit, app.Auth(), app.RateLimitPerClient(10, 20))

// Edit serves PATCH /api/groups/{id}: its body is the hand-written edit.
func Edit(ctx context.Context, in EditRequest) (roster.GroupView, error) {
	return edit(ctx, in)
}

var _ = Service.Endpoint("DELETE /api/groups/{id}", Delete, app.Auth(), app.RateLimitPerClient(10, 20))

// Delete serves DELETE /api/groups/{id}: its body is the hand-written deleteBody.
func Delete(ctx context.Context, in ByGroup) (NoInput, error) {
	return deleteBody(ctx, in)
}

var _ = Service.Endpoint("POST /api/groups/{id}/invitations", Invite, app.Auth(), app.RateLimitPerClient(10, 20))

// Invite serves POST /api/groups/{id}/invitations: its body is the hand-written invite.
func Invite(ctx context.Context, in InviteRequest) (roster.InvitationView, error) {
	return invite(ctx, in)
}

var _ = Service.Endpoint("GET /api/invitations", Invitations, app.Auth())

// Invitations serves GET /api/invitations: its body is the hand-written invitations.
func Invitations(ctx context.Context, in NoInput) (InvitationList, error) {
	return invitations(ctx, in)
}

var _ = Service.Endpoint("POST /api/invitations/{id}/accept", Accept, app.Auth(), app.RateLimitPerClient(10, 20))

// Accept serves POST /api/invitations/{id}/accept: its body is the hand-written accept.
func Accept(ctx context.Context, in ByInvitation) (roster.GroupView, error) {
	return accept(ctx, in)
}

var _ = Service.Endpoint("POST /api/invitations/{id}/decline", Decline, app.Auth(), app.RateLimitPerClient(10, 20))

// Decline serves POST /api/invitations/{id}/decline: its body is the hand-written decline.
func Decline(ctx context.Context, in ByInvitation) (NoInput, error) {
	return decline(ctx, in)
}

var _ = Service.Endpoint("DELETE /api/groups/{id}/members/{userId}", RemoveMember, app.Auth(), app.RateLimitPerClient(10, 20))

// RemoveMember serves DELETE /api/groups/{id}/members/{userId}: its body is the hand-written removeMember.
func RemoveMember(ctx context.Context, in ByMember) (NoInput, error) {
	return removeMember(ctx, in)
}

var _ = Service.Endpoint("PATCH /api/groups/{id}/members/{userId}", SetRole, app.Auth(), app.RateLimitPerClient(10, 20))

// SetRole serves PATCH /api/groups/{id}/members/{userId}: its body is the hand-written setRole.
func SetRole(ctx context.Context, in RoleRequest) (roster.GroupView, error) {
	return setRole(ctx, in)
}

// RoleQuery — The role of a member in a group, empty when they are not in
// it, for the components that decide who sees what.
var RoleQuery = Service.Query("role", AskRole)

// AskRole handles the query "role": its body is the hand-written askRole.
func AskRole(ctx context.Context, in RoleAsked) (RoleAnswer, error) {
	return askRole(ctx, in)
}

// MembershipsQuery — The groups a member belongs to and their role in each,
// for the components that follow them.
var MembershipsQuery = Service.Query("memberships", AskMemberships)

// AskMemberships handles the query "memberships": its body is the hand-written askMemberships.
func AskMemberships(ctx context.Context, in WhoseGroups) (Belongings, error) {
	return askMemberships(ctx, in)
}

// MembersQuery — The members of a group, for the components that tell them
// what happens to the group's tasks.
var MembersQuery = Service.Query("members", AskMembers)

// AskMembers handles the query "members": its body is the hand-written askMembers.
func AskMembers(ctx context.Context, in WhichGroup) (Members, error) {
	return askMembers(ctx, in)
}

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
	listGroups       roster.ListGroups
	createGroup      roster.CreateGroup
	getGroup         roster.GetGroup
	editGroup        roster.EditGroup
	deleteGroup      roster.DeleteGroup
	invite           roster.Invite
	listInvitations  roster.ListInvitations
	answerInvitation roster.AnswerInvitation
	removeMember     roster.RemoveMember
	setRole          roster.SetRole
	roleOf           roster.RoleOf
	memberships      roster.Memberships
	membersOf        roster.MembersOf
}

// Wire gives the handlers what they use: the role that mounts groupsapi calls it
// before its App runs.
func Wire(listGroups roster.ListGroups, createGroup roster.CreateGroup, getGroup roster.GetGroup, editGroup roster.EditGroup, deleteGroup roster.DeleteGroup, invite roster.Invite, listInvitations roster.ListInvitations, answerInvitation roster.AnswerInvitation, removeMember roster.RemoveMember, setRole roster.SetRole, roleOf roster.RoleOf, memberships roster.Memberships, membersOf roster.MembersOf) {
	wired.listGroups = listGroups
	wired.createGroup = createGroup
	wired.getGroup = getGroup
	wired.editGroup = editGroup
	wired.deleteGroup = deleteGroup
	wired.invite = invite
	wired.listInvitations = listInvitations
	wired.answerInvitation = answerInvitation
	wired.removeMember = removeMember
	wired.setRole = setRole
	wired.roleOf = roleOf
	wired.memberships = memberships
	wired.membersOf = membersOf
}
