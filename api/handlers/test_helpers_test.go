package handlers

import (
	"encoding/json"
	"testing"
)

func mustUnmarshalResponse(t *testing.T, data []byte, target interface{}) {
	t.Helper()

	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func mustScan(t *testing.T, err error, context string) {
	t.Helper()

	if err != nil {
		t.Fatalf("%s: %v", context, err)
	}
}
