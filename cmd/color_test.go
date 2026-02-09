package cmd

import (
	"strings"
	"testing"
)

func TestColorizer_Determinism(t *testing.T) {
	c := NewColorizer(true)
	first := c.Sprint("manager-a", "some text")
	second := c.Sprint("manager-a", "some text")
	if first != second {
		t.Errorf("expected deterministic output, got %q and %q", first, second)
	}
}

func TestColorizer_DifferentManagers(t *testing.T) {
	c := NewColorizer(true)
	a := c.Sprint("manager-a", "text")
	b := c.Sprint("manager-b", "text")
	// Different managers should (with high probability) produce different colored output.
	// Both should contain the original text.
	if !strings.Contains(a, "text") {
		t.Errorf("expected output to contain original text, got %q", a)
	}
	if !strings.Contains(b, "text") {
		t.Errorf("expected output to contain original text, got %q", b)
	}
}

func TestColorizer_Disabled(t *testing.T) {
	c := NewColorizer(false)
	got := c.Sprint("manager-a", "plain text")
	if got != "plain text" {
		t.Errorf("expected unmodified string when disabled, got %q", got)
	}
}

func TestColorizer_Nil(t *testing.T) {
	var c *Colorizer
	got := c.Sprint("manager-a", "plain text")
	if got != "plain text" {
		t.Errorf("expected unmodified string for nil colorizer, got %q", got)
	}
}

func TestGetInfoOr_NilColorizer(t *testing.T) {
	node := &Node{
		Managers: []ManagerInfo{
			{Manager: "mgr", Operation: "Apply"},
		},
	}
	got := getInfoOr(node, "default", nil)
	if got != "mgr Apply " {
		t.Errorf("unexpected result: %q", got)
	}
}
