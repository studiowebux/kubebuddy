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
	TotalComputes   int                       `json:"total_computes"`
	ActiveComputes  int                       `json:"active_computes"`
	TotalServices   int                       `json:"total_services"`
	TotalAssignments int                      `json:"total_assignments"`
	ComputeUtilization []ComputeUtilization   `json:"compute_utilization"`
}

type ComputeUtilization struct {
	Compute         *domain.Compute  `json:"compute"`
	Allocated       domain.Resources `json:"allocated"`
	Available       domain.Resources `json:"available"`
	UtilizationPct  float64          `json:"utilization_pct"`
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

		allocated := compute.GetAllocatedResources(assignments)
		available := compute.GetAvailableResources(allocated)

		// Calculate average utilization percentage
		totalUtil := 0.0
		resourceCount := 0

		for key, total := range compute.Resources {
			if alloc, ok := allocated[key]; ok {
				switch t := total.(type) {
				case int:
					if a, ok := alloc.(int); ok && t > 0 {
						totalUtil += (float64(a) / float64(t)) * 100
						resourceCount++
					}
				case float64:
					if a, ok := alloc.(float64); ok && t > 0 {
						totalUtil += (a / t) * 100
						resourceCount++
					}
				}
			}
		}

		avgUtil := 0.0
		if resourceCount > 0 {
			avgUtil = totalUtil / float64(resourceCount)
		}

		computeUtils = append(computeUtils, ComputeUtilization{
			Compute:        compute,
			Allocated:      allocated,
			Available:      available,
			UtilizationPct: avgUtil,
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
