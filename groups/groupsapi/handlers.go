// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package groupsapi

import (
	"context"
	"errors"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/groups/roster"
	"github.com/kitsunium/todo/groups/team"
)

// list is GET /api/groups's body; what it uses is in wired.
func list(ctx context.Context, in NoInput) (out GroupList, err error) {
	me, err := caller(ctx)
	if err != nil {
		return GroupList{}, err
	}
	res, err := wired.listGroups.Execute(ctx, roster.ListGroupsInput{User: me})
	if err != nil {
		return GroupList{}, refusal(err)
	}
	return GroupList{Groups: res.Groups}, nil
}

// create is POST /api/groups's body; what it uses is in wired.
func create(ctx context.Context, in CreateRequest) (out roster.GroupView, err error) {
	me, err := caller(ctx)
	if err != nil {
		return roster.GroupView{}, err
	}
	res, err := wired.createGroup.Execute(ctx, roster.CreateGroupInput{User: me, Name: in.Name, Color: in.Color})
	if err != nil {
		return roster.GroupView{}, refusal(err)
	}
	return res.Group, nil
}

// get is GET /api/groups/{id}'s body; what it uses is in wired.
func get(ctx context.Context, in ByGroup) (out roster.GroupView, err error) {
	me, err := caller(ctx)
	if err != nil {
		return roster.GroupView{}, err
	}
	res, err := wired.getGroup.Execute(ctx, roster.GetGroupInput{User: me, ID: in.ID})
	if err != nil {
		return roster.GroupView{}, refusal(err)
	}
	return res.Group, nil
}

// edit is PATCH /api/groups/{id}'s body; what it uses is in wired.
func edit(ctx context.Context, in EditRequest) (out roster.GroupView, err error) {
	me, err := caller(ctx)
	if err != nil {
		return roster.GroupView{}, err
	}
	res, err := wired.editGroup.Execute(ctx, roster.EditGroupInput{User: me, ID: in.ID, Name: in.Name, Color: in.Color})
	if err != nil {
		return roster.GroupView{}, refusal(err)
	}
	return res.Group, nil
}

// deleteBody is DELETE /api/groups/{id}'s body; what it uses is in wired.
func deleteBody(ctx context.Context, in ByGroup) (out NoInput, err error) {
	me, err := caller(ctx)
	if err != nil {
		return NoInput{}, err
	}
	if _, err := wired.deleteGroup.Execute(ctx, roster.DeleteGroupInput{User: me, ID: in.ID}); err != nil {
		return NoInput{}, refusal(err)
	}
	return NoInput{}, nil
}

// invite is POST /api/groups/{id}/invitations's body; what it uses is in wired.
func invite(ctx context.Context, in InviteRequest) (out roster.InvitationView, err error) {
	me, err := caller(ctx)
	if err != nil {
		return roster.InvitationView{}, err
	}
	if in.UserID == "" {
		return roster.InvitationView{}, invalid("userId", "required", "is required")
	}
	res, err := wired.invite.Execute(ctx, roster.InviteInput{User: me, ID: in.ID, Invitee: in.UserID})
	if err != nil {
		return roster.InvitationView{}, refusal(err)
	}
	return res.Invitation, nil
}

// invitations is GET /api/invitations's body; what it uses is in wired.
func invitations(ctx context.Context, in NoInput) (out InvitationList, err error) {
	me, err := caller(ctx)
	if err != nil {
		return InvitationList{}, err
	}
	res, err := wired.listInvitations.Execute(ctx, roster.ListInvitationsInput{User: me})
	if err != nil {
		return InvitationList{}, refusal(err)
	}
	return InvitationList{Invitations: res.Invitations}, nil
}

// accept is POST /api/invitations/{id}/accept's body; what it uses is in wired.
func accept(ctx context.Context, in ByInvitation) (out roster.GroupView, err error) {
	return answer(ctx, in.ID, team.InvitingOnAccept)
}

// decline is POST /api/invitations/{id}/decline's body; what it uses is in wired.
func decline(ctx context.Context, in ByInvitation) (out NoInput, err error) {
	_, err = answer(ctx, in.ID, team.InvitingOnDecline)
	return NoInput{}, err
}

// removeMember is DELETE /api/groups/{id}/members/{userId}'s body; what it uses is in wired.
func removeMember(ctx context.Context, in ByMember) (out NoInput, err error) {
	me, err := caller(ctx)
	if err != nil {
		return NoInput{}, err
	}
	if _, err := wired.removeMember.Execute(ctx, roster.RemoveMemberInput{User: me, ID: in.ID, Member: in.UserID}); err != nil {
		return NoInput{}, refusal(err)
	}
	return NoInput{}, nil
}

// setRole is PATCH /api/groups/{id}/members/{userId}'s body; what it uses is in wired.
func setRole(ctx context.Context, in RoleRequest) (out roster.GroupView, err error) {
	me, err := caller(ctx)
	if err != nil {
		return roster.GroupView{}, err
	}
	res, err := wired.setRole.Execute(ctx, roster.SetRoleInput{User: me, ID: in.ID, Member: in.UserID, Role: in.Role})
	if err != nil {
		return roster.GroupView{}, refusal(err)
	}
	return res.Group, nil
}

// askRole is the query "role"'s body.
func askRole(ctx context.Context, in RoleAsked) (out RoleAnswer, err error) {
	res, err := wired.roleOf.Execute(ctx, roster.RoleOfInput{Group: in.Group, User: in.User})
	return RoleAnswer{Role: res.Role}, err
}

// askMemberships is the query "memberships"'s body.
func askMemberships(ctx context.Context, in WhoseGroups) (out Belongings, err error) {
	res, err := wired.memberships.Execute(ctx, roster.MembershipsInput{User: in.User})
	return Belongings{Memberships: res.Memberships}, err
}

// askMembers is the query "members"'s body.
func askMembers(ctx context.Context, in WhichGroup) (out Members, err error) {
	res, err := wired.membersOf.Execute(ctx, roster.MembersOfInput{Group: in.Group})
	return Members{Users: res.Users}, err
}

// answer runs AnswerInvitation for one of the two routes that answer an invitation.
func answer(ctx context.Context, id string, event team.InvitingEvent) (roster.GroupView, error) {
	me, err := caller(ctx)
	if err != nil {
		return roster.GroupView{}, err
	}
	res, err := wired.answerInvitation.Execute(ctx, roster.AnswerInvitationInput{User: me, ID: id, Event: string(event)})
	if err != nil {
		return roster.GroupView{}, refusal(err)
	}
	return res.Group, nil
}

// caller is the member the request comes from; every route asks for one.
func caller(ctx context.Context) (string, error) {
	uid, ok := app.UserID(ctx)
	if !ok {
		return "", app.Unauthenticated("sign in first")
	}
	return string(uid), nil
}

// invalid is a 400 that names the field at fault.
func invalid(path, rule, message string) error {
	e := app.Invalid("the request is invalid")
	e.Violations = []app.Violation{{Path: path, Rule: rule, Message: message}}
	return e
}

// refusal is a use case's error as the API answers it: 404 for a group, an
// invitation or a member that is not the caller's, 403 for what a role does
// not allow, 409 for what the state of the group or the invitation does not
// allow, 400 with the field at fault for a name, a color or a role that is
// not valid.
func refusal(err error) error {
	switch {
	case errors.Is(err, team.ErrNotFound):
		return app.NotFound("no such group")
	case errors.Is(err, team.ErrNoInvitation):
		return app.NotFound("no such invitation")
	case errors.Is(err, team.ErrNoMember):
		return app.NotFound("no such member")
	case errors.Is(err, team.ErrNotManager):
		return app.Forbidden("only the group's owner and admins can do this")
	case errors.Is(err, team.ErrNotOwner):
		return app.Forbidden("only the group's owner can do this")
	case errors.Is(err, team.ErrNotContact):
		return app.Forbidden("you can only invite your contacts")
	case errors.Is(err, team.ErrCannotRemove):
		return app.Forbidden("you cannot remove this member")
	case errors.Is(err, team.ErrOwnerRole):
		return app.Forbidden("the owner's role cannot change")
	case errors.Is(err, team.ErrAlreadyMember):
		return app.Conflict("this person is already a member")
	case errors.Is(err, team.ErrAlreadyInvited):
		return app.Conflict("this person is already invited")
	case errors.Is(err, team.ErrOwnerStays):
		return app.Conflict("the owner cannot leave the group: delete it instead")
	case errors.Is(err, team.ErrNotPending):
		return app.Conflict("this invitation is not waiting for an answer any more")
	case errors.Is(err, team.ErrBadName):
		return invalid("name", "line", "must be one line of one to 60 characters")
	case errors.Is(err, team.ErrBadColor):
		return invalid("color", "oneof", "must be one of the group colors")
	case errors.Is(err, team.ErrBadRole):
		return invalid("role", "oneof", "must be admin or member")
	}
	return err
}

// groupInvitations is GET /api/groups/{id}/invitations's body; what it uses
// is in wired. A non-member is answered as for GET /api/groups/{id}, a plain
// member as for the invite route.
func groupInvitations(ctx context.Context, in ByGroup) (out PendingList, err error) {
	me, err := caller(ctx)
	if err != nil {
		return PendingList{}, err
	}
	res, err := wired.listGroupInvitations.Execute(ctx, roster.ListGroupInvitationsInput{User: me, ID: in.ID})
	if err != nil {
		return PendingList{}, refusal(err)
	}
	return PendingList{Invitations: res.Invitations}, nil
}

// withdraw is DELETE /api/groups/{id}/invitations/{invitationId}'s body; what
// it uses is in wired.
func withdraw(ctx context.Context, in ByGroupInvitation) (out NoInput, err error) {
	me, err := caller(ctx)
	if err != nil {
		return NoInput{}, err
	}
	if _, err := wired.withdrawInvitation.Execute(ctx, roster.WithdrawInvitationInput{User: me, ID: in.ID, InvitationID: in.InvitationID}); err != nil {
		return NoInput{}, refusal(err)
	}
	return NoInput{}, nil
}
