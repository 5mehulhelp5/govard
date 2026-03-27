package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"govard/internal/updater"
)

func TestShouldNotifyUpdateForTest(t *testing.T) {
	testCases := []struct {
		name           string
		currentVersion string
		latestTag      string
		want           bool
	}{
		{
			name:           "no latest tag",
			currentVersion: "1.0.0",
			latestTag:      "",
			want:           false,
		},
		{
			name:           "empty current version",
			currentVersion: "",
			latestTag:      "v1.0.1",
			want:           true,
		},
		{
			name:           "same semantic version",
			currentVersion: "1.0.1",
			latestTag:      "v1.0.1",
			want:           false,
		},
		{
			name:           "different semantic version",
			currentVersion: "1.0.1",
			latestTag:      "v1.0.2",
			want:           true,
		},
		{
			name:           "trimmed values",
			currentVersion: " 1.0.1 ",
			latestTag:      " v1.0.1 ",
			want:           false,
		},
		{
			name:           "dev build of same version",
			currentVersion: "1.1.0-2-gf2a0be7",
			latestTag:      "v1.1.0",
			want:           false,
		},
		{
			name:           "dev build of newer version available",
			currentVersion: "1.1.0-dev",
			latestTag:      "v1.2.0",
			want:           true,
		},
		{
			name:           "pre-release build of same version",
			currentVersion: "1.1.0-beta1",
			latestTag:      "v1.1.0",
			want:           false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			got := updater.ShouldNotifyUpdateForTest(testCase.currentVersion, testCase.latestTag)
			if got != testCase.want {
				t.Fatalf("ShouldNotifyUpdateForTest(%q, %q) = %v, want %v", testCase.currentVersion, testCase.latestTag, got, testCase.want)
			}
		})
	}
}

func TestCheckForUpdatesNotifiesWhenNewVersionAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v9.9.9"}`))
	}))
	defer server.Close()

	t.Setenv("GOVARD_UPDATE_CHECK_URL", server.URL)
	defer updater.SetUpdateCheckHTTPClientForTest(server.Client())()

	notifyCalls := 0
	var gotLatestTag, gotCurrentVersion string
	defer updater.SetUpdateCheckNotifierForTest(func(latestTag, currentVersion string) {
		notifyCalls++
		gotLatestTag = latestTag
		gotCurrentVersion = currentVersion
	})()

	updater.CheckForUpdates("1.0.0")

	if notifyCalls != 1 {
		t.Fatalf("expected notifier to be called once, got %d", notifyCalls)
	}
	if gotLatestTag != "v9.9.9" {
		t.Fatalf("latest tag = %q, want %q", gotLatestTag, "v9.9.9")
	}
	if gotCurrentVersion != "1.0.0" {
		t.Fatalf("current version = %q, want %q", gotCurrentVersion, "1.0.0")
	}
}

func TestCheckForUpdatesSkipsNotifierWhenVersionMatches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v1.0.0"}`))
	}))
	defer server.Close()

	t.Setenv("GOVARD_UPDATE_CHECK_URL", server.URL)
	defer updater.SetUpdateCheckHTTPClientForTest(server.Client())()

	notifyCalls := 0
	defer updater.SetUpdateCheckNotifierForTest(func(latestTag, currentVersion string) {
		notifyCalls++
	})()

	updater.CheckForUpdates("1.0.0")

	if notifyCalls != 0 {
		t.Fatalf("expected notifier to be skipped, got %d call(s)", notifyCalls)
	}
}

func TestCheckForUpdatesSkipsNotifierOnInvalidPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{`))
	}))
	defer server.Close()

	t.Setenv("GOVARD_UPDATE_CHECK_URL", server.URL)
	defer updater.SetUpdateCheckHTTPClientForTest(server.Client())()

	notifyCalls := 0
	defer updater.SetUpdateCheckNotifierForTest(func(latestTag, currentVersion string) {
		notifyCalls++
	})()

	updater.CheckForUpdates("1.0.0")

	if notifyCalls != 0 {
		t.Fatalf("expected notifier to be skipped for invalid payload, got %d call(s)", notifyCalls)
	}
}
