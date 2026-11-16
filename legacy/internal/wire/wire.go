// Package wire holds the conventions every service of the todo follows on
// the wire: what a timestamp looks like, what a one-line text may contain,
// which languages the product speaks, and how a request that breaks a rule
// is refused. It declares no building block — it is shared code, not a
// service, and owns no data.
package wire

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/kitsunium/sdk/framework/app"
)

// Now is the app's time, in UTC, to the millisecond: every timestamp the
// product stores or answers. It reads app.Now, so a test's manual clock and
// the Studio's clock reach every deadline the product computes.
func Now(ctx context.Context) time.Time {
	return app.Now(ctx).UTC().Truncate(time.Millisecond)
}

// Invalid refuses a request because of one field, the way the framework refuses a
// request whose validate tags fail: 400, code invalid_argument, and one
// violation a form can show beside the field.
func Invalid(path, rule, message string) error {
	return &app.Error{
		Status:     http.StatusBadRequest,
		Code:       app.WireInvalid,
		Message:    "the request is invalid",
		Violations: []app.Violation{{Path: path, Rule: rule, Message: message}},
	}
}

// Line cleans a one-line text field — a name, a title: it trims the spaces
// around it, and refuses it blank, longer than max characters, or holding a
// line break or another control character. A title travels into mail
// subjects and notifications, where a line break is an injection.
func Line(path, s string, max int) (string, error) {
	s = strings.TrimSpace(s)
	switch {
	case s == "":
		return "", Invalid(path, "required", "is required")
	case utf8.RuneCountInString(s) > max:
		return "", Invalid(path, "maxlen", "must be at most "+strconv.Itoa(max)+" characters long")
	case strings.IndexFunc(s, unicode.IsControl) >= 0:
		return "", Invalid(path, "line", "must be one line of text")
	}
	return s, nil
}

// Is reports whether err is a framework error with the given code: app.WireNotFound,
// app.WireConflict…
func Is(err error, code string) bool {
	var ke *app.Error
	return errors.As(err, &ke) && ke.Code == code
}
