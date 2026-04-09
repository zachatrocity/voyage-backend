package handlers

import "testing"

func TestIsLikelyMessageID(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "empty", in: "", want: false},
		{name: "spaces only", in: "   ", want: false},
		{name: "no at-sign", in: "nonexistent-hurl-message-id", want: false},
		{name: "contains spaces", in: "bad id@example.com", want: false},
		{name: "angle-bracket format", in: "<abc@example.com>", want: true},
		{name: "plain format", in: "abc@example.com", want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isLikelyMessageID(tc.in); got != tc.want {
				t.Fatalf("isLikelyMessageID(%q)=%v want %v", tc.in, got, tc.want)
			}
		})
	}
}
