package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"govard/internal/cmd"
)

func TestNormalizeReleaseTagForTest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "already prefixed", in: "v1.2.3", want: "v1.2.3"},
		{name: "without prefix", in: "1.2.3", want: "v1.2.3"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := cmd.NormalizeReleaseTagForTest(tt.in); got != tt.want {
				t.Fatalf("NormalizeReleaseTagForTest(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestBuildReleaseAssetNameForTest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		goos       string
		goarch     string
		wantAsset  string
		wantBinary string
		wantErr    bool
	}{
		{
			name:       "linux amd64",
			goos:       "linux",
			goarch:     "amd64",
			wantAsset:  "govard_1.0.1_Linux_amd64.tar.gz",
			wantBinary: "govard",
		},
		{
			name:       "darwin arm64",
			goos:       "darwin",
			goarch:     "arm64",
			wantAsset:  "govard_1.0.1_Darwin_arm64.tar.gz",
			wantBinary: "govard",
		},
		{
			name:       "windows amd64",
			goos:       "windows",
			goarch:     "amd64",
			wantAsset:  "govard_1.0.1_Windows_amd64.zip",
			wantBinary: "govard.exe",
		},
		{
			name:    "unsupported arch",
			goos:    "linux",
			goarch:  "386",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotAsset, gotBinary, err := cmd.BuildReleaseAssetNameForTest("govard", "v1.0.1", tt.goos, tt.goarch)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("BuildReleaseAssetNameForTest() expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("BuildReleaseAssetNameForTest() unexpected error: %v", err)
			}
			if gotAsset != tt.wantAsset {
				t.Fatalf("asset = %q, want %q", gotAsset, tt.wantAsset)
			}
			if gotBinary != tt.wantBinary {
				t.Fatalf("binary = %q, want %q", gotBinary, tt.wantBinary)
			}
		})
	}
}

func TestChecksumForAssetForTest(t *testing.T) {
	t.Parallel()

	const checksums = `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  govard_1.0.1_Linux_amd64.tar.gz
bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb  govard_1.0.1_Darwin_arm64.tar.gz
`
	got, err := cmd.ChecksumForAssetForTest(checksums, "govard_1.0.1_Darwin_arm64.tar.gz")
	if err != nil {
		t.Fatalf("ChecksumForAssetForTest() error = %v", err)
	}
	want := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	if got != want {
		t.Fatalf("checksum = %q, want %q", got, want)
	}
}

func TestChecksumForAssetForTestMissing(t *testing.T) {
	t.Parallel()

	_, err := cmd.ChecksumForAssetForTest("abc  foo.tar.gz\n", "missing.tar.gz")
	if err == nil {
		t.Fatal("ChecksumForAssetForTest() expected error for missing asset")
	}
}

func TestSelfUpdateLatestReleaseURLForTestUsesOverride(t *testing.T) {
	t.Setenv("GOVARD_SELF_UPDATE_LATEST_URL", "http://127.0.0.1:8080/latest/")

	got := cmd.SelfUpdateLatestReleaseURLForTest("ignored/repo")
	want := "http://127.0.0.1:8080/latest"
	if got != want {
		t.Fatalf("latest URL = %q, want %q", got, want)
	}
}

func TestSelfUpdateReleaseBaseURLForTestUsesOverride(t *testing.T) {
	t.Setenv("GOVARD_SELF_UPDATE_RELEASE_BASE_URL", "http://127.0.0.1:8080/releases/")

	got := cmd.SelfUpdateReleaseBaseURLForTest("ignored/repo", "v1.0.2")
	want := "http://127.0.0.1:8080/releases"
	if got != want {
		t.Fatalf("release base URL = %q, want %q", got, want)
	}
}

func TestResolveDesktopUpdateTargetsForTestIncludesPathAndSiblingTargets(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("self-update is not supported on windows")
	}

	tempDir := t.TempDir()
	pathDir := filepath.Join(tempDir, "path")
	cliDir := filepath.Join(tempDir, "cli")
	if err := os.MkdirAll(pathDir, 0o755); err != nil {
		t.Fatalf("mkdir path dir: %v", err)
	}
	if err := os.MkdirAll(cliDir, 0o755); err != nil {
		t.Fatalf("mkdir cli dir: %v", err)
	}

	pathDesktop := filepath.Join(pathDir, "govard-desktop")
	siblingDesktop := filepath.Join(cliDir, "govard-desktop")
	cliBinary := filepath.Join(cliDir, "govard")
	for _, candidate := range []string{pathDesktop, siblingDesktop, cliBinary} {
		if err := os.WriteFile(candidate, []byte(""), 0o755); err != nil {
			t.Fatalf("write %s: %v", candidate, err)
		}
	}

	t.Setenv("PATH", pathDir)

	got := cmd.ResolveDesktopUpdateTargetsForTest(cliBinary)
	if len(got) != 2 {
		t.Fatalf("expected 2 desktop targets, got %d: %v", len(got), got)
	}

	if !slices.Contains(got, pathDesktop) {
		t.Fatalf("expected PATH desktop target %q in %v", pathDesktop, got)
	}
	if !slices.Contains(got, siblingDesktop) {
		t.Fatalf("expected sibling desktop target %q in %v", siblingDesktop, got)
	}
}

func TestResolveDesktopUpdateTargetsForTestDeduplicatesTargets(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("self-update is not supported on windows")
	}

	tempDir := t.TempDir()
	cliDir := filepath.Join(tempDir, "cli")
	if err := os.MkdirAll(cliDir, 0o755); err != nil {
		t.Fatalf("mkdir cli dir: %v", err)
	}

	desktopBinary := filepath.Join(cliDir, "govard-desktop")
	cliBinary := filepath.Join(cliDir, "govard")
	for _, candidate := range []string{desktopBinary, cliBinary} {
		if err := os.WriteFile(candidate, []byte(""), 0o755); err != nil {
			t.Fatalf("write %s: %v", candidate, err)
		}
	}
	t.Setenv("PATH", cliDir)

	got := cmd.ResolveDesktopUpdateTargetsForTest(cliBinary)
	if len(got) != 1 {
		t.Fatalf("expected deduplicated desktop target, got %d: %v", len(got), got)
	}
	if got[0] != desktopBinary {
		t.Fatalf("unexpected desktop target %q, want %q", got[0], desktopBinary)
	}
}

func TestDetectMixedInstallChannelPairsForTestIncludesConflictingCopies(t *testing.T) {
	localDir := filepath.Join(t.TempDir(), "local-bin")
	systemDir := filepath.Join(t.TempDir(), "system-bin")
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		t.Fatalf("mkdir local dir: %v", err)
	}
	if err := os.MkdirAll(systemDir, 0o755); err != nil {
		t.Fatalf("mkdir system dir: %v", err)
	}

	localGovard := filepath.Join(localDir, "govard")
	systemGovard := filepath.Join(systemDir, "govard")
	if err := os.WriteFile(localGovard, []byte("local"), 0o755); err != nil {
		t.Fatalf("write local govard: %v", err)
	}
	if err := os.WriteFile(systemGovard, []byte("system"), 0o755); err != nil {
		t.Fatalf("write system govard: %v", err)
	}

	pairs := cmd.DetectMixedInstallChannelPairsForTest([]string{"govard"}, localDir, systemDir)
	if len(pairs) != 1 {
		t.Fatalf("expected 1 conflicting pair, got %d: %v", len(pairs), pairs)
	}
	if pairs[0][0] != localGovard || pairs[0][1] != systemGovard {
		t.Fatalf("unexpected pair %v", pairs[0])
	}
}

func TestDetectMixedInstallChannelPairsForTestSkipsSameTargetViaSymlink(t *testing.T) {
	localDir := filepath.Join(t.TempDir(), "local-bin")
	systemDir := filepath.Join(t.TempDir(), "system-bin")
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		t.Fatalf("mkdir local dir: %v", err)
	}
	if err := os.MkdirAll(systemDir, 0o755); err != nil {
		t.Fatalf("mkdir system dir: %v", err)
	}

	localGovard := filepath.Join(localDir, "govard")
	systemGovard := filepath.Join(systemDir, "govard")
	if err := os.WriteFile(localGovard, []byte("local"), 0o755); err != nil {
		t.Fatalf("write local govard: %v", err)
	}
	if err := os.Symlink(localGovard, systemGovard); err != nil {
		t.Fatalf("symlink system govard: %v", err)
	}

	pairs := cmd.DetectMixedInstallChannelPairsForTest([]string{"govard"}, localDir, systemDir)
	if len(pairs) != 0 {
		t.Fatalf("expected no conflicting pairs for shared symlink target, got %v", pairs)
	}
}
