// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:370a82d053439cd8e17f0439915150946927d383504b60ee20d20c234deff80b

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
