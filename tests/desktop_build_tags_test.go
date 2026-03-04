package tests

import (
	"runtime"
	"strings"
	"testing"

	cmdpkg "govard/internal/cmd"
)

func TestDesktopProductionBuildTagsForTest(t *testing.T) {
	tags := cmdpkg.DesktopProductionBuildTagsForTest()
	if !strings.Contains(tags, "desktop") {
		t.Fatalf("expected desktop tag in %q", tags)
	}
	if !strings.Contains(tags, "production") {
		t.Fatalf("expected production tag in %q", tags)
	}

	hasWebkitTag := strings.Contains(tags, "webkit2_41")
	if runtime.GOOS == "linux" && !hasWebkitTag {
		t.Fatalf("expected webkit2_41 tag on linux in %q", tags)
	}
	if runtime.GOOS != "linux" && hasWebkitTag {
		t.Fatalf("did not expect webkit2_41 tag on %s in %q", runtime.GOOS, tags)
	}
}
