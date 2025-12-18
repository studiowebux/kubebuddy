package domain

import (
	"time"
)

// Assignment represents a service deployed on a compute resource
type Assignment struct {
	ID        string    `json:"id"`
	ServiceID string    `json:"service_id"`
	ComputeID string    `json:"compute_id"`
	Quantity  int       `json:"quantity"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CanFitResources checks if required resources can fit within available resources
func CanFitResources(required Resources, available Resources) bool {
	for key, reqValue := range required {
		availValue, exists := available[key]
		if !exists {
			return false
		}

		// Compare numeric values
		switch req := reqValue.(type) {
		case int:
			if avail, ok := availValue.(int); ok {
				if req > avail {
					return false
				}
			} else if avail, ok := availValue.(float64); ok {
				if float64(req) > avail {
					return false
				}
			} else {
				return false
			}
		case float64:
			if avail, ok := availValue.(float64); ok {
				if req > avail {
					return false
				}
			} else if avail, ok := availValue.(int); ok {
				if req > float64(avail) {
					return false
				}
			} else {
				return false
			}
		default:
			// For non-numeric values, just check existence
			continue
		}
	}

	return true
}
