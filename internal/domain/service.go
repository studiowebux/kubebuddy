package domain

import "time"

// Service represents an application or workload with resource requirements
type Service struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	MinSpec   Resources           `json:"min_spec"`
	MaxSpec   Resources           `json:"max_spec"`
	Placement PlacementRules      `json:"placement"`
	Ports     []PortRequirement   `json:"ports,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// PlacementRules defines constraints for service placement
type PlacementRules struct {
	Affinity     []TagSelector `json:"affinity,omitempty"`
	AntiAffinity []TagSelector `json:"antiAffinity,omitempty"`
	SpreadMax    int           `json:"spreadMax,omitempty"` // Max instances per compute (0 = unlimited)
	TopologyKey  string        `json:"topologyKey,omitempty"` // Tag key to spread across
}

// TagSelector matches compute tags
type TagSelector struct {
	MatchLabels      map[string]string `json:"matchLabels,omitempty"`
	MatchExpressions []Expression      `json:"matchExpressions,omitempty"`
}

// Expression represents a tag matching expression
type Expression struct {
	Key      string   `json:"key"`
	Operator Operator `json:"operator"`
	Values   []string `json:"values,omitempty"`
}

// Operator defines tag matching operators
type Operator string

const (
	OperatorIn           Operator = "In"
	OperatorNotIn        Operator = "NotIn"
	OperatorExists       Operator = "Exists"
	OperatorDoesNotExist Operator = "DoesNotExist"
)

// Matches checks if a TagSelector matches the given tags
func (ts *TagSelector) Matches(tags map[string]string) bool {
	// Check exact label matches
	for key, value := range ts.MatchLabels {
		if tags[key] != value {
			return false
		}
	}

	// Check expressions
	for _, expr := range ts.MatchExpressions {
		if !expr.Matches(tags) {
			return false
		}
	}

	return true
}

// Matches checks if an Expression matches the given tags
func (e *Expression) Matches(tags map[string]string) bool {
	value, exists := tags[e.Key]

	switch e.Operator {
	case OperatorExists:
		return exists
	case OperatorDoesNotExist:
		return !exists
	case OperatorIn:
		if !exists {
			return false
		}
		for _, v := range e.Values {
			if v == value {
				return true
			}
		}
		return false
	case OperatorNotIn:
		if !exists {
			return true
		}
		for _, v := range e.Values {
			if v == value {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// CanPlaceOn checks if service can be placed on compute based on placement rules
func (s *Service) CanPlaceOn(compute *Compute, existingAssignments []*Assignment) bool {
	// Check affinity rules (must match)
	for _, selector := range s.Placement.Affinity {
		if !selector.Matches(compute.Tags) {
			return false
		}
	}

	// Check anti-affinity rules (must NOT match)
	for _, selector := range s.Placement.AntiAffinity {
		if selector.Matches(compute.Tags) {
			return false
		}
	}

	// Check spread constraint (max instances per compute)
	if s.Placement.SpreadMax > 0 {
		count := 0
		for _, assignment := range existingAssignments {
			if assignment.ServiceID == s.ID && assignment.ComputeID == compute.ID {
				count++
			}
		}
		if count >= s.Placement.SpreadMax {
			return false
		}
	}

	return true
}
