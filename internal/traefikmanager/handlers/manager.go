package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	tmconfig "github.com/hhftechnology/middleware-manager/internal/traefikmanager/config"
)

type ManagerHandler struct {
	client *http.Client
	repo   string
}

func NewManagerHandler(client *http.Client, repo string) *ManagerHandler {
	return &ManagerHandler{client: client, repo: repo}
}

func (h *ManagerHandler) Version(c *gin.Context) {
	resp, err := h.client.Get("https://api.github.com/repos/" + h.repo + "/releases/latest")
	if err != nil {
		respondJSON(c, gin.H{"version": "", "repo": h.repo})
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		respondJSON(c, gin.H{"version": "", "repo": h.repo})
		return
	}
	payload := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		respondJSON(c, gin.H{"version": "", "repo": h.repo})
		return
	}
	respondJSON(c, gin.H{"version": strings.TrimPrefix(tmconfig.StringFromAny(payload["tag_name"]), "v"), "repo": h.repo})
}
