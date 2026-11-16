// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:35933b95eb37968572e6d46dc1379b8281e3503fec0e8451a525bac3fb04f4af

// Package server — Serves the HTTP API, keeps the stores, archives what was
// done long ago, samples the vital signs and sends the reminders.
package server

import (
	"context"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/framework/model"
	"github.com/kitsunium/todo/activity/feed"
	"github.com/kitsunium/todo/contacts/circle"
	"github.com/kitsunium/todo/contacts/contactsapi"
	"github.com/kitsunium/todo/contacts/linkstore"
	"github.com/kitsunium/todo/identity/access"
	"github.com/kitsunium/todo/identity/authapi"
	"github.com/kitsunium/todo/identity/vault"
	"github.com/kitsunium/todo/notify/outbox"
	"github.com/kitsunium/todo/notify/post"
	"github.com/kitsunium/todo/notify/reminders"
	"github.com/kitsunium/todo/notify/watch"
	"github.com/kitsunium/todo/stats/census"
	"github.com/kitsunium/todo/tasks/api"
	"github.com/kitsunium/todo/tasks/planner"
	"github.com/kitsunium/todo/tasks/taskstore"
	"github.com/kitsunium/todo/web/site"
)

// App is role server: the framework App of vault, authapi, linkstore, contactsapi, taskstore, api, feed, post, outbox, watch, census, site, profile server.
var App = app.NewApp("server", vault.Service, authapi.Service, linkstore.Service, contactsapi.Service, taskstore.Service, api.Service, feed.Service, post.Service, outbox.Service, watch.Service, census.Service, site.Service).With(app.Profile(model.ProfileServer))

// Build is what the binary's build says: its variables, set by -ldflags -X.
type Build struct {
}

// Wire builds what role server's handlers use, and gives it to them.
func Wire(ctx context.Context, build Build) {
	accounts := vault.NewAccounts()
	stretcher := vault.NewStretcher()
	wallClock := vault.NewWallClock()
	oneTimeLinks := vault.NewOneTimeLinks()
	drawer := vault.NewDrawer()
	postman := vault.NewPostman(vault.BaseUrlSetting)
	signUp := access.NewSignUp(accounts, stretcher, wallClock, oneTimeLinks, drawer, postman)
	verifyEmail := access.NewVerifyEmail(drawer, oneTimeLinks, wallClock, accounts)
	resendVerification := access.NewResendVerification(accounts, wallClock, oneTimeLinks, drawer, postman)
	forgotPassword := access.NewForgotPassword(accounts, wallClock, oneTimeLinks, drawer, postman)
	sessions := vault.NewSessions()
	resetPassword := access.NewResetPassword(drawer, oneTimeLinks, wallClock, accounts, stretcher, sessions)
	changePassword := access.NewChangePassword(accounts, stretcher, wallClock, sessions)
	signIn := access.NewSignIn(accounts, wallClock, stretcher, drawer, sessions)
	signOut := access.NewSignOut(sessions)
	authenticate := access.NewAuthenticate(drawer, sessions, wallClock, accounts)
	links := linkstore.NewLinks()
	directory := linkstore.NewDirectory()
	listContacts := circle.NewListContacts(links, directory)
	linkstoreWallClock := linkstore.NewWallClock()
	announcer := linkstore.NewAnnouncer()
	requestContact := circle.NewRequestContact(directory, links, linkstoreWallClock, announcer)
	answerRequest := circle.NewAnswerRequest(links, linkstoreWallClock, directory, announcer)
	removeContact := circle.NewRemoveContact(links)
	appClock := taskstore.NewAppClock()
	tasks := taskstore.NewTasks()
	taskstoreAnnouncer := taskstore.NewAnnouncer()
	addTask := planner.NewAddTask(appClock, tasks, taskstoreAnnouncer)
	listTasks := planner.NewListTasks(tasks)
	moveTask := planner.NewMoveTask(tasks, appClock, taskstoreAnnouncer)
	archiveDone := planner.NewArchiveDone(appClock, tasks, taskstoreAnnouncer)
	countTasks := planner.NewCountTasks(appClock, tasks)
	dueSoon := planner.NewDueSoon(appClock, tasks)
	ledger := outbox.NewLedger()
	members := outbox.NewMembers()
	carrier := post.NewCarrier()
	remindDue := reminders.NewRemindDue(dueSoon, ledger, members, carrier)
	vault.Wire()
	authapi.Wire(signUp, authapi.VerifyAddressesSetting, verifyEmail, resendVerification, forgotPassword, resetPassword, changePassword, signIn, signOut, authenticate)
	linkstore.Wire()
	contactsapi.Wire(listContacts, requestContact, answerRequest, removeContact)
	taskstore.Wire()
	api.Wire(addTask, listTasks, moveTask, archiveDone, api.ArchiveAfterSetting, countTasks)
	feed.Wire()
	post.Wire()
	outbox.Wire()
	watch.Wire(remindDue, watch.RemindWithinSetting)
	census.Wire()
	site.Wire()
}
