package treehash

import "testing"

func TestParse(t *testing.T) {
	hash := "0123456789abcdef0123456789abcdef01234567"
	got, ok := Parse(hash)
	if !ok || got != Hash(hash) {
		t.Fatalf("Parse valid = %q, %v", got, ok)
	}
	for _, input := range []string{"", "xyz", "0123456789ABCDEF0123456789ABCDEF01234567"} {
		if got, ok := Parse(input); ok {
			t.Fatalf("Parse(%q) = %q, true", input, got)
		}
	}
}
