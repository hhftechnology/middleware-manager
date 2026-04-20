package handlers

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	tmconfig "github.com/hhftechnology/middleware-manager/internal/traefikmanager/config"
	tmtypes "github.com/hhftechnology/middleware-manager/internal/traefikmanager/types"
	"gopkg.in/yaml.v3"
)

type TraefikHandler struct {
	settings *tmconfig.SettingsStore
	files    *tmconfig.FileStore
	client   *http.Client
}

func NewTraefikHandler(settings *tmconfig.SettingsStore, files *tmconfig.FileStore, client *http.Client) *TraefikHandler {
	return &TraefikHandler{settings: settings, files: files, client: client}
}

func (h *TraefikHandler) Overview(c *gin.Context)    { h.proxyJSON(c, "/api/overview") }
func (h *TraefikHandler) Entrypoints(c *gin.Context) { h.proxyJSON(c, "/api/entrypoints") }
func (h *TraefikHandler) Version(c *gin.Context)     { h.proxyJSON(c, "/api/version") }

func (h *TraefikHandler) Routers(c *gin.Context) {
	h.multiProxy(c, []string{"/api/http/routers", "/api/tcp/routers", "/api/udp/routers"}, []string{"http", "tcp", "udp"})
}

func (h *TraefikHandler) Services(c *gin.Context) {
	h.multiProxy(c, []string{"/api/http/services", "/api/tcp/services", "/api/udp/services"}, []string{"http", "tcp", "udp"})
}

func (h *TraefikHandler) Middlewares(c *gin.Context) {
	h.multiProxy(c, []string{"/api/http/middlewares", "/api/tcp/middlewares"}, []string{"http", "tcp"})
}

func (h *TraefikHandler) Ping(c *gin.Context) {
	settings, _, err := h.settings.Load()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load settings", err)
		return
	}
	start := time.Now()
	resp, err := h.client.Get(strings.TrimRight(settings.TraefikAPIURL, "/") + "/ping")
	if err != nil {
		respondJSON(c, gin.H{"ok": false, "latency_ms": nil})
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		respondJSON(c, gin.H{"ok": false, "latency_ms": nil})
		return
	}
	respondJSON(c, gin.H{"ok": true, "latency_ms": time.Since(start).Milliseconds()})
}

func (h *TraefikHandler) RouterDetail(c *gin.Context) {
	protocol := strings.ToLower(c.Param("protocol"))
	if protocol != "http" && protocol != "tcp" && protocol != "udp" {
		respondError(c, http.StatusBadRequest, "invalid protocol", nil)
		return
	}
	h.proxyJSON(c, "/api/"+protocol+"/routers/"+c.Param("name"))
}

func (h *TraefikHandler) Certs(c *gin.Context) {
	settings, _, err := h.settings.Load()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load settings", err)
		return
	}
	certs, err := tmconfig.ParseACMECertificates(h.settings.EffectiveAcmePath(settings))
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to read certificates", err)
		return
	}
	for _, path := range h.files.ConfigPaths() {
		config, err := h.files.LoadConfig(path)
		if err != nil {
			continue
		}
		if certificates, ok := tmconfig.MapFromAny(config["tls"])["certificates"].([]any); ok {
			for _, item := range certificates {
				certFile := tmconfig.StringFromAny(tmconfig.MapFromAny(item)["certFile"])
				if certFile == "" {
					continue
				}
				info, err := tmconfig.ParseFileCertificate(certFile)
				if err == nil {
					certs = append(certs, info)
				}
			}
		}
	}
	sort.Slice(certs, func(i, j int) bool { return certs[i].Main < certs[j].Main })
	respondJSON(c, gin.H{"certs": certs})
}

func (h *TraefikHandler) Logs(c *gin.Context) {
	settings, _, err := h.settings.Load()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load settings", err)
		return
	}
	file, err := os.Open(h.settings.EffectiveAccessLogPath(settings))
	if err != nil {
		respondJSON(c, gin.H{"error": err.Error(), "lines": []string{}})
		return
	}
	defer func() { _ = file.Close() }()
	lines := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if len(lines) > 200 {
		lines = lines[len(lines)-200:]
	}
	respondJSON(c, gin.H{"lines": lines})
}

func (h *TraefikHandler) Plugins(c *gin.Context) {
	settings, _, err := h.settings.Load()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load settings", err)
		return
	}
	data, err := os.ReadFile(h.settings.EffectiveStaticConfigPath(settings))
	if err != nil {
		respondJSON(c, gin.H{"plugins": []tmtypes.PluginInfo{}, "error": err.Error()})
		return
	}
	raw := map[string]any{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to parse static config", err)
		return
	}
	plugins := make([]tmtypes.PluginInfo, 0)
	for name, value := range tmconfig.MapFromAny(tmconfig.MapFromAny(raw["experimental"])["plugins"]) {
		plugin := tmconfig.MapFromAny(value)
		plugins = append(plugins, tmtypes.PluginInfo{Name: name, ModuleName: tmconfig.StringFromAny(plugin["moduleName"]), Version: tmconfig.StringFromAny(plugin["version"]), Settings: tmconfig.MapFromAny(plugin["settings"])})
	}
	sort.Slice(plugins, func(i, j int) bool { return plugins[i].Name < plugins[j].Name })
	respondJSON(c, gin.H{"plugins": plugins})
}

func (h *TraefikHandler) proxyJSON(c *gin.Context, path string) {
	settings, _, err := h.settings.Load()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load settings", err)
		return
	}
	resp, err := h.client.Get(strings.TrimRight(settings.TraefikAPIURL, "/") + path)
	if err != nil {
		respondJSON(c, gin.H{})
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		respondJSON(c, gin.H{})
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		respondJSON(c, gin.H{})
		return
	}
	c.Data(http.StatusOK, "application/json", body)
}

func (h *TraefikHandler) multiProxy(c *gin.Context, paths, keys []string) {
	settings, _, err := h.settings.Load()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to load settings", err)
		return
	}
	result := gin.H{}
	for idx, path := range paths {
		resp, err := h.client.Get(strings.TrimRight(settings.TraefikAPIURL, "/") + path)
		if err != nil || resp.StatusCode != http.StatusOK {
			result[keys[idx]] = []any{}
			if resp != nil {
				_ = resp.Body.Close()
			}
			continue
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			result[keys[idx]] = []any{}
			continue
		}
		var payload any
		if err := json.Unmarshal(body, &payload); err != nil {
			result[keys[idx]] = []any{}
			continue
		}
		result[keys[idx]] = payload
	}
	respondJSON(c, result)
}
