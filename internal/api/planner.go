package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

func (s *Server) planCapacity(c *gin.Context) {
	var request domain.PlanRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Load all data for planning
	computes, err := s.store.Computes().List(c.Request.Context(), storage.ComputeFilters{})
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to load computes", err)
		return
	}

	// Populate compute resources from components
	for _, compute := range computes {
		// Get component assignments for this compute
		componentAssignments, err := s.store.ComputeComponents().ListByCompute(c.Request.Context(), compute.ID)
		if err != nil {
			continue // Skip on error
		}

		if len(componentAssignments) > 0 {
			// Load actual components
			components := make([]*domain.Component, 0, len(componentAssignments))
			for _, ca := range componentAssignments {
				comp, err := s.store.Components().Get(c.Request.Context(), ca.ComponentID)
				if err == nil {
					components = append(components, comp)
				}
			}

			// Calculate total resources from components
			compute.Resources = compute.GetTotalResourcesFromComponents(components, componentAssignments)
		}
	}

	services, err := s.store.Services().List(c.Request.Context())
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to load services", err)
		return
	}

	assignments, err := s.store.Assignments().List(c.Request.Context(), storage.AssignmentFilters{})
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to load assignments", err)
		return
	}

	// Create planner and execute
	planner := domain.NewCapacityPlanner(computes, services, assignments)
	result, err := planner.Plan(request)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to plan capacity", err)
		return
	}

	c.JSON(http.StatusOK, result)
}

type CapacityReportResponse struct {
	TotalComputes      int                    `json:"total_computes"`
	ActiveComputes     int                    `json:"active_computes"`
	TotalServices      int                    `json:"total_services"`
	TotalAssignments   int                    `json:"total_assignments"`
	ComputeUtilization []ComputeUtilization   `json:"compute_utilization"`
}

type ComputeUtilization struct {
	Compute         *domain.Compute       `json:"compute"`
	TotalResources  domain.Resources      `json:"total_resources"`
	Allocated       domain.Resources      `json:"allocated"`
	Available       domain.Resources      `json:"available"`
	UtilizationPct  float64               `json:"utilization_pct"`
	Statistics      *ResourceStatistics   `json:"statistics,omitempty"`
}

type ResourceStatistics struct {
	Min    domain.Resources `json:"min"`
	Max    domain.Resources `json:"max"`
	Avg    domain.Resources `json:"avg"`
	Median domain.Resources `json:"median"`
}

func (s *Server) capacityReport(c *gin.Context) {
	computes, err := s.store.Computes().List(c.Request.Context(), storage.ComputeFilters{})
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to load computes", err)
		return
	}

	// Populate compute resources from components
	for _, compute := range computes {
		// Get component assignments for this compute
		componentAssignments, err := s.store.ComputeComponents().ListByCompute(c.Request.Context(), compute.ID)
		if err != nil {
			continue // Skip on error
		}

		if len(componentAssignments) > 0 {
			// Load actual components
			components := make([]*domain.Component, 0, len(componentAssignments))
			for _, ca := range componentAssignments {
				comp, err := s.store.Components().Get(c.Request.Context(), ca.ComponentID)
				if err == nil {
					components = append(components, comp)
				}
			}

			// Calculate total resources from components
			compute.Resources = compute.GetTotalResourcesFromComponents(components, componentAssignments)
		}
	}

	services, err := s.store.Services().List(c.Request.Context())
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to load services", err)
		return
	}

	// Build services map for resource calculation
	servicesMap := make(map[string]*domain.Service)
	for _, svc := range services {
		servicesMap[svc.ID] = svc
	}

	assignments, err := s.store.Assignments().List(c.Request.Context(), storage.AssignmentFilters{})
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to load assignments", err)
		return
	}

	// Calculate utilization for each compute
	computeUtils := make([]ComputeUtilization, 0, len(computes))
	activeCount := 0

	for _, compute := range computes {
		if compute.State == domain.ComputeStateActive {
			activeCount++
		}

		allocated := compute.GetAllocatedResources(assignments, servicesMap)
		available := compute.GetAvailableResources(allocated)

		// Calculate average utilization percentage
		totalUtil := 0.0
		resourceCount := 0

		for key, total := range compute.Resources {
			if alloc, ok := allocated[key]; ok {
				// Convert both to float64 for comparison
				var totalFloat, allocFloat float64

				switch t := total.(type) {
				case int:
					totalFloat = float64(t)
				case float64:
					totalFloat = t
				default:
					continue
				}

				switch a := alloc.(type) {
				case int:
					allocFloat = float64(a)
				case float64:
					allocFloat = a
				default:
					continue
				}

				if totalFloat > 0 {
					totalUtil += (allocFloat / totalFloat) * 100
					resourceCount++
				}
			}
		}

		avgUtil := 0.0
		if resourceCount > 0 {
			avgUtil = totalUtil / float64(resourceCount)
		}

		// Calculate statistics for this compute's assignments
		computeAssignments := make([]*domain.Assignment, 0)
		for _, a := range assignments {
			if a.ComputeID == compute.ID {
				computeAssignments = append(computeAssignments, a)
			}
		}
		stats := calculateResourceStatistics(computeAssignments, servicesMap)

		computeUtils = append(computeUtils, ComputeUtilization{
			Compute:        compute,
			TotalResources: compute.Resources,
			Allocated:      allocated,
			Available:      available,
			UtilizationPct: avgUtil,
			Statistics:     stats,
		})
	}

	report := CapacityReportResponse{
		TotalComputes:      len(computes),
		ActiveComputes:     activeCount,
		TotalServices:      len(services),
		TotalAssignments:   len(assignments),
		ComputeUtilization: computeUtils,
	}

	c.JSON(http.StatusOK, report)
}

// calculateResourceStatistics calculates min/max/avg/median for resources across assignments
// All values are based on sum of max_spec (what services could use at maximum)
// Min = smallest max_spec across all assignments
// Max = sum of all max_spec (total if all services maxed out)
// Avg = average max_spec value
// Median = median max_spec value
func calculateResourceStatistics(assignments []*domain.Assignment, servicesMap map[string]*domain.Service) *ResourceStatistics {
	if len(assignments) == 0 {
		return nil
	}

	// Collect individual max values for statistics
	maxValues := make(map[string][]float64)

	for _, assignment := range assignments {
		service, ok := servicesMap[assignment.ServiceID]
		if !ok {
			continue
		}

		quantity := assignment.Quantity
		if quantity == 0 {
			quantity = 1
		}

		// Process MaxSpec
		for key, value := range service.MaxSpec {
			var floatVal float64
			switch v := value.(type) {
			case int:
				floatVal = float64(v) * float64(quantity)
			case float64:
				floatVal = v * float64(quantity)
			default:
				continue
			}
			maxValues[key] = append(maxValues[key], floatVal)
		}
	}

	// Build final statistics
	min := make(domain.Resources)
	max := make(domain.Resources)
	avg := make(domain.Resources)
	median := make(domain.Resources)

	for key, values := range maxValues {
		if len(values) == 0 {
			continue
		}

		// Min is the smallest max_spec value
		minVal := values[0]
		for _, v := range values {
			if v < minVal {
				minVal = v
			}
		}
		min[key] = minVal

		// Max is sum of all max_spec
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		max[key] = sum

		// Average
		avg[key] = sum / float64(len(values))

		// Median (sort values)
		sortedValues := make([]float64, len(values))
		copy(sortedValues, values)
		// Simple bubble sort for small arrays
		for i := 0; i < len(sortedValues); i++ {
			for j := i + 1; j < len(sortedValues); j++ {
				if sortedValues[i] > sortedValues[j] {
					sortedValues[i], sortedValues[j] = sortedValues[j], sortedValues[i]
				}
			}
		}

		if len(sortedValues)%2 == 0 {
			median[key] = (sortedValues[len(sortedValues)/2-1] + sortedValues[len(sortedValues)/2]) / 2
		} else {
			median[key] = sortedValues[len(sortedValues)/2]
		}
	}

	return &ResourceStatistics{
		Min:    min,
		Max:    max,
		Avg:    avg,
		Median: median,
	}
}
