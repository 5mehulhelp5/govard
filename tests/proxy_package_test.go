package tests

import (
	"testing"

	"govard/internal/proxy"
)

func TestProxyPkgEnsureTLSConfigForTest(t *testing.T) {
	cfg := map[string]interface{}{}
	changed := proxy.EnsureTLSConfigForTest(cfg)
	if !changed {
		t.Fatal("expected config to be changed")
	}
}

func TestProxyPkgUpsertAndRemoveDomainRouteForTest(t *testing.T) {
	cfg := map[string]interface{}{}
	proxy.EnsureTLSConfigForTest(cfg)
	if !proxy.UpsertDomainRouteForTest(cfg, "shop.demo.test", "web") {
		t.Fatal("expected first upsert to change config")
	}
	if !proxy.RemoveDomainRouteForTest(cfg, "shop.demo.test") {
		t.Fatal("expected route removal")
	}
}
