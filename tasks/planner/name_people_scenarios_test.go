// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"slices"
	"testing"

	"github.com/kitsunium/todo/tasks/task"
)

// names are the ids of some people, in order.
func names(ps []Person) []string {
	ids := make([]string, 0, len(ps))
	for _, p := range ps {
		ids = append(ids, p.ID)
	}
	return ids
}

// givenNamePeopleNamesEachPersonOnce builds NamePeople for "names each person once": the use case, its input, and
// what its output must be.
func givenNamePeopleNamesEachPersonOnce(t *testing.T) (NamePeople, NamePeopleInput, func(*testing.T, NamePeopleOutput)) {
	t.Helper()
	people := directory()
	first := task.Task{ID: "task_a", OwnerID: ada, AssigneeID: bob, SharedWith: []string{cy, bob}}
	second := task.Task{ID: "task_b", OwnerID: bob, AssigneeID: ada, SharedWith: []string{cy}}
	return NewNamePeople(people), NamePeopleInput{Tasks: []task.Task{first, second}}, func(t *testing.T, out NamePeopleOutput) {
		t.Helper()
		if got, want := names(out.People), []string{ada, bob, cy}; !slices.Equal(got, want) {
			t.Errorf("the people: %v, want %v, each once in the order they appear", got, want)
		}
		if out.People[0].Name != "Ada" || out.People[2].Name != "Cy" {
			t.Errorf("the names: %+v", out.People)
		}
		if got := slices.Clone(people.asked); len(got) != 3 {
			t.Errorf("the directory was asked %v, want one question for each of the three", got)
		}
	}
}

// givenNamePeopleLeavesOutAMemberTheDirectoryDoesNotKnow builds NamePeople for "leaves out a member the directory does not know": the use case, its input, and
// what its output must be.
func givenNamePeopleLeavesOutAMemberTheDirectoryDoesNotKnow(t *testing.T) (NamePeople, NamePeopleInput, func(*testing.T, NamePeopleOutput)) {
	t.Helper()
	people := directory()
	tk := task.Task{ID: "task_a", OwnerID: ada, AssigneeID: "user_gone", SharedWith: []string{"user_gone", dee}}
	return NewNamePeople(people), NamePeopleInput{Tasks: []task.Task{tk, tk}}, func(t *testing.T, out NamePeopleOutput) {
		t.Helper()
		if got := names(out.People); !slices.Equal(got, []string{ada}) {
			t.Errorf("the people: %v, want Ada alone", got)
		}
		if len(people.asked) != 3 {
			t.Errorf("the directory was asked %v, want one question for each of the three ids, the unknown ones included", people.asked)
		}
	}
}

// givenNamePeopleNamesNobodyForNoTask builds NamePeople for "names nobody for no task": the use case, its input, and
// what its output must be.
func givenNamePeopleNamesNobodyForNoTask(t *testing.T) (NamePeople, NamePeopleInput, func(*testing.T, NamePeopleOutput)) {
	t.Helper()
	people := directory()
	return NewNamePeople(people), NamePeopleInput{}, func(t *testing.T, out NamePeopleOutput) {
		t.Helper()
		if out.People == nil || len(out.People) != 0 || len(people.asked) != 0 {
			t.Errorf("the people: %#v, asked %v, want an empty list and no question", out.People, people.asked)
		}
	}
}
