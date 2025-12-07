package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

// AuthMiddleware validates API keys
func AuthMiddleware(store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing API key"})
			c.Abort()
			return
		}

		// Try to find the key by comparing hashes
		keys, err := store.APIKeys().List(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify API key"})
			c.Abort()
			return
		}

		var validKey *domain.APIKey
		for _, key := range keys {
			if err := bcrypt.CompareHashAndPassword([]byte(key.KeyHash), []byte(apiKey)); err == nil {
				validKey = key
				break
			}
		}

		if validKey == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			c.Abort()
			return
		}

		// Check if key is expired
		if validKey.IsExpired() {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key expired"})
			c.Abort()
			return
		}

		// Store the key info in context
		c.Set("api_key", validKey)
		c.Next()
	}
}

// RequireAdmin ensures the API key has admin scope
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		key, exists := c.Get("api_key")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		apiKey := key.(*domain.APIKey)
		if !apiKey.CanManageKeys() {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireWrite ensures the API key can modify resources
func RequireWrite() gin.HandlerFunc {
	return func(c *gin.Context) {
		key, exists := c.Get("api_key")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		apiKey := key.(*domain.APIKey)
		if !apiKey.CanWrite() {
			c.JSON(http.StatusForbidden, gin.H{"error": "write access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORS middleware
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// Helper to get current API key from context
func GetAPIKey(c *gin.Context) *domain.APIKey {
	key, _ := c.Get("api_key")
	return key.(*domain.APIKey)
}

// ParseTags helper for parsing tag query parameters (format: "key1=value1,key2=value2")
func ParseTags(tagsParam string) map[string]string {
	tags := make(map[string]string)
	if tagsParam == "" {
		return tags
	}

	pairs := strings.Split(tagsParam, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			tags[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	return tags
}
