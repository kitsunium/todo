// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/kitsunium/sdk/framework/app"
)

// client is Client's own state: the setting that names the server, and the
// HTTP client every call goes through.
type client struct {
	targetSetting *app.Setting[string]
	http          *http.Client
}

// newClient is NewClient's body.
func newClient(targetSetting *app.Setting[string]) *Client {
	return &Client{client{targetSetting: targetSetting, http: &http.Client{Timeout: 5 * time.Second}}}
}

// activity is Client.Activity's body: how many entries the feed answers.
func (a *Client) activity(ctx context.Context) (r0 int, err error) {
	var page struct {
		Entries []json.RawMessage `json:"entries"`
	}
	err = a.call(ctx, http.MethodGet, "/api/activity", nil, &page)
	return len(page.Entries), err
}

// add is Client.Add's body: the id of the task added.
func (a *Client) add(ctx context.Context, title string, priority string, assignee string, due time.Time) (r0 string, err error) {
	body := map[string]any{"title": title, "priority": priority, "assignee": assignee}
	if !due.IsZero() {
		body["due"] = due.UTC().Format(time.RFC3339)
	}
	var added struct {
		ID string `json:"id"`
	}
	err = a.call(ctx, http.MethodPost, "/api/tasks", body, &added)
	return added.ID, err
}

// list is Client.List's body: how many tasks the list answers.
func (a *Client) list(ctx context.Context) (r0 int, err error) {
	var list struct {
		Tasks []json.RawMessage `json:"tasks"`
	}
	err = a.call(ctx, http.MethodGet, "/api/tasks", nil, &list)
	return len(list.Tasks), err
}

// move is Client.Move's body.
func (a *Client) move(ctx context.Context, id string, event string) (err error) {
	return a.call(ctx, http.MethodPost, "/api/tasks/"+url.PathEscape(id)+"/"+url.PathEscape(event), nil, nil)
}

// stats is Client.Stats's body: how many samples the server keeps.
func (a *Client) stats(ctx context.Context) (r0 int, err error) {
	var report struct {
		History []json.RawMessage `json:"history"`
	}
	err = a.call(ctx, http.MethodGet, "/api/stats", nil, &report)
	return len(report.History), err
}

// call sends one request to the server target names and decodes its answer
// into out; an answer that is not a 2xx is an error that says its status.
func (a *Client) call(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		raw, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, "http://"+a.targetSetting.Get()+path, body)
	if err != nil {
		return err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := a.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("%s %s: %s", method, path, resp.Status)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(raw, out)
}
