package handlers

import (
	"net/http"
	"strings"
	"time"
)

type probeResult struct {
	ok        bool
	latencyMS int64
	errorText string
}

func probeTraefik(client *http.Client, baseURL string) probeResult {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		return probeResult{ok: false, errorText: "url is required"}
	}

	endpoints := []string{"/api/version", "/ping"}
	var lastErr string
	for _, endpoint := range endpoints {
		start := time.Now()
		resp, err := client.Get(base + endpoint)
		if err != nil {
			lastErr = "Connection failed"
			continue
		}
		func() {
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode == http.StatusOK {
				lastErr = ""
				return
			}
			lastErr = resp.Status
		}()
		if lastErr == "" {
			return probeResult{ok: true, latencyMS: time.Since(start).Milliseconds()}
		}
	}

	if lastErr == "" {
		lastErr = "Connection failed"
	}
	return probeResult{ok: false, errorText: lastErr}
}
