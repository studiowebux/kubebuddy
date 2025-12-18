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
// Min = sum of all min_spec, Max = sum of all max_spec
// Avg/Median = based on max_spec values only
func calculateResourceStatistics(assignments []*domain.Assignment, servicesMap map[string]*domain.Service) *ResourceStatistics {
	if len(assignments) == 0 {
		return nil
	}

	// Collect min and max totals, and individual max values for avg/median
	minTotals := make(map[string]float64)
	maxTotals := make(map[string]float64)
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

		// Process MinSpec for totals
		for key, value := range service.MinSpec {
			var floatVal float64
			switch v := value.(type) {
			case int:
				floatVal = float64(v) * float64(quantity)
			case float64:
				floatVal = v * float64(quantity)
			default:
				continue
			}
			minTotals[key] += floatVal
		}

		// Process MaxSpec for totals and individual values
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
			maxTotals[key] += floatVal
			maxValues[key] = append(maxValues[key], floatVal)
		}
	}

	// Build final statistics
	min := make(domain.Resources)
	max := make(domain.Resources)
	avg := make(domain.Resources)
	median := make(domain.Resources)

	// Get all unique resource keys
	allKeys := make(map[string]bool)
	for key := range minTotals {
		allKeys[key] = true
	}
	for key := range maxTotals {
		allKeys[key] = true
	}

	for key := range allKeys {
		// Min is sum of all min_spec
		if val, ok := minTotals[key]; ok {
			min[key] = val
		}

		// Max is sum of all max_spec
		if val, ok := maxTotals[key]; ok {
			max[key] = val
		}

		// Avg and Median based on max_spec values
		if values, ok := maxValues[key]; ok && len(values) > 0 {
			// Average
			sum := 0.0
			for _, v := range values {
				sum += v
			}
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
	}

	return &ResourceStatistics{
		Min:    min,
		Max:    max,
		Avg:    avg,
		Median: median,
	}
}
