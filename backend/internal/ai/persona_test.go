package ai

import (
	"strings"
	"testing"
)

func TestWithPersona(t *testing.T) {
	system := "you are Savio Copilot, grounded facts only"
	if got := withPersona(system, "balanced"); got != system {
		t.Fatalf("balanced must passthrough system unchanged, got %q", got)
	}
	if got := withPersona(system, "unknown"); got != system {
		t.Fatalf("unknown persona must passthrough system unchanged, got %q", got)
	}
	got := withPersona(system, "lenna")
	if !strings.Contains(got, "Lenna") || !strings.Contains(got, "facts only") {
		t.Fatalf("lenna must prepend persona block while keeping original grounding, got %q", got)
	}
}