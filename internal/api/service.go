package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studiowebux/kubebuddy/internal/domain"
)

func (s *Server) listServices(c *gin.Context) {
	services, err := s.store.Services().List(c.Request.Context())
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to list services", err)
		return
	}

	c.JSON(http.StatusOK, services)
}

func (s *Server) getService(c *gin.Context) {
	id := c.Param("id")

	service, err := s.store.Services().Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, http.StatusNotFound, "service not found", err)
		return
	}

	c.JSON(http.StatusOK, service)
}

func (s *Server) createService(c *gin.Context) {
	var service domain.Service

	if err := c.ShouldBindJSON(&service); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Initialize empty maps if nil
	if service.MinSpec == nil {
		service.MinSpec = make(domain.Resources)
	}
	if service.MaxSpec == nil {
		service.MaxSpec = make(domain.Resources)
	}

	// Check if service with same name already exists (upsert)
	existing, err := s.store.Services().GetByName(c.Request.Context(), service.Name)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to check existing service", err)
		return
	}

	if existing != nil {
		// Update existing service
		service.ID = existing.ID
		service.CreatedAt = existing.CreatedAt

		if err := s.store.Services().Update(c.Request.Context(), &service); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to update service", err)
			return
		}

		c.JSON(http.StatusOK, service)
	} else {
		// Create new service
		if service.ID == "" {
			service.ID = uuid.New().String()
		}

		if err := s.store.Services().Create(c.Request.Context(), &service); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to create service", err)
			return
		}

		c.JSON(http.StatusCreated, service)
	}
}

func (s *Server) updateService(c *gin.Context) {
	id := c.Param("id")

	// Check if service exists
	existing, err := s.store.Services().Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, http.StatusNotFound, "service not found", err)
		return
	}

	var service domain.Service
	if err := c.ShouldBindJSON(&service); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Preserve ID and timestamps
	service.ID = existing.ID
	service.CreatedAt = existing.CreatedAt

	if err := s.store.Services().Update(c.Request.Context(), &service); err != nil {
		handleError(c, http.StatusInternalServerError, "failed to update service", err)
		return
	}

	c.JSON(http.StatusOK, service)
}

func (s *Server) deleteService(c *gin.Context) {
	id := c.Param("id")

	if err := s.store.Services().Delete(c.Request.Context(), id); err != nil {
		handleError(c, http.StatusNotFound, "service not found", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "service deleted successfully"})
}
