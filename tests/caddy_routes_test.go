package tests

import (
	"testing"

	"govard/internal/proxy"
)

func TestUpsertDomainRouteIsIdempotent(t *testing.T) {
	config := map[string]interface{}{}

	if changed := proxy.UpsertDomainRouteForTest(config, "demo.test", "demo-web-1"); !changed {
		t.Fatal("expected first upsert to change config")
	}

	if changed := proxy.UpsertDomainRouteForTest(config, "demo.test", "demo-web-1"); changed {
		t.Fatal("expected second upsert with same target to be idempotent")
	}

	if changed := proxy.UpsertDomainRouteForTest(config, "demo.test", "demo-varnish-1"); !changed {
		t.Fatal("expected upsert with different target to change config")
	}

	routes := extractRoutes(t, config)
	if len(routes) != 1 {
		t.Fatalf("expected exactly one route for domain, got %d", len(routes))
	}
}

func TestRemoveDomainRoute(t *testing.T) {
	config := map[string]interface{}{}
	_ = proxy.UpsertDomainRouteForTest(config, "demo.test", "demo-web-1")

	if changed := proxy.RemoveDomainRouteForTest(config, "demo.test"); !changed {
		t.Fatal("expected remove to change config")
	}

	if changed := proxy.RemoveDomainRouteForTest(config, "demo.test"); changed {
		t.Fatal("expected second remove to be a no-op")
	}
}

func extractRoutes(t *testing.T, config map[string]interface{}) []interface{} {
	t.Helper()
	apps, ok := config["apps"].(map[string]interface{})
	if !ok {
		t.Fatal("missing apps map")
	}
	http, ok := apps["http"].(map[string]interface{})
	if !ok {
		t.Fatal("missing http map")
	}
	servers, ok := http["servers"].(map[string]interface{})
	if !ok {
		t.Fatal("missing servers map")
	}
	srv0, ok := servers["srv0"].(map[string]interface{})
	if !ok {
		t.Fatal("missing srv0 map")
	}
	routes, ok := srv0["routes"].([]interface{})
	if !ok {
		t.Fatal("missing routes slice")
	}
	return routes
}
