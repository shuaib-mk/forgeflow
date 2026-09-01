package validation

import "testing"

func TestSlug(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		value string
		want  bool
	}{{"forge-flow", true}, {"ForgeFlow", false}, {"-bad", false}, {"bad--slug", false}} {
		if got := Slug(tt.value); got != tt.want {
			t.Errorf("Slug(%q) = %v, want %v", tt.value, got, tt.want)
		}
	}
}
