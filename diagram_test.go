package main

import (
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/kitsunium/platform/analyzer"
	"github.com/kitsunium/platform/model"
)

// exercise uses the product the way people do, so that every service works
// and every edge the code can take is taken at least once.
func exercise(h *harness) {
	h.t.Helper()
	alice := h.signup("Alice", "alice@example.com")
	bob := h.signup("Bob", "bob@example.com")
	befriend(alice, bob, "bob@example.com")
	bobID := bob.id()

	// Accounts: a failed sign-in, a password change, a reset, sessions.
	h.client("anon").fails(http.StatusUnauthorized, "invalid_credentials", "POST", "/api/auth/login", map[string]any{"email": "bob@example.com", "password": "wrong password"})
	bob.expect(http.StatusOK, "PATCH", "/api/auth/me", map[string]any{"name": "Bob B.", "locale": "en"})
	bob.expect(http.StatusNoContent, "POST", "/api/auth/password", map[string]any{"current": testPassword, "password": testPassword})
	h.client("anon").expect(http.StatusOK, "POST", "/api/auth/password/forgot", map[string]any{"email": "alice@example.com"})
	h.client("anon").expect(http.StatusOK, "POST", "/api/auth/password/reset", map[string]any{"token": token(h.t, h.mail("alice@example.com", "/reset?token=")), "password": testPassword})
	alice.expect(http.StatusOK, "POST", "/api/auth/login", map[string]any{"email": "alice@example.com", "password": testPassword})
	phone := h.client("phone")
	phone.expect(http.StatusOK, "POST", "/api/auth/login", map[string]any{"email": "alice@example.com", "password": testPassword})
	for _, s := range call[struct {
		Sessions []struct {
			ID      string
			Current bool
		}
	}](alice, http.StatusOK, "GET", "/api/auth/sessions", nil).Sessions {
		if !s.Current {
			alice.expect(http.StatusNoContent, "DELETE", "/api/auth/sessions/"+s.ID, nil)
		}
	}
	h.client("anon").expect(http.StatusOK, "POST", "/api/auth/verify/resend", map[string]any{"email": "alice@example.com"})
	h.client("mallory").expect(http.StatusOK, "POST", "/api/auth/signup", map[string]any{"email": "alice@example.com", "name": "Mallory", "password": testPassword})

	// Contacts: an invitation claimed, a request declined, one cancelled, one removed.
	alice.expect(http.StatusOK, "POST", "/api/contacts", map[string]any{"email": "carol@example.com"})
	carol := h.signup("Carol", "carol@example.com")
	h.settle("carol's request", 50*time.Millisecond, func() bool {
		return len(call[contactList](carol, http.StatusOK, "GET", "/api/contacts", nil).Incoming) == 1
	})
	carol.expect(http.StatusNoContent, "POST", "/api/contacts/"+call[contactList](carol, http.StatusOK, "GET", "/api/contacts", nil).Incoming[0].ID+"/decline", nil)
	asked := call[struct{ Request struct{ ID string } }](carol, http.StatusOK, "POST", "/api/contacts", map[string]any{"email": "alice@example.com"})
	carol.expect(http.StatusNoContent, "POST", "/api/contacts/"+asked.Request.ID+"/cancel", nil)
	befriend(bob, carol, "carol@example.com")
	carol.expect(http.StatusNoContent, "DELETE", "/api/contacts/"+call[contactList](carol, http.StatusOK, "GET", "/api/contacts", nil).Contacts[0].ID, nil)

	// Groups: invited, joined, promoted, renamed, left, deleted.
	team := call[group](alice, http.StatusOK, "POST", "/api/groups", map[string]any{"name": "Launch", "color": "orange"})
	inv := call[struct{ ID string }](alice, http.StatusOK, "POST", "/api/groups/"+team.ID+"/invitations", map[string]any{"userId": bobID})
	bob.expect(http.StatusOK, "GET", "/api/invitations", nil)
	bob.expect(http.StatusOK, "POST", "/api/invitations/"+inv.ID+"/accept", nil)
	alice.expect(http.StatusOK, "PATCH", "/api/groups/"+team.ID+"/members/"+bobID, map[string]any{"role": "admin"})
	bob.expect(http.StatusOK, "PATCH", "/api/groups/"+team.ID, map[string]any{"name": "Launch team"})
	alice.expect(http.StatusOK, "GET", "/api/groups", nil)
	alice.expect(http.StatusOK, "GET", "/api/groups/"+team.ID, nil)
	side := call[group](alice, http.StatusOK, "POST", "/api/groups", map[string]any{"name": "Side project"})
	declined := call[struct{ ID string }](alice, http.StatusOK, "POST", "/api/groups/"+side.ID+"/invitations", map[string]any{"userId": bobID})
	bob.expect(http.StatusNoContent, "POST", "/api/invitations/"+declined.ID+"/decline", nil)
	alice.expect(http.StatusOK, "POST", "/api/groups/"+side.ID+"/invitations", map[string]any{"userId": bobID}) // revoked by the deletion
	alice.expect(http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Side task", "groupId": side.ID})
	alice.expect(http.StatusNoContent, "DELETE", "/api/groups/"+side.ID, nil)

	// Tasks: every endpoint, the guard, the timer, a reminder.
	due := h.clk.Now().Add(time.Hour).Format(time.RFC3339)
	post := call[task](alice, http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Write the post", "priority": 1, "due": due, "groupId": team.ID, "assigneeId": bobID})
	solo := call[task](alice, http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Renew the domain", "due": h.clk.Now().Add(-time.Hour).Format(time.RFC3339)})
	alice.expect(http.StatusOK, "POST", "/api/tasks/"+solo.ID+"/share", map[string]any{"userId": bobID})
	h.settle("the overdue guard", 200*time.Millisecond, func() bool {
		return call[task](alice, http.StatusOK, "GET", "/api/tasks/"+solo.ID, nil).Status == "overdue"
	})
	alice.expect(http.StatusOK, "PATCH", "/api/tasks/"+solo.ID, map[string]any{"due": due, "title": "Renew the domains"})
	bob.expect(http.StatusOK, "POST", "/api/tasks/"+solo.ID+"/complete", nil)
	bob.expect(http.StatusOK, "POST", "/api/tasks/"+solo.ID+"/reopen", nil)
	bob.expect(http.StatusOK, "POST", "/api/tasks/"+solo.ID+"/complete", nil)
	bob.expect(http.StatusOK, "POST", "/api/tasks/"+solo.ID+"/archive", nil)
	bob.expect(http.StatusOK, "POST", "/api/tasks/"+solo.ID+"/restore", nil)
	bob.expect(http.StatusOK, "DELETE", "/api/tasks/"+solo.ID+"/share/"+bobID, nil)
	for _, view := range []string{"", "?view=inbox", "?view=today", "?view=upcoming", "?view=shared", "?view=assigned", "?view=completed", "?view=group&group=" + team.ID} {
		bob.expect(http.StatusOK, "GET", "/api/tasks"+view, nil)
	}
	bob.expect(http.StatusOK, "GET", "/api/tasks/counts?tz=Europe/Paris", nil)
	bob.expect(http.StatusOK, "GET", "/api/activity", nil)
	bob.expect(http.StatusNoContent, "POST", "/api/activity/read", nil)
	h.mail("bob@example.com", "assigned you")
	h.clk.Set(h.clk.Now().Add(46 * time.Minute))
	h.mail("bob@example.com", "Due in")
	bob.expect(http.StatusOK, "POST", "/api/tasks/"+post.ID+"/complete", nil)
	h.clk.Advance(25 * time.Hour) // the auto-archive timer, a stats sample
	h.settle("the auto-archive timer", time.Second, func() bool {
		return call[task](alice, http.StatusOK, "GET", "/api/tasks/"+post.ID, nil).Status == "archived"
	})
	h.client("anon").expect(http.StatusOK, "GET", "/api/stats", nil)
	alice.expect(http.StatusNoContent, "DELETE", "/api/tasks/"+post.ID, nil)
	bob.expect(http.StatusNoContent, "DELETE", "/api/groups/"+team.ID+"/members/"+bobID, nil)
	alice.expect(http.StatusNoContent, "DELETE", "/api/groups/"+team.ID, nil)
	alice.expect(http.StatusNoContent, "POST", "/api/auth/logout", nil)
	h.drain()
}

// analyzed waits for the in-process static analysis and returns the graph.
func (h *harness) analyzed() *model.Graph {
	h.t.Helper()
	deadline := time.Now().Add(2 * time.Minute)
	for {
		g := h.app.Graph()
		if g.Analysis != nil && g.Analysis.Status != "running" {
			if g.Analysis.Status != "ok" {
				h.t.Fatalf("analysis %+v", g.Analysis)
			}
			return g
		}
		if time.Now().After(deadline) {
			h.t.Fatal("the static analysis never finished")
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// The diagram never lies. The product the process runs and the product the
// source declares are the same nodes; every edge the source proves is in
// the diagram; and every edge the running product took is one the source
// explains — declared by construction, or found in the code.
func TestTheDiagramMatchesTheCode(t *testing.T) {
	h := start(t, true)
	exercise(h)
	g := h.analyzed()
	static, err := analyzer.Analyze(t.Context(), analyzer.Options{Dir: "."})
	if err != nil {
		t.Fatal(err)
	}

	// The same nodes, both ways.
	ids := func(nodes []model.Node) []string {
		var out []string
		for _, n := range nodes {
			if n.Kind != model.KindExternal {
				out = append(out, n.ID)
			}
		}
		slices.Sort(out)
		return out
	}
	running, declared := ids(g.Nodes), ids(static.Nodes)
	for _, id := range running {
		if !slices.Contains(declared, id) {
			t.Errorf("%s runs, but the static analysis does not find it", id)
		}
	}
	for _, id := range declared {
		if !slices.Contains(running, id) {
			t.Errorf("%s is declared, but the product does not run it", id)
		}
	}
	if len(running) < 90 {
		t.Errorf("only %d nodes: the product is bigger than that", len(running))
	}

	// Every edge the code proves is drawn; every edge the product took is
	// explained by the code — but the requests from outside.
	for _, e := range static.Edges {
		if g.Edge(e.ID) == nil {
			t.Errorf("the code proves %s, the diagram does not draw it", e.ID)
		}
	}
	for _, e := range g.Edges {
		if g.Node(e.From) == nil || g.Node(e.To) == nil {
			t.Errorf("edge %s points outside the graph", e.ID)
		}
		from := g.Node(e.From)
		outside := from != nil && (from.Kind == model.KindExternal || from.Kind == model.KindFrontend)
		if e.Observed != nil && !e.Declared && len(e.Static) == 0 && !outside {
			t.Errorf("the product took %s, which the code does not explain", e.ID)
		}
	}

	// Every node says where it is, what it is, and — when it runs code —
	// where that code is.
	for _, n := range g.Nodes {
		if n.Kind == model.KindExternal {
			continue
		}
		if n.Source == nil || n.Source.File == "" {
			t.Errorf("%s has no source", n.ID)
		}
		if n.Kind != model.KindService && n.Doc == "" {
			t.Errorf("%s has no documentation", n.ID)
		}
		switch n.Kind {
		case model.KindEndpoint, model.KindJob, model.KindSubscription, model.KindLoop, model.KindAuth:
			if n.Handler == nil || n.Handler.EndLine == 0 {
				t.Errorf("%s has no handler range", n.ID)
			}
		}
	}
	for _, n := range g.Nodes {
		if n.Workflow == nil {
			continue
		}
		for _, tr := range n.Workflow.Transitions {
			if tr.Trigger == model.TriggerEvent && len(tr.Callers) == 0 {
				t.Errorf("nothing fires %q of %s", tr.Event, n.ID)
			}
		}
	}

	// The daemon's own loops: the reaper written by hand, whose select the
	// analysis reads; the reminders run by kit, whose wakes are data.
	if reaper := g.Node("identity/loop/session-reaper"); reaper == nil || reaper.Loop == nil || reaper.Loop.Style != model.LoopGoroutine {
		t.Errorf("the session reaper: %+v", reaper)
	} else {
		var kinds []string
		for _, c := range reaper.Loop.Selects {
			kinds = append(kinds, c.Kind)
		}
		if !slices.Contains(kinds, model.SelectTimer) || !slices.Contains(kinds, model.SelectDone) {
			t.Errorf("the reaper waits on %+v", reaper.Loop.Selects)
		}
	}
	if loop := g.Node("notify/loop/reminders"); loop == nil || loop.Loop == nil || loop.Loop.Style != model.LoopDeclared {
		t.Errorf("the reminders loop: %+v", loop)
	} else {
		wakes := map[string]model.WakeSource{}
		for _, w := range loop.Loop.Wakes {
			wakes[w.Kind] = w
		}
		if wakes[model.WakeInterval].Every != "1m0s" || wakes[model.WakeTopic].Topic != "tasks/topic/events" || wakes[model.WakeDeadline].Kind == "" {
			t.Errorf("the reminders wake on %+v", loop.Loop.Wakes)
		}
	}
	if e := g.Edge("tasks/topic/events|wakes|notify/loop/reminders"); e == nil || !e.Declared {
		t.Errorf("no declared wakes edge: %+v", e)
	}

	// Authentication: one handler, in front of every endpoint that asks.
	guarded := 0
	for _, n := range g.Nodes {
		if n.Endpoint != nil && n.Endpoint.Auth == model.AuthRequired {
			guarded++
			if len(n.Endpoint.Pipeline) == 0 || n.Endpoint.Pipeline[0].Kind != "auth" {
				t.Errorf("%s does not authenticate first: %+v", n.ID, n.Endpoint.Pipeline)
			}
		}
	}
	if auth := g.Node("identity/auth/session"); auth == nil || auth.Auth == nil || auth.Auth.Endpoints != guarded || guarded < 30 {
		t.Errorf("the session handler guards %+v of %d endpoints", auth, guarded)
	}

	// Mail: every sender reaches the one mailer.
	for _, from := range []string{"notify/endpoint/Send", "notify/subscription/task-mail", "notify/subscription/contact-mail", "notify/subscription/group-mail", "notify/loop/reminders"} {
		if e := g.Edge(from + "|sends|notify/mailer/mail"); e == nil || len(e.Static) == 0 || e.Observed == nil {
			t.Errorf("%s does not send through the mailer: %+v", from, e)
		}
	}

	// A call from the page is drawn from the page.
	req, _ := http.NewRequest("GET", h.app.URL()+"/api/stats", nil)
	req.Header.Set("Referer", h.app.URL()+"/app/today")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if e := h.app.Graph().Edge("web/frontend/app|calls|stats/endpoint/Current"); e == nil || e.Observed == nil {
		t.Error("a call from the page is not drawn from the page")
	}
	for _, d := range g.Diagnostics {
		if d.Severity == "error" || strings.Contains(d.Message, "directly") {
			t.Errorf("diagnostic: %s (%s)", d.Message, d.Node)
		}
	}
}
