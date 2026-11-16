// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:d9e6eb32619ad64ff259ee580b4b28ea598fe9a15bc77181628dbbd887d57f92

// Package server — Serves the HTTP API, keeps the stores, archives what was
// done long ago, samples the vital signs and sends the reminders.
package server

import (
	"context"
	"embed"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/framework/model"
	"github.com/kitsunium/todo/activity/feed"
	"github.com/kitsunium/todo/contacts/circle"
	"github.com/kitsunium/todo/contacts/contactsapi"
	"github.com/kitsunium/todo/contacts/linkstore"
	"github.com/kitsunium/todo/groups/groupsapi"
	"github.com/kitsunium/todo/groups/groupstore"
	"github.com/kitsunium/todo/groups/roster"
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

// configFiles are role server's configuration files, embedded from config: they
// are hand-written.
//
//go:embed config
var configFiles embed.FS

// App is role server: the framework App of vault, authapi, linkstore, contactsapi, groupstore, groupsapi, taskstore, api, feed, post, outbox, watch, census, site, profile server.
var App = app.NewApp("server", vault.Service, authapi.Service, linkstore.Service, contactsapi.Service, groupstore.Service, groupsapi.Service, taskstore.Service, api.Service, feed.Service, post.Service, outbox.Service, watch.Service, census.Service, site.Service).With(app.Profile(model.ProfileServer), app.ConfigFiles(configFiles))

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
	herald := vault.NewHerald()
	signUp := access.NewSignUp(accounts, stretcher, wallClock, oneTimeLinks, drawer, postman, herald)
	sessions := vault.NewSessions()
	verifyEmail := access.NewVerifyEmail(drawer, oneTimeLinks, wallClock, accounts, sessions, herald)
	resendVerification := access.NewResendVerification(accounts, wallClock, oneTimeLinks, drawer, postman)
	forgotPassword := access.NewForgotPassword(accounts, wallClock, oneTimeLinks, drawer, postman)
	resetPassword := access.NewResetPassword(drawer, oneTimeLinks, wallClock, accounts, stretcher, sessions, herald)
	changePassword := access.NewChangePassword(accounts, stretcher, wallClock, sessions)
	signIn := access.NewSignIn(accounts, wallClock, stretcher, drawer, sessions)
	signOut := access.NewSignOut(sessions)
	listSessions := access.NewListSessions(sessions, wallClock)
	revokeSession := access.NewRevokeSession(sessions)
	authenticate := access.NewAuthenticate(drawer, sessions, wallClock, accounts)
	links := linkstore.NewLinks()
	directory := linkstore.NewDirectory()
	invitations := linkstore.NewInvitations()
	listContacts := circle.NewListContacts(links, directory, invitations)
	linkstoreWallClock := linkstore.NewWallClock()
	announcer := linkstore.NewAnnouncer()
	courier := linkstore.NewCourier()
	requestContact := circle.NewRequestContact(directory, links, linkstoreWallClock, announcer, invitations, courier)
	answerRequest := circle.NewAnswerRequest(linkstoreWallClock, links, directory, announcer)
	cancelInvite := circle.NewCancelInvite(invitations)
	removeContact := circle.NewRemoveContact(links)
	claimInvites := circle.NewClaimInvites(invitations, links, linkstoreWallClock, directory, announcer)
	areContacts := circle.NewAreContacts(links)
	groups := groupstore.NewGroups()
	groupstoreDirectory := groupstore.NewDirectory()
	listGroups := roster.NewListGroups(groups, groupstoreDirectory)
	groupstoreWallClock := groupstore.NewWallClock()
	createGroup := roster.NewCreateGroup(groupstoreWallClock, groups, groupstoreDirectory)
	getGroup := roster.NewGetGroup(groups, groupstoreDirectory)
	editGroup := roster.NewEditGroup(groupstoreWallClock, groups, groupstoreDirectory)
	groupstoreInvitations := groupstore.NewInvitations()
	groupstoreAnnouncer := groupstore.NewAnnouncer()
	deleteGroup := roster.NewDeleteGroup(groups, groupstoreInvitations, groupstoreDirectory, groupstoreAnnouncer)
	friends := groupstore.NewFriends()
	invite := roster.NewInvite(groups, friends, groupstoreInvitations, groupstoreWallClock, groupstoreDirectory, groupstoreAnnouncer)
	listInvitations := roster.NewListInvitations(groupstoreInvitations, groups, groupstoreDirectory)
	answerInvitation := roster.NewAnswerInvitation(groupstoreWallClock, groupstoreInvitations, groups, groupstoreDirectory, groupstoreAnnouncer)
	removeMember := roster.NewRemoveMember(groupstoreWallClock, groups, groupstoreDirectory, groupstoreAnnouncer)
	setRole := roster.NewSetRole(groupstoreWallClock, groups, groupstoreDirectory, groupstoreAnnouncer)
	roleOf := roster.NewRoleOf(groups)
	memberships := roster.NewMemberships(groups)
	membersOf := roster.NewMembersOf(groups)
	tasks := taskstore.NewTasks()
	taskstoreGroups := taskstore.NewGroups()
	appClock := taskstore.NewAppClock()
	followGroup := planner.NewFollowGroup(tasks, taskstoreGroups, appClock)
	taskstoreAnnouncer := taskstore.NewAnnouncer()
	addTask := planner.NewAddTask(appClock, taskstoreGroups, tasks, taskstoreAnnouncer)
	listTasks := planner.NewListTasks(taskstoreGroups, appClock, tasks)
	countViews := planner.NewCountViews(taskstoreGroups, appClock, tasks)
	getTask := planner.NewGetTask(tasks, taskstoreGroups)
	editTask := planner.NewEditTask(tasks, taskstoreGroups, appClock, taskstoreAnnouncer)
	deleteTask := planner.NewDeleteTask(tasks, taskstoreGroups, taskstoreAnnouncer)
	taskstoreFriends := taskstore.NewFriends()
	shareTask := planner.NewShareTask(tasks, taskstoreGroups, taskstoreFriends, appClock, taskstoreAnnouncer)
	unshareTask := planner.NewUnshareTask(tasks, taskstoreGroups, appClock)
	moveTask := planner.NewMoveTask(tasks, taskstoreGroups, appClock, taskstoreAnnouncer)
	countTasks := planner.NewCountTasks(appClock, tasks)
	dueSoon := planner.NewDueSoon(appClock, tasks)
	ledger := outbox.NewLedger()
	members := outbox.NewMembers()
	carrier := post.NewCarrier()
	remindDue := reminders.NewRemindDue(dueSoon, ledger, members, carrier)
	vault.Wire()
	authapi.Wire(signUp, authapi.VerifyAddressesSetting, verifyEmail, resendVerification, forgotPassword, resetPassword, changePassword, signIn, signOut, listSessions, revokeSession, authenticate)
	linkstore.Wire()
	contactsapi.Wire(listContacts, requestContact, answerRequest, cancelInvite, removeContact, claimInvites, areContacts)
	groupstore.Wire()
	groupsapi.Wire(listGroups, createGroup, getGroup, editGroup, deleteGroup, invite, listInvitations, answerInvitation, removeMember, setRole, roleOf, memberships, membersOf)
	taskstore.Wire(followGroup)
	api.Wire(addTask, listTasks, countViews, getTask, editTask, deleteTask, shareTask, unshareTask, moveTask, countTasks)
	feed.Wire()
	post.Wire()
	outbox.Wire()
	watch.Wire(remindDue, watch.RemindWithinSetting)
	census.Wire()
	site.Wire()
}
