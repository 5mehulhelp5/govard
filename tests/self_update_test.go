package tests

import (
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
