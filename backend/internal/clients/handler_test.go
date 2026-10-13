package clients

import "testing"

func TestClientServiceID(t *testing.T) {
	got := clientServiceID("123", "00")
	if got != "123_00" {
		t.Fatalf("clientServiceID() = %q, want %q", got, "123_00")
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		maxLength int
		wantError bool
	}{
		{name: "keeps leading zeroes", value: "00123", maxLength: 30},
		{name: "rejects empty", value: "", maxLength: 30, wantError: true},
		{name: "rejects delimiter", value: "123_00", maxLength: 51, wantError: true},
		{name: "rejects whitespace", value: "12 3", maxLength: 30, wantError: true},
		{name: "rejects over limit", value: "1234", maxLength: 3, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateID(test.value, test.maxLength, "id")
			if (err != nil) != test.wantError {
				t.Fatalf("validateID() error = %v, wantError %t", err, test.wantError)
			}
		})
	}
}

func TestValidateServiceKey(t *testing.T) {
	tests := []struct {
		name       string
		clientID   string
		serviceKey string
		wantError  bool
	}{
		{name: "valid key keeps zeroes", clientID: "123", serviceKey: "123_00"},
		{name: "wrong client", clientID: "234", serviceKey: "123_00", wantError: true},
		{name: "extra delimiter", clientID: "123", serviceKey: "123_0_0", wantError: true},
		{name: "missing service id", clientID: "123", serviceKey: "123_", wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateServiceKey(test.clientID, test.serviceKey)
			if (err != nil) != test.wantError {
				t.Fatalf("validateServiceKey() error = %v, wantError %t", err, test.wantError)
			}
		})
	}
}
