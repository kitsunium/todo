// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package driver

import (
	"context"
	"errors"
	"testing"
	"time"
)

// countingAPI is APIV1 that counts its calls and fails the moves of one id.
type countingAPI struct {
	calls map[string]int
	next  int
}

func (a *countingAPI) Enter(context.Context, string, string, string) error {
	a.calls["Enter"]++
	return nil
}

func (a *countingAPI) Add(context.Context, string, string, time.Time) (string, error) {
	a.calls["Add"]++
	a.next++
	return "task_" + string(rune('a'+a.next)), nil
}

func (a *countingAPI) Move(_ context.Context, _ string, event string) error {
	a.calls["Move"]++
	if event == "archive" {
		return errors.New("refused")
	}
	return nil
}

func (a *countingAPI) List(context.Context) (int, error)     { a.calls["List"]++; return 0, nil }
func (a *countingAPI) Activity(context.Context) (int, error) { a.calls["Activity"]++; return 0, nil }
func (a *countingAPI) Stats(context.Context) (int, error)    { a.calls["Stats"]++; return 0, nil }

// givenDriveDrivesEveryCallOfTheAPI builds Drive for "drives every call of the api": the use case, its input, and
// what its output must be.
func givenDriveDrivesEveryCallOfTheAPI(t *testing.T) (Drive, DriveInput, func(*testing.T, DriveOutput)) {
	t.Helper()
	api := &countingAPI{calls: map[string]int{}}
	return NewDrive(api), DriveInput{For: 0, Rate: 1000}, func(t *testing.T, out DriveOutput) {
		t.Helper()
		for _, call := range []string{"Enter", "Add", "List", "Activity", "Stats"} {
			if api.calls[call] != 1 {
				t.Errorf("%s was called %d time(s) in one round, want 1", call, api.calls[call])
			}
		}
		if api.calls["Move"] != 2 {
			t.Errorf("Move was called %d time(s) in one round, want 2", api.calls["Move"])
		}
		if out.Requests != 7 || out.Failures != 1 {
			t.Errorf("the load counted %d request(s) and %d failure(s), want 7 and the archive that failed", out.Requests, out.Failures)
		}
	}
}

// givenDriveRefusesARateOfZero builds Drive for "refuses a rate of zero": the use case and the input it
// refuses with ErrNoRate.
func givenDriveRefusesARateOfZero(t *testing.T) (Drive, DriveInput) {
	t.Helper()
	return NewDrive(&countingAPI{calls: map[string]int{}}), DriveInput{For: time.Second}
}
