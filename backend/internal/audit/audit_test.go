package audit

import "testing"

func TestNormalizeActorType(t *testing.T) {
	if got := normalizeActorType(""); got != "USER" {
		t.Fatalf("expected USER, got %q", got)
	}
	if got := normalizeActorType("AI"); got != "AI" {
		t.Fatalf("expected AI, got %q", got)
	}
}
