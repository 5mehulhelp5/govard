package tests

import (
	"reflect"
	"testing"

	"govard/internal/desktop"
)

func TestDesktopPkgResolveRequestedLogTargetsForTest(t *testing.T) {
	t.Run("default uses discovered targets", func(t *testing.T) {
		targets := desktop.ResolveRequestedLogTargetsForTest("", []string{"web", "php", "db"})
		if !reflect.DeepEqual(targets, []string{"web", "php", "db"}) {
			t.Fatalf("expected discovered targets, got %#v", targets)
		}
	})

	t.Run("all uses discovered targets", func(t *testing.T) {
		targets := desktop.ResolveRequestedLogTargetsForTest("all", []string{"web", "php"})
		if !reflect.DeepEqual(targets, []string{"web", "php"}) {
			t.Fatalf("expected discovered targets, got %#v", targets)
		}
	})

	t.Run("specific service returns single target", func(t *testing.T) {
		targets := desktop.ResolveRequestedLogTargetsForTest("php", []string{"web", "php"})
		if !reflect.DeepEqual(targets, []string{"php"}) {
			t.Fatalf("expected specific target, got %#v", targets)
		}
	})

	t.Run("missing discovered targets fall back to web", func(t *testing.T) {
		targets := desktop.ResolveRequestedLogTargetsForTest("all", nil)
		if !reflect.DeepEqual(targets, []string{"web"}) {
			t.Fatalf("expected fallback web target, got %#v", targets)
		}
	})
}

func TestDesktopPkgPrefixServiceLogLinesForTest(t *testing.T) {
	raw := "line one\nline two"
	prefixed := desktop.PrefixServiceLogLinesForTest("php", raw)
	expected := "[php] line one\n[php] line two"
	if prefixed != expected {
		t.Fatalf("expected %q, got %q", expected, prefixed)
	}
}
