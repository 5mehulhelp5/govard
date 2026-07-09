package tests

import (
	"testing"

	"govard/internal/proxy"
)

func TestEnsureSearchServerConfigAddsListenPort(t *testing.T) {
	config := map[string]interface{}{}

	changed := proxy.EnsureSearchServerConfigForTest(config)
	if !changed {
		t.Fatalf("Expected ensureSearchServerConfig to report changes")
	}

	srvSearch := extractServer(t, config, "srv_search")

	listen, ok := srvSearch["listen"].([]interface{})
	if !ok {
		t.Fatalf("Expected listen to be a slice")
	}
	if !proxy.StringSliceContainsForTest(listen, ":9200") {
		t.Fatalf("Expected srv_search to include :9200")
	}

	autoHTTPS, ok := srvSearch["automatic_https"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected srv_search to have automatic_https map, got %#v", srvSearch["automatic_https"])
	}
	if disable, ok := autoHTTPS["disable"].(bool); !ok || !disable {
		t.Fatalf("Expected srv_search automatic_https.disable to be true, got %#v", autoHTTPS["disable"])
	}
}

func TestUpsertSearchRouteIsIdempotentAndIsolatedFromWebRoute(t *testing.T) {
	config := map[string]interface{}{}

	if changed := proxy.UpsertSearchRouteForTest(config, "demo.test", "demo-elasticsearch-1"); !changed {
		t.Fatal("expected first upsert to change config")
	}
	if changed := proxy.UpsertSearchRouteForTest(config, "demo.test", "demo-elasticsearch-1"); changed {
		t.Fatal("expected second upsert with same target to be idempotent")
	}

	// A web route for the same domain must coexist with, not replace, the search route.
	if changed := proxy.UpsertDomainRouteForTest(config, "demo.test", "demo-web-1"); !changed {
		t.Fatal("expected web route upsert to change config")
	}

	searchRoutes := extractRoutesForServer(t, config, "srv_search")
	if len(searchRoutes) != 1 {
		t.Fatalf("expected exactly one search route, got %d", len(searchRoutes))
	}
	webRoutes := extractRoutesForServer(t, config, "srv0")
	if len(webRoutes) != 1 {
		t.Fatalf("expected exactly one web route, got %d", len(webRoutes))
	}
}

func TestRemoveSearchRoute(t *testing.T) {
	config := map[string]interface{}{}
	_ = proxy.UpsertSearchRouteForTest(config, "demo.test", "demo-elasticsearch-1")

	if changed := proxy.RemoveSearchRouteForTest(config, "demo.test"); !changed {
		t.Fatal("expected remove to change config")
	}
	if changed := proxy.RemoveSearchRouteForTest(config, "demo.test"); changed {
		t.Fatal("expected second remove to be a no-op")
	}
}

func extractServer(t *testing.T, config map[string]interface{}, serverName string) map[string]interface{} {
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
	server, ok := servers[serverName].(map[string]interface{})
	if !ok {
		t.Fatalf("missing %s map", serverName)
	}
	return server
}

func extractRoutesForServer(t *testing.T, config map[string]interface{}, serverName string) []interface{} {
	t.Helper()
	server := extractServer(t, config, serverName)
	routes, ok := server["routes"].([]interface{})
	if !ok {
		t.Fatal("missing routes slice")
	}
	return routes
}
