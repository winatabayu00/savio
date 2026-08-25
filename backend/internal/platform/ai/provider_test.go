package ai

import (
	"context"

	"testing"
	"time"
)

func TestMockReturnsStructuredJSON(t *testing.T) {
	p := Mock{}
	ctx := context.Background()
	for _, marker := range []string{"CATEGORIZE", "INSIGHT", "COPILOT", "EXPLAIN"} {
		out, err := p.Complete(ctx, "", marker)
		if err != nil {
			t.Fatalf("%s: %v", marker, err)
		}
		m, err := ExtractJSON(out)
		if err != nil {
			t.Fatalf("%s: %v", marker, err)
		}
		if len(m) == 0 {
			t.Fatalf("%s: empty object", marker)
		}
	}
}

func TestExtractJSONHandlesFencedProse(t *testing.T) {
	m, err := ExtractJSON("Sure! Here is the result:\n\n```json\n{\"answer\":\"ok\"}\n```")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if m["answer"] != "ok" {
		t.Fatalf("answer = %v", m["answer"])
	}
	if _, err := ExtractJSON("no object here"); err == nil {
		t.Fatalf("expected error for non-JSON")
	}
}

func TestInvalidOutputRejected(t *testing.T) {
	out, err := Mock{}.Complete(context.Background(), "", "BROKEN")
	if err != nil {
		t.Fatalf("mock: %v", err)
	}
	if _, err := ExtractJSON(out); err == nil {
		t.Fatalf("broken output must fail validation")
	}
}

func TestRequireStringEnforcesAllowedValues(t *testing.T) {
	if _, err := RequireString(map[string]any{"kind": "EXPENSE"}, "kind", "INCOME", "EXPENSE"); err != nil {
		t.Fatalf("valid: %v", err)
	}
	if _, err := RequireString(map[string]any{"kind": "HACK"}, "kind", "INCOME", "EXPENSE"); err == nil {
		t.Fatalf("unknown enum must be rejected")
	}
}

func TestConfigAdapterDefaults(t *testing.T) {
	p := NewProvider(cfgStub{})
	if _, ok := p.(Mock); !ok {
		t.Fatalf("empty config must produce Mock, got %T", p)
	}
}

type cfgStub struct{}

func (cfgStub) AIProvider() string       { return "mock" }
func (cfgStub) AIBaseURL() string        { return "" }
func (cfgStub) AIAPIKey() string         { return "" }
func (cfgStub) AIModel() string          { return "" }
func (cfgStub) AITimeout() time.Duration { return time.Second }
