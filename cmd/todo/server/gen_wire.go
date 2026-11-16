// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:6d83e3d5800ec2321096deeccd3c573f533b4ecf81275cccebf1ad5e2b7143c2

// Package server — Serves the HTTP API, keeps the stores, archives what was
// done long ago, samples the vital signs and sends the reminders.
package server

import (
	"context"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/framework/model"
	"github.com/kitsunium/todo/activity/feed"
	"github.com/kitsunium/todo/identity/access"
	"github.com/kitsunium/todo/identity/authapi"
	"github.com/kitsunium/todo/identity/vault"
	"github.com/kitsunium/todo/notify/outbox"
	"github.com/kitsunium/todo/notify/reminders"
	"github.com/kitsunium/todo/notify/watch"
	"github.com/kitsunium/todo/stats/census"
	"github.com/kitsunium/todo/tasks/api"
	"github.com/kitsunium/todo/tasks/planner"
	"github.com/kitsunium/todo/tasks/taskstore"
)

// App is role server: the framework App of vault, authapi, taskstore, api, feed, outbox, watch, census, profile server.
var App = app.NewApp("server", vault.Service, authapi.Service, taskstore.Service, api.Service, feed.Service, outbox.Service, watch.Service, census.Service).With(app.Profile(model.ProfileServer))

// Build is what the binary's build says: its variables, set by -ldflags -X.
type Build struct {
}

// Wire builds what role server's handlers use, and gives it to them.
func Wire(ctx context.Context, build Build) {
	accounts := vault.NewAccounts()
	stretcher := vault.NewStretcher()
	wallClock := vault.NewWallClock()
	signUp := access.NewSignUp(accounts, stretcher, wallClock)
	drawer := vault.NewDrawer()
	sessions := vault.NewSessions()
	signIn := access.NewSignIn(accounts, wallClock, stretcher, drawer, sessions)
	signOut := access.NewSignOut(sessions)
	authenticate := access.NewAuthenticate(drawer, sessions, wallClock, accounts)
	appClock := taskstore.NewAppClock()
	tasks := taskstore.NewTasks()
	announcer := taskstore.NewAnnouncer()
	addTask := planner.NewAddTask(appClock, tasks, announcer)
	listTasks := planner.NewListTasks(tasks)
	moveTask := planner.NewMoveTask(tasks, appClock, announcer)
	archiveDone := planner.NewArchiveDone(appClock, tasks, announcer)
	dueSoon := planner.NewDueSoon(appClock, tasks)
	ledger := outbox.NewLedger()
	relay := outbox.NewRelay(outbox.SmtpAddrSetting, outbox.MailFromSetting)
	remindDue := reminders.NewRemindDue(dueSoon, ledger, relay)
	countTasks := planner.NewCountTasks(appClock, tasks)
	vault.Wire()
	authapi.Wire(signUp, signIn, signOut, authenticate)
	taskstore.Wire()
	api.Wire(addTask, listTasks, moveTask, archiveDone, api.ArchiveAfterSetting)
	feed.Wire()
	outbox.Wire()
	watch.Wire(remindDue, watch.RemindWithinSetting)
	census.Wire(countTasks)
}
