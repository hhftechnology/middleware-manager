package config

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	tmtypes "github.com/hhftechnology/middleware-manager/internal/traefikmanager/types"
	"gopkg.in/yaml.v3"
)

const SelfRouteFilename = "traefik-manager-self.yml"

var backupFilenamePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+\.(yml|yaml)\.\d{8}_\d{6}\.bak$`)

type FileStore struct {
	cfg         tmtypes.RuntimeConfig
	configPaths []string
	activeDir   string
}

func NewFileStore(cfg tmtypes.RuntimeConfig) (*FileStore, error) {
	store := &FileStore{cfg: cfg}
	if err := store.Refresh(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *FileStore) Refresh() error {
	switch {
	case strings.TrimSpace(s.cfg.ConfigDir) != "":
		s.activeDir = strings.TrimSpace(s.cfg.ConfigDir)
		paths := make([]string, 0)
		err := filepath.Walk(s.activeDir, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil || info == nil || info.IsDir() {
				return walkErr
			}
			if strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml") {
				paths = append(paths, path)
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("scan config dir: %w", err)
		}
		sort.Strings(paths)
		if len(paths) == 0 {
			paths = []string{filepath.Join(s.activeDir, "dynamic.yml")}
		}
		s.configPaths = paths
	case len(s.cfg.ConfigPaths) > 0:
		s.activeDir = ""
		s.configPaths = append([]string(nil), s.cfg.ConfigPaths...)
	default:
		s.activeDir = ""
		path := strings.TrimSpace(s.cfg.ConfigPath)
		if path == "" {
			path = "/app/config/dynamic.yml"
		}
		s.configPaths = []string{path}
	}
	return nil
}

func (s *FileStore) ConfigPaths() []string {
	return append([]string(nil), s.configPaths...)
}

func (s *FileStore) ActiveConfigDir() string {
	return s.activeDir
}

func (s *FileStore) MultiConfig() bool {
	return len(s.configPaths) > 1
}

func (s *FileStore) ResolveConfigPath(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		if len(s.configPaths) == 0 {
			return ""
		}
		return s.configPaths[0]
	}
	for _, path := range s.configPaths {
		if trimmed == path || trimmed == filepath.Base(path) {
			return path
		}
	}
	if s.activeDir != "" && !strings.ContainsAny(trimmed, `/\`) {
		name := trimmed
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			name += ".yml"
		}
		candidate := filepath.Join(s.activeDir, name)
		if s.IsSafePath(candidate) {
			return candidate
		}
	}
	return ""
}

func (s *FileStore) RegisterConfigPath(path string) {
	if strings.TrimSpace(path) == "" {
		return
	}
	for _, existing := range s.configPaths {
		if existing == path {
			return
		}
	}
	s.configPaths = append(s.configPaths, path)
	sort.Strings(s.configPaths)
}

func (s *FileStore) IsSafePath(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	resolved := filepath.Clean(path)
	if s.activeDir != "" {
		root := filepath.Clean(s.activeDir)
		return resolved == root || strings.HasPrefix(resolved, root+string(os.PathSeparator))
	}
	for _, allowed := range s.configPaths {
		dir := filepath.Dir(filepath.Clean(allowed))
		if resolved == dir || strings.HasPrefix(resolved, dir+string(os.PathSeparator)) {
			return true
		}
	}
	backupDir := filepath.Clean(s.cfg.BackupDir)
	return resolved == backupDir || strings.HasPrefix(resolved, backupDir+string(os.PathSeparator))
}

func (s *FileStore) LoadConfig(path string) (map[string]any, error) {
	target := path
	if strings.TrimSpace(target) == "" {
		target = s.ResolveConfigPath("")
	}
	if target == "" {
		return emptyDynamicConfig(), nil
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return emptyDynamicConfig(), nil
	}
	data, err := os.ReadFile(target)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	out := map[string]any{}
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parse yaml config: %w", err)
	}
	if len(out) == 0 {
		return emptyDynamicConfig(), nil
	}
	normalized, ok := NormalizeYAML(out).(map[string]any)
	if !ok {
		return emptyDynamicConfig(), nil
	}
	return normalized, nil
}

func (s *FileStore) SaveConfig(path string, data map[string]any) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("config path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	encoded, err := yaml.Marshal(StripEmptySections(data))
	if err != nil {
		return fmt.Errorf("encode yaml config: %w", err)
	}
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, encoded, 0o644); err != nil {
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename config: %w", err)
	}
	s.RegisterConfigPath(path)
	return nil
}

func (s *FileStore) CreateBackup(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", nil
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", nil
	}
	if err := os.MkdirAll(s.cfg.BackupDir, 0o755); err != nil {
		return "", fmt.Errorf("create backup dir: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read config for backup: %w", err)
	}
	name := fmt.Sprintf("%s.%s.bak", filepath.Base(path), time.Now().Format("20060102_150405"))
	destination := filepath.Join(s.cfg.BackupDir, name)
	if err := os.WriteFile(destination, data, 0o644); err != nil {
		return "", fmt.Errorf("write backup: %w", err)
	}
	return destination, nil
}

func (s *FileStore) ListBackups() ([]tmtypes.BackupEntry, error) {
	if err := os.MkdirAll(s.cfg.BackupDir, 0o755); err != nil {
		return nil, fmt.Errorf("create backup dir: %w", err)
	}
	entries, err := os.ReadDir(s.cfg.BackupDir)
	if err != nil {
		return nil, fmt.Errorf("read backup dir: %w", err)
	}
	backups := make([]tmtypes.BackupEntry, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".bak") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		backups = append(backups, tmtypes.BackupEntry{
			Name:     entry.Name(),
			Size:     info.Size(),
			Modified: info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}
	sort.Slice(backups, func(i, j int) bool { return backups[i].Name > backups[j].Name })
	return backups, nil
}

func (s *FileStore) ValidateBackupPath(filename string) (string, error) {
	if !backupFilenamePattern.MatchString(filename) {
		return "", errors.New("invalid backup filename")
	}
	fullPath := filepath.Clean(filepath.Join(s.cfg.BackupDir, filename))
	root := filepath.Clean(s.cfg.BackupDir)
	if fullPath != root && !strings.HasPrefix(fullPath, root+string(os.PathSeparator)) {
		return "", errors.New("invalid backup path")
	}
	return fullPath, nil
}

func (s *FileStore) ResolveTargetForBackup(filename string) string {
	for _, path := range s.configPaths {
		if strings.HasPrefix(filename, filepath.Base(path)+".") {
			return path
		}
	}
	if len(s.configPaths) > 0 {
		return s.configPaths[0]
	}
	return ""
}

func (s *FileStore) RestoreBackup(filename string) error {
	source, err := s.ValidateBackupPath(filename)
	if err != nil {
		return err
	}
	target := s.ResolveTargetForBackup(filename)
	if target == "" {
		return errors.New("could not resolve backup target")
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read backup file: %w", err)
	}
	if _, err := s.CreateBackup(target); err != nil {
		return err
	}
	return os.WriteFile(target, data, 0o644)
}

func (s *FileStore) DeleteBackup(filename string) error {
	path, err := s.ValidateBackupPath(filename)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

func StripEmptySections(config map[string]any) map[string]any {
	out := MapFromAny(config)
	for _, proto := range []string{"http", "tcp", "udp"} {
		section := MapFromAny(out[proto])
		if len(section) == 0 {
			delete(out, proto)
			continue
		}
		for _, key := range []string{"routers", "services", "middlewares"} {
			items := MapFromAny(section[key])
			if len(items) == 0 {
				delete(section, key)
			}
		}
		if len(section) == 0 {
			delete(out, proto)
			continue
		}
		out[proto] = section
	}
	return out
}

func BuildApps(config map[string]any, configFile string, multiConfig bool) []tmtypes.App {
	apps := make([]tmtypes.App, 0)
	for _, proto := range []string{"http", "tcp", "udp"} {
		section := MapFromAny(config[proto])
		routers := MapFromAny(section["routers"])
		services := MapFromAny(section["services"])
		transports := MapFromAny(section["serversTransports"])
		for name, rawRouter := range routers {
			router := MapFromAny(rawRouter)
			serviceName := StringFromAny(router["service"])
			id := name
			if multiConfig && configFile != "" {
				id = configFile + "::" + name
			}
			app := tmtypes.App{
				ID:          id,
				Name:        name,
				Rule:        StringFromAny(router["rule"]),
				ServiceName: serviceName,
				Middlewares: StringSliceFromAny(router["middlewares"]),
				EntryPoints: StringSliceFromAny(router["entryPoints"]),
				Protocol:    proto,
				TLS:         hasTLS(router["tls"]),
				Enabled:     true,
				ConfigFile:  configFile,
				Provider:    "file",
			}
			switch proto {
			case "http":
				service := MapFromAny(services[serviceName])
				loadBalancer := MapFromAny(service["loadBalancer"])
				if servers, ok := loadBalancer["servers"].([]any); ok && len(servers) > 0 {
					app.Target = StringFromAny(MapFromAny(servers[0])["url"])
				}
				app.PassHostHeader = true
				if passHost, ok := loadBalancer["passHostHeader"].(bool); ok {
					app.PassHostHeader = passHost
				}
				if tls := MapFromAny(router["tls"]); len(tls) > 0 {
					app.CertResolver = StringFromAny(tls["certResolver"])
				}
				if transportName := StringFromAny(loadBalancer["serversTransport"]); transportName != "" {
					app.InsecureSkipVerify = BoolFromAny(MapFromAny(transports[transportName])["insecureSkipVerify"])
				}
			case "tcp", "udp":
				service := MapFromAny(services[serviceName])
				loadBalancer := MapFromAny(service["loadBalancer"])
				if servers, ok := loadBalancer["servers"].([]any); ok && len(servers) > 0 {
					app.Target = StringFromAny(MapFromAny(servers[0])["address"])
				}
				if tls := MapFromAny(router["tls"]); len(tls) > 0 {
					app.CertResolver = StringFromAny(tls["certResolver"])
				}
			}
			if app.Target == "" {
				app.Target = "N/A"
			}
			apps = append(apps, app)
		}
	}
	return apps
}

func BuildMiddlewares(config map[string]any, configFile string) ([]tmtypes.MiddlewareEntry, error) {
	entries := make([]tmtypes.MiddlewareEntry, 0)
	for name, raw := range MapFromAny(MapFromAny(config["http"])["middlewares"]) {
		buf := bytes.Buffer{}
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)
		if err := encoder.Encode(raw); err != nil {
			_ = encoder.Close()
			return nil, fmt.Errorf("encode middleware yaml: %w", err)
		}
		_ = encoder.Close()
		entries = append(entries, tmtypes.MiddlewareEntry{
			Name:       name,
			YAML:       strings.TrimSpace(buf.String()),
			Type:       "http",
			ConfigFile: configFile,
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries, nil
}

func ApplyDisabledRoutes(settings tmtypes.Settings, apps []tmtypes.App) []tmtypes.App {
	for routeID, disabled := range settings.DisabledRoutes {
		app := tmtypes.App{
			ID:          routeID,
			Name:        RouteNameFromID(routeID),
			Rule:        StringFromAny(disabled.Router["rule"]),
			ServiceName: StringFromAny(disabled.Router["service"]),
			Middlewares: StringSliceFromAny(disabled.Router["middlewares"]),
			EntryPoints: StringSliceFromAny(disabled.Router["entryPoints"]),
			Protocol:    disabled.Protocol,
			TLS:         hasTLS(disabled.Router["tls"]),
			Enabled:     false,
			ConfigFile:  disabled.ConfigFile,
			Provider:    "file",
		}
		loadBalancer := MapFromAny(disabled.Service["loadBalancer"])
		if disabled.Protocol == "http" {
			if servers, ok := loadBalancer["servers"].([]any); ok && len(servers) > 0 {
				app.Target = StringFromAny(MapFromAny(servers[0])["url"])
			}
			app.PassHostHeader = true
			if passHost, ok := loadBalancer["passHostHeader"].(bool); ok {
				app.PassHostHeader = passHost
			}
		} else if servers, ok := loadBalancer["servers"].([]any); ok && len(servers) > 0 {
			app.Target = StringFromAny(MapFromAny(servers[0])["address"])
		}
		if app.Target == "" {
			app.Target = "N/A"
		}
		apps = append(apps, app)
	}
	return apps
}

func ParseACMECertificates(path string) ([]tmtypes.CertificateInfo, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read acme json: %w", err)
	}
	raw := map[string]any{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse acme json: %w", err)
	}
	certs := make([]tmtypes.CertificateInfo, 0)
	for resolverName, resolverValue := range raw {
		resolver := MapFromAny(resolverValue)
		items, ok := resolver["Certificates"].([]any)
		if !ok {
			items, _ = resolver["certificates"].([]any)
		}
		for _, item := range items {
			entry := MapFromAny(item)
			domain := MapFromAny(entry["domain"])
			certs = append(certs, tmtypes.CertificateInfo{
				Resolver: resolverName,
				Main:     StringFromAny(domain["main"]),
				Sans:     StringSliceFromAny(domain["sans"]),
				NotAfter: parseEncodedCertExpiry(StringFromAny(entry["certificate"])),
			})
		}
	}
	return certs, nil
}

func ParseFileCertificate(path string) (tmtypes.CertificateInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return tmtypes.CertificateInfo{}, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return tmtypes.CertificateInfo{}, errors.New("invalid pem certificate")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return tmtypes.CertificateInfo{}, err
	}
	return tmtypes.CertificateInfo{
		Resolver: "file",
		Main:     filepath.Base(path),
		Sans:     cert.DNSNames,
		NotAfter: cert.NotAfter.Format("Jan 02 15:04:05 2006 MST"),
		CertFile: path,
	}, nil
}

func emptyDynamicConfig() map[string]any {
	return map[string]any{
		"http": map[string]any{
			"routers":     map[string]any{},
			"services":    map[string]any{},
			"middlewares": map[string]any{},
		},
	}
}

func hasTLS(input any) bool {
	switch value := input.(type) {
	case bool:
		return value
	case map[string]any:
		return len(value) > 0
	case map[any]any:
		return len(value) > 0
	default:
		return false
	}
}

func parseEncodedCertExpiry(raw string) string {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	block, _ := pem.Decode(decoded)
	if block == nil {
		return ""
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return ""
	}
	return cert.NotAfter.Format("Jan 02 15:04:05 2006 MST")
}
