package services

import "testing"

func TestMapToOrderedTLSPreservesDomainObjects(t *testing.T) {
	cp := &ConfigProxy{}
	tls := map[string]interface{}{
		"certResolver": "letsencrypt",
		"domains": []interface{}{
			map[string]interface{}{"main": "*.example.com"},
			"example.com",
		},
	}

	ordered := cp.mapToOrderedTLS(tls)
	if ordered.CertResolver != "letsencrypt" {
		t.Fatalf("certResolver = %q", ordered.CertResolver)
	}
	if len(ordered.Domains) != 2 {
		t.Fatalf("expected 2 domains, got %#v", ordered.Domains)
	}
	obj, ok := ordered.Domains[0].(map[string]interface{})
	if !ok || obj["main"] != "*.example.com" {
		t.Fatalf("structured domain was dropped: %#v", ordered.Domains[0])
	}
	if ordered.Domains[1] != "example.com" {
		t.Fatalf("string domain = %#v", ordered.Domains[1])
	}
}
