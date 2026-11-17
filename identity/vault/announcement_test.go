package vault

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

// The topic "accounts" tells the id of the member whose address was proven,
// never the address: a listener that needs it asks the directory. Every field
// of the announcement is filled with what a real one holds, and its JSON is
// read back as the topic carries it.
func TestAnAnnouncementCarriesNoAddress(t *testing.T) {
	msg := Announcement{ID: "account_1", Kind: KindVerified, UserID: "user_ada", At: time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC)}
	typ := reflect.TypeOf(msg)
	for i := range typ.NumField() {
		if f := typ.Field(i); strings.Contains(strings.ToLower(f.Name), "mail") || strings.Contains(strings.ToLower(f.Tag.Get("json")), "mail") {
			t.Errorf("the field %s of the announcement reads as an address", f.Name)
		}
	}
	raw, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	var members map[string]any
	if err := json.Unmarshal(raw, &members); err != nil {
		t.Fatal(err)
	}
	if _, has := members["email"]; has {
		t.Errorf("the announcement %s has an email member", raw)
	}
	for name, v := range members {
		if s, ok := v.(string); ok && strings.Contains(s, "@") {
			t.Errorf("the member %q holds %q, which reads as an address", name, s)
		}
	}
	if len(members) != typ.NumField() {
		t.Errorf("the announcement says %d members for %d fields: %s", len(members), typ.NumField(), raw)
	}
}
