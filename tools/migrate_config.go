package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"

    "github.com/hhftechnology/middleware-manager/services"
    "github.com/hhftechnology/middleware-manager/database"
)

// ConfigMigration handles migration from single file to separate files
type ConfigMigration struct {
    confDir       string
    backupDir     string
    dryRun        bool
}

func main() {
    var (
        confDir   = flag.String("conf-dir", "/conf", "Traefik configuration directory")
        backupDir = flag.String("backup-dir", "/backup", "Directory to store backup of old config")
        dryRun    = flag.Bool("dry-run", false, "Show what would be done without making changes")
        dbPath    = flag.String("db-path", "/data/middleware.db", "Path to the database file")
    )
    flag.Parse()

    migration := &ConfigMigration{
        confDir:   *confDir,
        backupDir: *backupDir,
        dryRun:    *dryRun,
    }

    if err := migration.Run(*dbPath); err != nil {
        log.Fatalf("Migration failed: %v", err)
    }
}

// Run executes the migration process
func (cm *ConfigMigration) Run(dbPath string) error {
    log.Println("Starting configuration migration to separate files...")

    // Check if old config exists
    oldConfigPath := filepath.Join(cm.confDir, "resource-overrides.yml")
    if _, err := os.Stat(oldConfigPath); os.IsNotExist(err) {
        log.Println("No old resource-overrides.yml found, nothing to migrate")
        return cm.generateSeparateFiles(dbPath)
    }

    // Create backup directory
    if !cm.dryRun {
        if err := os.MkdirAll(cm.backupDir, 0755); err != nil {
            return fmt.Errorf("failed to create backup directory: %w", err)
        }
    }

    // Backup old configuration
    if err := cm.backupOldConfig(); err != nil {
        return fmt.Errorf("failed to backup old config: %w", err)
    }

    // Generate separate files
    if err := cm.generateSeparateFiles(dbPath); err != nil {
        return fmt.Errorf("failed to generate separate files: %w", err)
    }

    // Clean up old file
    if err := cm.cleanupOldConfig(); err != nil {
        return fmt.Errorf("failed to cleanup old config: %w", err)
    }

    log.Println("Migration completed successfully!")
    return nil
}

// backupOldConfig creates a backup of the existing resource-overrides.yml
func (cm *ConfigMigration) backupOldConfig() error {
    oldConfigPath := filepath.Join(cm.confDir, "resource-overrides.yml")
    backupPath := filepath.Join(cm.backupDir, "resource-overrides.yml.backup")

    if cm.dryRun {
        log.Printf("DRY RUN: Would backup %s to %s", oldConfigPath, backupPath)
        return nil
    }

    log.Printf("Backing up %s to %s", oldConfigPath, backupPath)

    data, err := os.ReadFile(oldConfigPath)
    if err != nil {
        return fmt.Errorf("failed to read old config: %w", err)
    }

    if err := os.WriteFile(backupPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write backup: %w", err)
    }

    log.Println("Backup created successfully")
    return nil
}

// generateSeparateFiles creates the new separate configuration files
func (cm *ConfigMigration) generateSeparateFiles(dbPath string) error {
    log.Println("Generating separate configuration files...")

    if cm.dryRun {
        log.Println("DRY RUN: Would generate separate configuration files")
        return cm.simulateGeneration()
    }

    // Initialize database
    db, err := database.NewDB(dbPath)
    if err != nil {
        return fmt.Errorf("failed to initialize database: %w", err)
    }
    defer db.Close()

    // Create config manager (minimal implementation for migration)
    configManager := &MockConfigManager{}
    
    // Create config generator with separate file support
    generator := services.NewConfigGenerator(db, cm.confDir, configManager)
    
    // Create separate config manager
    separateManager := services.NewSeparateConfigManager(cm.confDir)
    
    // Ensure directories exist
    if err := separateManager.EnsureDirectories(); err != nil {
        return fmt.Errorf("failed to create directories: %w", err)
    }

    // Generate configuration
    if err := generator.Start(0); err != nil {
        return fmt.Errorf("failed to generate config: %w", err)
    }

    log.Println("Separate configuration files generated successfully")
    return nil
}

// simulateGeneration shows what would be generated without creating files
func (cm *ConfigMigration) simulateGeneration() error {
    log.Println("DRY RUN: Separate files that would be generated:")
    log.Println("  - middlewares/ (directory for individual middleware files)")
    log.Println("  - http-routers.yml")
    log.Println("  - tcp-routers.yml (if TCP routers exist)")
    log.Println("  - http-services.yml")
    log.Println("  - tcp-services.yml (if TCP services exist)")
    log.Println("  - udp-services.yml (if UDP services exist)")
    return nil
}

// cleanupOldConfig removes the old resource-overrides.yml file
func (cm *ConfigMigration) cleanupOldConfig() error {
    oldConfigPath := filepath.Join(cm.confDir, "resource-overrides.yml")

    if cm.dryRun {
        log.Printf("DRY RUN: Would remove old config file %s", oldConfigPath)
        return nil
    }

    log.Printf("Removing old config file %s", oldConfigPath)

    if err := os.Remove(oldConfigPath); err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("failed to remove old config: %w", err)
    }

    log.Println("Old config file removed successfully")
    return nil
}

// MockConfigManager provides a minimal config manager for migration
type MockConfigManager struct{}

func (m *MockConfigManager) GetActiveDataSourceConfig() (*MockDataSourceConfig, error) {
    return &MockDataSourceConfig{
        Type: "pangolin",
        URL:  "http://localhost:3001/api/v1",
    }, nil
}

// MockDataSourceConfig represents a data source configuration
type MockDataSourceConfig struct {
    Type string
    URL  string
}

// Additional utility functions

// validateMigration checks that the migration was successful
func (cm *ConfigMigration) validateMigration() error {
    separateManager := services.NewSeparateConfigManager(cm.confDir)
    
    if err := separateManager.ValidateConfigStructure(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    summary, err := separateManager.GetConfigSummary()
    if err != nil {
        return fmt.Errorf("failed to get config summary: %w", err)
    }

    log.Println("Migration validation successful:")
    for configType, count := range summary {
        log.Printf("  - %s: %d files", configType, count)
    }

    return nil
}

// rollback restores the old configuration in case of issues
func (cm *ConfigMigration) rollback() error {
    log.Println("Rolling back migration...")

    backupPath := filepath.Join(cm.backupDir, "resource-overrides.yml.backup")
    oldConfigPath := filepath.Join(cm.confDir, "resource-overrides.yml")

    if cm.dryRun {
        log.Printf("DRY RUN: Would restore %s from %s", oldConfigPath, backupPath)
        return nil
    }

    data, err := os.ReadFile(backupPath)
    if err != nil {
        return fmt.Errorf("failed to read backup: %w", err)
    }

    if err := os.WriteFile(oldConfigPath, data, 0644); err != nil {
        return fmt.Errorf("failed to restore config: %w", err)
    }

    // Remove separate files
    separateManager := services.NewSeparateConfigManager(cm.confDir)
    files, err := separateManager.ListConfigFiles()
    if err != nil {
        log.Printf("Warning: failed to list config files for cleanup: %v", err)
    } else {
        for _, file := range files {
            if err := separateManager.RemoveConfigFile(file); err != nil {
                log.Printf("Warning: failed to remove %s: %v", file, err)
            }
        }
    }

    log.Println("Rollback completed")
    return nil
}