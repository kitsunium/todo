package todos

import (
	"cmp"
	"context"
	"slices"
	"strings"
	"time"

	"github.com/kitsunium/platform/kit"
)

// The todo list's API. Every write is rate limited: each one persists the
// store and publishes a change.
var (
	_ = Service.Endpoint("GET /todos", List)
	_ = Service.Endpoint("POST /todos", Create, kit.RateLimit(5, 10))
	_ = Service.Endpoint("POST /todos/{id}/complete", Complete, kit.RateLimit(10, 20))
	_ = Service.Endpoint("POST /todos/{id}/reopen", Reopen, kit.RateLimit(10, 20))
	_ = Service.Endpoint("POST /todos/{id}/archive", Archive, kit.RateLimit(10, 20))
	_ = Service.Endpoint("DELETE /todos/{id}", Delete, kit.RateLimit(10, 20))

	// CensusAPI lets other services count todos without reaching into the
	// store: a service's data is its own.
	CensusAPI = Service.Endpoint("GET /internal/census", Census, kit.Private())
)

// ListInput filters the list.
type ListInput struct {
	// Status keeps only the todos in this state. Empty keeps them all.
	Status Status `query:"status"`
}

// ListOutput is the todo list, newest first.
type ListOutput struct {
	Todos []Todo `json:"todos"`
}

// List returns the todos, newest first, optionally filtered by status.
func List(ctx context.Context, in ListInput) (ListOutput, error) {
	all, err := Todos.List(ctx)
	if err != nil {
		return ListOutput{}, err
	}
	out := make([]Todo, 0, len(all))
	for _, t := range all {
		if in.Status == "" || t.Status == in.Status {
			out = append(out, t)
		}
	}
	slices.SortFunc(out, func(a, b Todo) int { return cmp.Compare(b.CreatedAt.UnixNano(), a.CreatedAt.UnixNano()) })
	return ListOutput{Todos: out}, nil
}

// CreateInput is a new todo.
type CreateInput struct {
	Title string     `json:"title" validate:"required,maxlen=200"`
	Due   *time.Time `json:"due,omitempty"`
}

// Create adds a todo. It enters the lifecycle as open.
func Create(ctx context.Context, in CreateInput) (Todo, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return Todo{}, kit.Invalid("the title cannot be blank")
	}
	return Lifecycle.Start(ctx, Todo{
		ID:        kit.NewID("todo"),
		Title:     title,
		Due:       in.Due,
		CreatedAt: time.Now().UTC(),
	})
}

// ByID addresses one todo.
type ByID struct {
	ID string `path:"id"`
}

// Complete marks a todo as done.
func Complete(ctx context.Context, in ByID) (Todo, error) {
	return Lifecycle.Fire(ctx, in.ID, "complete")
}

// Reopen puts a done todo back on the list.
func Reopen(ctx context.Context, in ByID) (Todo, error) {
	return Lifecycle.Fire(ctx, in.ID, "reopen")
}

// Archive takes a done todo off the list, keeping it for the record.
func Archive(ctx context.Context, in ByID) (Todo, error) {
	return Lifecycle.Fire(ctx, in.ID, "archive")
}

// Delete removes a todo for good.
func Delete(ctx context.Context, in ByID) (kit.Empty, error) {
	return kit.Empty{}, Todos.Delete(ctx, in.ID)
}

// CensusOutput counts the todos per state.
type CensusOutput struct {
	Counts map[Status]int `json:"counts"`
	Total  int            `json:"total"`
}

// Census counts the todos per state.
func Census(ctx context.Context, _ kit.Empty) (CensusOutput, error) {
	counts, err := Lifecycle.Census(ctx)
	if err != nil {
		return CensusOutput{}, err
	}
	total := 0
	for _, n := range counts {
		total += n
	}
	return CensusOutput{Counts: counts, Total: total}, nil
}
