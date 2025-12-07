package domain

import "time"

// APIKeyScope defines the permissions level of an API key
type APIKeyScope string

const (
	APIKeyScopeAdmin    APIKeyScope = "admin"  // Can manage other API keys
	APIKeyScopeReadWrite APIKeyScope = "readwrite" // Can read and modify resources
	APIKeyScopeReadOnly APIKeyScope = "readonly"  // Can only read resources
)

// APIKey represents an API authentication key
type APIKey struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	KeyHash     string      `json:"-"` // bcrypt hash of the key (not exposed in JSON)
	Scope       APIKeyScope `json:"scope"`
	Description string      `json:"description,omitempty"`
	CreatedBy   string      `json:"created_by,omitempty"` // ID of admin key that created this
	CreatedAt   time.Time   `json:"created_at"`
	ExpiresAt   *time.Time  `json:"expires_at,omitempty"`
}

// IsExpired checks if the API key has expired
func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

// CanManageKeys checks if this API key can manage other keys
func (k *APIKey) CanManageKeys() bool {
	return k.Scope == APIKeyScopeAdmin
}

// CanWrite checks if this API key can modify resources
func (k *APIKey) CanWrite() bool {
	return k.Scope == APIKeyScopeAdmin || k.Scope == APIKeyScopeReadWrite
}
