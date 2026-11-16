package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
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
	// The role's dev configuration trusts addresses; a test proves them, as a
	// deployment with a mail relay does, unless it says otherwise.
	opts = append([]app.AppOption{app.Set("verify-addresses", true)}, opts...)
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

// signUp makes an account and proves its address with the link it is
// mailed, which signs it in.
func (p *product) signUp(name, email string) *member {
	p.t.Helper()
	m := p.stranger()
	creds := map[string]any{"email": email, "name": name, "password": passphrase}
	m.call(http.StatusOK, "POST", "/api/auth/signup", creds, nil)
	// The link signs its member in: no password is stretched a second time.
	var in struct {
		Token string `json:"token"`
		User  struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	m.call(http.StatusOK, "POST", "/api/auth/verify", map[string]any{"token": p.link(strings.ToLower(email), "Confirmez votre adresse")}, &in)
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
	ada.call(http.StatusOK, "POST", "/api/tasks", map[string]any{"title": " Write the design ", "priority": "high", "assigneeId": ada.id}, &added)
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
	ada.call(http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Ship the release", "due": epoch.Add(6 * time.Hour)}, &soon)
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

// groupView is a group as the API answers it.
type groupView struct {
	ID, Name, Color, Role string
	Members               []struct {
		User struct{ ID, Name string }
		Role string
	}
}

// befriend makes two members contacts: the first asks, the second accepts.
func (m *member) befriend(other *member, otherEmail string) {
	m.p.t.Helper()
	var l struct{ ID string }
	m.call(http.StatusOK, "POST", "/api/contacts", map[string]any{"email": otherEmail}, &l)
	other.call(http.StatusOK, "POST", "/api/contacts/"+l.ID+"/accept", nil, nil)
}

// roles are who is in a group and what each is, by name.
func (g groupView) roles() map[string]string {
	out := map[string]string{}
	for _, m := range g.Members {
		out[m.User.Name] = m.Role
	}
	return out
}

// Members form a group: its creator owns it, invites only their contacts, an
// invitation is answered by its invitee alone, roles bound what each may do,
// members leave or are removed, each is told in their feed, and an
// invitation nobody answers for a week expires.
func TestMembersFormAGroup(t *testing.T) {
	p := start(t)
	ada, bob, cy, dee := p.signUp("Ada", "ada@example.com"), p.signUp("Bob", "bob@example.com"), p.signUp("Cy", "cy@example.com"), p.signUp("Dee", "dee@example.com")
	p.stranger().call(http.StatusUnauthorized, "GET", "/api/groups", nil, nil)
	p.stranger().call(http.StatusUnauthorized, "GET", "/api/invitations", nil, nil)

	// A group is created by a name of one line and a color of the palette, or none.
	ada.call(http.StatusBadRequest, "POST", "/api/groups", map[string]any{"name": "  "}, nil)
	ada.call(http.StatusBadRequest, "POST", "/api/groups", map[string]any{"name": "a\nb"}, nil)
	ada.call(http.StatusBadRequest, "POST", "/api/groups", map[string]any{"name": "Core", "color": "chartreuse"}, nil)
	var g groupView
	ada.call(http.StatusOK, "POST", "/api/groups", map[string]any{"name": " Core ", "color": "blue"}, &g)
	if g.ID == "" || g.Name != "Core" || g.Color != "blue" || g.Role != "owner" || !maps.Equal(g.roles(), map[string]string{"Ada": "owner"}) {
		t.Fatalf("the group created: %+v", g)
	}
	var picked groupView
	ada.call(http.StatusOK, "POST", "/api/groups", map[string]any{"name": "Side"}, &picked)
	if picked.Color == "" {
		t.Errorf("a group with no color has one picked from its name: %+v", picked)
	}
	bob.call(http.StatusNotFound, "GET", "/api/groups/"+g.ID, nil, nil) // not in it: as if there were none
	var list struct{ Groups []groupView }
	ada.call(http.StatusOK, "GET", "/api/groups", nil, &list)
	if len(list.Groups) != 2 || list.Groups[0].Name != "Core" || list.Groups[1].Name != "Side" {
		t.Errorf("Ada's groups: %+v, want Core then Side", list.Groups)
	}
	bob.call(http.StatusOK, "GET", "/api/groups", nil, &list)
	if len(list.Groups) != 0 {
		t.Errorf("Bob's groups: %+v, want none", list.Groups)
	}

	// Only a contact is invited, once, and only the invitee answers.
	ada.befriend(bob, "bob@example.com")
	ada.befriend(cy, "cy@example.com")
	invite := func(by *member, group string, who *member, status int) (inv struct {
		ID, State string
		Group     struct{ ID, Name string }
		Inviter   struct{ Name string }
		Invitee   struct{ Name string }
	}) {
		t.Helper()
		by.call(status, "POST", "/api/groups/"+group+"/invitations", map[string]any{"userId": who.id}, &inv)
		return inv
	}
	invite(ada, g.ID, dee, http.StatusForbidden) // no contact of hers
	invite(ada, g.ID, ada, http.StatusConflict)  // already in
	invite(bob, g.ID, cy, http.StatusNotFound)   // not in the group yet
	ada.call(http.StatusBadRequest, "POST", "/api/groups/"+g.ID+"/invitations", map[string]any{}, nil)
	inv := invite(ada, g.ID, bob, http.StatusOK)
	if inv.ID == "" || inv.State != "pending" || inv.Group.Name != "Core" || inv.Inviter.Name != "Ada" || inv.Invitee.Name != "Bob" {
		t.Fatalf("the invitation: %+v", inv)
	}
	invite(ada, g.ID, bob, http.StatusConflict) // already waiting

	var invitations struct {
		Invitations []struct {
			ID, State string
			Group     struct{ ID, Name string }
			Inviter   struct{ Name string }
		}
	}
	bob.call(http.StatusOK, "GET", "/api/invitations", nil, &invitations)
	if len(invitations.Invitations) != 1 || invitations.Invitations[0].ID != inv.ID || invitations.Invitations[0].Group.Name != "Core" || invitations.Invitations[0].Inviter.Name != "Ada" {
		t.Fatalf("Bob's invitations: %+v", invitations.Invitations)
	}
	cy.call(http.StatusOK, "GET", "/api/invitations", nil, &invitations)
	if len(invitations.Invitations) != 0 {
		t.Errorf("Cy sees the invitation of Bob: %+v", invitations.Invitations)
	}
	cy.call(http.StatusNotFound, "POST", "/api/invitations/"+inv.ID+"/accept", nil, nil)
	ada.call(http.StatusNotFound, "POST", "/api/invitations/"+inv.ID+"/decline", nil, nil)
	bob.call(http.StatusNotFound, "POST", "/api/invitations/invitation_nobody/accept", nil, nil)
	bob.call(http.StatusOK, "POST", "/api/invitations/"+inv.ID+"/accept", nil, &g)
	if g.Role != "member" || !maps.Equal(g.roles(), map[string]string{"Ada": "owner", "Bob": "member"}) {
		t.Fatalf("the group Bob joined: %+v", g)
	}
	bob.call(http.StatusConflict, "POST", "/api/invitations/"+inv.ID+"/accept", nil, nil)
	bob.call(http.StatusConflict, "POST", "/api/invitations/"+inv.ID+"/decline", nil, nil)

	// A declined invitation keeps nobody out of a later one.
	inv = invite(ada, g.ID, cy, http.StatusOK)
	cy.call(http.StatusOK, "POST", "/api/invitations/"+inv.ID+"/decline", nil, nil)
	cy.call(http.StatusNotFound, "GET", "/api/groups/"+g.ID, nil, nil)
	cy.call(http.StatusConflict, "POST", "/api/invitations/"+inv.ID+"/accept", nil, nil)
	inv = invite(ada, g.ID, cy, http.StatusOK)
	cy.call(http.StatusOK, "POST", "/api/invitations/"+inv.ID+"/accept", nil, &g)
	if !maps.Equal(g.roles(), map[string]string{"Ada": "owner", "Bob": "member", "Cy": "member"}) {
		t.Fatalf("the group Cy joined: %+v", g)
	}
	if g.Members[0].User.Name != "Ada" || g.Members[1].User.Name != "Bob" || g.Members[2].User.Name != "Cy" {
		t.Errorf("the members are ranked, the owner first, then by when they joined: %+v", g.Members)
	}

	// What a role may do: a member edits nothing, the owner alone gives roles,
	// the owner's own never changes, an admin runs the group.
	bob.call(http.StatusForbidden, "PATCH", "/api/groups/"+g.ID, map[string]any{"name": "Mine"}, nil)
	bob.call(http.StatusForbidden, "POST", "/api/groups/"+g.ID+"/invitations", map[string]any{"userId": dee.id}, nil)
	bob.call(http.StatusForbidden, "PATCH", "/api/groups/"+g.ID+"/members/"+cy.id, map[string]any{"role": "admin"}, nil)
	dee.call(http.StatusNotFound, "PATCH", "/api/groups/"+g.ID, map[string]any{"name": "Mine"}, nil)
	ada.call(http.StatusBadRequest, "PATCH", "/api/groups/"+g.ID, map[string]any{"name": " "}, nil)
	ada.call(http.StatusBadRequest, "PATCH", "/api/groups/"+g.ID, map[string]any{"color": ""}, nil)
	ada.call(http.StatusBadRequest, "PATCH", "/api/groups/"+g.ID, map[string]any{"color": "chartreuse"}, nil)
	ada.call(http.StatusOK, "PATCH", "/api/groups/"+g.ID, map[string]any{"name": "Platform"}, &g)
	if g.Name != "Platform" || g.Color != "blue" || len(g.Members) != 3 {
		t.Errorf("a rename leaves the rest: %+v", g)
	}
	ada.call(http.StatusBadRequest, "PATCH", "/api/groups/"+g.ID+"/members/"+bob.id, map[string]any{"role": "owner"}, nil)
	ada.call(http.StatusBadRequest, "PATCH", "/api/groups/"+g.ID+"/members/"+bob.id, map[string]any{}, nil)
	ada.call(http.StatusForbidden, "PATCH", "/api/groups/"+g.ID+"/members/"+ada.id, map[string]any{"role": "member"}, nil)
	ada.call(http.StatusNotFound, "PATCH", "/api/groups/"+g.ID+"/members/"+dee.id, map[string]any{"role": "admin"}, nil)
	ada.call(http.StatusOK, "PATCH", "/api/groups/"+g.ID+"/members/"+bob.id, map[string]any{"role": "admin"}, &g)
	if !maps.Equal(g.roles(), map[string]string{"Ada": "owner", "Bob": "admin", "Cy": "member"}) {
		t.Fatalf("the roles after Bob is made an admin: %+v", g)
	}
	bob.call(http.StatusOK, "PATCH", "/api/groups/"+g.ID, map[string]any{"color": "red"}, &g)
	if g.Color != "red" || g.Role != "admin" {
		t.Errorf("the group an admin recolored: %+v", g)
	}
	bob.befriend(dee, "dee@example.com")
	inv = invite(bob, g.ID, dee, http.StatusOK) // an admin invites
	dee.call(http.StatusOK, "POST", "/api/invitations/"+inv.ID+"/accept", nil, &g)
	if g.Role != "member" || len(g.Members) != 4 {
		t.Fatalf("the group Dee joined: %+v", g)
	}

	// Who may remove whom: nobody removes the owner, an admin only members,
	// anyone but the owner leaves, and the owner deletes the group instead.
	bob.call(http.StatusForbidden, "DELETE", "/api/groups/"+g.ID+"/members/"+ada.id, nil, nil)
	cy.call(http.StatusForbidden, "DELETE", "/api/groups/"+g.ID+"/members/"+bob.id, nil, nil)
	ada.call(http.StatusConflict, "DELETE", "/api/groups/"+g.ID+"/members/"+ada.id, nil, nil)
	ada.call(http.StatusNotFound, "DELETE", "/api/groups/"+g.ID+"/members/user_nobody", nil, nil)
	p.stranger().call(http.StatusUnauthorized, "DELETE", "/api/groups/"+g.ID+"/members/"+cy.id, nil, nil)
	bob.call(http.StatusOK, "DELETE", "/api/groups/"+g.ID+"/members/"+dee.id, nil, nil) // removed
	dee.call(http.StatusNotFound, "GET", "/api/groups/"+g.ID, nil, nil)
	cy.call(http.StatusOK, "DELETE", "/api/groups/"+g.ID+"/members/"+cy.id, nil, nil) // leaves
	cy.call(http.StatusNotFound, "GET", "/api/groups/"+g.ID, nil, nil)
	ada.call(http.StatusOK, "GET", "/api/groups/"+g.ID, nil, &g)
	if !maps.Equal(g.roles(), map[string]string{"Ada": "owner", "Bob": "admin"}) {
		t.Errorf("the group after Dee was removed and Cy left: %+v", g)
	}

	// Each member is told what concerns them, and not what they did themselves.
	kindsIn := func(m *member, want ...string) []string {
		t.Helper()
		var kinds []string
		p.settle("the group entries in a feed", 50*time.Millisecond, func() bool {
			var feed struct {
				Entries []struct{ Kind, Title string }
			}
			m.call(http.StatusOK, "GET", "/api/activity", nil, &feed)
			kinds = kinds[:0]
			for _, e := range feed.Entries {
				if strings.HasPrefix(e.Kind, "group-") {
					kinds = append(kinds, e.Kind+" "+e.Title)
				}
			}
			return len(kinds) >= len(want)
		})
		slices.Sort(kinds)
		slices.Sort(want)
		return kinds
	}
	check := func(who string, got []string, want ...string) {
		t.Helper()
		if !slices.Equal(got, want) {
			t.Errorf("%s's feed: %q, want %q", who, got, want)
		}
	}
	adaTold := []string{"group-joined Core", "group-joined Core", "group-joined Platform", "group-removed Platform", "group-left Platform"}
	check("Ada", kindsIn(ada, adaTold...), "group-joined Core", "group-joined Core", "group-joined Platform", "group-left Platform", "group-removed Platform")
	check("Dee", kindsIn(dee, "group-invited Platform", "group-removed Platform"), "group-invited Platform", "group-removed Platform")
	check("Bob", kindsIn(bob, "group-invited Core", "group-joined Core", "group-joined Platform", "group-left Platform", "group-role Platform"),
		"group-invited Core", "group-joined Core", "group-joined Platform", "group-left Platform", "group-role Platform")

	// A group is deleted by its owner alone, and the invitations into it go
	// with it.
	side := picked.ID
	inv = invite(ada, side, bob, http.StatusOK)
	bob.call(http.StatusOK, "POST", "/api/invitations/"+inv.ID+"/accept", nil, nil)
	inv = invite(ada, side, cy, http.StatusOK) // waits when the group goes
	bob.call(http.StatusForbidden, "DELETE", "/api/groups/"+g.ID, nil, nil)
	dee.call(http.StatusNotFound, "DELETE", "/api/groups/"+g.ID, nil, nil)
	bob.call(http.StatusForbidden, "DELETE", "/api/groups/"+side, nil, nil)
	ada.call(http.StatusOK, "DELETE", "/api/groups/"+side, nil, nil)
	ada.call(http.StatusNotFound, "GET", "/api/groups/"+side, nil, nil)
	cy.call(http.StatusNotFound, "POST", "/api/invitations/"+inv.ID+"/accept", nil, nil)
	cy.call(http.StatusOK, "GET", "/api/invitations", nil, &invitations)
	if len(invitations.Invitations) != 0 {
		t.Errorf("the invitations into a deleted group: %+v", invitations.Invitations)
	}
	p.settle("the deletion in Bob's feed", 50*time.Millisecond, func() bool {
		var feed struct {
			Entries []struct{ Kind, Title string }
		}
		bob.call(http.StatusOK, "GET", "/api/activity", nil, &feed)
		return len(feed.Entries) > 0 && feed.Entries[0].Kind == "group-deleted" && feed.Entries[0].Title == "Side"
	})

	// An invitation nobody answers for a week expires, and may be sent anew.
	inv = invite(bob, g.ID, dee, http.StatusOK)
	invitedAt := p.clk.Now()
	p.settle("the invitation expired", 24*time.Hour, func() bool {
		dee.call(http.StatusOK, "GET", "/api/invitations", nil, &invitations)
		return len(invitations.Invitations) == 0
	})
	if got := p.clk.Now().Sub(invitedAt); got < 7*24*time.Hour || got > 9*24*time.Hour {
		t.Errorf("the invitation expired after %s, want a week", got)
	}
	dee.call(http.StatusConflict, "POST", "/api/invitations/"+inv.ID+"/accept", nil, nil)
	invite(bob, g.ID, dee, http.StatusOK)
}

// groupTask is a task as a group's members read it.
type groupTask struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Priority   string `json:"priority"`
	State      string `json:"state"`
	OwnerID    string `json:"ownerId"`
	GroupID    string `json:"groupId"`
	AssigneeID string `json:"assigneeId"`
}

// viewCounts is what GET /api/tasks/counts answers.
type viewCounts struct {
	Inbox, Today, Upcoming, Shared, Assigned, Overdue, CompletedThisWeek int
	Groups                                                               map[string]int
}

// A group shares a list: a task added to it is seen, edited and moved by
// every member and by nobody else; its owner or an admin of the group deletes
// it, its owner alone shares it; it is assigned to its owner, to a member of
// the group, or, when it is personal, to a contact, who then sees it; the
// views and the counts follow what each member sees; a member removed from
// the group stops doing its tasks, and a group deleted gives its tasks back
// to their owners.
func TestAGroupSharesAList(t *testing.T) {
	p := start(t)
	ada, bob, cy, dee, eve := p.signUp("Ada", "ada@example.com"), p.signUp("Bob", "bob@example.com"), p.signUp("Cy", "cy@example.com"), p.signUp("Dee", "dee@example.com"), p.signUp("Eve", "eve@example.com")
	ada.befriend(bob, "bob@example.com")
	ada.befriend(cy, "cy@example.com")
	ada.befriend(eve, "eve@example.com")
	var g groupView
	ada.call(http.StatusOK, "POST", "/api/groups", map[string]any{"name": "Core"}, &g)
	join := func(who *member, role string) {
		t.Helper()
		var inv struct{ ID string }
		ada.call(http.StatusOK, "POST", "/api/groups/"+g.ID+"/invitations", map[string]any{"userId": who.id}, &inv)
		who.call(http.StatusOK, "POST", "/api/invitations/"+inv.ID+"/accept", nil, nil)
		if role != "member" {
			ada.call(http.StatusOK, "PATCH", "/api/groups/"+g.ID+"/members/"+who.id, map[string]any{"role": role}, nil)
		}
	}
	join(bob, "admin")
	join(cy, "member")

	// Only a member adds a task to the group's list, and assigns it to who is
	// in the group; a personal task is assigned to a member it is shared with.
	soon, later := epoch.Add(3*time.Hour), epoch.Add(72*time.Hour)
	var first, second, mine groupTask
	dee.call(http.StatusBadRequest, "POST", "/api/tasks", map[string]any{"title": "x", "groupId": g.ID}, nil)
	ada.call(http.StatusBadRequest, "POST", "/api/tasks", map[string]any{"title": "x", "groupId": g.ID, "assigneeId": eve.id}, nil) // a contact, not in the group
	ada.call(http.StatusBadRequest, "POST", "/api/tasks", map[string]any{"title": "x", "assigneeId": dee.id}, nil)                  // sees nothing of it
	ada.call(http.StatusBadRequest, "POST", "/api/tasks", map[string]any{"title": "x", "assigneeId": eve.id}, nil)                  // a contact, with whom it is not shared
	ada.call(http.StatusBadRequest, "POST", "/api/tasks", map[string]any{"title": "x", "assignee": "Grace"}, nil)                   // the free text is gone
	ada.call(http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Plan the sprint", "groupId": g.ID, "assigneeId": bob.id, "due": soon}, &first)
	ada.call(http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Write the notes", "groupId": g.ID, "due": later}, &second)
	ada.call(http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Water the plants"}, &mine)
	ada.call(http.StatusOK, "POST", "/api/tasks/"+mine.ID+"/share", map[string]any{"userId": eve.id}, nil)
	ada.call(http.StatusOK, "PATCH", "/api/tasks/"+mine.ID, map[string]any{"assigneeId": eve.id}, &mine) // shared with her first, then assigned
	if first.GroupID != g.ID || first.AssigneeID != bob.id || first.OwnerID != ada.id || mine.AssigneeID != eve.id || mine.GroupID != "" {
		t.Fatalf("the tasks added: %+v %+v", first, mine)
	}
	eve.call(http.StatusOK, "GET", "/api/tasks/"+mine.ID, nil, nil) // her assignee sees it
	dee.call(http.StatusNotFound, "GET", "/api/tasks/"+mine.ID, nil, nil)
	cy.call(http.StatusNotFound, "GET", "/api/tasks/"+mine.ID, nil, nil)

	// Every member sees the group's tasks, nobody else does.
	var list struct{ Tasks []groupTask }
	for _, m := range []*member{ada, bob, cy} {
		m.call(http.StatusOK, "GET", "/api/tasks/"+first.ID, nil, nil)
	}
	dee.call(http.StatusNotFound, "GET", "/api/tasks/"+first.ID, nil, nil)
	eve.call(http.StatusNotFound, "GET", "/api/tasks/"+first.ID, nil, nil)
	cy.call(http.StatusOK, "GET", "/api/tasks", nil, &list)
	if len(list.Tasks) != 2 {
		t.Errorf("Cy's list: %+v, want the two tasks of the group", list.Tasks)
	}
	dee.call(http.StatusOK, "GET", "/api/tasks", nil, &list)
	if len(list.Tasks) != 0 {
		t.Errorf("Dee's list: %+v, want none", list.Tasks)
	}

	// The views and the counts are each member's.
	bob.call(http.StatusBadRequest, "GET", "/api/tasks?view=someday", nil, nil)
	bob.call(http.StatusBadRequest, "GET", "/api/tasks?view=group", nil, nil)
	bob.call(http.StatusBadRequest, "GET", "/api/tasks?view=today&tz=Mars/Olympus", nil, nil)
	bob.call(http.StatusBadRequest, "GET", "/api/tasks/counts?tz=Mars/Olympus", nil, nil)
	dee.call(http.StatusNotFound, "GET", "/api/tasks?group="+g.ID, nil, nil)
	titlesOf := func(m *member, query string) []string {
		t.Helper()
		var out struct{ Tasks []groupTask }
		m.call(http.StatusOK, "GET", "/api/tasks?"+query, nil, &out)
		var titles []string
		for _, x := range out.Tasks {
			titles = append(titles, x.Title)
		}
		return titles
	}
	for query, want := range map[string][]string{
		"view=today&tz=Europe/Paris":  {"Plan the sprint"},
		"view=upcoming":               {"Write the notes"},
		"view=assigned":               {"Plan the sprint"},
		"view=inbox":                  nil,
		"view=shared":                 nil,
		"view=completed":              nil,
		"group=" + g.ID:               {"Plan the sprint", "Write the notes"},
		"view=group&group=" + g.ID:    {"Plan the sprint", "Write the notes"},
		"view=assigned&group=" + g.ID: {"Plan the sprint"},
		"view=today&group=" + g.ID + "&tz=Europe/Paris": {"Plan the sprint"},
	} {
		if got := titlesOf(bob, query); !slices.Equal(got, want) {
			t.Errorf("Bob's %s: %q, want %q", query, got, want)
		}
	}
	countsOf := func(m *member, query string) (counts viewCounts) {
		t.Helper()
		m.call(http.StatusOK, "GET", "/api/tasks/counts"+query, nil, &counts)
		return counts
	}
	counts := countsOf(bob, "?tz=Europe/Paris")
	if counts.Today != 1 || counts.Upcoming != 1 || counts.Assigned != 1 || counts.Inbox != 0 || counts.Shared != 0 || counts.Overdue != 0 || counts.Groups[g.ID] != 2 {
		t.Errorf("Bob's counts: %+v, want one today, one upcoming, one assigned and two on the group", counts)
	}
	counts = countsOf(ada, "")
	if counts.Inbox != 1 || counts.Groups[g.ID] != 2 { // the personal task, and the two of the group
		t.Errorf("Ada's counts: %+v, want one in the inbox and two on the group", counts)
	}
	counts = countsOf(dee, "")
	if counts.Inbox != 0 || len(counts.Groups) != 0 {
		t.Errorf("Dee's counts: %+v, want nothing", counts)
	}

	// Every member edits and moves, and may assign a task to one in the group;
	// only the owner shares it; the owner or an admin deletes it.
	var edited groupTask
	cy.call(http.StatusOK, "PATCH", "/api/tasks/"+first.ID, map[string]any{"priority": "high"}, &edited)
	if edited.Priority != "high" || edited.AssigneeID != bob.id || edited.GroupID != g.ID {
		t.Errorf("the task Cy edited: %+v, want the priority alone changed", edited)
	}
	cy.call(http.StatusBadRequest, "PATCH", "/api/tasks/"+first.ID, map[string]any{"assigneeId": dee.id}, nil)
	cy.call(http.StatusBadRequest, "PATCH", "/api/tasks/"+first.ID, map[string]any{"assigneeId": eve.id}, nil) // a contact of Ada's, and of nobody in the group
	cy.call(http.StatusOK, "PATCH", "/api/tasks/"+first.ID, map[string]any{"assigneeId": cy.id}, &edited)
	cy.call(http.StatusOK, "PATCH", "/api/tasks/"+first.ID, map[string]any{"assigneeId": bob.id}, &edited)
	cy.call(http.StatusOK, "POST", "/api/tasks/"+second.ID+"/complete", nil, &edited)
	cy.call(http.StatusForbidden, "POST", "/api/tasks/"+first.ID+"/share", map[string]any{"userId": dee.id}, nil)
	bob.call(http.StatusForbidden, "POST", "/api/tasks/"+first.ID+"/share", map[string]any{"userId": dee.id}, nil)
	cy.call(http.StatusForbidden, "DELETE", "/api/tasks/"+first.ID, nil, nil)
	dee.call(http.StatusNotFound, "DELETE", "/api/tasks/"+first.ID, nil, nil)
	if got := titlesOf(cy, "view=completed"); !slices.Equal(got, []string{"Write the notes"}) {
		t.Errorf("Cy's completed tasks: %q", got)
	}
	counts = countsOf(cy, "")
	if counts.CompletedThisWeek != 1 || counts.Groups[g.ID] != 1 {
		t.Errorf("Cy's counts: %+v, want one finished this week and one open on the group", counts)
	}
	var gone groupTask
	cy.call(http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Tidy the board", "groupId": g.ID}, &gone)
	bob.call(http.StatusOK, "DELETE", "/api/tasks/"+gone.ID, nil, nil) // an admin deletes a member's task
	cy.call(http.StatusNotFound, "GET", "/api/tasks/"+gone.ID, nil, nil)

	// Each member of the group is told of its tasks, and nobody else.
	feedKinds := func(m *member, want int) []string {
		t.Helper()
		var kinds []string
		p.settle("the group's tasks in a feed", 50*time.Millisecond, func() bool {
			var feed struct {
				Entries []struct{ Kind, TaskID string }
			}
			m.call(http.StatusOK, "GET", "/api/activity", nil, &feed)
			kinds = kinds[:0]
			for _, e := range feed.Entries {
				if e.TaskID == first.ID {
					kinds = append(kinds, e.Kind)
				}
			}
			return len(kinds) >= want
		})
		return kinds
	}
	if got := feedKinds(cy, 2); got[len(got)-1] != "added" {
		t.Errorf("what Cy's feed tells of the first task, newest first: %q, want it to end with added", got)
	}
	if got := feedKinds(dee, 0); len(got) != 0 {
		t.Errorf("what Dee's feed tells of the first task: %q, want nothing", got)
	}

	// A member removed from the group stops doing its tasks and no more sees
	// them; the tasks of a deleted group go back to their owners.
	ada.call(http.StatusOK, "DELETE", "/api/groups/"+g.ID+"/members/"+bob.id, nil, nil)
	// The tasks follow the group by a message: until it is heard, Bob still
	// sees the task he is assigned, as its assignee.
	p.settle("the task unassigned from Bob", 50*time.Millisecond, func() bool {
		ada.call(http.StatusOK, "GET", "/api/tasks/"+first.ID, nil, &edited)
		return edited.AssigneeID == ""
	})
	bob.call(http.StatusNotFound, "GET", "/api/tasks/"+first.ID, nil, nil)
	if edited.GroupID != g.ID {
		t.Errorf("the task Bob stopped doing: %+v, want it still on the group", edited)
	}
	cy.call(http.StatusOK, "PATCH", "/api/tasks/"+first.ID, map[string]any{"assigneeId": cy.id}, nil)
	ada.call(http.StatusConflict, "DELETE", "/api/groups/"+g.ID+"/members/"+ada.id, nil, nil)
	cy.call(http.StatusForbidden, "DELETE", "/api/groups/"+g.ID, nil, nil)
	ada.call(http.StatusOK, "DELETE", "/api/groups/"+g.ID, nil, nil)
	p.settle("the tasks back on their owner's list", 50*time.Millisecond, func() bool {
		ada.call(http.StatusOK, "GET", "/api/tasks/"+first.ID, nil, &edited)
		return edited.GroupID == ""
	})
	if edited.AssigneeID != "" || edited.OwnerID != ada.id {
		t.Errorf("the task of the deleted group: %+v, want it Ada's alone, unassigned", edited)
	}
	cy.call(http.StatusNotFound, "GET", "/api/tasks/"+first.ID, nil, nil)
	cy.call(http.StatusNotFound, "GET", "/api/tasks/"+second.ID, nil, nil)
	if got := titlesOf(ada, "view=inbox"); !slices.Equal(got, []string{"Plan the sprint", "Water the plants"}) { // the one Cy made high first
		t.Errorf("Ada's inbox: %q, want the two open tasks back", got)
	}
	ada.call(http.StatusNotFound, "GET", "/api/tasks?group="+g.ID, nil, nil)
}

// status sends one request and returns its status, with no retry and no
// failure of its own: what a goroutine of a test calls.
func (m *member) status(method, path string, in any) int {
	var body io.Reader
	if in != nil {
		raw, _ := json.Marshal(in)
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, m.p.url+path, body)
	if err != nil {
		return 0
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if m.token != "" {
		req.Header.Set("Authorization", "Bearer "+m.token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	return resp.StatusCode
}

// together runs every call at the same instant and returns their statuses.
func together(calls ...func() int) []int {
	out := make([]int, len(calls))
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i, call := range calls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			out[i] = call()
		}()
	}
	close(start)
	wg.Wait()
	return out
}

// Invitees who accept at the same instant all join: no acceptance writes the
// group back over another's.
func TestAcceptancesAtTheSameInstantAllJoin(t *testing.T) {
	p := start(t)
	ada := p.signUp("Ada", "ada@example.com")
	var g groupView
	ada.call(http.StatusOK, "POST", "/api/groups", map[string]any{"name": "Core"}, &g)
	var accept []func() int
	for _, name := range []string{"bob", "cy", "dee", "eve", "fay"} {
		who := p.signUp(name, name+"@example.com")
		ada.befriend(who, name+"@example.com")
		var inv struct{ ID string }
		ada.call(http.StatusOK, "POST", "/api/groups/"+g.ID+"/invitations", map[string]any{"userId": who.id}, &inv)
		accept = append(accept, func() int { return who.status("POST", "/api/invitations/"+inv.ID+"/accept", nil) })
	}
	for i, code := range together(accept...) {
		if code != http.StatusOK {
			t.Errorf("acceptance %d: %d, want 200", i, code)
		}
	}
	ada.call(http.StatusOK, "GET", "/api/groups/"+g.ID, nil, &g)
	if len(g.Members) != 6 {
		t.Errorf("the group has %d members after five acceptances at once, want 6: %v", len(g.Members), g.roles())
	}
}

// Two changes of one task made at the same instant both stay: neither
// writes the task back as it was before the other.
func TestChangesOfATaskAtTheSameInstantBothStay(t *testing.T) {
	p := start(t)
	ada := p.signUp("Ada", "ada@example.com")
	var tk struct{ ID, Title, Notes, State string }
	ada.call(http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Start"}, &tk)
	for round := range 8 {
		title, notes := "title "+strconv.Itoa(round), "notes "+strconv.Itoa(round)
		codes := together(
			func() int { return ada.status("PATCH", "/api/tasks/"+tk.ID, map[string]any{"title": title}) },
			func() int { return ada.status("PATCH", "/api/tasks/"+tk.ID, map[string]any{"notes": notes}) },
		)
		if codes[0] != http.StatusOK || codes[1] != http.StatusOK {
			t.Fatalf("round %d: %v, want two 200", round, codes)
		}
		ada.call(http.StatusOK, "GET", "/api/tasks/"+tk.ID, nil, &tk)
		if tk.Title != title || tk.Notes != notes {
			t.Fatalf("round %d: the task holds %q and %q, want %q and %q: one change was lost", round, tk.Title, tk.Notes, title, notes)
		}
	}
	// A move and an edit at once: the move stays, and so does the edit.
	codes := together(
		func() int { return ada.status("POST", "/api/tasks/"+tk.ID+"/complete", nil) },
		func() int { return ada.status("PATCH", "/api/tasks/"+tk.ID, map[string]any{"notes": "last"}) },
	)
	ada.call(http.StatusOK, "GET", "/api/tasks/"+tk.ID, nil, &tk)
	if codes[0] != http.StatusOK || codes[1] != http.StatusOK || tk.State != "done" || tk.Notes != "last" {
		t.Errorf("a move and an edit at once: %v, the task is %s with notes %q, want it done with the last notes", codes, tk.State, tk.Notes)
	}
}

// Two members who ask each other at the same instant make one request, not
// two: the second is answered that one is waiting. And a wrong password
// tried five times at once counts five times.
func TestRequestsAtTheSameInstantMakeOneLink(t *testing.T) {
	p := start(t)
	ada, bob := p.signUp("Ada", "ada@example.com"), p.signUp("Bob", "bob@example.com")
	codes := together(
		func() int { return ada.status("POST", "/api/contacts", map[string]any{"email": "bob@example.com"}) },
		func() int { return bob.status("POST", "/api/contacts", map[string]any{"email": "ada@example.com"}) },
	)
	slices.Sort(codes)
	if !slices.Equal(codes, []int{http.StatusOK, http.StatusConflict}) {
		t.Errorf("two requests at once: %v, want one 200 and one 409", codes)
	}
	var mine struct{ Incoming, Outgoing []struct{ ID string } }
	ada.call(http.StatusOK, "GET", "/api/contacts", nil, &mine)
	if len(mine.Incoming)+len(mine.Outgoing) != 1 {
		t.Errorf("Ada has %d request(s) received and %d sent, want one in all", len(mine.Incoming), len(mine.Outgoing))
	}

	wrong := func() int {
		return p.stranger().status("POST", "/api/auth/login", map[string]any{"email": "ada@example.com", "password": "not the right one"})
	}
	codes = together(wrong, wrong, wrong, wrong, wrong)
	for _, code := range codes {
		if code != http.StatusUnauthorized {
			t.Fatalf("five wrong passwords at once: %v, want five 401", codes)
		}
	}
	// The account is locked: the right password is refused with a 403.
	p.stranger().call(http.StatusForbidden, "POST", "/api/auth/login", map[string]any{"email": "ada@example.com", "password": passphrase}, nil)
}

// Two sign-ups of one address at the same instant make one account, and
// both are answered alike.
func TestSignUpsAtTheSameInstantMakeOneAccount(t *testing.T) {
	p := start(t)
	signUp := func(name string) func() int {
		return func() int {
			return p.stranger().status("POST", "/api/auth/signup", map[string]any{"email": "ada@example.com", "name": name, "password": passphrase})
		}
	}
	if codes := together(signUp("Ada"), signUp("Another Ada")); codes[0] != http.StatusOK || codes[1] != http.StatusOK {
		t.Fatalf("two sign-ups at once: %v, want two 200", codes)
	}
	p.settle("the two mails", 50*time.Millisecond, func() bool {
		return len(p.mails("ada@example.com", "Confirmez votre adresse")) == 1 && len(p.mails("ada@example.com", "Vous avez déjà un compte")) == 1
	})
}
