// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:0c616a2ddc211ada253896609000dcaab6758455e2991084b86e140dd79690fe

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

// CircleV1 says who is a contact of whom.
type CircleV1 interface {
	AreContacts(ctx context.Context, one string, other string) (bool, error)
}

// ClockV1 tells the time.
type ClockV1 interface {
	Now(ctx context.Context) (time.Time, error)
}

// JournalV1 is told what happened to a task, for whoever keeps a feed of it.
type JournalV1 interface {
	Record(ctx context.Context, kind string, t task.Task) error
}

// TaskStoreV1 keeps the tasks.
type TaskStoreV1 interface {
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (task.Task, error)
	List(ctx context.Context) ([]task.Task, error)
	NewID(ctx context.Context) (string, error)
	Save(ctx context.Context, t task.Task) error
}

// AddTask adds an open task to its owner's list.
type AddTask interface {
	Execute(ctx context.Context, in AddTaskInput) (AddTaskOutput, error)
}

// AddTaskInput is AddTask's input: a field added breaks no caller.
type AddTaskInput struct {
	Owner    string
	Title    string
	Notes    string
	Priority string
	Assignee string
	Due      time.Time
}

// AddTaskOutput is AddTask's output.
type AddTaskOutput struct {
	Task task.Task
}

// AddTaskHandler runs AddTask: its body is the hand-written execute.
type AddTaskHandler struct {
	clock     ClockV1
	taskStore TaskStoreV1
	journal   JournalV1
}

// NewAddTask builds AddTask on what it calls.
func NewAddTask(clock ClockV1, taskStore TaskStoreV1, journal JournalV1) *AddTaskHandler {
	return &AddTaskHandler{clock: clock, taskStore: taskStore, journal: journal}
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

// DeleteTask deletes a task, for its owner alone.
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
	journal   JournalV1
}

// NewDeleteTask builds DeleteTask on what it calls.
func NewDeleteTask(taskStore TaskStoreV1, journal JournalV1) *DeleteTaskHandler {
	return &DeleteTaskHandler{taskStore: taskStore, journal: journal}
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
// due date, each only when given, for a member who sees it.
type EditTask interface {
	Execute(ctx context.Context, in EditTaskInput) (EditTaskOutput, error)
}

// EditTaskInput is EditTask's input: a field added breaks no caller.
type EditTaskInput struct {
	User     string
	ID       string
	Title    *string
	Notes    *string
	Priority *string
	Assignee *string
	Due      *time.Time
}

// EditTaskOutput is EditTask's output.
type EditTaskOutput struct {
	Task task.Task
}

// EditTaskHandler runs EditTask: its body is the hand-written execute.
type EditTaskHandler struct {
	taskStore TaskStoreV1
	clock     ClockV1
	journal   JournalV1
}

// NewEditTask builds EditTask on what it calls.
func NewEditTask(taskStore TaskStoreV1, clock ClockV1, journal JournalV1) *EditTaskHandler {
	return &EditTaskHandler{taskStore: taskStore, clock: clock, journal: journal}
}

// Execute runs EditTask.
func (h *EditTaskHandler) Execute(ctx context.Context, in EditTaskInput) (EditTaskOutput, error) {
	return h.execute(ctx, &in)
}

var _ EditTask = (*EditTaskHandler)(nil)

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
}

// NewGetTask builds GetTask on what it calls.
func NewGetTask(taskStore TaskStoreV1) *GetTaskHandler {
	return &GetTaskHandler{taskStore: taskStore}
}

// Execute runs GetTask.
func (h *GetTaskHandler) Execute(ctx context.Context, in GetTaskInput) (GetTaskOutput, error) {
	return h.execute(ctx, &in)
}

var _ GetTask = (*GetTaskHandler)(nil)

// ListTasks lists the tasks a member sees, theirs and those shared with them,
// of a state, or every one not archived, the most urgent first.
type ListTasks interface {
	Execute(ctx context.Context, in ListTasksInput) (ListTasksOutput, error)
}

// ListTasksInput is ListTasks's input: a field added breaks no caller.
type ListTasksInput struct {
	Owner string
	State string
}

// ListTasksOutput is ListTasks's output.
type ListTasksOutput struct {
	Tasks []task.Task
}

// ListTasksHandler runs ListTasks: its body is the hand-written execute.
type ListTasksHandler struct {
	taskStore TaskStoreV1
}

// NewListTasks builds ListTasks on what it calls.
func NewListTasks(taskStore TaskStoreV1) *ListTasksHandler {
	return &ListTasksHandler{taskStore: taskStore}
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
	clock     ClockV1
	journal   JournalV1
}

// NewMoveTask builds MoveTask on what it calls.
func NewMoveTask(taskStore TaskStoreV1, clock ClockV1, journal JournalV1) *MoveTaskHandler {
	return &MoveTaskHandler{taskStore: taskStore, clock: clock, journal: journal}
}

// Execute runs MoveTask.
func (h *MoveTaskHandler) Execute(ctx context.Context, in MoveTaskInput) (MoveTaskOutput, error) {
	return h.execute(ctx, &in)
}

var _ MoveTask = (*MoveTaskHandler)(nil)

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
	circle    CircleV1
	clock     ClockV1
	journal   JournalV1
}

// NewShareTask builds ShareTask on what it calls.
func NewShareTask(taskStore TaskStoreV1, circle CircleV1, clock ClockV1, journal JournalV1) *ShareTaskHandler {
	return &ShareTaskHandler{taskStore: taskStore, circle: circle, clock: clock, journal: journal}
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
	clock     ClockV1
}

// NewUnshareTask builds UnshareTask on what it calls.
func NewUnshareTask(taskStore TaskStoreV1, clock ClockV1) *UnshareTaskHandler {
	return &UnshareTaskHandler{taskStore: taskStore, clock: clock}
}

// Execute runs UnshareTask.
func (h *UnshareTaskHandler) Execute(ctx context.Context, in UnshareTaskInput) (UnshareTaskOutput, error) {
	return h.execute(ctx, &in)
}

var _ UnshareTask = (*UnshareTaskHandler)(nil)
