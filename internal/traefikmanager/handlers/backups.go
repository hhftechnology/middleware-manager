package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	tmconfig "github.com/hhftechnology/middleware-manager/internal/traefikmanager/config"
)

type BackupHandler struct {
	files *tmconfig.FileStore
}

func NewBackupHandler(files *tmconfig.FileStore) *BackupHandler {
	return &BackupHandler{files: files}
}

func (h *BackupHandler) List(c *gin.Context) {
	backups, err := h.files.ListBackups()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to list backups", err)
		return
	}
	respondJSON(c, backups)
}

func (h *BackupHandler) Create(c *gin.Context) {
	path := h.files.ResolveConfigPath("")
	name, err := h.files.CreateBackup(path)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to create backup", err)
		return
	}
	respondJSON(c, gin.H{"success": true, "name": filepath.Base(name)})
}

func (h *BackupHandler) Restore(c *gin.Context) {
	if err := h.files.RestoreBackup(c.Param("filename")); err != nil {
		respondError(c, http.StatusBadRequest, "Failed to restore backup", err)
		return
	}
	respondJSON(c, gin.H{"success": true})
}

func (h *BackupHandler) Delete(c *gin.Context) {
	if err := h.files.DeleteBackup(c.Param("filename")); err != nil {
		respondError(c, http.StatusBadRequest, "Failed to delete backup", err)
		return
	}
	respondJSON(c, gin.H{"success": true})
}
