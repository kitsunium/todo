package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kitsunium/sdk/framework/kit"
	"github.com/kitsunium/sdk/framework/model"
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

	// Tasks: every endpoint, the timers, a reminder.
	due := h.clk.Now().Add(time.Hour).Format(time.RFC3339)
	post := call[task](alice, http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Write the post", "priority": 1, "due": due, "groupId": team.ID, "assigneeId": bobID})
	solo := call[task](alice, http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Renew the domain", "due": h.clk.Now().Add(-time.Hour).Format(time.RFC3339)})
	alice.expect(http.StatusOK, "POST", "/api/tasks/"+solo.ID+"/share", map[string]any{"userId": bobID})
	h.settle("the overdue timer", 200*time.Millisecond, func() bool {
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
//
// What the running product says of itself is checked everywhere. What only
// the source says is kit's static analysis: checked when kit is here, and
// skipped, by name, when it is not.
func TestTheDiagramMatchesTheCode(t *testing.T) {
	source := kitSource()
	var analysis kit.AnalyzeFunc
	if source != nil {
		analysis = source.analyze
	}
	h := start(t, analysis)
	exercise(h)
	g := h.app.Graph()
	if source != nil {
		g = h.analyzed()
	}
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

	// The product as it runs: every node in it, and every edge between two.
	running := ids(g.Nodes)
	if len(running) < 90 {
		t.Errorf("only %d nodes: the product is bigger than that", len(running))
	}
	for _, e := range g.Edges {
		if g.Node(e.From) == nil || g.Node(e.To) == nil {
			t.Errorf("edge %s points outside the graph", e.ID)
		}
	}

	// Every node says where it is.
	for _, n := range g.Nodes {
		if n.Kind != model.KindExternal && (n.Source == nil || n.Source.File == "") {
			t.Errorf("%s has no source", n.ID)
		}
	}

	// The daemon's own loops: the reaper written by hand; the reminders run
	// by kit, whose wakes are data.
	if reaper := g.Node("identity/loop/session-reaper"); reaper == nil || reaper.Loop == nil || reaper.Loop.Style != model.LoopGoroutine {
		t.Errorf("the session reaper: %+v", reaper)
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

	// Mail: every sender reached the one mailer.
	senders := []string{"notify/command/send", "notify/subscription/task-mail", "notify/subscription/contact-mail", "notify/subscription/group-mail", "notify/loop/reminders"}
	for _, from := range senders {
		if e := g.Edge(from + "|sends|notify/mailer/mail"); e == nil || e.Observed == nil {
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

	if source == nil {
		t.Skip(withoutKit + "what only the source says: the same nodes both ways, every edge explained, each node's documentation and handler range, who fires each transition, the reaper's select, the mailer's senders in the code")
	}
	static := source.graph(t)

	// The same nodes, both ways.
	declared := ids(static.Nodes)
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

	// Every edge the code proves is drawn; every edge the product took is
	// explained by the code — but the requests from outside.
	for _, e := range static.Edges {
		if g.Edge(e.ID) == nil {
			t.Errorf("the code proves %s, the diagram does not draw it", e.ID)
		}
	}
	for _, e := range g.Edges {
		from := g.Node(e.From)
		outside := from != nil && (from.Kind == model.KindExternal || from.Kind == model.KindFrontend)
		if e.Observed != nil && !e.Declared && len(e.Static) == 0 && !outside {
			t.Errorf("the product took %s, which the code does not explain", e.ID)
		}
	}

	// Every node says what it is, and — when it runs code — where that code
	// is.
	for _, n := range g.Nodes {
		if n.Kind == model.KindExternal {
			continue
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

	// The reaper, written by hand: the analysis reads its select.
	if reaper := g.Node("identity/loop/session-reaper"); reaper != nil && reaper.Loop != nil && reaper.Loop.Style == model.LoopGoroutine {
		var kinds []string
		for _, c := range reaper.Loop.Selects {
			kinds = append(kinds, c.Kind)
		}
		if !slices.Contains(kinds, model.SelectTimer) || !slices.Contains(kinds, model.SelectDone) {
			t.Errorf("the reaper waits on %+v", reaper.Loop.Selects)
		}
	}

	// Mail: the code, too, says every sender reaches the one mailer.
	for _, from := range senders {
		if e := g.Edge(from + "|sends|notify/mailer/mail"); e != nil && len(e.Static) == 0 {
			t.Errorf("%s sends through the mailer, which the code does not explain: %+v", from, e)
		}
	}
}

// The static half of the diagram is kit's. The analyzer that reads the
// source is the platform's, and the todo links none of it: the framework
// links no analyzer (SDK ADR 0147 §1), and a product imports nothing from the
// platform (the platform's ADR 0010, D2). The tests run the kit tool instead,
// as a process, and give the product what it read through the framework's own
// hook, kit.Analyzer — as the product's graph had it when it linked the
// analyzer.

// withoutKit says which half of the diagram suite did not run, and why.
const withoutKit = "kit's static analysis reads the source, and there is no kit here — put kit on PATH, or set KIT to its binary — to check "

// kitSource is kit's analysis of the todo's source, by the kit binary $KIT
// names, else by kit on PATH; nil when there is neither.
func kitSource() *sourceAnalysis {
	bin := os.Getenv("KIT")
	if bin == "" {
		var err error
		if bin, err = exec.LookPath("kit"); err != nil {
			return nil
		}
	}
	return &sourceAnalysis{bin: bin}
}

// sourceAnalysis runs `kit graph -static -format json` on the product's
// module: the graph kit's analyzer reads from the source alone.
type sourceAnalysis struct {
	bin string
	mu  sync.Mutex
	raw []byte // the graph kit printed last
	err error  // why kit's last run gave none
}

// analyze is the product's kit.AnalyzeFunc: kit run on the module rooted at
// dir.
func (s *sourceAnalysis) analyze(ctx context.Context, dir string, modules []string) (*model.Graph, error) {
	raw, g, err := s.run(ctx, dir, modules)
	s.mu.Lock()
	s.raw, s.err = raw, err
	s.mu.Unlock()
	return g, err
}

// run runs kit, and reads what it printed.
func (s *sourceAnalysis) run(ctx context.Context, dir string, modules []string) ([]byte, *model.Graph, error) {
	if len(modules) > 0 {
		// kit graph -static reads one module: the nodes of a mounted one
		// would be missing, and the diagram would lie.
		return nil, nil, fmt.Errorf("kit graph -static reads the product's module alone, and the app mounts %v", modules)
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, s.bin, "graph", "-static", "-format", "json", dir)
	cmd.Stdout, cmd.Stderr, cmd.WaitDelay = &stdout, &stderr, 5*time.Second
	if err := cmd.Run(); err != nil {
		return nil, nil, fmt.Errorf("%s graph -static %s: %w: %s", s.bin, dir, err, bytes.TrimSpace(stderr.Bytes()))
	}
	g, err := decodeGraph(stdout.Bytes())
	if err != nil {
		return nil, nil, err
	}
	return stdout.Bytes(), g, nil
}

// graph is the graph kit printed last, decoded afresh — the product merged
// its own copy into the graph it runs —, or the test's end when kit gave
// none.
func (s *sourceAnalysis) graph(t *testing.T) *model.Graph {
	t.Helper()
	s.mu.Lock()
	raw, err := s.raw, s.err
	s.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if raw == nil {
		t.Fatal("kit has not analyzed the source")
	}
	g, err := decodeGraph(raw)
	if err != nil {
		t.Fatal(err)
	}
	return g
}

// decodeGraph reads the graph kit printed, which must be in the model the
// framework go.mod requires speaks.
func decodeGraph(raw []byte) (*model.Graph, error) {
	var g model.Graph
	if err := json.Unmarshal(raw, &g); err != nil {
		return nil, fmt.Errorf("kit's graph: %w", err)
	}
	if g.Version != model.Version {
		return nil, fmt.Errorf("kit's graph is model version %d and the framework reads version %d: use a kit built on the framework go.mod requires", g.Version, model.Version)
	}
	return &g, nil
}
