package wordpress

import (
	"testing"
)

func TestParseSites(t *testing.T) {
	raw := []map[string]interface{}{
		{
			"name":   "Main Blog",
			"path":   "/var/www/html/blog",
			"run_as": "www-data",
		},
		{
			"path": "/var/www/html/shop",
		},
		{
			"name": "No Path Site",
			// Missing path — should be skipped
		},
	}

	sites := parseSites(raw)

	if len(sites) != 2 {
		t.Fatalf("expected 2 sites, got %d", len(sites))
	}

	if sites[0].Name != "Main Blog" {
		t.Errorf("expected name 'Main Blog', got %q", sites[0].Name)
	}
	if sites[0].Path != "/var/www/html/blog" {
		t.Errorf("expected path '/var/www/html/blog', got %q", sites[0].Path)
	}
	if sites[0].RunAs != "www-data" {
		t.Errorf("expected run_as 'www-data', got %q", sites[0].RunAs)
	}

	// Second site should use path as name (fallback)
	if sites[1].Name != "/var/www/html/shop" {
		t.Errorf("expected name to fall back to path, got %q", sites[1].Name)
	}
	// No default run_as (empty means run as current user, no sudo)
	if sites[1].RunAs != "" {
		t.Errorf("expected empty run_as, got %q", sites[1].RunAs)
	}
}
