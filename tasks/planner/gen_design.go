// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:c09190927b7a702a68883efedfd3b487bc8023ea307d3b6eeda776c8d6820e96

// Package planner — What the team does with tasks, and what that needs.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package planner

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/tasks/task"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 1, 2, 255), "NOT_IMPLEMENTED", "not implemented yet", "planner: a hand-written body is still a skeleton")

// Person — A member as a task names them, for a page that has no other way
// to.
type Person struct {
	ID   string `json:"id" app:"public"`
	Name string `json:"name" app:"personal"`
}

// CircleV1 says who is a contact of whom.
type CircleV1 interface {
	AreContacts(ctx context.Context, one string, other string) (bool, error)
}

// ClockV1 tells the time.
type ClockV1 interface {
	Now(ctx context.Context) (time.Time, error)
}

// GroupsV1 says who belongs to which group, and what they are in it.
type GroupsV1 interface {
	Memberships(ctx context.Context, user string) (map[string]string, error)
	RoleOf(ctx context.Context, group string, user string) (string, error)
}

// JournalV1 is told what happened to a task, for whoever keeps a feed of it.
type JournalV1 interface {
	// Involve says to everyone who sees a task that a member brought another one
	// into it; the actor and the member brought in are in the event.
	Involve(ctx context.Context, kind string, t task.Task, actor string, user string) error
	Record(ctx context.Context, kind string, t task.Task) error
	// Tell says what a member did to a task to one member alone: the actor and the
	// member told are in the event.
	Tell(ctx context.Context, kind string, t task.Task, actor string, to string) error
}

// PeopleV1 is the directory of the members, asked for a name alone.
type PeopleV1 interface {
	ByID(ctx context.Context, id string) (Person, bool, error)
}

// TaskStoreV1 keeps the tasks.
type TaskStoreV1 interface {
	// Change applies a change to the task of an id as it is kept at that instant,
	// alone among the writers of the store, and keeps the result; an error of the
	// change keeps nothing.
	Change(ctx context.Context, id string, change func(*task.Task) error) (task.Task, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (task.Task, error)
	List(ctx context.Context) ([]task.Task, error)
	NewID(ctx context.Context) (string, error)
	Save(ctx context.Context, t task.Task) error
}

// AddTask adds an open task to its owner's list, or to the list of a group the
// owner belongs to, assigned or not.
type AddTask interface {
	Execute(ctx context.Context, in AddTaskInput) (AddTaskOutput, error)
}

// AddTaskInput is AddTask's input: a field added breaks no caller.
type AddTaskInput struct {
	Owner      string
	Title      string
	Notes      string
	Priority   string
	GroupID    string
	AssigneeID string
	Due        time.Time
}

// AddTaskOutput is AddTask's output.
type AddTaskOutput struct {
	Task task.Task
}

// AddTaskHandler runs AddTask: its body is the hand-written execute.
type AddTaskHandler struct {
	clock     ClockV1
	groups    GroupsV1
	taskStore TaskStoreV1
	journal   JournalV1
}

// NewAddTask builds AddTask on what it calls.
func NewAddTask(clock ClockV1, groups GroupsV1, taskStore TaskStoreV1, journal JournalV1) *AddTaskHandler {
	return &AddTaskHandler{clock: clock, groups: groups, taskStore: taskStore, journal: journal}
}

// Execute runs AddTask.
func (h *AddTaskHandler) Execute(ctx context.Context, in AddTaskInput) (AddTaskOutput, error) {
	return h.execute(ctx, &in)
}

var _ AddTask = (*AddTaskHandler)(nil)

// CountTasks counts the tasks per state, and the open ones past their due
// date.
type CountTasks interface {
	Execute(ctx context.Context, in CountTasksInput) (CountTasksOutput, error)
}

// CountTasksInput is CountTasks's input: a field added breaks no caller.
type CountTasksInput struct {
}

// CountTasksOutput is CountTasks's output.
type CountTasksOutput struct {
	Counts  map[string]int
	Total   int
	Overdue int
}

// CountTasksHandler runs CountTasks: its body is the hand-written execute.
type CountTasksHandler struct {
	clock     ClockV1
	taskStore TaskStoreV1
}

// NewCountTasks builds CountTasks on what it calls.
func NewCountTasks(clock ClockV1, taskStore TaskStoreV1) *CountTasksHandler {
	return &CountTasksHandler{clock: clock, taskStore: taskStore}
}

// Execute runs CountTasks.
func (h *CountTasksHandler) Execute(ctx context.Context, in CountTasksInput) (CountTasksOutput, error) {
	return h.execute(ctx, &in)
}

var _ CountTasks = (*CountTasksHandler)(nil)

// CountViews counts what a member's sidebar shows, the tasks left to do in
// each view, the overdue ones, those finished this week, and the open ones of
// each of their groups.
type CountViews interface {
	Execute(ctx context.Context, in CountViewsInput) (CountViewsOutput, error)
}

// CountViewsInput is CountViews's input: a field added breaks no caller.
type CountViewsInput struct {
	User string
	TZ   string
}

// CountViewsOutput is CountViews's output.
type CountViewsOutput struct {
	Inbox             int
	Today             int
	Upcoming          int
	Shared            int
	Assigned          int
	Overdue           int
	CompletedThisWeek int
	Groups            map[string]int
}

// CountViewsHandler runs CountViews: its body is the hand-written execute.
type CountViewsHandler struct {
	groups    GroupsV1
	clock     ClockV1
	taskStore TaskStoreV1
}

// NewCountViews builds CountViews on what it calls.
func NewCountViews(groups GroupsV1, clock ClockV1, taskStore TaskStoreV1) *CountViewsHandler {
	return &CountViewsHandler{groups: groups, clock: clock, taskStore: taskStore}
}

// Execute runs CountViews.
func (h *CountViewsHandler) Execute(ctx context.Context, in CountViewsInput) (CountViewsOutput, error) {
	return h.execute(ctx, &in)
}

var _ CountViews = (*CountViewsHandler)(nil)

// DeleteTask deletes a task, for its owner, or for an owner or an admin of its
// group.
type DeleteTask interface {
	Execute(ctx context.Context, in DeleteTaskInput) (DeleteTaskOutput, error)
}

// DeleteTaskInput is DeleteTask's input: a field added breaks no caller.
type DeleteTaskInput struct {
	User string
	ID   string
}

// DeleteTaskOutput is DeleteTask's output.
type DeleteTaskOutput struct {
}

// DeleteTaskHandler runs DeleteTask: its body is the hand-written execute.
type DeleteTaskHandler struct {
	taskStore TaskStoreV1
	groups    GroupsV1
	journal   JournalV1
}

// NewDeleteTask builds DeleteTask on what it calls.
func NewDeleteTask(taskStore TaskStoreV1, groups GroupsV1, journal JournalV1) *DeleteTaskHandler {
	return &DeleteTaskHandler{taskStore: taskStore, groups: groups, journal: journal}
}

// Execute runs DeleteTask.
func (h *DeleteTaskHandler) Execute(ctx context.Context, in DeleteTaskInput) (DeleteTaskOutput, error) {
	return h.execute(ctx, &in)
}

var _ DeleteTask = (*DeleteTaskHandler)(nil)

// DueSoon lists the open tasks due within a delay, the soonest first.
type DueSoon interface {
	Execute(ctx context.Context, in DueSoonInput) (DueSoonOutput, error)
}

// DueSoonInput is DueSoon's input: a field added breaks no caller.
type DueSoonInput struct {
	Within time.Duration
}

// DueSoonOutput is DueSoon's output.
type DueSoonOutput struct {
	Tasks []task.Task
}

// DueSoonHandler runs DueSoon: its body is the hand-written execute.
type DueSoonHandler struct {
	clock     ClockV1
	taskStore TaskStoreV1
}

// NewDueSoon builds DueSoon on what it calls.
func NewDueSoon(clock ClockV1, taskStore TaskStoreV1) *DueSoonHandler {
	return &DueSoonHandler{clock: clock, taskStore: taskStore}
}

// Execute runs DueSoon.
func (h *DueSoonHandler) Execute(ctx context.Context, in DueSoonInput) (DueSoonOutput, error) {
	return h.execute(ctx, &in)
}

var _ DueSoon = (*DueSoonHandler)(nil)

// EditTask changes what a task says, its title, notes, priority, assignee or
// due date, each only when given, for a member who sees it; an empty assignee
// clears it. Its owner alone moves it to another group they are in, or to
// none, and an assignee who is not in the new group is cleared.
type EditTask interface {
	Execute(ctx context.Context, in EditTaskInput) (EditTaskOutput, error)
}

// EditTaskInput is EditTask's input: a field added breaks no caller.
type EditTaskInput struct {
	User       string
	ID         string
	Title      *string
	Notes      *string
	Priority   *string
	AssigneeID *string
	GroupID    *string
	Due        *time.Time
}

// EditTaskOutput is EditTask's output.
type EditTaskOutput struct {
	Task task.Task
}

// EditTaskHandler runs EditTask: its body is the hand-written execute.
type EditTaskHandler struct {
	taskStore TaskStoreV1
	groups    GroupsV1
	clock     ClockV1
	journal   JournalV1
}

// NewEditTask builds EditTask on what it calls.
func NewEditTask(taskStore TaskStoreV1, groups GroupsV1, clock ClockV1, journal JournalV1) *EditTaskHandler {
	return &EditTaskHandler{taskStore: taskStore, groups: groups, clock: clock, journal: journal}
}

// Execute runs EditTask.
func (h *EditTaskHandler) Execute(ctx context.Context, in EditTaskInput) (EditTaskOutput, error) {
	return h.execute(ctx, &in)
}

var _ EditTask = (*EditTaskHandler)(nil)

// FollowGroup keeps the tasks consistent with a group that changed: the tasks
// of a deleted group go back to their owners' lists, and a task assigned to a
// member who is no longer in its group, and no more its owner's or shared,
// loses its assignee. Doing it twice changes nothing.
type FollowGroup interface {
	Execute(ctx context.Context, in FollowGroupInput) (FollowGroupOutput, error)
}

// FollowGroupInput is FollowGroup's input: a field added breaks no caller.
type FollowGroupInput struct {
	Group   string
	Deleted bool
}

// FollowGroupOutput is FollowGroup's output.
type FollowGroupOutput struct {
}

// FollowGroupHandler runs FollowGroup: its body is the hand-written execute.
type FollowGroupHandler struct {
	taskStore TaskStoreV1
	groups    GroupsV1
	clock     ClockV1
}

// NewFollowGroup builds FollowGroup on what it calls.
func NewFollowGroup(taskStore TaskStoreV1, groups GroupsV1, clock ClockV1) *FollowGroupHandler {
	return &FollowGroupHandler{taskStore: taskStore, groups: groups, clock: clock}
}

// Execute runs FollowGroup.
func (h *FollowGroupHandler) Execute(ctx context.Context, in FollowGroupInput) (FollowGroupOutput, error) {
	return h.execute(ctx, &in)
}

var _ FollowGroup = (*FollowGroupHandler)(nil)

// GetTask reads one task, for a member who sees it.
type GetTask interface {
	Execute(ctx context.Context, in GetTaskInput) (GetTaskOutput, error)
}

// GetTaskInput is GetTask's input: a field added breaks no caller.
type GetTaskInput struct {
	User string
	ID   string
}

// GetTaskOutput is GetTask's output.
type GetTaskOutput struct {
	Task task.Task
}

// GetTaskHandler runs GetTask: its body is the hand-written execute.
type GetTaskHandler struct {
	taskStore TaskStoreV1
	groups    GroupsV1
}

// NewGetTask builds GetTask on what it calls.
func NewGetTask(taskStore TaskStoreV1, groups GroupsV1) *GetTaskHandler {
	return &GetTaskHandler{taskStore: taskStore, groups: groups}
}

// Execute runs GetTask.
func (h *GetTaskHandler) Execute(ctx context.Context, in GetTaskInput) (GetTaskOutput, error) {
	return h.execute(ctx, &in)
}

var _ GetTask = (*GetTaskHandler)(nil)

// ListTasks lists the tasks a member sees, theirs, those shared with them,
// assigned to them and on their groups' lists, of a view (inbox, today,
// upcoming, shared, assigned, completed or group), of a state, or every one
// not archived, the most urgent first; a group narrows the list, and a time
// zone says when the day ends.
type ListTasks interface {
	Execute(ctx context.Context, in ListTasksInput) (ListTasksOutput, error)
}

// ListTasksInput is ListTasks's input: a field added breaks no caller.
type ListTasksInput struct {
	Owner string
	State string
	View  string
	Group string
	TZ    string
}

// ListTasksOutput is ListTasks's output.
type ListTasksOutput struct {
	Tasks []task.Task
}

// ListTasksHandler runs ListTasks: its body is the hand-written execute.
type ListTasksHandler struct {
	groups    GroupsV1
	clock     ClockV1
	taskStore TaskStoreV1
}

// NewListTasks builds ListTasks on what it calls.
func NewListTasks(groups GroupsV1, clock ClockV1, taskStore TaskStoreV1) *ListTasksHandler {
	return &ListTasksHandler{groups: groups, clock: clock, taskStore: taskStore}
}

// Execute runs ListTasks.
func (h *ListTasksHandler) Execute(ctx context.Context, in ListTasksInput) (ListTasksOutput, error) {
	return h.execute(ctx, &in)
}

var _ ListTasks = (*ListTasksHandler)(nil)

// MoveTask moves a task through its lifecycle by one event (complete, reopen,
// archive, restore), for a member who sees it; a task they do not see is not
// found.
type MoveTask interface {
	Execute(ctx context.Context, in MoveTaskInput) (MoveTaskOutput, error)
}

// MoveTaskInput is MoveTask's input: a field added breaks no caller.
type MoveTaskInput struct {
	Owner string
	ID    string
	Event string
}

// MoveTaskOutput is MoveTask's output.
type MoveTaskOutput struct {
	Task task.Task
}

// MoveTaskHandler runs MoveTask: its body is the hand-written execute.
type MoveTaskHandler struct {
	taskStore TaskStoreV1
	groups    GroupsV1
	clock     ClockV1
	journal   JournalV1
}

// NewMoveTask builds MoveTask on what it calls.
func NewMoveTask(taskStore TaskStoreV1, groups GroupsV1, clock ClockV1, journal JournalV1) *MoveTaskHandler {
	return &MoveTaskHandler{taskStore: taskStore, groups: groups, clock: clock, journal: journal}
}

// Execute runs MoveTask.
func (h *MoveTaskHandler) Execute(ctx context.Context, in MoveTaskInput) (MoveTaskOutput, error) {
	return h.execute(ctx, &in)
}

var _ MoveTask = (*MoveTaskHandler)(nil)

// NamePeople names the people of the tasks it is given, every owner, assignee
// and member a task is shared with, each once, in the order they first appear;
// the directory is asked once for each, and one it does not know is left out.
type NamePeople interface {
	Execute(ctx context.Context, in NamePeopleInput) (NamePeopleOutput, error)
}

// NamePeopleInput is NamePeople's input: a field added breaks no caller.
type NamePeopleInput struct {
	Tasks []task.Task
}

// NamePeopleOutput is NamePeople's output.
type NamePeopleOutput struct {
	People []Person
}

// NamePeopleHandler runs NamePeople: its body is the hand-written execute.
type NamePeopleHandler struct {
	people PeopleV1
}

// NewNamePeople builds NamePeople on what it calls.
func NewNamePeople(people PeopleV1) *NamePeopleHandler {
	return &NamePeopleHandler{people: people}
}

// Execute runs NamePeople.
func (h *NamePeopleHandler) Execute(ctx context.Context, in NamePeopleInput) (NamePeopleOutput, error) {
	return h.execute(ctx, &in)
}

var _ NamePeople = (*NamePeopleHandler)(nil)

// ShareTask shares a task with a contact of its owner, who then sees it.
type ShareTask interface {
	Execute(ctx context.Context, in ShareTaskInput) (ShareTaskOutput, error)
}

// ShareTaskInput is ShareTask's input: a field added breaks no caller.
type ShareTaskInput struct {
	User string
	ID   string
	With string
}

// ShareTaskOutput is ShareTask's output.
type ShareTaskOutput struct {
	Task task.Task
}

// ShareTaskHandler runs ShareTask: its body is the hand-written execute.
type ShareTaskHandler struct {
	taskStore TaskStoreV1
	groups    GroupsV1
	circle    CircleV1
	clock     ClockV1
	journal   JournalV1
}

// NewShareTask builds ShareTask on what it calls.
func NewShareTask(taskStore TaskStoreV1, groups GroupsV1, circle CircleV1, clock ClockV1, journal JournalV1) *ShareTaskHandler {
	return &ShareTaskHandler{taskStore: taskStore, groups: groups, circle: circle, clock: clock, journal: journal}
}

// Execute runs ShareTask.
func (h *ShareTaskHandler) Execute(ctx context.Context, in ShareTaskInput) (ShareTaskOutput, error) {
	return h.execute(ctx, &in)
}

var _ ShareTask = (*ShareTaskHandler)(nil)

// UnshareTask ends the sharing of a task with a member, for its owner, or for
// that member who leaves it.
type UnshareTask interface {
	Execute(ctx context.Context, in UnshareTaskInput) (UnshareTaskOutput, error)
}

// UnshareTaskInput is UnshareTask's input: a field added breaks no caller.
type UnshareTaskInput struct {
	User string
	ID   string
	With string
}

// UnshareTaskOutput is UnshareTask's output.
type UnshareTaskOutput struct {
	Task task.Task
}

// UnshareTaskHandler runs UnshareTask: its body is the hand-written execute.
type UnshareTaskHandler struct {
	taskStore TaskStoreV1
	groups    GroupsV1
	clock     ClockV1
}

// NewUnshareTask builds UnshareTask on what it calls.
func NewUnshareTask(taskStore TaskStoreV1, groups GroupsV1, clock ClockV1) *UnshareTaskHandler {
	return &UnshareTaskHandler{taskStore: taskStore, groups: groups, clock: clock}
}

// Execute runs UnshareTask.
func (h *UnshareTaskHandler) Execute(ctx context.Context, in UnshareTaskInput) (UnshareTaskOutput, error) {
	return h.execute(ctx, &in)
}

var _ UnshareTask = (*UnshareTaskHandler)(nil)
