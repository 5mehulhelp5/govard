package tests

import (
	"testing"

	"govard/internal/proxy"
)

func TestEnsureTLSConfigAddsTestPolicyAndListenPorts(t *testing.T) {
	config := map[string]interface{}{}

	changed := proxy.EnsureTLSConfigForTest(config)
	if !changed {
		t.Fatalf("Expected ensureTLSConfig to report changes")
	}

	apps, ok := config["apps"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected apps to be a map")
	}

	http, ok := apps["http"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected http to be a map")
	}

	servers, ok := http["servers"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected servers to be a map")
	}

	srv0, ok := servers["srv0"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected srv0 to be a map")
	}

	listen, ok := srv0["listen"].([]interface{})
	if !ok {
		t.Fatalf("Expected listen to be a slice")
	}

	if !proxy.StringSliceContainsForTest(listen, ":80") || !proxy.StringSliceContainsForTest(listen, ":443") {
		t.Fatalf("Expected listen to include :80 and :443")
	}

	tls, ok := apps["tls"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected tls to be a map")
	}
	automation, ok := tls["automation"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected automation to be a map")
	}
	policies, ok := automation["policies"].([]interface{})
	if !ok {
		t.Fatalf("Expected policies to be a slice")
	}

	if !proxy.PolicyIncludesSubjectForTest(policies, "*.test") {
		t.Fatalf("Expected policies to include *.test")
	}
	if !proxy.PolicyIncludesSubjectForTest(policies, "*.govard.test") {
		t.Fatalf("Expected policies to include *.govard.test")
	}
}
