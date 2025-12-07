package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

func (s *Server) listComputes(c *gin.Context) {
	filters := storage.ComputeFilters{
		Type:     c.Query("type"),
		Provider: c.Query("provider"),
		Region:   c.Query("region"),
		State:    c.Query("state"),
		Tags:     ParseTags(c.Query("tags")),
	}

	computes, err := s.store.Computes().List(c.Request.Context(), filters)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to list computes", err)
		return
	}

	c.JSON(http.StatusOK, computes)
}

func (s *Server) getCompute(c *gin.Context) {
	id := c.Param("id")

	compute, err := s.store.Computes().Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, http.StatusNotFound, "compute not found", err)
		return
	}

	c.JSON(http.StatusOK, compute)
}

func (s *Server) createCompute(c *gin.Context) {
	var compute domain.Compute

	if err := c.ShouldBindJSON(&compute); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Initialize empty maps if nil
	if compute.Tags == nil {
		compute.Tags = make(map[string]string)
	}
	if compute.Resources == nil {
		compute.Resources = make(domain.Resources)
	}

	// Check if compute with same name+provider+region+type already exists (upsert)
	existing, err := s.store.Computes().GetByNameProviderRegionType(
		c.Request.Context(),
		compute.Name,
		compute.Provider,
		compute.Region,
		string(compute.Type),
	)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to check existing compute", err)
		return
	}

	if existing != nil {
		// Update existing compute
		compute.ID = existing.ID
		compute.CreatedAt = existing.CreatedAt
		compute.UpdatedAt = time.Now()

		if err := s.store.Computes().Update(c.Request.Context(), &compute); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to update compute", err)
			return
		}

		c.JSON(http.StatusOK, compute)
	} else {
		// Create new compute
		if compute.ID == "" {
			compute.ID = uuid.New().String()
		}

		// Set default state
		if compute.State == "" {
			compute.State = domain.ComputeStateActive
		}

		if err := s.store.Computes().Create(c.Request.Context(), &compute); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to create compute", err)
			return
		}

		c.JSON(http.StatusCreated, compute)
	}
}

func (s *Server) updateCompute(c *gin.Context) {
	id := c.Param("id")

	// Check if compute exists
	existing, err := s.store.Computes().Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, http.StatusNotFound, "compute not found", err)
		return
	}

	var compute domain.Compute
	if err := c.ShouldBindJSON(&compute); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Preserve ID and timestamps
	compute.ID = existing.ID
	compute.CreatedAt = existing.CreatedAt

	if err := s.store.Computes().Update(c.Request.Context(), &compute); err != nil {
		handleError(c, http.StatusInternalServerError, "failed to update compute", err)
		return
	}

	c.JSON(http.StatusOK, compute)
}

func (s *Server) deleteCompute(c *gin.Context) {
	id := c.Param("id")

	if err := s.store.Computes().Delete(c.Request.Context(), id); err != nil {
		handleError(c, http.StatusNotFound, "compute not found", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "compute deleted successfully"})
}
