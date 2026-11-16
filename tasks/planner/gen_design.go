// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:dd6d9d76cd70cb3afe4d348b4ed0e2bdf2f8aa9f13db84b61b2cb1c34cdd43a0

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

// ArchiveDone archives the tasks done for longer than a delay.
type ArchiveDone interface {
	Execute(ctx context.Context, in ArchiveDoneInput) (ArchiveDoneOutput, error)
}

// ArchiveDoneInput is ArchiveDone's input: a field added breaks no caller.
type ArchiveDoneInput struct {
	After time.Duration
}

// ArchiveDoneOutput is ArchiveDone's output.
type ArchiveDoneOutput struct {
	Archived int
}

// ArchiveDoneHandler runs ArchiveDone: its body is the hand-written execute.
type ArchiveDoneHandler struct {
	clock     ClockV1
	taskStore TaskStoreV1
	journal   JournalV1
}

// NewArchiveDone builds ArchiveDone on what it calls.
func NewArchiveDone(clock ClockV1, taskStore TaskStoreV1, journal JournalV1) *ArchiveDoneHandler {
	return &ArchiveDoneHandler{clock: clock, taskStore: taskStore, journal: journal}
}

// Execute runs ArchiveDone.
func (h *ArchiveDoneHandler) Execute(ctx context.Context, in ArchiveDoneInput) (ArchiveDoneOutput, error) {
	return h.execute(ctx, &in)
}

var _ ArchiveDone = (*ArchiveDoneHandler)(nil)

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

// ListTasks lists a member's tasks of a state, or every one not archived, the
// most urgent first.
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

// MoveTask moves a task of its owner through its lifecycle by one event
// (complete, reopen, archive, restore); another member's task is not found.
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
