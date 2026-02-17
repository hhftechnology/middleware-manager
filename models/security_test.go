package models

import (
	"testing"
)

func TestDefaultSecureHeaders(t *testing.T) {
	h := DefaultSecureHeaders()

	checks := map[string]string{
		"XContentTypeOptions": h.XContentTypeOptions,
		"XFrameOptions":       h.XFrameOptions,
		"XXSSProtection":      h.XXSSProtection,
		"HSTS":                h.HSTS,
		"ReferrerPolicy":      h.ReferrerPolicy,
	}

	expected := map[string]string{
		"XContentTypeOptions": "nosniff",
		"XFrameOptions":       "SAMEORIGIN",
		"XXSSProtection":      "1; mode=block",
		"HSTS":                "max-age=31536000; includeSubDomains",
		"ReferrerPolicy":      "strict-origin-when-cross-origin",
	}

	for field, got := range checks {
		want := expected[field]
		if got != want {
			t.Errorf("%s = %q, want %q", field, got, want)
		}
	}

	// CSP and PermissionsPolicy default to empty
	if h.CSP != "" {
		t.Errorf("CSP = %q, want empty", h.CSP)
	}
	if h.PermissionsPolicy != "" {
		t.Errorf("PermissionsPolicy = %q, want empty", h.PermissionsPolicy)
	}
}

func TestTLSHardeningOptions(t *testing.T) {
	opts := TLSHardeningOptions()

	if opts["minVersion"] != "VersionTLS12" {
		t.Errorf("minVersion = %v", opts["minVersion"])
	}
	if opts["maxVersion"] != "VersionTLS13" {
		t.Errorf("maxVersion = %v", opts["maxVersion"])
	}
	if opts["sniStrict"] != true {
		t.Errorf("sniStrict = %v", opts["sniStrict"])
	}

	ciphers, ok := opts["cipherSuites"].([]string)
	if !ok {
		t.Fatal("cipherSuites should be []string")
	}
	if len(ciphers) != 6 {
		t.Errorf("len(cipherSuites) = %d, want 6", len(ciphers))
	}

	curves, ok := opts["curvePreferences"].([]string)
	if !ok {
		t.Fatal("curvePreferences should be []string")
	}
	if len(curves) != 3 {
		t.Errorf("len(curvePreferences) = %d, want 3", len(curves))
	}
}
