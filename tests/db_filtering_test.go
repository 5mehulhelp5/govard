package tests

import (
	"govard/internal/cmd"
	"testing"
)

func TestBuildIgnoredTableArgsIsolation(t *testing.T) {
	tests := []struct {
		name      string
		noNoise   bool
		noPII     bool
		framework string
		want      []string
		notWant   []string
	}{
		{
			name:      "Magento: only no-noise",
			noNoise:   true,
			noPII:     false,
			framework: "magento2",
			want:      []string{"--ignore-table=magento.cache_tag"},
			notWant:   []string{"--ignore-table=magento.customer_entity", "--ignore-table=magento.admin_user"},
		},
		{
			name:      "Magento: only no-pii",
			noNoise:   false,
			noPII:     true,
			framework: "magento2",
			want:      []string{"--ignore-table=magento.customer_entity", "--ignore-table=magento.admin_user"},
			notWant:   []string{"--ignore-table=magento.cache_tag"},
		},
		{
			name:      "Magento: both",
			noNoise:   true,
			noPII:     true,
			framework: "magento2",
			want:      []string{"--ignore-table=magento.cache_tag", "--ignore-table=magento.customer_entity"},
		},
		{
			name:      "Laravel: only no-noise",
			noNoise:   true,
			noPII:     false,
			framework: "laravel",
			want:      []string{"--ignore-table=magento.sessions"},
			notWant:   []string{"--ignore-table=magento.users"},
		},
		{
			name:      "Wordpress: only no-pii",
			noNoise:   false,
			noPII:     true,
			framework: "wordpress",
			want:      []string{"--ignore-table=magento.users"},
			notWant:   []string{"--ignore-table=magento.wflogs"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cmd.BuildIgnoredTableArgsForTest("magento", "", tt.noNoise, tt.noPII, tt.framework)

			for _, w := range tt.want {
				found := false
				for _, g := range got {
					if g == w {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected %s in output, but missing. Got: %v", w, got)
				}
			}

			for _, nw := range tt.notWant {
				for _, g := range got {
					if g == nw {
						t.Errorf("did NOT expect %s in output, but found it. Got: %v", nw, got)
					}
				}
			}
		})
	}
}
