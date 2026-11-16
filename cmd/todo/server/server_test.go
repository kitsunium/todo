package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/framework/model"
	"github.com/kitsunium/sdk/pkg/v1/clock"
	loadrole "github.com/kitsunium/todo/cmd/todo/load"
	"github.com/kitsunium/todo/cmd/todo/server"
	"github.com/kitsunium/todo/notify/post"

	// What the binary's main links for this role: the HTTP engine and the
	// subsystems its components declare. A test program links them itself.
	_ "github.com/kitsunium/sdk/framework/app/jobs"
	_ "github.com/kitsunium/sdk/framework/app/mail"
	_ "github.com/kitsunium/sdk/framework/app/queue"
	_ "github.com/kitsunium/sdk/framework/app/server"
	_ "github.com/kitsunium/sdk/framework/app/store"
	_ "github.com/kitsunium/sdk/framework/app/workflow"
)

// epoch is the instant the product's clock starts at.
var epoch = time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC)

// product is the server role, whole, for one test: in memory, on a free
// loopback port, on a manual clock at epoch.
type product struct {
	t      *testing.T
	url    string
	clk    *clock.ManualClock
	before int // the mails captured before this test's product started
}

// start wires the role as the binary does and runs its App until the test
// ends.
func start(t *testing.T, opts ...app.AppOption) *product {
	t.Helper()
	clk := clock.NewManualClock(epoch)
	server.Wire(t.Context(), server.Build{})
	a := server.App.With(append([]app.AppOption{app.InMemory(), app.Listen("127.0.0.1:0"), app.Env(app.EnvDev), app.Logs(io.Discard), app.Clock(clk)}, opts...)...)
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
	return &product{t: t, url: a.URL(), clk: clk, before: len(post.MailMailer.Captured())}
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

// secretIn is the one-time secret of a link.
var secretIn = regexp.MustCompile(`token=([A-Za-z0-9_-]+)`)

// link waits for a mail to addr whose text holds want, and returns the
// secret of the link in the newest one.
func (p *product) link(addr, want string) string {
	p.t.Helper()
	var found []model.MailMessage
	p.settle("a mail to "+addr+" with "+want, 50*time.Millisecond, func() bool {
		found = p.mails(addr, want)
		return len(found) > 0
	})
	match := secretIn.FindStringSubmatch(found[len(found)-1].Text)
	if match == nil {
		p.t.Fatalf("no link with a token in %q", found[len(found)-1].Text)
	}
	return match[1]
}

// signUp makes an account, proves its address with the link it is mailed,
// and signs it in.
func (p *product) signUp(name, email string) *member {
	p.t.Helper()
	m := p.stranger()
	creds := map[string]any{"email": email, "name": name, "password": passphrase}
	m.call(http.StatusOK, "POST", "/api/auth/signup", creds, nil)
	m.call(http.StatusOK, "POST", "/api/auth/verify", map[string]any{"token": p.link(strings.ToLower(email), "Confirmez votre adresse")}, nil)
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
	m.p.t.Helper()
	return m.retry(0, status, method, path, in, out)
}

// retry is call, after retries refusals of the rate limiter.
func (m *member) retry(retries, status int, method, path string, in, out any) http.Header {
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
	if resp.StatusCode == http.StatusTooManyRequests && status != http.StatusTooManyRequests && retries < 60 {
		// The limits are real and a test goes fast: the buckets refill on the
		// product's clock, so a refusal moves it by what one token costs the
		// slowest bucket, and the request is sent again.
		p.clk.Advance(500 * time.Millisecond)
		time.Sleep(2 * time.Millisecond)
		return m.retry(retries+1, status, method, path, in, out)
	}
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

// mails are the mails the product sent to addr since this test started whose
// subject or text holds want, oldest first: with no relay named, the mailer
// captures what it would deliver.
func (p *product) mails(addr, want string) []model.MailMessage {
	var out []model.MailMessage
	for _, m := range post.MailMailer.Captured()[p.before:] {
		text := strings.Join(strings.Fields(m.Subject+" "+m.Text), " ") // the text part is wrapped
		if slices.ContainsFunc(m.To, func(to string) bool { return strings.Contains(to, addr) }) && strings.Contains(text, want) {
			out = append(out, m)
		}
	}
	return out
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
	if moved.State != "done" || moved.DoneAt.Before(epoch) || moved.DoneAt.After(epoch.Add(time.Minute)) { // the clock moved a little while the mails went
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

// The jobs run on the product's clock: the vital signs are sampled, the owner of
// a task due soon is reminded of it once, and a task done for a day is archived.
func TestTheJobsRunOnTheClock(t *testing.T) {
	p := start(t)
	ada := p.signUp("Ada", "ada@example.com")
	var soon, other taskView
	ada.call(http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Ship the release", "assignee": "Grace", "due": epoch.Add(6 * time.Hour)}, &soon)
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

	reminders := func() []model.MailMessage { return p.mails("ada@example.com", "Bientôt à échéance") }
	p.settle("the reminder", 10*time.Second, func() bool { return len(reminders()) > 0 })
	p.clk.Advance(5 * time.Minute) // more runs of the job: the task is reminded of once
	time.Sleep(20 * time.Millisecond)
	if got := reminders(); len(got) != 1 || got[0].Subject != "Bientôt à échéance : Ship the release" {
		t.Errorf("the reminders: %+v, want the one", got)
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
	var feed struct {
		Entries []struct{ Kind, TaskID string }
	}
	p.settle("the archiving in the feed", 50*time.Millisecond, func() bool {
		ada.call(http.StatusOK, "GET", "/api/activity?limit=1", nil, &feed)
		return len(feed.Entries) == 1 && feed.Entries[0].Kind == "archive"
	})
	if feed.Entries[0].TaskID != other.ID {
		t.Errorf("the feed tells the archiving of %s, want %s", feed.Entries[0].TaskID, other.ID)
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
	// An address that has an account is answered like a new one, and its member told.
	var again struct{ Email, Status string }
	anon.call(http.StatusOK, "POST", "/api/auth/signup", map[string]any{"email": "ada@example.com", "name": "Someone", "password": "another passphrase"}, &again)
	if again.Email != "ada@example.com" || again.Status != "verification_sent" {
		t.Errorf("signing up twice answers %+v, want what a new address gets", again)
	}
	p.settle("the mail that says so", 50*time.Millisecond, func() bool {
		return len(p.mails("ada@example.com", "Vous avez déjà un compte")) == 1
	})
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

	// Her sessions: the one of this client, and the two the browser opened; she
	// ends one of the others, and nobody ends hers.
	var list struct {
		Sessions []struct {
			ID      string
			Current bool
		}
	}
	ada.call(http.StatusOK, "GET", "/api/auth/sessions", nil, &list)
	other := ""
	for _, s := range list.Sessions {
		if !s.Current {
			other = s.ID
		}
	}
	if len(list.Sessions) != 3 || other == "" {
		t.Fatalf("her sessions: %+v, want three, one of them current", list.Sessions)
	}
	bob := p.signUp("Bob", "bob@example.com")
	bob.call(http.StatusNotFound, "DELETE", "/api/auth/sessions/"+other, nil, nil)
	ada.call(http.StatusOK, "DELETE", "/api/auth/sessions/"+other, nil, nil)
	ada.call(http.StatusOK, "GET", "/api/auth/sessions", nil, &list)
	if len(list.Sessions) != 2 {
		t.Errorf("her sessions after one was ended: %+v, want two", list.Sessions)
	}

	ada.call(http.StatusOK, "POST", "/api/auth/logout", nil, nil)
	ada.call(http.StatusUnauthorized, "GET", "/api/auth/me", nil, nil)
}

// The load role, whole, drives the server role: it signs in as its own
// member and one round makes every call of the API, none refused (its exit
// status says so).
func TestTheLoadDrivesTheServer(t *testing.T) {
	p := start(t, app.Set("verify-addresses", false)) // the load reads no mail: its server trusts addresses
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

// Two members become contacts by a request one sends and the other
// accepts; each is told in their feed, a stranger to the link finds none,
// and either ends it.
func TestMembersBecomeContacts(t *testing.T) {
	p := start(t)
	ada, bob, cy := p.signUp("Ada", "ada@example.com"), p.signUp("Bob", "bob@example.com"), p.signUp("Cy", "cy@example.com")
	p.stranger().call(http.StatusUnauthorized, "GET", "/api/contacts", nil, nil)
	ada.call(http.StatusNotFound, "POST", "/api/contacts", map[string]any{"email": "nobody@example.com"}, nil)
	ada.call(http.StatusBadRequest, "POST", "/api/contacts", map[string]any{"email": "ada@example.com"}, nil)

	var l struct {
		ID, State string
	}
	ada.call(http.StatusOK, "POST", "/api/contacts", map[string]any{"email": "Bob@Example.com"}, &l)
	if l.ID == "" || l.State != "pending" {
		t.Fatalf("the request: %+v", l)
	}
	bob.call(http.StatusConflict, "POST", "/api/contacts", map[string]any{"email": "ada@example.com"}, nil) // already waiting
	cy.call(http.StatusNotFound, "POST", "/api/contacts/"+l.ID+"/accept", nil, nil)
	ada.call(http.StatusConflict, "POST", "/api/contacts/"+l.ID+"/accept", nil, nil) // hers to cancel, not to accept

	type listing struct {
		Contacts, Incoming, Outgoing []struct {
			ID   string
			User struct{ ID, Name, Email string }
		}
	}
	var mine, his listing
	ada.call(http.StatusOK, "GET", "/api/contacts", nil, &mine)
	bob.call(http.StatusOK, "GET", "/api/contacts", nil, &his)
	if len(mine.Outgoing) != 1 || mine.Outgoing[0].User.Name != "Bob" || len(his.Incoming) != 1 || his.Incoming[0].User.Email != "ada@example.com" {
		t.Fatalf("the request waiting: Ada sees %+v, Bob sees %+v", mine, his)
	}

	bob.call(http.StatusOK, "POST", "/api/contacts/"+l.ID+"/accept", nil, &l)
	if l.State != "accepted" {
		t.Fatalf("the request accepted: %+v", l)
	}
	ada.call(http.StatusOK, "GET", "/api/contacts", nil, &mine)
	if len(mine.Contacts) != 1 || mine.Contacts[0].User.ID != bob.id || len(mine.Outgoing) != 0 {
		t.Errorf("Ada's contacts: %+v, want Bob alone", mine)
	}

	var feed struct {
		Entries []struct{ Kind, Title string }
	}
	told := func(m *member, kind, by string) {
		t.Helper()
		p.settle(kind+" in the feed", 50*time.Millisecond, func() bool {
			m.call(http.StatusOK, "GET", "/api/activity", nil, &feed)
			return len(feed.Entries) > 0
		})
		if len(feed.Entries) != 1 || feed.Entries[0].Kind != kind || feed.Entries[0].Title != by {
			t.Errorf("the feed: %+v, want %s by %s alone", feed.Entries, kind, by)
		}
	}
	told(bob, "contact-requested", "Ada")
	told(ada, "contact-accepted", "Bob")

	cy.call(http.StatusNotFound, "DELETE", "/api/contacts/"+l.ID, nil, nil)
	bob.call(http.StatusOK, "DELETE", "/api/contacts/"+l.ID, nil, nil)
	ada.call(http.StatusOK, "GET", "/api/contacts", nil, &mine)
	if len(mine.Contacts) != 0 {
		t.Errorf("Ada still has %+v after Bob ended the contact", mine.Contacts)
	}
	ada.call(http.StatusOK, "POST", "/api/contacts", map[string]any{"email": "bob@example.com"}, nil) // and may ask again

	// A request nobody answers for fourteen days expires, and may be sent anew.
	p.settle("the request expired", 24*time.Hour, func() bool {
		ada.call(http.StatusOK, "GET", "/api/contacts", nil, &mine)
		return len(mine.Outgoing) == 0
	})
	if got := p.clk.Now().Sub(epoch); got < 14*24*time.Hour || got > 16*24*time.Hour {
		t.Errorf("the request expired after %s, want fourteen days", got)
	}
	ada.call(http.StatusOK, "POST", "/api/contacts", map[string]any{"email": "bob@example.com"}, nil)
}

// Signing in is limited per client: a burst of five, then two a second on
// the product's clock. The sixth attempt in a row is refused with a 429,
// and one more is served once the clock has moved.
func TestSigningInIsRateLimited(t *testing.T) {
	p := start(t)
	attempt := func() int {
		t.Helper()
		resp, err := http.Post(p.url+"/api/auth/login", "application/json", strings.NewReader(`{"email":"nobody@example.com","password":"whatever it is"}`))
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		return resp.StatusCode
	}
	for i := range 5 {
		if code := attempt(); code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: %d, want 401", i+1, code)
		}
	}
	if code := attempt(); code != http.StatusTooManyRequests {
		t.Fatalf("the sixth attempt in a row: %d, want 429", code)
	}
	p.clk.Advance(time.Second)
	if code := attempt(); code != http.StatusUnauthorized {
		t.Errorf("a second later: %d, want 401 again", code)
	}
}

// An account signs nobody in until its address is proven by the link it was
// mailed; a member who forgot their password chooses a new one by a link
// that works once, which ends their sessions; and one who knows theirs
// changes it.
func TestAnAddressIsProvenAndAPasswordReset(t *testing.T) {
	p := start(t)
	anon := p.stranger()
	creds := map[string]any{"email": "ada@example.com", "password": passphrase}
	anon.call(http.StatusOK, "POST", "/api/auth/signup", map[string]any{"email": "ada@example.com", "name": "Ada", "password": passphrase}, nil)
	anon.call(http.StatusForbidden, "POST", "/api/auth/login", creds, nil) // email_unverified
	first := p.link("ada@example.com", "Confirmez votre adresse")
	anon.call(http.StatusOK, "POST", "/api/auth/verify/resend", map[string]any{"email": "ada@example.com"}, nil)
	var second string
	p.settle("a second link", 50*time.Millisecond, func() bool {
		second = p.link("ada@example.com", "Confirmez votre adresse")
		return second != first
	})
	anon.call(http.StatusBadRequest, "POST", "/api/auth/verify", map[string]any{"token": first}, nil) // replaced by the second
	anon.call(http.StatusOK, "POST", "/api/auth/verify", map[string]any{"token": second}, nil)
	anon.call(http.StatusBadRequest, "POST", "/api/auth/verify", map[string]any{"token": second}, nil) // works once

	var in struct{ Token string }
	anon.call(http.StatusOK, "POST", "/api/auth/login", creds, &in)
	ada := &member{p: p, token: in.Token}
	ada.call(http.StatusOK, "GET", "/api/auth/me", nil, nil)

	// Forgotten: the same answer for an address nobody has, a link for Ada.
	anon.call(http.StatusOK, "POST", "/api/auth/password/forgot", map[string]any{"email": "nobody@example.com"}, nil)
	anon.call(http.StatusOK, "POST", "/api/auth/password/forgot", map[string]any{"email": "Ada@example.com"}, nil)
	reset := p.link("ada@example.com", "Choisissez un nouveau mot de passe")
	anon.call(http.StatusBadRequest, "POST", "/api/auth/password/reset", map[string]any{"token": reset, "password": "short"}, nil)
	anon.call(http.StatusBadRequest, "POST", "/api/auth/verify", map[string]any{"token": reset}, nil) // a reset link verifies nothing
	renewed := map[string]any{"email": "ada@example.com", "password": "a brand new passphrase"}
	anon.call(http.StatusOK, "POST", "/api/auth/password/reset", map[string]any{"token": reset, "password": renewed["password"]}, nil)
	anon.call(http.StatusBadRequest, "POST", "/api/auth/password/reset", map[string]any{"token": reset, "password": "yet another passphrase"}, nil)
	ada.call(http.StatusUnauthorized, "GET", "/api/auth/me", nil, nil) // her session ended with the reset
	anon.call(http.StatusUnauthorized, "POST", "/api/auth/login", creds, nil)
	anon.call(http.StatusOK, "POST", "/api/auth/login", renewed, &in)
	ada.token = in.Token

	// Changed: by who knows the current one, and the other sessions end.
	anon.call(http.StatusOK, "POST", "/api/auth/login", renewed, &in)
	elsewhere := &member{p: p, token: in.Token}
	ada.call(http.StatusUnauthorized, "POST", "/api/auth/password", map[string]any{"current": "not the right one", "password": "one more passphrase"}, nil)
	ada.call(http.StatusOK, "POST", "/api/auth/password", map[string]any{"current": renewed["password"], "password": "one more passphrase"}, nil)
	ada.call(http.StatusOK, "GET", "/api/auth/me", nil, nil)
	elsewhere.call(http.StatusUnauthorized, "GET", "/api/auth/me", nil, nil)
}

// A task is shared with a contact of its owner, who then sees it, edits it
// and moves it, and is told of it in their feed; only the owner shares and
// deletes; the member it is shared with may leave it.
func TestATaskIsSharedWithAContact(t *testing.T) {
	p := start(t)
	ada, bob, cy := p.signUp("Ada", "ada@example.com"), p.signUp("Bob", "bob@example.com"), p.signUp("Cy", "cy@example.com")
	var l struct{ ID string }
	ada.call(http.StatusOK, "POST", "/api/contacts", map[string]any{"email": "bob@example.com"}, &l)
	bob.call(http.StatusOK, "POST", "/api/contacts/"+l.ID+"/accept", nil, nil)

	type shared struct {
		ID, Title, Priority, State, OwnerID string
		SharedWith                          []string
	}
	var tk shared
	ada.call(http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Plan the launch"}, &tk)
	bob.call(http.StatusNotFound, "GET", "/api/tasks/"+tk.ID, nil, nil)
	ada.call(http.StatusConflict, "POST", "/api/tasks/"+tk.ID+"/share", map[string]any{"userId": cy.id}, nil) // no contact of hers
	ada.call(http.StatusOK, "POST", "/api/tasks/"+tk.ID+"/share", map[string]any{"userId": bob.id}, &tk)
	if len(tk.SharedWith) != 1 || tk.SharedWith[0] != bob.id {
		t.Fatalf("the task shared: %+v", tk)
	}

	var list struct{ Tasks []shared }
	bob.call(http.StatusOK, "GET", "/api/tasks", nil, &list)
	if len(list.Tasks) != 1 || list.Tasks[0].ID != tk.ID || list.Tasks[0].OwnerID != ada.id {
		t.Fatalf("Bob's list: %+v, want Ada's task", list.Tasks)
	}
	bob.call(http.StatusOK, "PATCH", "/api/tasks/"+tk.ID, map[string]any{"priority": "high"}, &tk)
	if tk.Priority != "high" || tk.Title != "Plan the launch" {
		t.Errorf("the task Bob edited: %+v, want the priority alone changed", tk)
	}
	bob.call(http.StatusBadRequest, "PATCH", "/api/tasks/"+tk.ID, map[string]any{"title": " "}, nil)
	bob.call(http.StatusOK, "POST", "/api/tasks/"+tk.ID+"/complete", nil, &tk)
	bob.call(http.StatusForbidden, "POST", "/api/tasks/"+tk.ID+"/share", map[string]any{"userId": cy.id}, nil)
	bob.call(http.StatusForbidden, "DELETE", "/api/tasks/"+tk.ID, nil, nil)
	cy.call(http.StatusNotFound, "DELETE", "/api/tasks/"+tk.ID, nil, nil)

	var feed struct {
		Entries []struct{ Kind, TaskID string }
	}
	p.settle("the sharing, the edit and the move in Bob's feed", 50*time.Millisecond, func() bool {
		bob.call(http.StatusOK, "GET", "/api/activity", nil, &feed)
		return len(feed.Entries) >= 4 // accepted contact aside: shared, edited, complete
	})
	var kinds []string
	for _, e := range feed.Entries {
		if e.TaskID == tk.ID {
			kinds = append(kinds, e.Kind)
		}
	}
	if !slices.Equal(kinds, []string{"complete", "edited", "shared"}) {
		t.Errorf("what Bob's feed tells of the task, newest first: %q", kinds)
	}

	bob.call(http.StatusOK, "DELETE", "/api/tasks/"+tk.ID+"/share/"+bob.id, nil, nil) // he leaves it
	bob.call(http.StatusNotFound, "GET", "/api/tasks/"+tk.ID, nil, nil)
	ada.call(http.StatusOK, "DELETE", "/api/tasks/"+tk.ID, nil, nil)
	ada.call(http.StatusNotFound, "GET", "/api/tasks/"+tk.ID, nil, nil)
}
