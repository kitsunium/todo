package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/clock"
	loadrole "github.com/kitsunium/todo/cmd/todo/load"
	"github.com/kitsunium/todo/cmd/todo/server"

	// What the binary's main links for this role: the HTTP engine and the
	// subsystems its components declare. A test program links them itself.
	_ "github.com/kitsunium/sdk/framework/app/jobs"
	_ "github.com/kitsunium/sdk/framework/app/queue"
	_ "github.com/kitsunium/sdk/framework/app/server"
	_ "github.com/kitsunium/sdk/framework/app/store"
)

// epoch is the instant the product's clock starts at.
var epoch = time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC)

// product is the server role, whole, for one test: in memory, on a free
// loopback port, on a manual clock at epoch.
type product struct {
	t   *testing.T
	url string
	clk *clock.ManualClock
}

// start wires the role as the binary does and runs its App until the test
// ends.
func start(t *testing.T) *product {
	t.Helper()
	clk := clock.NewManualClock(epoch)
	server.Wire(t.Context(), server.Build{})
	a := server.App.With(app.InMemory(), app.Listen("127.0.0.1:0"), app.Env(app.EnvDev), app.Logs(io.Discard), app.Clock(clk))
	if err := a.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.Stop(ctx); err != nil {
			t.Error(err)
		}
	})
	return &product{t: t, url: a.URL(), clk: clk}
}

// member is one caller of the API: a stranger while token is empty, a
// member signed in once it holds a session's secret.
type member struct {
	p     *product
	token string
	id    string
}

// stranger is a caller with no session.
func (p *product) stranger() *member { return &member{p: p} }

// passphrase is every test member's.
const passphrase = "correct horse battery"

// signUp makes an account and signs it in.
func (p *product) signUp(name, email string) *member {
	p.t.Helper()
	m := p.stranger()
	creds := map[string]any{"email": email, "name": name, "password": passphrase}
	m.call(http.StatusOK, "POST", "/api/auth/signup", creds, nil)
	var in struct {
		Token string `json:"token"`
		User  struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	delete(creds, "name") // a request is decoded strictly
	m.call(http.StatusOK, "POST", "/api/auth/login", creds, &in)
	m.token, m.id = in.Token, in.User.ID
	return m
}

// call sends one request, checks its status and decodes its answer into out.
func (m *member) call(status int, method, path string, in, out any) http.Header {
	p := m.p
	p.t.Helper()
	var body io.Reader
	if in != nil {
		raw, err := json.Marshal(in)
		if err != nil {
			p.t.Fatal(err)
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(p.t.Context(), method, p.url+path, body)
	if err != nil {
		p.t.Fatal(err)
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if m.token != "" {
		req.Header.Set("Authorization", "Bearer "+m.token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		p.t.Fatalf("%s %s: %v", method, path, err)
	}
	raw, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != status {
		p.t.Fatalf("%s %s: %d %s, want %d", method, path, resp.StatusCode, raw, status)
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			p.t.Fatalf("%s %s: decode %s: %v", method, path, raw, err)
		}
	}
	return resp.Header
}

// settle moves the product's clock forward by step until cond holds: the
// jobs wait on that clock, so nothing scheduled happens unless it moves.
func (p *product) settle(what string, step time.Duration, cond func() bool) {
	p.t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			p.t.Fatalf("timed out waiting for %s", what)
		}
		p.clk.Advance(step)
		time.Sleep(2 * time.Millisecond)
	}
}

// taskView is a task as the API answers it.
type taskView struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Priority string    `json:"priority"`
	State    string    `json:"state"`
	DoneAt   time.Time `json:"doneAt"`
}

// A member's task is added, listed, moved through its lifecycle, and every
// move is in their feed, and in nobody else's; what the API refuses, it refuses with the status a client
// can act on.
func TestATaskGoesThroughItsLifecycle(t *testing.T) {
	p := start(t)
	ada, bob := p.signUp("Ada", "ada@example.com"), p.signUp("Bob", "bob@example.com")
	p.stranger().call(http.StatusUnauthorized, "GET", "/api/tasks", nil, nil)
	p.stranger().call(http.StatusUnauthorized, "POST", "/api/tasks", map[string]any{"title": "x"}, nil)
	var added taskView
	ada.call(http.StatusOK, "POST", "/api/tasks", map[string]any{"title": " Write the design ", "priority": "high", "assignee": "Ada"}, &added)
	if added.ID == "" || added.Title != "Write the design" || added.Priority != "high" || added.State != "open" {
		t.Fatalf("the task added: %+v", added)
	}
	ada.call(http.StatusBadRequest, "POST", "/api/tasks", map[string]any{"title": "  "}, nil)
	ada.call(http.StatusBadRequest, "POST", "/api/tasks", map[string]any{"title": "x", "priority": "urgent"}, nil)
	ada.call(http.StatusNotFound, "POST", "/api/tasks/task_nobody/complete", nil, nil)
	ada.call(http.StatusConflict, "POST", "/api/tasks/"+added.ID+"/archive", nil, nil)
	bob.call(http.StatusNotFound, "POST", "/api/tasks/"+added.ID+"/complete", nil, nil) // not his: as if there were none

	var moved taskView
	ada.call(http.StatusOK, "POST", "/api/tasks/"+added.ID+"/complete", nil, &moved)
	if moved.State != "done" || !moved.DoneAt.Equal(epoch) {
		t.Fatalf("the task completed: %+v", moved)
	}
	ada.call(http.StatusOK, "POST", "/api/tasks/"+added.ID+"/reopen", nil, &moved)
	ada.call(http.StatusOK, "POST", "/api/tasks/"+added.ID+"/complete", nil, &moved)
	ada.call(http.StatusOK, "POST", "/api/tasks/"+added.ID+"/archive", nil, &moved)
	if moved.State != "archived" {
		t.Fatalf("the task archived: %+v", moved)
	}

	var list struct {
		Tasks []taskView `json:"tasks"`
	}
	ada.call(http.StatusOK, "GET", "/api/tasks", nil, &list)
	if len(list.Tasks) != 0 {
		t.Errorf("the list still holds the archived task: %+v", list.Tasks)
	}
	ada.call(http.StatusOK, "GET", "/api/tasks?state=archived", nil, &list)
	if len(list.Tasks) != 1 || list.Tasks[0].ID != added.ID {
		t.Errorf("the archived tasks: %+v, want the one task", list.Tasks)
	}
	ada.call(http.StatusOK, "POST", "/api/tasks/"+added.ID+"/restore", nil, &moved)
	if moved.State != "open" || !moved.DoneAt.IsZero() {
		t.Errorf("the task restored: %+v", moved)
	}

	var feed struct {
		Entries []struct {
			Kind   string `json:"kind"`
			TaskID string `json:"taskId"`
		} `json:"entries"`
	}
	// The feed hears of the moves through a topic: later, on the product's clock.
	p.settle("the six entries of the feed", 50*time.Millisecond, func() bool {
		ada.call(http.StatusOK, "GET", "/api/activity", nil, &feed)
		return len(feed.Entries) >= 6
	})
	var kinds []string
	for _, e := range feed.Entries {
		if e.TaskID != added.ID {
			t.Errorf("an entry of another task: %+v", e)
		}
		kinds = append(kinds, e.Kind)
	}
	if len(kinds) != 6 || kinds[0] != "restore" || kinds[5] != "added" {
		t.Errorf("the feed, newest first: %q, want restore … added, six entries", kinds)
	}
	ada.call(http.StatusOK, "GET", "/api/activity?limit=2", nil, &feed)
	if len(feed.Entries) != 2 {
		t.Errorf("a page of 2 holds %d entries", len(feed.Entries))
	}
	bob.call(http.StatusOK, "GET", "/api/activity", nil, &feed)
	bob.call(http.StatusOK, "GET", "/api/tasks?state=open", nil, &list)
	if len(feed.Entries) != 0 || len(list.Tasks) != 0 {
		t.Errorf("Bob sees %d entries and %d tasks of Ada's", len(feed.Entries), len(list.Tasks))
	}
}

// The jobs run on the product's clock: the vital signs are sampled, a task
// due soon is reminded of once, and a task done for a day is archived.
func TestTheJobsRunOnTheClock(t *testing.T) {
	p := start(t)
	ada := p.signUp("Ada", "ada@example.com")
	var soon, other taskView
	ada.call(http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Ship the release", "assignee": "ada@example.com", "due": epoch.Add(6 * time.Hour)}, &soon)
	ada.call(http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Tidy the backlog"}, &other)
	ada.call(http.StatusOK, "POST", "/api/tasks/"+other.ID+"/complete", nil, nil)

	var report struct {
		Latest *struct {
			Counts         map[string]int `json:"counts"`
			Total          int            `json:"total"`
			CompletionRate float64        `json:"completionRate"`
		} `json:"latest"`
	}
	p.settle("a sample of the vital signs", 5*time.Second, func() bool {
		p.stranger().call(http.StatusOK, "GET", "/api/stats", nil, &report)
		return report.Latest != nil
	})
	if l := report.Latest; l.Total != 2 || l.Counts["open"] != 1 || l.Counts["done"] != 1 || l.CompletionRate != 0.5 {
		t.Errorf("the sample: %+v", *l)
	}

	var outbox struct {
		Mails []struct {
			To      string `json:"to"`
			Subject string `json:"subject"`
			Relayed bool   `json:"relayed"`
		} `json:"mails"`
	}
	p.settle("the reminder", 10*time.Second, func() bool {
		ada.call(http.StatusOK, "GET", "/api/outbox", nil, &outbox)
		return len(outbox.Mails) > 0
	})
	p.clk.Advance(5 * time.Minute) // more runs of the job: the task is reminded of once
	time.Sleep(20 * time.Millisecond)
	ada.call(http.StatusOK, "GET", "/api/outbox", nil, &outbox)
	if len(outbox.Mails) != 1 || outbox.Mails[0].To != "ada@example.com" || outbox.Mails[0].Subject != "Due soon: Ship the release" || outbox.Mails[0].Relayed {
		t.Errorf("the outbox: %+v, want the one reminder, kept and not relayed", outbox.Mails)
	}

	var list struct {
		Tasks []taskView `json:"tasks"`
	}
	p.settle("the done task archived", time.Hour, func() bool {
		ada.call(http.StatusOK, "GET", "/api/tasks?state=archived", nil, &list)
		return len(list.Tasks) == 1
	})
	if list.Tasks[0].ID != other.ID {
		t.Errorf("archived %+v, want the task done a day ago", list.Tasks)
	}
}

// An account signs in with its password, by a bearer token or the session
// cookie; a wrong password is refused like an unknown address, five in a
// row lock the account for fifteen minutes, and signing out ends the
// session.
func TestAMemberSignsInAndOut(t *testing.T) {
	p := start(t)
	anon := p.stranger()
	ada := p.signUp("Ada Lovelace", "Ada@Example.com")
	anon.call(http.StatusConflict, "POST", "/api/auth/signup", map[string]any{"email": "ada@example.com", "name": "Ada", "password": passphrase}, nil)
	anon.call(http.StatusBadRequest, "POST", "/api/auth/signup", map[string]any{"email": "grace@example.com", "name": "Grace", "password": "short"}, nil)
	anon.call(http.StatusBadRequest, "POST", "/api/auth/signup", map[string]any{"email": "not an address", "name": "Grace", "password": passphrase}, nil)

	var me struct {
		ID, Email, Name string
	}
	ada.call(http.StatusOK, "GET", "/api/auth/me", nil, &me)
	if me.ID != ada.id || me.Email != "ada@example.com" || me.Name != "Ada Lovelace" {
		t.Errorf("me: %+v", me)
	}
	anon.call(http.StatusUnauthorized, "GET", "/api/auth/me", nil, nil)
	(&member{p: p, token: "not-a-session"}).call(http.StatusUnauthorized, "GET", "/api/auth/me", nil, nil)

	// A browser: the cookie the sign-in sets opens the same session.
	creds := map[string]any{"email": "ada@example.com", "password": passphrase}
	cookie := anon.call(http.StatusOK, "POST", "/api/auth/login", creds, nil).Get("Set-Cookie")
	if !strings.HasPrefix(cookie, "todo_session=") || !strings.Contains(cookie, "HttpOnly") {
		t.Fatalf("the sign-in sets %q, want the todo_session cookie, HttpOnly", cookie)
	}
	req, _ := http.NewRequestWithContext(t.Context(), "GET", p.url+"/api/auth/me", nil)
	req.Header.Set("Cookie", strings.SplitN(cookie, ";", 2)[0])
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("me with the cookie: %d, want 200", resp.StatusCode)
	}

	wrong := map[string]any{"email": "ada@example.com", "password": "not the right one"}
	anon.call(http.StatusUnauthorized, "POST", "/api/auth/login", map[string]any{"email": "nobody@example.com", "password": passphrase}, nil)
	for range 5 {
		anon.call(http.StatusUnauthorized, "POST", "/api/auth/login", wrong, nil)
	}
	anon.call(http.StatusForbidden, "POST", "/api/auth/login", creds, nil)
	p.clk.Advance(15*time.Minute + time.Second)
	anon.call(http.StatusOK, "POST", "/api/auth/login", creds, nil)

	ada.call(http.StatusOK, "POST", "/api/auth/logout", nil, nil)
	ada.call(http.StatusUnauthorized, "GET", "/api/auth/me", nil, nil)
}

// The load role, whole, drives the server role: it signs in as its own
// member and one round makes every call of the API, none refused (its exit
// status says so).
func TestTheLoadDrivesTheServer(t *testing.T) {
	p := start(t)
	loadrole.Wire(t.Context(), loadrole.Build{})
	a := loadrole.App.With(app.Set("target", strings.TrimPrefix(p.url, "http://")))
	if code := a.Main(t.Context(), []string{"-for", "0s", "-rate", "200"}); code != 0 {
		t.Fatalf("the load exits %d, want 0: every call answered", code)
	}
	var report struct {
		Latest *struct {
			Total int `json:"total"`
		} `json:"latest"`
	}
	p.settle("a sample that counts the load's task", 5*time.Second, func() bool {
		p.stranger().call(http.StatusOK, "GET", "/api/stats", nil, &report)
		return report.Latest != nil
	})
	if report.Latest.Total != 1 {
		t.Errorf("the server counts %d task(s), want the one the load added", report.Latest.Total)
	}
}

// The server serves the web app: its page at the root and at any path a
// browser may be at, its style and its script beside it, and the API under
// it still answers as the API.
func TestTheServerServesTheWebApp(t *testing.T) {
	p := start(t)
	get := func(path string) (int, string, string) {
		t.Helper()
		resp, err := http.Get(p.url + path)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return resp.StatusCode, resp.Header.Get("Content-Type"), string(raw)
	}
	for _, path := range []string{"/", "/tasks"} {
		if code, kind, body := get(path); code != http.StatusOK || !strings.HasPrefix(kind, "text/html") || !strings.Contains(body, `<script src="/app.js" defer>`) {
			t.Errorf("GET %s: %d %s, want the page", path, code, kind)
		}
	}
	if code, kind, body := get("/app.js"); code != http.StatusOK || !strings.Contains(kind, "javascript") || !strings.Contains(body, "/api/auth/login") {
		t.Errorf("GET /app.js: %d %s, want the script", code, kind)
	}
	if code, kind, _ := get("/app.css"); code != http.StatusOK || !strings.HasPrefix(kind, "text/css") {
		t.Errorf("GET /app.css: %d %s, want the style", code, kind)
	}
	if code, kind, _ := get("/api/tasks"); code != http.StatusUnauthorized || !strings.Contains(kind, "json") {
		t.Errorf("GET /api/tasks: %d %s, want the API's 401", code, kind)
	}
}
