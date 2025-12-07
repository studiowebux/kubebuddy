package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

func (s *Server) listComponents(c *gin.Context) {
	filters := storage.ComponentFilters{
		Type:         c.Query("type"),
		Manufacturer: c.Query("manufacturer"),
	}

	components, err := s.store.Components().List(c.Request.Context(), filters)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to list components", err)
		return
	}

	c.JSON(http.StatusOK, components)
}

func (s *Server) getComponent(c *gin.Context) {
	id := c.Param("id")

	component, err := s.store.Components().Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, http.StatusNotFound, "component not found", err)
		return
	}

	c.JSON(http.StatusOK, component)
}

func (s *Server) createComponent(c *gin.Context) {
	var component domain.Component

	if err := c.ShouldBindJSON(&component); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if component.Specs == nil {
		component.Specs = make(map[string]interface{})
	}

	// Check if component with same manufacturer and model already exists (upsert)
	existing, err := s.store.Components().GetByManufacturerAndModel(c.Request.Context(), component.Manufacturer, component.Model)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to check existing component", err)
		return
	}

	if existing != nil {
		// Update existing component
		component.ID = existing.ID
		component.CreatedAt = existing.CreatedAt
		component.UpdatedAt = time.Now()

		if err := s.store.Components().Update(c.Request.Context(), &component); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to update component", err)
			return
		}

		c.JSON(http.StatusOK, component)
	} else {
		// Create new component
		if component.ID == "" {
			component.ID = uuid.New().String()
		}

		now := time.Now()
		component.CreatedAt = now
		component.UpdatedAt = now

		if err := s.store.Components().Create(c.Request.Context(), &component); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to create component", err)
			return
		}

		c.JSON(http.StatusCreated, component)
	}
}

func (s *Server) updateComponent(c *gin.Context) {
	id := c.Param("id")

	existing, err := s.store.Components().Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, http.StatusNotFound, "component not found", err)
		return
	}

	var component domain.Component
	if err := c.ShouldBindJSON(&component); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	component.ID = existing.ID
	component.CreatedAt = existing.CreatedAt
	component.UpdatedAt = time.Now()

	if err := s.store.Components().Update(c.Request.Context(), &component); err != nil {
		handleError(c, http.StatusInternalServerError, "failed to update component", err)
		return
	}

	c.JSON(http.StatusOK, component)
}

func (s *Server) deleteComponent(c *gin.Context) {
	id := c.Param("id")

	if err := s.store.Components().Delete(c.Request.Context(), id); err != nil {
		handleError(c, http.StatusNotFound, "component not found", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "component deleted successfully"})
}

func (s *Server) assignComponent(c *gin.Context) {
	var assignment domain.ComputeComponent

	if err := c.ShouldBindJSON(&assignment); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if assignment.ID == "" {
		assignment.ID = uuid.New().String()
	}

	if assignment.Quantity == 0 {
		assignment.Quantity = 1
	}

	assignment.CreatedAt = time.Now()

	// Verify compute exists
	if _, err := s.store.Computes().Get(c.Request.Context(), assignment.ComputeID); err != nil {
		handleError(c, http.StatusNotFound, "compute not found", err)
		return
	}

	// Verify component exists
	if _, err := s.store.Components().Get(c.Request.Context(), assignment.ComponentID); err != nil {
		handleError(c, http.StatusNotFound, "component not found", err)
		return
	}

	if err := s.store.ComputeComponents().Assign(c.Request.Context(), &assignment); err != nil {
		handleError(c, http.StatusInternalServerError, "failed to assign component", err)
		return
	}

	c.JSON(http.StatusCreated, assignment)
}

func (s *Server) unassignComponent(c *gin.Context) {
	id := c.Param("id")

	if err := s.store.ComputeComponents().Unassign(c.Request.Context(), id); err != nil {
		handleError(c, http.StatusNotFound, "assignment not found", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "component unassigned successfully"})
}

func (s *Server) listComputeComponents(c *gin.Context) {
	computeID := c.Query("compute_id")
	componentID := c.Query("component_id")

	var assignments []*domain.ComputeComponent
	var err error

	if computeID != "" {
		assignments, err = s.store.ComputeComponents().ListByCompute(c.Request.Context(), computeID)
	} else if componentID != "" {
		assignments, err = s.store.ComputeComponents().ListByComponent(c.Request.Context(), componentID)
	} else {
		handleError(c, http.StatusBadRequest, "compute_id or component_id query parameter required", nil)
		return
	}

	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to list assignments", err)
		return
	}

	c.JSON(http.StatusOK, assignments)
}
