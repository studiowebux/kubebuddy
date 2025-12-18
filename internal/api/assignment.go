package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

func (s *Server) listAssignments(c *gin.Context) {
	filters := storage.AssignmentFilters{
		ServiceID: c.Query("service_id"),
		ComputeID: c.Query("compute_id"),
	}

	assignments, err := s.store.Assignments().List(c.Request.Context(), filters)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to list assignments", err)
		return
	}

	c.JSON(http.StatusOK, assignments)
}

func (s *Server) getAssignment(c *gin.Context) {
	id := c.Param("id")

	assignment, err := s.store.Assignments().Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, http.StatusNotFound, "assignment not found", err)
		return
	}

	c.JSON(http.StatusOK, assignment)
}

func (s *Server) createAssignment(c *gin.Context) {
	var assignment domain.Assignment

	if err := c.ShouldBindJSON(&assignment); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Verify service exists
	service, err := s.store.Services().Get(c.Request.Context(), assignment.ServiceID)
	if err != nil {
		handleError(c, http.StatusBadRequest, "service not found", err)
		return
	}

	// Verify compute exists
	compute, err := s.store.Computes().Get(c.Request.Context(), assignment.ComputeID)
	if err != nil {
		handleError(c, http.StatusBadRequest, "compute not found", err)
		return
	}

	// Populate compute resources from components
	componentAssignments, err := s.store.ComputeComponents().ListByCompute(c.Request.Context(), compute.ID)
	if err == nil && len(componentAssignments) > 0 {
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

	// Check if assignment already exists first (for upsert logic)
	existing, err := s.store.Assignments().GetByComputeAndService(c.Request.Context(), assignment.ComputeID, assignment.ServiceID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to check existing assignment", err)
		return
	}

	// Check for force flag to bypass validation
	force := c.Query("force") == "true"

	// Get all existing assignments
	allAssignments, err := s.store.Assignments().List(c.Request.Context(), storage.AssignmentFilters{})
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to verify capacity", err)
		return
	}

	// Get all services to calculate allocated resources
	allServices, err := s.store.Services().List(c.Request.Context())
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to load services", err)
		return
	}
	servicesMap := make(map[string]*domain.Service)
	for _, svc := range allServices {
		servicesMap[svc.ID] = svc
	}

	if !force {
		// Check placement rules
		if !service.CanPlaceOn(compute, allAssignments) {
			handleError(c, http.StatusBadRequest, "placement rules violated", nil)
			return
		}

		// For updates, exclude the existing assignment from capacity check
		assignmentsForCapacity := allAssignments
		if existing != nil {
			assignmentsForCapacity = make([]*domain.Assignment, 0, len(allAssignments)-1)
			for _, a := range allAssignments {
				if a.ID != existing.ID {
					assignmentsForCapacity = append(assignmentsForCapacity, a)
				}
			}
		}

		// Check if resources are available
		allocated := compute.GetAllocatedResources(assignmentsForCapacity, servicesMap)
		available := compute.GetAvailableResources(allocated)

		// Use service max spec for capacity planning, multiplied by requested quantity
		quantity := assignment.Quantity
		if quantity == 0 {
			quantity = 1
		}

		requiredResources := make(domain.Resources)
		for key, value := range service.MaxSpec {
			switch v := value.(type) {
			case int:
				requiredResources[key] = v * quantity
			case float64:
				requiredResources[key] = v * float64(quantity)
			default:
				requiredResources[key] = value
			}
		}

		if !domain.CanFitResources(requiredResources, available) {
			handleError(c, http.StatusBadRequest, "insufficient resources available", nil)
			return
		}
	}

	if existing != nil {
		// Update existing assignment
		assignment.ID = existing.ID
		assignment.CreatedAt = existing.CreatedAt
		if err := s.store.Assignments().Update(c.Request.Context(), &assignment); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to update assignment", err)
			return
		}
		c.JSON(http.StatusOK, assignment)
	} else {
		// Create new assignment
		assignment.ID = uuid.New().String()
		if err := s.store.Assignments().Create(c.Request.Context(), &assignment); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to create assignment", err)
			return
		}
		c.JSON(http.StatusCreated, assignment)
	}
}

func (s *Server) deleteAssignment(c *gin.Context) {
	id := c.Param("id")

	if err := s.store.Assignments().Delete(c.Request.Context(), id); err != nil {
		handleError(c, http.StatusNotFound, "assignment not found", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "assignment deleted successfully"})
}
