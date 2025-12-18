package domain

// PlanRequest represents a capacity planning request
type PlanRequest struct {
	ServiceID   string      `json:"service_id"`
	Constraints Constraints `json:"constraints,omitempty"`
}

// Constraints defines optional filters for capacity planning
type Constraints struct {
	ComputeID string  `json:"compute_id,omitempty"`
	Provider  string  `json:"provider,omitempty"`
	Region    string  `json:"region,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
	MinBuffer float64 `json:"min_buffer,omitempty"` // Minimum % of resources to keep available (0.0-1.0)
}

// PlanResult contains the result of capacity planning
type PlanResult struct {
	Feasible        bool              `json:"feasible"`
	Candidates      []Candidate       `json:"candidates,omitempty"`
	Recommendations []Recommendation  `json:"recommendations,omitempty"`
	Message         string            `json:"message,omitempty"`
}

// Candidate represents a compute resource that can accommodate the service
type Candidate struct {
	Compute         *Compute  `json:"compute"`
	UtilizationAfter float64  `json:"utilization_after"` // 0.0-1.0
	AvailableAfter  Resources `json:"available_after"`
	Score           float64   `json:"score"` // Higher is better fit
}

// Recommendation suggests what to purchase if no capacity is available
type Recommendation struct {
	Type      ComputeType `json:"type"`
	Spec      Resources   `json:"spec"`
	Quantity  int         `json:"quantity"`
	Rationale string      `json:"rationale"`
}

// CapacityPlanner handles capacity planning logic
type CapacityPlanner struct {
	computes    []*Compute
	services    []*Service
	assignments []*Assignment
}

// NewCapacityPlanner creates a new capacity planner
func NewCapacityPlanner(computes []*Compute, services []*Service, assignments []*Assignment) *CapacityPlanner {
	return &CapacityPlanner{
		computes:    computes,
		services:    services,
		assignments: assignments,
	}
}

// Plan evaluates capacity for a service
func (cp *CapacityPlanner) Plan(request PlanRequest) (*PlanResult, error) {
	// Find the service
	var service *Service
	for _, s := range cp.services {
		if s.ID == request.ServiceID {
			service = s
			break
		}
	}

	if service == nil {
		return &PlanResult{
			Feasible: false,
			Message:  "service not found",
		}, nil
	}

	// Filter compute resources
	candidates := make([]Candidate, 0)

	for _, compute := range cp.computes {
		// Skip inactive compute
		if compute.State != ComputeStateActive {
			continue
		}

		// Apply constraint filters
		if request.Constraints.ComputeID != "" && compute.ID != request.Constraints.ComputeID {
			continue
		}
		if request.Constraints.Provider != "" && compute.Provider != request.Constraints.Provider {
			continue
		}
		if request.Constraints.Region != "" && compute.Region != request.Constraints.Region {
			continue
		}
		if len(request.Constraints.Tags) > 0 && !compute.MatchesTags(request.Constraints.Tags) {
			continue
		}

		// Check placement rules (skip if specific compute requested)
		if request.Constraints.ComputeID == "" && !service.CanPlaceOn(compute, cp.assignments) {
			continue
		}

		// Build services map for resource calculation
		servicesMap := make(map[string]*Service)
		for _, svc := range cp.services {
			servicesMap[svc.ID] = svc
		}

		// Calculate available resources
		allocated := compute.GetAllocatedResources(cp.assignments, servicesMap)
		available := compute.GetAvailableResources(allocated)

		// Check if service min spec fits
		if !CanFitResources(service.MinSpec, available) {
			continue
		}

		// Apply buffer constraint
		if request.Constraints.MinBuffer > 0 {
			// Check if placing this service would leave enough buffer
			tempAllocated := make(Resources)
			for k, v := range allocated {
				tempAllocated[k] = v
			}
			for k, v := range service.MinSpec {
				if existing, ok := tempAllocated[k]; ok {
					// Handle type conversions for both int and float64
					switch e := existing.(type) {
					case int:
						switch val := v.(type) {
						case int:
							tempAllocated[k] = e + val
						case float64:
							tempAllocated[k] = e + int(val)
						}
					case float64:
						switch val := v.(type) {
						case int:
							tempAllocated[k] = e + float64(val)
						case float64:
							tempAllocated[k] = e + val
						}
					}
				} else {
					tempAllocated[k] = v
				}
			}

			// Calculate utilization after placement
			totalUtilization := 0.0
			resourceCount := 0
			for key, total := range compute.Resources {
				if alloc, ok := tempAllocated[key]; ok {
					switch t := total.(type) {
					case int:
						if t > 0 {
							switch a := alloc.(type) {
							case int:
								totalUtilization += float64(a) / float64(t)
								resourceCount++
							case float64:
								totalUtilization += a / float64(t)
								resourceCount++
							}
						}
					case float64:
						if t > 0 {
							switch a := alloc.(type) {
							case int:
								totalUtilization += float64(a) / t
								resourceCount++
							case float64:
								totalUtilization += a / t
								resourceCount++
							}
						}
					}
				}
			}

			avgUtilization := 0.0
			if resourceCount > 0 {
				avgUtilization = totalUtilization / float64(resourceCount)
			}

			if avgUtilization > (1.0 - request.Constraints.MinBuffer) {
				continue
			}
		}

		// Calculate score (prefer lower utilization for better headroom)
		totalUtilization := 0.0
		resourceCount := 0
		availableAfter := make(Resources)

		for key, total := range compute.Resources {
			allocAfter := allocated[key]
			if minReq, ok := service.MinSpec[key]; ok {
				// Add min spec to allocated
				switch total.(type) {
				case int:
					currentAlloc := 0
					// Handle type mismatches in allocated
					switch a := allocAfter.(type) {
					case int:
						currentAlloc = a
					case float64:
						currentAlloc = int(a)
					}
					// Handle both int and float64 from JSON in minReq
					switch m := minReq.(type) {
					case int:
						allocAfter = currentAlloc + m
					case float64:
						allocAfter = currentAlloc + int(m)
					}
				case float64:
					currentAlloc := 0.0
					// Handle type mismatches in allocated
					switch a := allocAfter.(type) {
					case int:
						currentAlloc = float64(a)
					case float64:
						currentAlloc = a
					}
					// Handle both int and float64 from JSON in minReq
					switch m := minReq.(type) {
					case int:
						allocAfter = currentAlloc + float64(m)
					case float64:
						allocAfter = currentAlloc + m
					}
				}
			}

			// Calculate utilization
			switch totalVal := total.(type) {
			case int:
				if totalVal > 0 {
					switch a := allocAfter.(type) {
					case int:
						util := float64(a) / float64(totalVal)
						totalUtilization += util
						resourceCount++
						availableAfter[key] = totalVal - a
					case float64:
						util := a / float64(totalVal)
						totalUtilization += util
						resourceCount++
						availableAfter[key] = totalVal - int(a)
					}
				}
			case float64:
				if totalVal > 0 {
					switch a := allocAfter.(type) {
					case int:
						util := float64(a) / totalVal
						totalUtilization += util
						resourceCount++
						availableAfter[key] = totalVal - float64(a)
					case float64:
						util := a / totalVal
						totalUtilization += util
						resourceCount++
						availableAfter[key] = totalVal - a
					}
				}
			}
		}

		avgUtilization := 0.0
		if resourceCount > 0 {
			avgUtilization = totalUtilization / float64(resourceCount)
		}

		// Score: prefer balanced utilization (not too empty, not too full)
		// Ideal target is around 60-70% utilization
		targetUtilization := 0.65
		score := 100.0 - (100.0 * abs(avgUtilization-targetUtilization))

		candidates = append(candidates, Candidate{
			Compute:          compute,
			UtilizationAfter: avgUtilization,
			AvailableAfter:   availableAfter,
			Score:            score,
		})
	}

	// Sort candidates by score (highest first)
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].Score > candidates[i].Score {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	if len(candidates) > 0 {
		return &PlanResult{
			Feasible:   true,
			Candidates: candidates,
			Message:    "found suitable compute resources",
		}, nil
	}

	// No candidates found, generate recommendations
	recommendations := cp.generateRecommendations(service)

	return &PlanResult{
		Feasible:        false,
		Recommendations: recommendations,
		Message:         "no suitable compute resources found, recommendations generated",
	}, nil
}

func (cp *CapacityPlanner) generateRecommendations(service *Service) []Recommendation {
	recommendations := make([]Recommendation, 0)

	// Recommend based on max spec (user already set their desired max)
	spec := make(Resources)
	for key, value := range service.MaxSpec {
		spec[key] = value
	}

	// Check affinity for preferred type
	preferredType := ComputeTypeBaremetal // Default to baremetal
	for _, selector := range service.Placement.Affinity {
		if computeType, ok := selector.MatchLabels["type"]; ok {
			switch computeType {
			case "vps":
				preferredType = ComputeTypeVPS
			case "vm":
				preferredType = ComputeTypeVM
			}
		}
	}

	recommendations = append(recommendations, Recommendation{
		Type:      preferredType,
		Spec:      spec,
		Quantity:  1,
		Rationale: "based on service max spec",
	})

	return recommendations
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
