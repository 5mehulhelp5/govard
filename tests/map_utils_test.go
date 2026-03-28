package tests

import (
	"testing"

	"govard/internal/engine"
)

func TestMergeMap_NilValues(t *testing.T) {
	dst := map[string]any{
		"volumes": map[string]any{
			"db-data": nil,
		},
	}
	src := map[string]any{
		"volumes": map[string]any{
			"search-data": nil,
		},
	}

	engine.MergeMap(dst, src)

	vols := dst["volumes"].(map[string]any)
	if _, ok := vols["search-data"]; !ok {
		t.Errorf("MergeMap lost 'search-data' which had a nil value")
	}
	if _, ok := vols["db-data"]; !ok {
		t.Errorf("MergeMap lost 'db-data' which was in dst")
	}
}
