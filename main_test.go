package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/platform/model"
	"github.com/kitsunium/sdk/pkg/v1/clock"
	"github.com/kitsunium/todo/activity"
	"github.com/kitsunium/todo/stats"
	"github.com/kitsunium/todo/todos"
	"github.com/kitsunium/todo/web"
)

// run starts the whole product in dev, in memory, on a manual clock.
func run(t *testing.T, analyze bool) (*kit.App, *clock.ManualClock) {
	t.Helper()
	clk := clock.NewManualClock(time.Now().UTC())
	app := App.With(kit.InMemory(), kit.Listen("127.0.0.1:0"), kit.Env(kit.EnvDev), kit.Analyze(analyze), kit.Logs(io.Discard), kit.Clock(clk))
	if err := app.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := app.Stop(ctx); err != nil {
			t.Error(err)
		}
	})
	return app, clk
}

func do(t *testing.T, app *kit.App, method, path string, body any) (int, []byte) {
	t.Helper()
	var rd io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		rd = bytes.NewReader(raw)
	}
	req, _ := http.NewRequestWithContext(t.Context(), method, app.URL()+path, rd)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, raw
}

func decode[T any](t *testing.T, raw []byte) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	return v
}

// until advances the clock by step until cond holds.
func until(t *testing.T, clk *clock.ManualClock, step time.Duration, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		clk.Advance(step)
		time.Sleep(10 * time.Millisecond)
	}
}

func TestTheTodoList(t *testing.T) {
	app, clk := run(t, false)

	status, raw := do(t, app, "POST", "/todos", map[string]any{"title": "  Write the kit README  "})
	if status != http.StatusOK {
		t.Fatalf("create: %d %s", status, raw)
	}
	readme := decode[todos.Todo](t, raw)
	if readme.Title != "Write the kit README" || readme.Status != todos.Open || !strings.HasPrefix(readme.ID, "todo_") {
		t.Fatalf("created %+v", readme)
	}
	past := clk.Now().Add(-time.Hour)
	_, raw = do(t, app, "POST", "/todos", map[string]any{"title": "Renew the domain", "due": past})
	domain := decode[todos.Todo](t, raw)

	if status, raw := do(t, app, "POST", "/todos", map[string]any{"title": "   "}); status != http.StatusBadRequest {
		t.Errorf("a blank title: %d %s", status, raw)
	}
	if status, _ := do(t, app, "POST", "/todos/"+readme.ID+"/archive", nil); status != http.StatusConflict {
		t.Errorf("archiving an open todo must conflict, got %d", status)
	}

	// The overdue guard is the workflow's own loop at work.
	until(t, clk, time.Second, "the overdue guard", func() bool {
		_, raw := do(t, app, "GET", "/todos?status=overdue", nil)
		return len(decode[todos.ListOutput](t, raw).Todos) == 1
	})

	for _, step := range []struct {
		event string
		want  todos.Status
	}{{"complete", todos.Done}, {"reopen", todos.Open}, {"complete", todos.Done}} {
		status, raw := do(t, app, "POST", "/todos/"+readme.ID+"/"+step.event, nil)
		got := decode[todos.Todo](t, raw)
		if status != http.StatusOK || got.Status != step.want {
			t.Fatalf("%s: %d %s", step.event, status, raw)
		}
		if step.want == todos.Done && got.DoneAt == nil {
			t.Errorf("%s: OnEnter(Done) did not stamp DoneAt", step.event)
		}
	}

	// The auto-archive timer: two minutes after being done.
	until(t, clk, 30*time.Second, "the auto-archive timer", func() bool {
		_, raw := do(t, app, "GET", "/todos?status=archived", nil)
		return len(decode[todos.ListOutput](t, raw).Todos) == 1
	})

	// Every transition reached the activity feed through the topic.
	until(t, clk, 50*time.Millisecond, "the activity feed", func() bool {
		_, raw := do(t, app, "GET", "/activity", nil)
		return len(decode[activity.RecentOutput](t, raw).Entries) == 7
	})
	_, raw = do(t, app, "GET", "/activity", nil)
	events := map[string]int{}
	for _, e := range decode[activity.RecentOutput](t, raw).Entries {
		events[e.Event]++
	}
	if events["create"] != 2 || events["complete"] != 2 || events["reopen"] != 1 || events["overdue"] != 1 || events["auto-archive"] != 1 {
		t.Errorf("feed events %v", events)
	}

	// The stats loop samples through the todos service's private API.
	until(t, clk, 10*time.Second, "a stats sample", func() bool {
		_, raw := do(t, app, "GET", "/stats", nil)
		return decode[struct{ Latest *struct{ Total int } }](t, raw).Latest != nil
	})

	if status, _ := do(t, app, "DELETE", "/todos/"+domain.ID, nil); status != http.StatusNoContent {
		t.Errorf("delete: %d", status)
	}
	if status, _ := do(t, app, "DELETE", "/todos/"+domain.ID, nil); status != http.StatusNotFound {
		t.Errorf("second delete: %d", status)
	}
}

func TestTheUserInterfaceIsServed(t *testing.T) {
	app, _ := run(t, false)
	for _, path := range []string{"/", "/app.js", "/style.css", "/favicon.svg", "/some/client/route"} {
		resp, err := http.Get(app.URL() + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s: %d", path, resp.StatusCode)
		}
		if csp := resp.Header.Get("Content-Security-Policy"); !strings.Contains(csp, "default-src 'self'") {
			t.Errorf("%s: CSP %q", path, csp)
		}
	}
}

// The diagram never lies: every node the source declares is a node the
// product serves, every edge the source proves is in the product's graph,
// and every call the page makes shows up as an edge from the page.
func TestTheDiagramMatchesTheCode(t *testing.T) {
	app, _ := run(t, true)
	deadline := time.Now().Add(60 * time.Second)
	var g *model.Graph
	for {
		g = app.Graph()
		if g.Analysis != nil && g.Analysis.Status != "running" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the static analysis never finished")
		}
		time.Sleep(100 * time.Millisecond)
	}
	if g.Analysis.Status != "ok" {
		t.Fatalf("analysis %+v", g.Analysis)
	}
	var static int
	for _, e := range g.Edges {
		if len(e.Static) > 0 {
			static++
		}
		if g.Node(e.From) == nil || g.Node(e.To) == nil {
			t.Errorf("edge %s points outside the graph", e.ID)
		}
	}
	if static < 14 {
		t.Errorf("only %d edges carry static evidence", static)
	}
	for _, n := range g.Nodes {
		if n.Kind == model.KindExternal {
			continue
		}
		if n.Source == nil || n.Source.File == "" {
			t.Errorf("%s has no source", n.ID)
		}
		if (n.Kind == model.KindEndpoint || n.Kind == model.KindJob || n.Kind == model.KindSubscription) &&
			(n.Handler == nil || n.Handler.EndLine == 0) {
			t.Errorf("%s has no handler range", n.ID)
		}
		if n.Kind != model.KindService && n.Doc == "" {
			t.Errorf("%s has no documentation", n.ID)
		}
	}
	for _, tr := range g.Node("todos/workflow/lifecycle").Workflow.Transitions {
		if tr.Trigger == model.TriggerEvent && len(tr.Callers) == 0 {
			t.Errorf("no endpoint fires %q", tr.Event)
		}
	}

	req, _ := http.NewRequest("GET", app.URL()+"/todos", nil)
	req.Header.Set("Referer", app.URL()+"/")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if e := app.Graph().Edge("web/frontend/ui|calls|todos/endpoint/List"); e == nil || e.Observed == nil {
		t.Error("a call from the page is not drawn from the page")
	}
}

// In production the Studio does not exist, and /_kit/ must say so rather than
// fall through to the page: the page shows its link to the diagram when
// /_kit/api/graph answers.
func TestNoStudioInProduction(t *testing.T) {
	app := App.With(kit.InMemory(), kit.Listen("127.0.0.1:0"), kit.Env(kit.EnvProduction), kit.Logs(io.Discard))
	if err := app.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer app.Stop(context.Background())
	for path, want := range map[string]int{"/_kit/api/graph": 404, "/_kit/": 404, "/_kit/api/source?file=main.go": 404, "/_kit/health/ready": 200, "/": 200} {
		resp, err := http.Get(app.URL() + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
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
	start := strings.Index(readme, "```mermaid\nflowchart LR")
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
	code := kit.NewApp("todo", todos.Service, activity.Service, stats.Service, web.Service).
		Main(t.Context(), []string{"graph", "-format", "mermaid"})
	os.Stdout = stdout
	w.Close()
	got := <-done
	if code != 0 {
		t.Fatalf("graph exited %d", code)
	}
	if got != body {
		t.Fatalf("the README's diagram is stale; regenerate it with `go run . graph -format mermaid`.\nwant:\n%s", got)
	}
}
