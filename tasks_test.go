package main

import (
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/kitsunium/todo/tasks/api"
)

// task is a task as the API shows it.
type task struct {
	ID, Title, Notes string
	Priority         int
	Status           string
	Due              *time.Time
	Owner            user
	Assignee         *user
	Group            *struct{ ID, Name, Color string }
	SharedWith       []user
	CreatedAt        time.Time
	CompletedAt      *time.Time
	CompletedBy      *user
	Can              struct{ Edit, Delete, Share bool }
}

// taskList is GET /api/tasks.
type taskList struct{ Tasks []task }

// counts is GET /api/tasks/counts.
type counts struct {
	Inbox, Today, Upcoming, Shared, Assigned, Overdue, CompletedThisWeek int
	Groups                                                               map[string]int
	Unread                                                               int
}

// titles lists the titles of a view, in order.
func (c *client) titles(query string) []string {
	c.h.t.Helper()
	var out []string
	for _, t := range call[taskList](c, http.StatusOK, "GET", "/api/tasks"+query, nil).Tasks {
		out = append(out, t.Title)
	}
	return out
}

// The task list of one user: tasks created with priorities and due dates,
// the views and their counts, edits, the lifecycle — with the workflow's own
// guard marking a task overdue and its own timer archiving a done one.
func TestTheTaskList(t *testing.T) {
	h := start(t, false)
	alice := h.signup("Alice", "alice@example.com")
	at := func(d time.Duration) string { return epoch.Add(d).Format(time.RFC3339) }

	readme := call[task](alice, http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "  Write the kit README  ", "priority": 1, "due": at(8 * time.Hour)})
	if readme.Title != "Write the kit README" || readme.Status != "open" || readme.Priority != 1 || !strings.HasPrefix(readme.ID, "task_") ||
		readme.Owner.Name != "Alice" || readme.SharedWith == nil || !readme.Can.Edit || !readme.Can.Delete || !readme.Can.Share {
		t.Fatalf("created %+v", readme)
	}
	domain := call[task](alice, http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Renew the domain", "priority": 2, "due": at(-24 * time.Hour)})
	offsite := call[task](alice, http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Plan the offsite", "notes": "Somewhere green.", "due": at(72 * time.Hour)})
	later := call[task](alice, http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Read later", "priority": 4})

	alice.violates("title", "POST", "/api/tasks", map[string]any{"title": "   "})
	alice.violates("priority", "POST", "/api/tasks", map[string]any{"title": "x", "priority": 7})
	alice.violates("groupId", "POST", "/api/tasks", map[string]any{"title": "x", "groupId": "group_nope"})
	alice.violates("assigneeId", "POST", "/api/tasks", map[string]any{"title": "x", "assigneeId": "user_nope"})
	alice.violates("view", "GET", "/api/tasks?view=someday", nil)
	alice.violates("tz", "GET", "/api/tasks?view=today&tz=Mars/Olympus", nil)

	// The overdue guard is the workflow's own loop at work.
	h.settle("the overdue guard", 200*time.Millisecond, func() bool {
		return call[task](alice, http.StatusOK, "GET", "/api/tasks/"+domain.ID, nil).Status == "overdue"
	})
	if got := alice.titles("?view=today"); !slices.Equal(got, []string{"Renew the domain", "Write the kit README"}) {
		t.Errorf("today %q", got)
	}
	if got := alice.titles("?view=upcoming"); !slices.Equal(got, []string{"Plan the offsite"}) {
		t.Errorf("upcoming %q", got)
	}
	if got := alice.titles("?view=inbox"); len(got) != 4 || got[3] != "Read later" {
		t.Errorf("inbox %q", got)
	}
	// Today ends at midnight where the user is. The README, due at 17:00
	// UTC, is due today in New York (13:00), but tomorrow in Tokyo (02:00)
	// and in Honolulu, where it is still yesterday evening.
	for tz, want := range map[string][]string{
		"America/New_York": {"Renew the domain", "Write the kit README"},
		"Asia/Tokyo":       {"Renew the domain"},
		"Pacific/Honolulu": {"Renew the domain"},
	} {
		if got := alice.titles("?view=today&tz=" + tz); !slices.Equal(got, want) {
			t.Errorf("today in %s: %q, want %q", tz, got, want)
		}
	}
	if got := alice.titles("?view=upcoming&tz=Asia/Tokyo"); !slices.Equal(got, []string{"Write the kit README", "Plan the offsite"}) {
		t.Errorf("upcoming in Tokyo %q", got)
	}
	c := call[counts](alice, http.StatusOK, "GET", "/api/tasks/counts", nil)
	if c.Inbox != 4 || c.Today != 2 || c.Upcoming != 1 || c.Overdue != 1 || c.Shared != 0 || c.Assigned != 0 || c.CompletedThisWeek != 0 || c.Groups == nil {
		t.Errorf("counts %+v", c)
	}

	// Edits: an absent member stays, an empty one clears; an overdue task
	// moved to a later date is to do again.
	edited := call[task](alice, http.StatusOK, "PATCH", "/api/tasks/"+readme.ID, map[string]any{"notes": "Mention the Studio.", "priority": 0, "due": ""})
	if edited.Title != "Write the kit README" || edited.Notes != "Mention the Studio." || edited.Priority != 0 || edited.Due != nil {
		t.Fatalf("edited %+v", edited)
	}
	rescheduled := call[task](alice, http.StatusOK, "PATCH", "/api/tasks/"+domain.ID, map[string]any{"due": at(48 * time.Hour)})
	if rescheduled.Status != "open" {
		t.Fatalf("rescheduled %+v", rescheduled)
	}
	alice.violates("due", "PATCH", "/api/tasks/"+domain.ID, map[string]any{"due": "tomorrow"})
	alice.violates("priority", "PATCH", "/api/tasks/"+domain.ID, map[string]any{"priority": -1})
	alice.violates("title", "PATCH", "/api/tasks/"+domain.ID, map[string]any{"title": ""})
	alice.fails(http.StatusBadRequest, "invalid_argument", "PATCH", "/api/tasks/"+domain.ID, map[string]any{"status": "done"})

	// The lifecycle, through its endpoints.
	for _, step := range []struct{ event, want string }{
		{"complete", "done"}, {"reopen", "open"}, {"complete", "done"}, {"archive", "archived"}, {"restore", "open"}, {"complete", "done"},
	} {
		got := call[task](alice, http.StatusOK, "POST", "/api/tasks/"+readme.ID+"/"+step.event, nil)
		if got.Status != step.want {
			t.Fatalf("%s: %+v", step.event, got)
		}
		finished := step.want == "done" || step.want == "archived" // an archived task keeps its completion
		if (got.CompletedAt != nil) != finished || (finished && (got.CompletedBy == nil || got.CompletedBy.Name != "Alice")) {
			t.Fatalf("%s: completion %v by %v", step.event, got.CompletedAt, got.CompletedBy)
		}
	}
	alice.fails(http.StatusConflict, "conflict", "POST", "/api/tasks/"+readme.ID+"/complete", nil)
	alice.fails(http.StatusConflict, "conflict", "POST", "/api/tasks/"+offsite.ID+"/archive", nil)
	if c := call[counts](alice, http.StatusOK, "GET", "/api/tasks/counts", nil); c.CompletedThisWeek != 1 || c.Inbox != 3 {
		t.Errorf("counts after completing %+v", c)
	}
	if got := alice.titles(""); len(got) != 4 || got[3] != "Write the kit README" {
		t.Errorf("the whole list keeps a done task, last: %q", got)
	}

	// The auto-archive timer: a day after being done.
	h.clk.Advance(api.ArchiveAfter - time.Minute)
	h.drain()
	if got := call[task](alice, http.StatusOK, "GET", "/api/tasks/"+readme.ID, nil); got.Status != "done" {
		t.Fatalf("archived too early: %s", got.Status)
	}
	h.settle("the auto-archive timer", 5*time.Second, func() bool {
		return call[task](alice, http.StatusOK, "GET", "/api/tasks/"+readme.ID, nil).Status == "archived"
	})
	if got := alice.titles("?view=completed"); !slices.Equal(got, []string{"Write the kit README"}) {
		t.Errorf("completed %q", got)
	}
	if got := alice.titles(""); slices.Contains(got, "Write the kit README") {
		t.Errorf("the whole list shows an archived task: %q", got)
	}

	alice.expect(http.StatusNoContent, "DELETE", "/api/tasks/"+later.ID, nil)
	alice.fails(http.StatusNotFound, "not_found", "GET", "/api/tasks/"+later.ID, nil)
	alice.fails(http.StatusNotFound, "not_found", "DELETE", "/api/tasks/"+later.ID, nil)
}

// Sharing needs a contact; a stranger does not see a task — 404, not 403 —
// and whoever sees it may work on it, but only its owner shares or deletes
// it. Everyone concerned hears about it: by mail, and in their feed.
func TestSharingATask(t *testing.T) {
	h := start(t, false)
	alice := h.signup("Alice", "alice@example.com")
	bob := h.signup("Bob", "bob@example.com")
	dave := h.signup("Dave", "dave@example.com")
	befriend(alice, bob, "bob@example.com")
	bobID, daveID := bob.id(), dave.id()

	milk := call[task](alice, http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Buy milk", "priority": 3})
	alice.fails(http.StatusForbidden, "permission_denied", "POST", "/api/tasks/"+milk.ID+"/share", map[string]any{"userId": daveID})
	alice.violates("userId", "POST", "/api/tasks/"+milk.ID+"/share", map[string]any{"userId": alice.id()})
	dave.fails(http.StatusNotFound, "not_found", "GET", "/api/tasks/"+milk.ID, nil)
	dave.fails(http.StatusNotFound, "not_found", "POST", "/api/tasks/"+milk.ID+"/complete", nil)
	dave.fails(http.StatusNotFound, "not_found", "PATCH", "/api/tasks/"+milk.ID, map[string]any{"title": "Mine"})
	dave.fails(http.StatusNotFound, "not_found", "DELETE", "/api/tasks/"+milk.ID, nil)
	dave.fails(http.StatusNotFound, "not_found", "POST", "/api/tasks/"+milk.ID+"/share", map[string]any{"userId": daveID})

	shared := call[task](alice, http.StatusOK, "POST", "/api/tasks/"+milk.ID+"/share", map[string]any{"userId": bobID})
	if len(shared.SharedWith) != 1 || shared.SharedWith[0].Email != "bob@example.com" {
		t.Fatalf("shared %+v", shared)
	}
	if again := call[task](alice, http.StatusOK, "POST", "/api/tasks/"+milk.ID+"/share", map[string]any{"userId": bobID}); len(again.SharedWith) != 1 {
		t.Fatalf("shared twice %+v", again)
	}
	mail := h.mail("bob@example.com", "Alice a partagé « Buy milk » avec vous")
	if !strings.Contains(mail.Text, "http://localhost:4000/app/tasks/"+milk.ID) || !strings.Contains(mail.HTML, "Buy milk") {
		t.Errorf("the shared mail: %q", mail.Text)
	}
	bob.hasEntry("task.shared", "Alice shared “Buy milk” with you.")

	seen := call[task](bob, http.StatusOK, "GET", "/api/tasks/"+milk.ID, nil)
	if !seen.Can.Edit || seen.Can.Delete || seen.Can.Share {
		t.Fatalf("bob may %+v", seen.Can)
	}
	if got := bob.titles("?view=shared"); !slices.Equal(got, []string{"Buy milk"}) {
		t.Errorf("bob's shared view %q", got)
	}
	if c := call[counts](bob, http.StatusOK, "GET", "/api/tasks/counts", nil); c.Shared != 1 || c.Unread == 0 {
		t.Errorf("bob's counts %+v", c)
	}
	bob.expect(http.StatusNoContent, "POST", "/api/activity/read", nil)
	if c := call[counts](bob, http.StatusOK, "GET", "/api/tasks/counts", nil); c.Unread != 0 {
		t.Errorf("unread after reading %d", c.Unread)
	}

	// Assigned to a sharee; completed by them.
	assigned := call[task](alice, http.StatusOK, "PATCH", "/api/tasks/"+milk.ID, map[string]any{"assigneeId": bobID})
	if assigned.Assignee == nil || assigned.Assignee.ID != bobID {
		t.Fatalf("assigned %+v", assigned)
	}
	h.mail("bob@example.com", "Alice vous a attribué « Buy milk »")
	if got := bob.titles("?view=assigned"); !slices.Equal(got, []string{"Buy milk"}) {
		t.Errorf("bob's assigned view %q", got)
	}
	done := call[task](bob, http.StatusOK, "POST", "/api/tasks/"+milk.ID+"/complete", nil)
	if done.CompletedBy == nil || done.CompletedBy.ID != bobID {
		t.Fatalf("completed by %+v", done.CompletedBy)
	}
	alice.hasEntry("task.completed", "Bob completed “Buy milk”.")

	bob.fails(http.StatusForbidden, "permission_denied", "DELETE", "/api/tasks/"+milk.ID, nil)
	bob.fails(http.StatusForbidden, "permission_denied", "POST", "/api/tasks/"+milk.ID+"/share", map[string]any{"userId": daveID})
	bobs := call[struct{ ID string }](bob, http.StatusOK, "POST", "/api/groups", map[string]any{"name": "Bob's"})
	bob.fails(http.StatusForbidden, "permission_denied", "PATCH", "/api/tasks/"+milk.ID, map[string]any{"groupId": bobs.ID})

	// A sharee leaves: they no longer see it, nor do it.
	left := call[task](bob, http.StatusOK, "DELETE", "/api/tasks/"+milk.ID+"/share/"+bobID, nil)
	if len(left.SharedWith) != 0 || left.Assignee != nil || left.Can.Edit {
		t.Fatalf("after leaving %+v", left)
	}
	bob.fails(http.StatusNotFound, "not_found", "GET", "/api/tasks/"+milk.ID, nil)
	alice.hasEntry("task.unshared", "Bob left “Buy milk”.")
}

// A task's reminder goes out fifteen minutes before it is due, once, to its
// owner and to the users it is shared with.
func TestRemindersGoOutBeforeTheDueDate(t *testing.T) {
	h := start(t, false)
	alice := h.signup("Alice", "alice@example.com")
	bob := h.signup("Bob", "bob@example.com")
	bob.expect(http.StatusOK, "PATCH", "/api/auth/me", map[string]any{"locale": "en"}) // each is reminded in their language
	befriend(alice, bob, "bob@example.com")

	due := h.clk.Now().Add(time.Hour).Truncate(time.Second)
	review := call[task](alice, http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Review the launch post", "priority": 1, "due": due.Format(time.RFC3339)})
	alice.expect(http.StatusOK, "POST", "/api/tasks/"+review.ID+"/share", map[string]any{"userId": bob.id()})
	quiet := call[task](alice, http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Done before its time", "due": due.Format(time.RFC3339)})
	alice.expect(http.StatusOK, "POST", "/api/tasks/"+quiet.ID+"/complete", nil)
	h.drain()

	h.clk.Set(due.Add(-16 * time.Minute))
	h.drain()
	if n := len(mailsTo("alice@example.com", "Échéance dans")); n != 0 {
		t.Fatalf("%d reminders sixteen minutes before", n)
	}
	h.clk.Set(due.Add(-15 * time.Minute))
	h.mail("alice@example.com", "Échéance dans 15 minutes : Review the launch post")
	h.mail("bob@example.com", "Due in 15 minutes: Review the launch post")
	h.clk.Advance(10 * time.Minute)
	h.drain()
	for _, addr := range []string{"alice@example.com", "bob@example.com"} {
		if got := append(mailsTo(addr, "Due in"), mailsTo(addr, "Échéance dans")...); len(got) != 1 {
			t.Errorf("%s got %d reminders", addr, len(got))
		}
	}
	if got := mailsTo("alice@example.com", "Done before its time"); len(got) != 0 {
		t.Errorf("a done task was reminded: %d mails", len(got))
	}
}
