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

// execute is RemindDue's body: the owner of each open task due within
// Within is mailed about it once, at the address the directory gives. The mail leaves before the
// task is marked, so a mail that fails is tried again at the next run, and
// one that left is never sent twice unless the mark itself fails.
func (h *RemindDueHandler) execute(ctx context.Context, in *RemindDueInput) (RemindDueOutput, error) {
	due, err := h.dueSoon.Execute(ctx, planner.DueSoonInput{Within: in.Within})
	if err != nil {
		return RemindDueOutput{}, err
	}
	var out RemindDueOutput
	for _, t := range due.Tasks {
		seen, err := h.reminderLog.Seen(ctx, t.ID)
		if err != nil {
			return out, err
		}
		if seen {
			continue
		}
		to, found, err := h.directory.Address(ctx, t.OwnerID)
		if err != nil {
			return out, err
		}
		if !found {
			continue // the owner's account is gone: nobody to remind
		}
		if err := h.mailer.Send(ctx, reminder(t, to)); err != nil {
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
func reminder(t task.Task, to string) Mail {
	title := strings.Join(strings.Fields(t.Title), " ")
	return Mail{
		To:      to,
		Subject: "Bientôt à échéance : " + title,
		Body:    "Bonjour,\n\nVotre tâche « " + title + " » arrive à échéance le " + t.Due.UTC().Format("02/01/2006 à 15:04") + " UTC.\n",
	}
}
