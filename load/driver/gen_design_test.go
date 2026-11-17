// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:390717aeedfce4c9881bfea93cf567e5dfb6e0f4b99777661b165bb23d396f17

package driver

import (
	"errors"
	"testing"
)

// TestDriveDrivesEveryCallOfTheAPI is Drive's scenario "drives every call of the api" (success).
func TestDriveDrivesEveryCallOfTheAPI(t *testing.T) {
	uc, in, then := givenDriveDrivesEveryCallOfTheAPI(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("Drive: %v, want a success", err)
	}
	then(t, out)
}

// TestDriveRefusesARateOfZero is Drive's scenario "refuses a rate of zero" (refusal).
func TestDriveRefusesARateOfZero(t *testing.T) {
	uc, in := givenDriveRefusesARateOfZero(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, ErrNoRate) {
		t.Fatalf("Drive: %v, want ErrNoRate", err)
	}
}
