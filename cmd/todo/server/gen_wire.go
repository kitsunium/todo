// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:2a2e1ae2f3c14afcf45b67e8c18923c85e5975df67c02bb2610e585bf5526ed2

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
	listSessions := access.NewListSessions(sessions, wallClock)
	revokeSession := access.NewRevokeSession(sessions)
	authenticate := access.NewAuthenticate(drawer, sessions, wallClock, accounts)
	links := linkstore.NewLinks()
	directory := linkstore.NewDirectory()
	listContacts := circle.NewListContacts(links, directory)
	linkstoreWallClock := linkstore.NewWallClock()
	announcer := linkstore.NewAnnouncer()
	requestContact := circle.NewRequestContact(directory, links, linkstoreWallClock, announcer)
	answerRequest := circle.NewAnswerRequest(links, linkstoreWallClock, directory, announcer)
	removeContact := circle.NewRemoveContact(links)
	areContacts := circle.NewAreContacts(links)
	appClock := taskstore.NewAppClock()
	tasks := taskstore.NewTasks()
	taskstoreAnnouncer := taskstore.NewAnnouncer()
	addTask := planner.NewAddTask(appClock, tasks, taskstoreAnnouncer)
	listTasks := planner.NewListTasks(tasks)
	getTask := planner.NewGetTask(tasks)
	editTask := planner.NewEditTask(tasks, appClock, taskstoreAnnouncer)
	deleteTask := planner.NewDeleteTask(tasks, taskstoreAnnouncer)
	friends := taskstore.NewFriends()
	shareTask := planner.NewShareTask(tasks, friends, appClock, taskstoreAnnouncer)
	unshareTask := planner.NewUnshareTask(tasks, appClock)
	moveTask := planner.NewMoveTask(tasks, appClock, taskstoreAnnouncer)
	countTasks := planner.NewCountTasks(appClock, tasks)
	dueSoon := planner.NewDueSoon(appClock, tasks)
	ledger := outbox.NewLedger()
	members := outbox.NewMembers()
	carrier := post.NewCarrier()
	remindDue := reminders.NewRemindDue(dueSoon, ledger, members, carrier)
	vault.Wire()
	authapi.Wire(signUp, authapi.VerifyAddressesSetting, verifyEmail, resendVerification, forgotPassword, resetPassword, changePassword, signIn, signOut, listSessions, revokeSession, authenticate)
	linkstore.Wire()
	contactsapi.Wire(listContacts, requestContact, answerRequest, removeContact, areContacts)
	taskstore.Wire()
	api.Wire(addTask, listTasks, getTask, editTask, deleteTask, shareTask, unshareTask, moveTask, countTasks)
	feed.Wire()
	post.Wire()
	outbox.Wire()
	watch.Wire(remindDue, watch.RemindWithinSetting)
	census.Wire()
	site.Wire()
}
