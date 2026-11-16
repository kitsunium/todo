package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/clock"
	"github.com/kitsunium/todo/cmd/todo/server"

	// What the binary's main links for this role: the HTTP engine and the
	// subsystems its components declare. A test program links them itself.
	_ "github.com/kitsunium/sdk/framework/app/jobs"
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

// call sends one request, checks its status and decodes its answer into out.
func (p *product) call(status int, method, path string, in, out any) {
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

// A task is added, listed, moved through its lifecycle, and every move is
// in the feed; what the API refuses, it refuses with the status a client
// can act on.
func TestATaskGoesThroughItsLifecycle(t *testing.T) {
	p := start(t)
	var added taskView
	p.call(http.StatusOK, "POST", "/api/tasks", map[string]any{"title": " Write the design ", "priority": "high", "assignee": "Ada"}, &added)
	if added.ID == "" || added.Title != "Write the design" || added.Priority != "high" || added.State != "open" {
		t.Fatalf("the task added: %+v", added)
	}
	p.call(http.StatusBadRequest, "POST", "/api/tasks", map[string]any{"title": "  "}, nil)
	p.call(http.StatusBadRequest, "POST", "/api/tasks", map[string]any{"title": "x", "priority": "urgent"}, nil)
	p.call(http.StatusNotFound, "POST", "/api/tasks/task_nobody/complete", nil, nil)
	p.call(http.StatusConflict, "POST", "/api/tasks/"+added.ID+"/archive", nil, nil)

	var moved taskView
	p.call(http.StatusOK, "POST", "/api/tasks/"+added.ID+"/complete", nil, &moved)
	if moved.State != "done" || !moved.DoneAt.Equal(epoch) {
		t.Fatalf("the task completed: %+v", moved)
	}
	p.call(http.StatusOK, "POST", "/api/tasks/"+added.ID+"/reopen", nil, &moved)
	p.call(http.StatusOK, "POST", "/api/tasks/"+added.ID+"/complete", nil, &moved)
	p.call(http.StatusOK, "POST", "/api/tasks/"+added.ID+"/archive", nil, &moved)
	if moved.State != "archived" {
		t.Fatalf("the task archived: %+v", moved)
	}

	var list struct {
		Tasks []taskView `json:"tasks"`
	}
	p.call(http.StatusOK, "GET", "/api/tasks", nil, &list)
	if len(list.Tasks) != 0 {
		t.Errorf("the list still holds the archived task: %+v", list.Tasks)
	}
	p.call(http.StatusOK, "GET", "/api/tasks?state=archived", nil, &list)
	if len(list.Tasks) != 1 || list.Tasks[0].ID != added.ID {
		t.Errorf("the archived tasks: %+v, want the one task", list.Tasks)
	}
	p.call(http.StatusOK, "POST", "/api/tasks/"+added.ID+"/restore", nil, &moved)
	if moved.State != "open" || !moved.DoneAt.IsZero() {
		t.Errorf("the task restored: %+v", moved)
	}

	var feed struct {
		Entries []struct {
			Kind   string `json:"kind"`
			TaskID string `json:"taskId"`
		} `json:"entries"`
	}
	p.call(http.StatusOK, "GET", "/api/activity", nil, &feed)
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
	p.call(http.StatusOK, "GET", "/api/activity?limit=2", nil, &feed)
	if len(feed.Entries) != 2 {
		t.Errorf("a page of 2 holds %d entries", len(feed.Entries))
	}
}

// The jobs run on the product's clock: the vital signs are sampled, a task
// due soon is reminded of once, and a task done for a day is archived.
func TestTheJobsRunOnTheClock(t *testing.T) {
	p := start(t)
	var soon, other taskView
	p.call(http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Ship the release", "assignee": "ada@example.com", "due": epoch.Add(6 * time.Hour)}, &soon)
	p.call(http.StatusOK, "POST", "/api/tasks", map[string]any{"title": "Tidy the backlog"}, &other)
	p.call(http.StatusOK, "POST", "/api/tasks/"+other.ID+"/complete", nil, nil)

	var report struct {
		Latest *struct {
			Counts         map[string]int `json:"counts"`
			Total          int            `json:"total"`
			CompletionRate float64        `json:"completionRate"`
		} `json:"latest"`
	}
	p.settle("a sample of the vital signs", 5*time.Second, func() bool {
		p.call(http.StatusOK, "GET", "/api/stats", nil, &report)
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
		p.call(http.StatusOK, "GET", "/api/outbox", nil, &outbox)
		return len(outbox.Mails) > 0
	})
	p.clk.Advance(5 * time.Minute) // more runs of the job: the task is reminded of once
	time.Sleep(20 * time.Millisecond)
	p.call(http.StatusOK, "GET", "/api/outbox", nil, &outbox)
	if len(outbox.Mails) != 1 || outbox.Mails[0].To != "ada@example.com" || outbox.Mails[0].Subject != "Due soon: Ship the release" || outbox.Mails[0].Relayed {
		t.Errorf("the outbox: %+v, want the one reminder, kept and not relayed", outbox.Mails)
	}

	var list struct {
		Tasks []taskView `json:"tasks"`
	}
	p.settle("the done task archived", time.Hour, func() bool {
		p.call(http.StatusOK, "GET", "/api/tasks?state=archived", nil, &list)
		return len(list.Tasks) == 1
	})
	if list.Tasks[0].ID != other.ID {
		t.Errorf("archived %+v, want the task done a day ago", list.Tasks)
	}
}
