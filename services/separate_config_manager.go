package services

import (
    "fmt"
    "io/fs"
    "log"
    "os"
    "path/filepath"
    "strings"
)

// SeparateConfigManager manages separate Traefik configuration files
type SeparateConfigManager struct {
    confDir string
}

// NewSeparateConfigManager creates a new separate config manager
func NewSeparateConfigManager(confDir string) *SeparateConfigManager {
    return &SeparateConfigManager{
        confDir: confDir,
    }
}

// EnsureDirectories creates necessary directories for separate config files
func (scm *SeparateConfigManager) EnsureDirectories() error {
    dirs := []string{
        filepath.Join(scm.confDir, "middlewares"),
        filepath.Join(scm.confDir, "routers"),
        filepath.Join(scm.confDir, "services"),
    }

    for _, dir := range dirs {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return fmt.Errorf("failed to create directory %s: %w", dir, err)
        }
    }

    return nil
}

// ListConfigFiles returns a list of all configuration files
func (scm *SeparateConfigManager) ListConfigFiles() ([]string, error) {
    var files []string

    err := filepath.WalkDir(scm.confDir, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        if !d.IsDir() && strings.HasSuffix(d.Name(), ".yml") {
            relPath, err := filepath.Rel(scm.confDir, path)
            if err != nil {
                return err
            }
            files = append(files, relPath)
        }

        return nil
    })

    if err != nil {
        return nil, fmt.Errorf("failed to walk config directory: %w", err)
    }

    return files, nil
}

// GetConfigFilesByType returns configuration files grouped by type
func (scm *SeparateConfigManager) GetConfigFilesByType() (map[string][]string, error) {
    files, err := scm.ListConfigFiles()
    if err != nil {
        return nil, err
    }

    configTypes := map[string][]string{
        "middlewares": {},
        "routers":     {},
        "services":    {},
    }

    for _, file := range files {
        switch {
        case strings.HasPrefix(file, "middlewares/"):
            configTypes["middlewares"] = append(configTypes["middlewares"], file)
        case strings.HasSuffix(file, "-routers.yml"):
            configTypes["routers"] = append(configTypes["routers"], file)
        case strings.HasSuffix(file, "-services.yml"):
            configTypes["services"] = append(configTypes["services"], file)
        default:
            // Handle other config files
            if strings.Contains(file, "middleware") {
                configTypes["middlewares"] = append(configTypes["middlewares"], file)
            } else if strings.Contains(file, "router") {
                configTypes["routers"] = append(configTypes["routers"], file)
            } else if strings.Contains(file, "service") {
                configTypes["services"] = append(configTypes["services"], file)
            }
        }
    }

    return configTypes, nil
}

// CleanupOldFiles removes old configuration files before regeneration
func (scm *SeparateConfigManager) CleanupOldFiles() error {
    // Remove the old single resource-overrides.yml file if it exists
    oldConfigFile := filepath.Join(scm.confDir, "resource-overrides.yml")
    if err := os.Remove(oldConfigFile); err != nil && !os.IsNotExist(err) {
        log.Printf("Warning: failed to remove old config file %s: %v", oldConfigFile, err)
    }

    return nil
}

// GetConfigSummary returns a summary of all configuration files
func (scm *SeparateConfigManager) GetConfigSummary() (map[string]int, error) {
    configTypes, err := scm.GetConfigFilesByType()
    if err != nil {
        return nil, err
    }

    summary := make(map[string]int)
    for configType, files := range configTypes {
        summary[configType] = len(files)
    }

    return summary, nil
}

// ValidateConfigStructure validates that the configuration structure is correct
func (scm *SeparateConfigManager) ValidateConfigStructure() error {
    requiredDirs := []string{
        "middlewares",
    }

    for _, dir := range requiredDirs {
        dirPath := filepath.Join(scm.confDir, dir)
        if _, err := os.Stat(dirPath); os.IsNotExist(err) {
            return fmt.Errorf("required directory %s does not exist", dir)
        }
    }

    return nil
}

// GetFileSize returns the size of a specific config file
func (scm *SeparateConfigManager) GetFileSize(filename string) (int64, error) {
    filePath := filepath.Join(scm.confDir, filename)
    info, err := os.Stat(filePath)
    if err != nil {
        return 0, err
    }
    return info.Size(), nil
}

// RemoveConfigFile removes a specific configuration file
func (scm *SeparateConfigManager) RemoveConfigFile(filename string) error {
    filePath := filepath.Join(scm.confDir, filename)
    return os.Remove(filePath)
}

// FileExists checks if a configuration file exists
func (scm *SeparateConfigManager) FileExists(filename string) bool {
    filePath := filepath.Join(scm.confDir, filename)
    _, err := os.Stat(filePath)
    return err == nil
}