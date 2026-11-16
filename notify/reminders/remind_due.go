// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package reminders

import (
	"context"
	"strings"

	"github.com/kitsunium/todo/tasks/planner"
	"github.com/kitsunium/todo/tasks/task"
)

// execute is RemindDue's body: each open task due within Within and
// assigned to an address is mailed about once. The mail leaves before the
// task is marked, so a mail that fails is tried again at the next run, and
// one that left is never sent twice unless the mark itself fails.
func (h *RemindDueHandler) execute(ctx context.Context, in *RemindDueInput) (RemindDueOutput, error) {
	due, err := h.dueSoon.Execute(ctx, planner.DueSoonInput{Within: in.Within})
	if err != nil {
		return RemindDueOutput{}, err
	}
	var out RemindDueOutput
	for _, t := range due.Tasks {
		if !strings.Contains(t.Assignee, "@") {
			continue
		}
		seen, err := h.reminderLog.Seen(ctx, t.ID)
		if err != nil {
			return out, err
		}
		if seen {
			continue
		}
		if err := h.mailer.Send(ctx, reminder(t)); err != nil {
			return out, err
		}
		if err := h.reminderLog.Mark(ctx, t.ID); err != nil {
			return out, err
		}
		out.Sent++
	}
	return out, nil
}

// reminder is the mail a task due soon is reminded of with. The title goes
// into the subject as one line: a line break there would be a header.
func reminder(t task.Task) Mail {
	title := strings.Join(strings.Fields(t.Title), " ")
	return Mail{
		UserID:  t.OwnerID,
		To:      t.Assignee,
		Subject: "Due soon: " + title,
		Body:    "\"" + title + "\" is due on " + t.Due.UTC().Format("Monday 2 January 2006 at 15:04 UTC") + ".\n",
	}
}
