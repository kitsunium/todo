// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package driver

import (
	"context"
	"strconv"
	"time"
)

// What a load adds: a few titles, the three priorities, and assignees of
// whom one has an address, so that the reminders have something to send.
var (
	titles     = []string{"Write the release notes", "Review the pull request", "Call the supplier", "Plan the sprint", "Fix the flaky test", "Update the runbook"}
	priorities = []string{"normal", "high", "low"}
	assignees  = []string{"Ada", "ada@example.com", "Grace", ""}
)

// The load's own member: it signs up the first time, and signs in.
const (
	member     = "load@example.com"
	passphrase = "a load that only adds tasks"
)

// execute is Drive's body: it enters as the load's member, then makes one call of the API per tick, Rate ticks a
// second, until For has passed or the context ends. The calls go round in
// six steps — add a task, list them, complete the oldest one added, read
// the feed, read the vital signs, then archive or reopen that task — so the
// first round makes every call of the API whatever For is. Entering that
// fails is counted like any call: every call after it is then refused. A call that
// fails is counted and the load goes on.
func (h *DriveHandler) execute(ctx context.Context, in *DriveInput) (DriveOutput, error) {
	if in.Rate <= 0 {
		return DriveOutput{}, ErrNoRate
	}
	var out DriveOutput
	count := func(err error) {
		out.Requests++
		if err != nil {
			out.Failures++
		}
	}
	tick := time.NewTicker(time.Second / time.Duration(in.Rate))
	defer tick.Stop()
	deadline := time.Now().Add(in.For)
	count(h.api.Enter(ctx, member, "Load", passphrase))
	var added []string // the tasks added and not yet moved to their end
	for n := 0; ; n++ {
		round := n / 6
		switch n % 6 {
		case 0:
			due := time.Time{}
			if round%2 == 0 {
				due = time.Now().Add(12 * time.Hour)
			}
			id, err := h.api.Add(ctx, titles[round%len(titles)]+" #"+strconv.Itoa(round+1), priorities[round%len(priorities)], assignees[round%len(assignees)], due)
			count(err)
			if err == nil {
				added = append(added, id)
			}
		case 1:
			_, err := h.api.List(ctx)
			count(err)
		case 2:
			if len(added) > 0 {
				count(h.api.Move(ctx, added[0], "complete"))
			}
		case 3:
			_, err := h.api.Activity(ctx)
			count(err)
		case 4:
			_, err := h.api.Stats(ctx)
			count(err)
		case 5:
			if len(added) > 0 {
				event := "archive"
				if round%3 == 2 {
					event = "reopen"
				}
				count(h.api.Move(ctx, added[0], event))
				added = added[1:]
			}
		}
		if n >= 5 && !time.Now().Before(deadline) {
			return out, nil // the first round is always whole
		}
		select {
		case <-ctx.Done():
			return out, nil
		case <-tick.C:
		}
	}
}
