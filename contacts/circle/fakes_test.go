// The ports the use cases' tests run on: links and members in memory, a
// clock that stands still, notices that are remembered.

package circle

import (
	"context"
	"strconv"
	"time"

	"github.com/kitsunium/todo/contacts/link"
)

// epoch is the instant every test starts at.
var epoch = time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC)

// world is every port of the domain, in memory, with three members.
type world struct {
	links   map[string]link.Link
	people  map[string]Person
	notices []string
	n       int
}

func newWorld() *world {
	w := &world{links: map[string]link.Link{}, people: map[string]Person{}}
	for _, name := range []string{"ada", "bob", "cy"} {
		w.people["user_"+name] = Person{ID: "user_" + name, Email: name + "@example.com", Name: name}
	}
	return w
}

// linked adds a link from one member to another, in a state.
func (w *world) linked(from, to string, state link.RequestsState) link.Link {
	w.n++
	l := link.Link{ID: "link_" + strconv.Itoa(w.n), Pair: link.PairOf(from, to), RequesterID: from, AddresseeID: to, State: state, CreatedAt: epoch, UpdatedAt: epoch}
	w.links[l.ID] = l
	return l
}

func (w *world) NewID(context.Context) (string, error) {
	w.n++
	return "link_" + strconv.Itoa(w.n), nil
}

func (w *world) Save(_ context.Context, l link.Link) error {
	w.links[l.ID] = l
	return nil
}

func (w *world) Get(_ context.Context, id string) (link.Link, bool, error) {
	l, ok := w.links[id]
	return l, ok, nil
}

func (w *world) ByPair(_ context.Context, pair string) (link.Link, bool, error) {
	for _, l := range w.links {
		if l.Pair == pair {
			return l, true, nil
		}
	}
	return link.Link{}, false, nil
}

func (w *world) Of(_ context.Context, user string) ([]link.Link, error) {
	var out []link.Link
	for _, l := range w.links {
		if _, mine := l.Other(user); mine {
			out = append(out, l)
		}
	}
	return out, nil
}

func (w *world) Change(_ context.Context, id string, change func(*link.Link) error) (link.Link, error) {
	l, ok := w.links[id]
	if !ok {
		return link.Link{}, link.ErrNotFound
	}
	if err := change(&l); err != nil {
		return link.Link{}, err
	}
	w.links[id] = l
	return l, nil
}

func (w *world) Delete(_ context.Context, id string) error {
	delete(w.links, id)
	return nil
}

func (w *world) ByEmail(_ context.Context, email string) (Person, bool, error) {
	for _, p := range w.people {
		if p.Email == email {
			return p, true, nil
		}
	}
	return Person{}, false, nil
}

func (w *world) ByID(_ context.Context, id string) (Person, bool, error) {
	p, ok := w.people[id]
	return p, ok, nil
}

func (w *world) Tell(_ context.Context, kind string, l link.Link, actor Person) error {
	w.notices = append(w.notices, kind+" by "+actor.ID+" on "+l.ID)
	return nil
}

func (w *world) Now(context.Context) (time.Time, error) { return epoch.Add(time.Hour), nil }

func (w *world) request() RequestContact { return NewRequestContact(w, w, w, w) }
func (w *world) answer() AnswerRequest   { return NewAnswerRequest(w, w, w, w) }
