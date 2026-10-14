package services

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/hhftechnology/middleware-manager/database"
	"github.com/hhftechnology/middleware-manager/models"
)

// chainRefMode controls how a resolved local middleware reference is qualified.
type chainRefMode int

const (
	// chainRefProxy emits references for the HTTP provider document. Local
	// middlewares live in that same document, so a stored @file suffix is removed.
	chainRefProxy chainRefMode = iota
	// chainRefFile emits references for resource-overrides.yml. Local
	// middlewares need an explicit @file provider suffix.
	chainRefFile
)

// middlewareRecord is one row from the middlewares table.
type middlewareRecord struct {
	ID     string
	Name   string
	Type   string
	Config string
}

// ResolveMiddlewareRef maps a chain reference to the middleware ID.
// Display names (including names with spaces or capitals) resolve to the
// stored ID. A provider suffix such as @file or @http is preserved.
// References that match neither an ID nor a name are returned unchanged.
func ResolveMiddlewareRef(ref string, ids map[string]struct{}, names map[string]string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ref
	}
	base, suffix, hasSuffix := splitMiddlewareRef(ref)
	if _, ok := ids[base]; ok {
		return joinMiddlewareRef(base, suffix, hasSuffix)
	}
	if id, ok := names[base]; ok {
		return joinMiddlewareRef(id, suffix, hasSuffix)
	}
	lower := strings.ToLower(base)
	for name, id := range names {
		if strings.ToLower(name) == lower {
			return joinMiddlewareRef(id, suffix, hasSuffix)
		}
	}
	return ref
}

func splitMiddlewareRef(ref string) (base, suffix string, hasSuffix bool) {
	if i := strings.LastIndex(ref, "@"); i > 0 {
		return ref[:i], ref[i+1:], true
	}
	return ref, "", false
}

func joinMiddlewareRef(base, suffix string, hasSuffix bool) string {
	if hasSuffix {
		return base + "@" + suffix
	}
	return base
}

func indexMiddlewareRecords(records []middlewareRecord) (map[string]struct{}, map[string]string) {
	ids := make(map[string]struct{}, len(records))
	names := make(map[string]string, len(records))
	for _, mw := range records {
		ids[mw.ID] = struct{}{}
		if mw.Name != "" {
			names[mw.Name] = mw.ID
		}
	}
	return ids, names
}

func chainRefsFromConfig(configStr string) []string {
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(configStr), &cfg); err != nil {
		return nil
	}
	return chainRefStrings(cfg)
}

func chainRefStrings(cfg map[string]interface{}) []string {
	raw, ok := cfg["middlewares"]
	if !ok {
		return nil
	}
	switch list := raw.(type) {
	case []interface{}:
		out := make([]string, 0, len(list))
		for _, item := range list {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return list
	default:
		return nil
	}
}

// expandChainMemberIDs adds chain members so a chain assigned to a router
// also emits the middlewares it references.
func expandChainMemberIDs(records []middlewareRecord, allowed map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{}, len(allowed))
	for id := range allowed {
		out[id] = struct{}{}
	}
	ids, names := indexMiddlewareRecords(records)
	for pass := 0; pass < len(records)+1; pass++ {
		added := false
		for _, mw := range records {
			if mw.Type != "chain" {
				continue
			}
			if _, ok := out[mw.ID]; !ok {
				continue
			}
			for _, ref := range chainRefsFromConfig(mw.Config) {
				resolved := ResolveMiddlewareRef(ref, ids, names)
				base, _, _ := splitMiddlewareRef(resolved)
				if _, known := ids[base]; !known {
					continue
				}
				if _, exists := out[base]; exists {
					continue
				}
				out[base] = struct{}{}
				added = true
			}
		}
		if !added {
			break
		}
	}
	return out
}

func rewriteChainRefs(cfg map[string]interface{}, ids map[string]struct{}, names map[string]string, mode chainRefMode) {
	raw, ok := cfg["middlewares"].([]interface{})
	if !ok {
		return
	}
	for i, item := range raw {
		s, ok := item.(string)
		if !ok {
			continue
		}
		resolved := ResolveMiddlewareRef(s, ids, names)
		raw[i] = qualifyChainRef(resolved, ids, mode)
	}
}

func qualifyChainRef(resolved string, ids map[string]struct{}, mode chainRefMode) string {
	base, suffix, hasSuffix := splitMiddlewareRef(resolved)
	_, known := ids[base]
	switch mode {
	case chainRefFile:
		if known && !hasSuffix {
			return base + "@file"
		}
	case chainRefProxy:
		if known && hasSuffix && suffix == "file" {
			return base
		}
	}
	return resolved
}

func loadMiddlewareRecords(db *database.DB) ([]middlewareRecord, error) {
	rows, err := db.Query("SELECT id, name, type, config FROM middlewares")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []middlewareRecord
	for rows.Next() {
		var mw middlewareRecord
		if err := rows.Scan(&mw.ID, &mw.Name, &mw.Type, &mw.Config); err != nil {
			log.Printf("Failed to scan middleware: %v", err)
			continue
		}
		records = append(records, mw)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

// emitMiddlewareConfigs builds Traefik middleware objects keyed by ID.
// Chain members that use a display name are rewritten to that ID.
func emitMiddlewareConfigs(records []middlewareRecord, mode chainRefMode) map[string]interface{} {
	ids, names := indexMiddlewareRecords(records)
	out := make(map[string]interface{}, len(records))
	for _, mw := range records {
		var middlewareConfig map[string]interface{}
		if err := json.Unmarshal([]byte(mw.Config), &middlewareConfig); err != nil {
			log.Printf("Failed to parse middleware config for %s: %v", mw.ID, err)
			continue
		}
		middlewareConfig = models.ProcessMiddlewareConfig(mw.Type, middlewareConfig)
		if mw.Type == "chain" {
			rewriteChainRefs(middlewareConfig, ids, names, mode)
		}
		out[mw.ID] = map[string]interface{}{
			mw.Type: middlewareConfig,
		}
	}
	return out
}
