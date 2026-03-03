package tests

import (
	"errors"
	"strings"
	"testing"

	"govard/internal/proxy"
)

func TestIsDefaultFileServerRouteForTest(t *testing.T) {
	positive := map[string]interface{}{
		"handle": []interface{}{
			map[string]interface{}{
				"handler": "vars",
				"root":    "/usr/share/caddy",
			},
			map[string]interface{}{
				"handler": "file_server",
			},
		},
	}

	if !proxy.IsDefaultFileServerRouteForTest(positive) {
		t.Fatal("expected route to be detected as default file-server route")
	}

	negative := map[string]interface{}{
		"match": []interface{}{
			map[string]interface{}{
				"host": []interface{}{"sample.test"},
			},
		},
		"handle": []interface{}{
			map[string]interface{}{
				"handler": "reverse_proxy",
			},
		},
	}

	if proxy.IsDefaultFileServerRouteForTest(negative) {
		t.Fatal("expected matched reverse-proxy route not to be treated as default route")
	}
}

func TestInitCaddyForTestUsesRunnerAndSeedsExpectedConfig(t *testing.T) {
	var gotContainer string
	var gotPayload string

	defer proxy.SetInitCaddyCommandRunnerForTest(func(container string, initJSON string) error {
		gotContainer = container
		gotPayload = initJSON
		return nil
	})()

	err := proxy.InitCaddyForTest("govard-proxy-caddy")
	if err != nil {
		t.Fatalf("InitCaddyForTest() error = %v", err)
	}

	if gotContainer != "govard-proxy-caddy" {
		t.Fatalf("container = %q, want %q", gotContainer, "govard-proxy-caddy")
	}
	if !strings.Contains(gotPayload, `"listen": [":80", ":443"]`) {
		t.Fatalf("init payload missing listen ports: %s", gotPayload)
	}
	if !strings.Contains(gotPayload, `"subjects": ["*.test"]`) {
		t.Fatalf("init payload missing *.test policy: %s", gotPayload)
	}
	if !strings.Contains(gotPayload, `"subjects": ["*.govard.test"]`) {
		t.Fatalf("init payload missing *.govard.test policy: %s", gotPayload)
	}
}

func TestInitCaddyForTestPropagatesRunnerError(t *testing.T) {
	defer proxy.SetInitCaddyCommandRunnerForTest(func(container string, initJSON string) error {
		return errors.New("load failed")
	})()

	err := proxy.InitCaddyForTest("govard-proxy-caddy")
	if err == nil {
		t.Fatal("expected init caddy error")
	}
	if !strings.Contains(err.Error(), "load failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}
