package domain

import "time"

// Predefined journal categories
const (
	JournalCategoryMaintenance = "maintenance"
	JournalCategoryIncident    = "incident"
	JournalCategoryDeployment  = "deployment"
	JournalCategoryHardware    = "hardware"
	JournalCategoryNetwork     = "network"
	JournalCategoryOther       = "other"
)

// JournalEntry represents a log entry for a compute resource
type JournalEntry struct {
	ID        string    `json:"id"`
	ComputeID string    `json:"compute_id"`
	Category  string    `json:"category"`
	Content   string    `json:"content"` // Plain text or markdown
	CreatedBy string    `json:"created_by"` // API key name that created this entry
	CreatedAt time.Time `json:"created_at"`
}

// PredefinedCategories returns list of predefined journal categories
func PredefinedCategories() []string {
	return []string{
		JournalCategoryMaintenance,
		JournalCategoryIncident,
		JournalCategoryDeployment,
		JournalCategoryHardware,
		JournalCategoryNetwork,
		JournalCategoryOther,
	}
}
