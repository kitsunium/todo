package main

// The todo is the end-to-end test of kit: every test runs the whole product
// in-process — eight services, in memory, on a free port, in dev (the
// capture mailer, the Studio), on a manual clock the test moves — and talks
// to it over HTTP like the web app does, one cookie jar per user.

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/platform/model"
	"github.com/kitsunium/sdk/pkg/v1/clock"
	"github.com/kitsunium/todo/notify"
)

// testPassword is every test user's password.
const testPassword = "correct horse battery staple"

// epoch is when every test starts: a Thursday morning, UTC.
var epoch = time.Date(2026, 9, 24, 9, 0, 0, 0, time.UTC)

// harness is one run of the product.
type harness struct {
	t   *testing.T
	app *kit.App
	clk *clock.ManualClock
}

// logBuffer keeps the product's logs, shown when a test fails.
type logBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (l *logBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buf.Write(p)
}

func (l *logBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buf.String()
}

// start runs the whole product for one test: in dev, in memory, on a free
// port, on a manual clock at epoch, with the capture mailer.
func start(t *testing.T, analyze bool) *harness {
	t.Helper()
	t.Setenv("KIT_SMTP_URL", "") // never a real relay, whatever the machine says
	t.Setenv("TODO_BASE_URL", "")
	clk := clock.NewManualClock(epoch)
	logs := &logBuffer{}
	app := App.With(kit.InMemory(), kit.Listen("127.0.0.1:0"), kit.Env(kit.EnvDev), kit.Analyze(analyze), kit.Logs(logs), kit.Clock(clk))
	if err := app.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := app.Stop(ctx); err != nil {
			t.Error(err)
		}
		if t.Failed() {
			t.Logf("the product's logs:\n%s", logs)
		}
	})
	return &harness{t: t, app: app, clk: clk}
}

// settle moves the app's clock forward by step until cond holds. The
// subscriptions' consumers, the outbox, the workflow sweeps and the loops all
// wait on that clock, so nothing asynchronous happens unless it moves.
func (h *harness) settle(what string, step time.Duration, cond func() bool) {
	h.t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			h.t.Fatalf("timed out waiting for %s", what)
		}
		h.clk.Advance(step)
		time.Sleep(2 * time.Millisecond)
	}
}

// drain lets the asynchronous work already queued happen: it moves the clock
// a little, several times.
func (h *harness) drain() {
	for range 20 {
		h.clk.Advance(50 * time.Millisecond)
		time.Sleep(2 * time.Millisecond)
	}
}

// client is one user's browser: a cookie jar, and the product's address.
type client struct {
	h      *harness
	name   string
	hc     *http.Client
	bearer string
	// languages is the browser's Accept-Language, if any.
	languages string
}

// client opens a browser with an empty cookie jar.
func (h *harness) client(name string) *client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		h.t.Fatal(err)
	}
	return &client{h: h, name: name, hc: &http.Client{Jar: jar, Timeout: 30 * time.Second}}
}

// do sends a request and returns the status, the body and the headers. A
// rate-limited request is retried: the limits are real, and tests go fast.
func (c *client) do(method, path string, body any) (int, []byte, http.Header) {
	c.h.t.Helper()
	var raw []byte
	if body != nil {
		var err error
		if raw, err = json.Marshal(body); err != nil {
			c.h.t.Fatal(err)
		}
	}
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(c.h.t.Context(), method, c.h.app.URL()+path, bytes.NewReader(raw))
		if err != nil {
			c.h.t.Fatal(err)
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if c.bearer != "" {
			req.Header.Set("Authorization", "Bearer "+c.bearer)
		}
		if c.languages != "" {
			req.Header.Set("Accept-Language", c.languages)
		}
		resp, err := c.hc.Do(req)
		if err != nil {
			c.h.t.Fatalf("%s: %s %s: %v", c.name, method, path, err)
		}
		out, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 60 {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		return resp.StatusCode, out, resp.Header
	}
}

// expect sends a request that must answer status, and returns the body.
func (c *client) expect(status int, method, path string, body any) []byte {
	c.h.t.Helper()
	got, raw, _ := c.do(method, path, body)
	if got != status {
		c.h.t.Fatalf("%s: %s %s: %d %s, want %d", c.name, method, path, got, raw, status)
	}
	return raw
}

// call sends a request that must answer status, and decodes the answer.
func call[T any](c *client, status int, method, path string, body any) T {
	c.h.t.Helper()
	raw := c.expect(status, method, path, body)
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		c.h.t.Fatalf("%s: %s %s: decode %s: %v", c.name, method, path, raw, err)
	}
	return v
}

// apiError is an error answer.
type apiError struct {
	Error struct {
		Code       string
		Message    string
		Violations []struct{ Path, Rule, Message string }
	}
}

// fails sends a request that must fail with status and code.
func (c *client) fails(status int, code, method, path string, body any) apiError {
	c.h.t.Helper()
	e := call[apiError](c, status, method, path, body)
	if e.Error.Code != code {
		c.h.t.Fatalf("%s: %s %s: code %q (%s), want %q", c.name, method, path, e.Error.Code, e.Error.Message, code)
	}
	return e
}

// violates sends a request that must be refused for its field path.
func (c *client) violates(path, method, route string, body any) {
	c.h.t.Helper()
	e := c.fails(http.StatusBadRequest, kit.CodeInvalid, method, route, body)
	if !slices.ContainsFunc(e.Error.Violations, func(v struct{ Path, Rule, Message string }) bool { return v.Path == path }) {
		c.h.t.Fatalf("%s: %s %s: violations %+v, want one on %q", c.name, method, route, e.Error.Violations, path)
	}
}

// user is a user as the API shows one.
type user struct {
	ID        string
	Name      string
	Email     string
	Locale    string
	CreatedAt time.Time
}

// me is the answer of the endpoints that return the caller.
type me struct{ User user }

// mailsTo returns the captured mails to addr whose subject or text holds
// want, oldest first.
func mailsTo(addr, want string) []model.MailMessage {
	var out []model.MailMessage
	for _, m := range notify.Mail.Captured() {
		text := strings.Join(strings.Fields(m.Subject+" "+m.Text), " ") // the text part is wrapped
		if slices.ContainsFunc(m.To, func(to string) bool { return strings.Contains(to, addr) }) && strings.Contains(text, want) {
			out = append(out, m)
		}
	}
	return out
}

// mail waits for a mail to addr whose text holds want, and returns the
// newest.
func (h *harness) mail(addr, want string) model.MailMessage {
	h.t.Helper()
	var found []model.MailMessage
	h.settle("a mail to "+addr+" with "+want, 50*time.Millisecond, func() bool {
		found = mailsTo(addr, want)
		return len(found) > 0
	})
	return found[len(found)-1]
}

// secretIn is the one-time secret of a link.
var secretIn = regexp.MustCompile(`token=([A-Za-z0-9_-]{43})`)

// token reads the one-time secret of the link in a mail.
func token(t *testing.T, m model.MailMessage) string {
	t.Helper()
	match := secretIn.FindStringSubmatch(m.Text)
	if match == nil {
		t.Fatalf("no link with a token in %q", m.Text)
	}
	return match[1]
}

// signup opens a verified account and returns its signed-in browser. The
// browser sends no Accept-Language: the account reads French.
func (h *harness) signup(name, email string) *client {
	h.t.Helper()
	c := h.client(name)
	c.expect(http.StatusOK, "POST", "/api/auth/signup", map[string]any{"email": email, "name": name, "password": testPassword})
	c.expect(http.StatusOK, "POST", "/api/auth/verify", map[string]any{"token": token(h.t, h.mail(email, "/verify?token="))})
	return c
}

// id is the signed-in user's ID.
func (c *client) id() string {
	c.h.t.Helper()
	return call[me](c, http.StatusOK, "GET", "/api/auth/me", nil).User.ID
}

// befriend makes two users contacts: a asks, b accepts.
func befriend(a, b *client, bEmail string) {
	a.h.t.Helper()
	added := call[struct {
		Kind    string
		Request struct{ ID string }
	}](a, http.StatusOK, "POST", "/api/contacts", map[string]any{"email": bEmail})
	if added.Kind != "request" {
		a.h.t.Fatalf("adding %s made a %q", bEmail, added.Kind)
	}
	b.expect(http.StatusOK, "POST", "/api/contacts/"+added.Request.ID+"/accept", nil)
}

// In production the Studio does not exist, and /_kit/ must say so rather than
// fall through to the web app.
func TestNoStudioInProduction(t *testing.T) {
	t.Setenv("KIT_SMTP_URL", "")
	app := App.With(kit.InMemory(), kit.Listen("127.0.0.1:0"), kit.Env(kit.EnvProduction), kit.Logs(io.Discard), kit.Clock(clock.NewManualClock(epoch)))
	if err := app.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer app.Stop(context.Background())
	for path, want := range map[string]int{
		"/_kit/api/graph": 404, "/_kit/": 404, "/_kit/api/source?file=main.go": 404, "/_kit/api/mail": 404,
		"/_kit/health/ready": 200, "/": 200, "/app/today": 200, "/api/auth/me": 401,
	} {
		resp, err := http.Get(app.URL() + path)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != want {
			t.Errorf("%s: %d, want %d", path, resp.StatusCode, want)
		}
	}
}

// The README's architecture diagram is generated, and stays so: this fails
// when the code changes the graph and the README was not regenerated with
//
//	go run . graph -format mermaid
func TestTheReadmeDiagramIsCurrent(t *testing.T) {
	raw, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	readme := string(raw)
	start := strings.Index(readme, "```mermaid\nflowchart")
	if start < 0 {
		t.Fatal("the README has no architecture diagram")
	}
	body := readme[start+len("```mermaid\n"):]
	body = body[:strings.Index(body, "```")]
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	done := make(chan string)
	go func() {
		out, _ := io.ReadAll(r)
		done <- string(out)
	}()
	code := kit.NewApp("todo", services()...).Main(t.Context(), []string{"graph", "-format", "mermaid"})
	os.Stdout = stdout
	_ = w.Close()
	got := <-done
	if code != 0 {
		t.Fatalf("graph exited %d", code)
	}
	if got != body {
		t.Fatalf("the README's diagram is stale; regenerate it with `go run . graph -format mermaid`.\nwant:\n%s", got)
	}
}
