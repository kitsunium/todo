// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:8d40d0873f391460466a0aa214adea3d028e9293117fc8bd920c20c5af44af3e

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
