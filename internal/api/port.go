package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

func (s *Server) listPortAssignments(c *gin.Context) {
	filters := storage.PortAssignmentFilters{
		AssignmentID: c.Query("assignment_id"),
		IPID:         c.Query("ip_id"),
		Protocol:     c.Query("protocol"),
	}

	assignments, err := s.store.PortAssignments().List(c.Request.Context(), filters)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to list port assignments", err)
		return
	}

	if assignments == nil {
		assignments = []*domain.PortAssignment{}
	}

	c.JSON(http.StatusOK, assignments)
}

func (s *Server) getPortAssignment(c *gin.Context) {
	id := c.Param("id")

	assignment, err := s.store.PortAssignments().Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, http.StatusNotFound, "port assignment not found", err)
		return
	}

	c.JSON(http.StatusOK, assignment)
}

func (s *Server) createPortAssignment(c *gin.Context) {
	var assignment domain.PortAssignment

	if err := c.ShouldBindJSON(&assignment); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Check if port assignment with same ip_id+port+protocol already exists (upsert)
	existing, err := s.store.PortAssignments().GetByIPPortProtocol(c.Request.Context(), assignment.IPID, assignment.Port, string(assignment.Protocol))
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to check existing port assignment", err)
		return
	}

	if existing != nil {
		// Update existing port assignment
		assignment.ID = existing.ID
		assignment.CreatedAt = existing.CreatedAt

		if err := s.store.PortAssignments().Update(c.Request.Context(), &assignment); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to update port assignment", err)
			return
		}

		c.JSON(http.StatusOK, assignment)
	} else {
		// Create new port assignment
		if assignment.ID == "" {
			assignment.ID = uuid.New().String()
		}

		assignment.CreatedAt = time.Now()

		if err := s.store.PortAssignments().Create(c.Request.Context(), &assignment); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to create port assignment", err)
			return
		}

		c.JSON(http.StatusCreated, assignment)
	}
}

func (s *Server) updatePortAssignment(c *gin.Context) {
	id := c.Param("id")

	existing, err := s.store.PortAssignments().Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, http.StatusNotFound, "port assignment not found", err)
		return
	}

	var assignment domain.PortAssignment
	if err := c.ShouldBindJSON(&assignment); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	assignment.ID = existing.ID
	assignment.CreatedAt = existing.CreatedAt

	if err := s.store.PortAssignments().Update(c.Request.Context(), &assignment); err != nil {
		handleError(c, http.StatusInternalServerError, "failed to update port assignment", err)
		return
	}

	c.JSON(http.StatusOK, assignment)
}

func (s *Server) deletePortAssignment(c *gin.Context) {
	id := c.Param("id")

	if err := s.store.PortAssignments().Delete(c.Request.Context(), id); err != nil {
		handleError(c, http.StatusNotFound, "port assignment not found", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "port assignment deleted successfully"})
}
