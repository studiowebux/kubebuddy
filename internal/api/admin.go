package api

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type CreateAPIKeyRequest struct {
	Name        string              `json:"name" binding:"required"`
	Scope       domain.APIKeyScope  `json:"scope" binding:"required"`
	Description string              `json:"description"`
	ExpiresAt   *time.Time          `json:"expires_at"`
}

type CreateAPIKeyResponse struct {
	APIKey *domain.APIKey `json:"api_key"`
	Key    string         `json:"key"` // Plain text key (only returned once)
}

func (s *Server) listAPIKeys(c *gin.Context) {
	keys, err := s.store.APIKeys().List(c.Request.Context())
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to list API keys", err)
		return
	}

	c.JSON(http.StatusOK, keys)
}

func (s *Server) getAPIKey(c *gin.Context) {
	id := c.Param("id")

	key, err := s.store.APIKeys().Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, http.StatusNotFound, "API key not found", err)
		return
	}

	c.JSON(http.StatusOK, key)
}

func (s *Server) createAPIKey(c *gin.Context) {
	var req CreateAPIKeyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Validate scope
	if req.Scope != domain.APIKeyScopeAdmin &&
		req.Scope != domain.APIKeyScopeReadWrite &&
		req.Scope != domain.APIKeyScopeReadOnly {
		handleError(c, http.StatusBadRequest, "invalid scope", nil)
		return
	}

	// Generate random API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		handleError(c, http.StatusInternalServerError, "failed to generate API key", err)
		return
	}
	plainKey := base64.URLEncoding.EncodeToString(keyBytes)

	// Hash the key
	keyHash, err := bcrypt.GenerateFromPassword([]byte(plainKey), bcrypt.DefaultCost)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to hash API key", err)
		return
	}

	// Get current API key for created_by
	currentKey := GetAPIKey(c)

	apiKey := &domain.APIKey{
		ID:          uuid.New().String(),
		Name:        req.Name,
		KeyHash:     string(keyHash),
		Scope:       req.Scope,
		Description: req.Description,
		CreatedBy:   currentKey.ID,
		ExpiresAt:   req.ExpiresAt,
	}

	if err := s.store.APIKeys().Create(c.Request.Context(), apiKey); err != nil {
		handleError(c, http.StatusInternalServerError, "failed to create API key", err)
		return
	}

	response := CreateAPIKeyResponse{
		APIKey: apiKey,
		Key:    plainKey,
	}

	c.JSON(http.StatusCreated, response)
}

func (s *Server) deleteAPIKey(c *gin.Context) {
	id := c.Param("id")

	// Prevent deleting own key
	currentKey := GetAPIKey(c)
	if currentKey.ID == id {
		handleError(c, http.StatusBadRequest, "cannot delete your own API key", nil)
		return
	}

	if err := s.store.APIKeys().Delete(c.Request.Context(), id); err != nil {
		handleError(c, http.StatusNotFound, "API key not found", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key deleted successfully"})
}
