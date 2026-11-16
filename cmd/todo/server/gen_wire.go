// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:370a82d053439cd8e17f0439915150946927d383504b60ee20d20c234deff80b

// Package server — Serves the HTTP API, keeps the stores, archives what was
// done long ago, samples the vital signs and sends the reminders.
package server

import (
	"context"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/framework/model"
	"github.com/kitsunium/todo/activity/feed"
	"github.com/kitsunium/todo/notify/outbox"
	"github.com/kitsunium/todo/notify/reminders"
	"github.com/kitsunium/todo/notify/watch"
	"github.com/kitsunium/todo/stats/census"
	"github.com/kitsunium/todo/tasks/api"
	"github.com/kitsunium/todo/tasks/planner"
	"github.com/kitsunium/todo/tasks/taskstore"
)

// App is role server: the framework App of taskstore, api, feed, outbox, watch, census, profile server.
var App = app.NewApp("server", taskstore.Service, api.Service, feed.Service, outbox.Service, watch.Service, census.Service).With(app.Profile(model.ProfileServer))

// Build is what the binary's build says: its variables, set by -ldflags -X.
type Build struct {
}

// Wire builds what role server's handlers use, and gives it to them.
func Wire(ctx context.Context, build Build) {
	appClock := taskstore.NewAppClock()
	tasks := taskstore.NewTasks()
	feedFeed := feed.NewFeed()
	addTask := planner.NewAddTask(appClock, tasks, feedFeed)
	listTasks := planner.NewListTasks(tasks)
	moveTask := planner.NewMoveTask(tasks, appClock, feedFeed)
	archiveDone := planner.NewArchiveDone(appClock, tasks, feedFeed)
	dueSoon := planner.NewDueSoon(appClock, tasks)
	ledger := outbox.NewLedger()
	relay := outbox.NewRelay(outbox.SmtpAddrSetting, outbox.MailFromSetting)
	remindDue := reminders.NewRemindDue(dueSoon, ledger, relay)
	countTasks := planner.NewCountTasks(appClock, tasks)
	taskstore.Wire()
	api.Wire(addTask, listTasks, moveTask, archiveDone, api.ArchiveAfterSetting)
	feed.Wire()
	outbox.Wire()
	watch.Wire(remindDue, watch.RemindWithinSetting)
	census.Wire(countTasks)
}
