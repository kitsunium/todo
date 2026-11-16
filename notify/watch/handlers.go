// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package watch

import (
	"context"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/observe/logger"
	"github.com/kitsunium/todo/notify/reminders"
)

// remind is the job "remind-due"'s body: an error is the run's failure.
func remind(ctx context.Context) error {
	res, err := wired.remindDue.Execute(ctx, reminders.RemindDueInput{Within: wired.remindWithinSetting.Get()})
	if res.Sent > 0 {
		logger.Info(ctx, app.Log(ctx), "reminders sent", logger.Int("mails", res.Sent))
	}
	return err
}
