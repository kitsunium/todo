package main

import (
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"
)

// contactList is GET /api/contacts.
type contactList struct {
	Contacts []struct {
		ID    string
		User  user
		Since time.Time
	}
	Incoming, Outgoing []struct {
		ID     string
		User   user
		Status string
	}
	Invites []struct{ ID, Email, Status string }
}

// entry is one line of a feed.
type entry struct {
	ID, Kind, Text string
	Read           bool
	Actor          *user
	Task           *struct{ ID, Title string }
	Group          *struct{ ID, Name, Color string }
}

// feed is GET /api/activity.
type feed struct{ Entries []entry }

// hasEntry waits for an entry of kind whose text is text in c's feed.
func (c *client) hasEntry(kind, text string) {
	c.h.t.Helper()
	c.h.settle(c.name+"'s feed to say "+text, 50*time.Millisecond, func() bool {
		return slices.ContainsFunc(call[feed](c, http.StatusOK, "GET", "/api/activity", nil).Entries, func(e entry) bool {
			return e.Kind == kind && e.Text == text
		})
	})
}

// Two users become contacts when one asks and the other accepts; someone
// without an account is invited by mail, and the invitation turns into a
// request once they sign up and verify their address.
func TestContactsAndInvitations(t *testing.T) {
	h := start(t, false)
	alice := h.signup("Alice", "alice@example.com")
	bob := h.signup("Bob", "bob@example.com")

	added := call[struct {
		Kind    string
		Request struct {
			ID     string
			User   user
			Status string
		}
	}](alice, http.StatusOK, "POST", "/api/contacts", map[string]any{"email": "BOB@example.com"})
	if added.Kind != "request" || added.Request.User.Email != "bob@example.com" || added.Request.Status != "pending" {
		t.Fatalf("adding bob: %+v", added)
	}
	alice.fails(http.StatusConflict, "conflict", "POST", "/api/contacts", map[string]any{"email": "bob@example.com"})
	bob.fails(http.StatusConflict, "conflict", "POST", "/api/contacts", map[string]any{"email": "alice@example.com"})
	alice.violates("email", "POST", "/api/contacts", map[string]any{"email": "alice@example.com"})
	alice.violates("email", "POST", "/api/contacts", map[string]any{"email": "nope"})

	if l := call[contactList](bob, http.StatusOK, "GET", "/api/contacts", nil); len(l.Incoming) != 1 || l.Incoming[0].User.Name != "Alice" {
		t.Fatalf("bob's contacts %+v", l)
	}
	if l := call[contactList](alice, http.StatusOK, "GET", "/api/contacts", nil); len(l.Outgoing) != 1 || len(l.Contacts) != 0 {
		t.Fatalf("alice's contacts %+v", l)
	}
	h.mail("bob@example.com", "Alice wants to add you as a contact")
	bob.hasEntry("contact.requested", "Alice wants to add you as a contact.")

	alice.fails(http.StatusForbidden, "permission_denied", "POST", "/api/contacts/"+added.Request.ID+"/accept", nil)
	accepted := call[struct {
		Contact struct {
			ID   string
			User user
		}
	}](bob, http.StatusOK, "POST", "/api/contacts/"+added.Request.ID+"/accept", nil)
	if accepted.Contact.User.Email != "alice@example.com" {
		t.Fatalf("accepted %+v", accepted)
	}
	bob.fails(http.StatusConflict, "conflict", "POST", "/api/contacts/"+added.Request.ID+"/accept", nil)
	for _, c := range []*client{alice, bob} {
		if l := call[contactList](c, http.StatusOK, "GET", "/api/contacts", nil); len(l.Contacts) != 1 || len(l.Incoming)+len(l.Outgoing) != 0 {
			t.Fatalf("%s's contacts %+v", c.name, l)
		}
	}
	h.mail("alice@example.com", "Bob accepted your contact request")
	alice.hasEntry("contact.accepted", "Bob accepted your contact request.")

	// Someone without an account is invited by mail.
	invited := call[struct {
		Kind   string
		Invite struct{ ID, Email, Status string }
	}](alice, http.StatusOK, "POST", "/api/contacts", map[string]any{"email": "dan@example.com"})
	if invited.Kind != "invite" || invited.Invite.Email != "dan@example.com" || invited.Invite.Status != "pending" {
		t.Fatalf("inviting dan: %+v", invited)
	}
	alice.fails(http.StatusConflict, "conflict", "POST", "/api/contacts", map[string]any{"email": "dan@example.com"})
	join := h.mail("dan@example.com", "Alice invited you to Todo")
	if !strings.Contains(join.Text, "http://localhost:4000/signup?email=dan%40example.com") {
		t.Errorf("the invitation links to %q", join.Text)
	}
	dan := h.signup("Dan", "dan@example.com")
	h.settle("the invitation to become a request", 50*time.Millisecond, func() bool {
		return len(call[contactList](dan, http.StatusOK, "GET", "/api/contacts", nil).Incoming) == 1
	})
	incoming := call[contactList](dan, http.StatusOK, "GET", "/api/contacts", nil).Incoming[0]
	if incoming.User.Email != "alice@example.com" {
		t.Fatalf("dan's request comes from %+v", incoming.User)
	}
	if l := call[contactList](alice, http.StatusOK, "GET", "/api/contacts", nil); len(l.Invites) != 0 || len(l.Outgoing) != 1 {
		t.Fatalf("alice's invitation did not become a request: %+v", l)
	}

	// Declined, asked again; cancelled; removed.
	dan.expect(http.StatusNoContent, "POST", "/api/contacts/"+incoming.ID+"/decline", nil)
	if l := call[contactList](alice, http.StatusOK, "GET", "/api/contacts", nil); len(l.Outgoing) != 0 {
		t.Fatalf("a declined request is still out: %+v", l)
	}
	again := call[struct{ Request struct{ ID string } }](alice, http.StatusOK, "POST", "/api/contacts", map[string]any{"email": "dan@example.com"})
	dan.fails(http.StatusForbidden, "permission_denied", "POST", "/api/contacts/"+again.Request.ID+"/cancel", nil)
	alice.expect(http.StatusNoContent, "POST", "/api/contacts/"+again.Request.ID+"/cancel", nil)
	bob.fails(http.StatusNotFound, "not_found", "DELETE", "/api/contacts/"+again.Request.ID, nil)
	bob.expect(http.StatusNoContent, "DELETE", "/api/contacts/"+accepted.Contact.ID, nil)
	if l := call[contactList](alice, http.StatusOK, "GET", "/api/contacts", nil); len(l.Contacts) != 0 {
		t.Fatalf("a removed contact stays: %+v", l)
	}

	// An invitation by email is withdrawn by its sender only.
	erin := call[struct{ Invite struct{ ID string } }](alice, http.StatusOK, "POST", "/api/contacts", map[string]any{"email": "erin@example.com"})
	bob.fails(http.StatusNotFound, "not_found", "POST", "/api/contacts/"+erin.Invite.ID+"/cancel", nil)
	alice.expect(http.StatusNoContent, "POST", "/api/contacts/"+erin.Invite.ID+"/cancel", nil)
}

// group is a group as its members see it.
type group struct {
	ID, Name, Color, Role string
	Members               []struct {
		User user
		Role string
	}
	OpenTasks int
}

// A group: created by its owner, joined by invitation — contacts only —,
// run by its owner and admins, left by its members, deleted by its owner,
// whose tasks then go back to their owners' lists.
func TestGroups(t *testing.T) {
	h := start(t, false)
	alice := h.signup("Alice", "alice@example.com")
	bob := h.signup("Bob", "bob@example.com")
	carol := h.signup("Carol", "carol@example.com")
	befriend(alice, bob, "bob@example.com")
	bobID, carolID := bob.id(), carol.id()

	design := call[group](alice, http.StatusOK, "POST", "/api/groups", map[string]any{"name": "Design", "color": "violet"})
	if design.Role != "owner" || design.Color != "violet" || len(design.Members) != 1 || design.OpenTasks != 0 {
		t.Fatalf("created %+v", design)
	}
	alice.violates("color", "POST", "/api/groups", map[string]any{"name": "Ops", "color": "chartreuse"})
	if picked := call[group](alice, http.StatusOK, "POST", "/api/groups", map[string]any{"name": "Ops"}); picked.Color == "" {
		t.Error("a group without a color got none")
	}

	alice.fails(http.StatusForbidden, "permission_denied", "POST", "/api/groups/"+design.ID+"/invitations", map[string]any{"userId": carolID})
	invitation := call[struct {
		ID      string
		Group   struct{ ID, Name, Color string }
		Inviter user
		Invitee *user
		Status  string
	}](alice, http.StatusOK, "POST", "/api/groups/"+design.ID+"/invitations", map[string]any{"userId": bobID})
	if invitation.Group.Name != "Design" || invitation.Inviter.Name != "Alice" || invitation.Invitee == nil || invitation.Invitee.ID != bobID || invitation.Status != "pending" {
		t.Fatalf("invitation %+v", invitation)
	}
	alice.fails(http.StatusConflict, "conflict", "POST", "/api/groups/"+design.ID+"/invitations", map[string]any{"userId": bobID})
	h.mail("bob@example.com", "Alice invited you to join Design")
	bob.hasEntry("group.invited", "Alice invited you to join Design.")

	bob.fails(http.StatusNotFound, "not_found", "GET", "/api/groups/"+design.ID, nil)
	pending := call[struct{ Invitations []struct{ ID string } }](bob, http.StatusOK, "GET", "/api/invitations", nil)
	if len(pending.Invitations) != 1 || pending.Invitations[0].ID != invitation.ID {
		t.Fatalf("bob's invitations %+v", pending)
	}
	carol.fails(http.StatusNotFound, "not_found", "POST", "/api/invitations/"+invitation.ID+"/accept", nil)
	joined := call[group](bob, http.StatusOK, "POST", "/api/invitations/"+invitation.ID+"/accept", nil)
	if joined.Role != "member" || len(joined.Members) != 2 || joined.Members[0].Role != "owner" {
		t.Fatalf("joined %+v", joined)
	}
	alice.hasEntry("group.joined", "Bob joined Design.")

	// The group's list: every member sees its tasks.
	task := call[struct{ ID string }](alice, http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Draw the logo", "groupId": design.ID, "assigneeId": bobID})
	bob.expect(http.StatusOK, "GET", "/api/tasks/"+task.ID, nil)
	if l := call[struct{ Tasks []struct{ ID string } }](bob, http.StatusOK, "GET", "/api/tasks?view=group&group="+design.ID, nil); len(l.Tasks) != 1 {
		t.Fatalf("bob's view of the group %+v", l)
	}
	if g := call[group](bob, http.StatusOK, "GET", "/api/groups/"+design.ID, nil); g.OpenTasks != 1 {
		t.Errorf("open tasks %d", g.OpenTasks)
	}
	h.mail("bob@example.com", "Alice assigned you “Draw the logo”")
	carol.fails(http.StatusNotFound, "not_found", "GET", "/api/tasks/"+task.ID, nil)
	carol.fails(http.StatusNotFound, "not_found", "GET", "/api/tasks?view=group&group="+design.ID, nil)

	// Roles.
	bob.fails(http.StatusForbidden, "permission_denied", "PATCH", "/api/groups/"+design.ID, map[string]any{"name": "Bob's"})
	bob.fails(http.StatusForbidden, "permission_denied", "PATCH", "/api/groups/"+design.ID+"/members/"+bobID, map[string]any{"role": "admin"})
	promoted := call[group](alice, http.StatusOK, "PATCH", "/api/groups/"+design.ID+"/members/"+bobID, map[string]any{"role": "admin"})
	if promoted.Members[1].Role != "admin" {
		t.Fatalf("promoted %+v", promoted)
	}
	alice.violates("role", "PATCH", "/api/groups/"+design.ID+"/members/"+bobID, map[string]any{"role": "owner"})
	if renamed := call[group](bob, http.StatusOK, "PATCH", "/api/groups/"+design.ID, map[string]any{"name": "Design team"}); renamed.Name != "Design team" || renamed.Role != "admin" {
		t.Fatalf("renamed %+v", renamed)
	}
	bob.fails(http.StatusForbidden, "permission_denied", "DELETE", "/api/groups/"+design.ID, nil)
	bob.fails(http.StatusForbidden, "permission_denied", "DELETE", "/api/groups/"+design.ID+"/members/"+alice.id(), nil)

	// Leaving: anyone but the owner. The one who leaves stops doing the group's tasks.
	alice.fails(http.StatusConflict, "conflict", "DELETE", "/api/groups/"+design.ID+"/members/"+alice.id(), nil)
	bob.expect(http.StatusNoContent, "DELETE", "/api/groups/"+design.ID+"/members/"+bobID, nil)
	bob.fails(http.StatusNotFound, "not_found", "GET", "/api/tasks/"+task.ID, nil)
	if l := call[struct{ Groups []group }](bob, http.StatusOK, "GET", "/api/groups", nil); len(l.Groups) != 0 {
		t.Fatalf("bob still has %+v", l)
	}
	alice.hasEntry("group.left", "Bob left Design team.")
	h.settle("the task to lose its assignee", 50*time.Millisecond, func() bool {
		return call[struct{ Assignee *user }](alice, http.StatusOK, "GET", "/api/tasks/"+task.ID, nil).Assignee == nil
	})

	// Deleting: the tasks go back to their owners' lists.
	alice.expect(http.StatusNoContent, "DELETE", "/api/groups/"+design.ID, nil)
	alice.fails(http.StatusNotFound, "not_found", "GET", "/api/groups/"+design.ID, nil)
	h.settle("the task to go back to alice's inbox", 50*time.Millisecond, func() bool {
		l := call[struct{ Tasks []struct{ ID string } }](alice, http.StatusOK, "GET", "/api/tasks?view=inbox", nil)
		return len(l.Tasks) == 1 && l.Tasks[0].ID == task.ID
	})
	if got := call[struct{ Group *struct{ ID string } }](alice, http.StatusOK, "GET", "/api/tasks/"+task.ID, nil); got.Group != nil {
		t.Errorf("the task is still on %+v", got.Group)
	}
}
