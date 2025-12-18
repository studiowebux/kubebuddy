package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

func (s *Server) listIPAddresses(c *gin.Context) {
	filters := storage.IPAddressFilters{
		Type:     c.Query("type"),
		Provider: c.Query("provider"),
		Region:   c.Query("region"),
		State:    c.Query("state"),
	}

	ips, err := s.store.IPAddresses().List(c.Request.Context(), filters)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to list IP addresses", err)
		return
	}

	if ips == nil {
		ips = []*domain.IPAddress{}
	}

	c.JSON(http.StatusOK, ips)
}

func (s *Server) getIPAddress(c *gin.Context) {
	idOrAddress := c.Param("id")

	// Try to get by ID first
	ip, err := s.store.IPAddresses().Get(c.Request.Context(), idOrAddress)
	if err != nil {
		// If not found by ID, try by address
		ip, err = s.store.IPAddresses().GetByAddress(c.Request.Context(), idOrAddress)
		if err != nil {
			handleError(c, http.StatusNotFound, "IP address not found", err)
			return
		}
	}

	c.JSON(http.StatusOK, ip)
}

func (s *Server) createIPAddress(c *gin.Context) {
	var ip domain.IPAddress

	if err := c.ShouldBindJSON(&ip); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if ip.DNSServers == nil {
		ip.DNSServers = []string{}
	}

	// Check if IP with same address already exists (upsert)
	existing, err := s.store.IPAddresses().GetByAddress(c.Request.Context(), ip.Address)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to check existing IP", err)
		return
	}

	if existing != nil {
		// Update existing IP
		ip.ID = existing.ID
		ip.CreatedAt = existing.CreatedAt
		ip.UpdatedAt = time.Now()

		if err := s.store.IPAddresses().Update(c.Request.Context(), &ip); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to update IP address", err)
			return
		}

		c.JSON(http.StatusOK, ip)
	} else {
		// Create new IP
		if ip.ID == "" {
			ip.ID = uuid.New().String()
		}

		now := time.Now()
		ip.CreatedAt = now
		ip.UpdatedAt = now

		if err := s.store.IPAddresses().Create(c.Request.Context(), &ip); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to create IP address", err)
			return
		}

		c.JSON(http.StatusCreated, ip)
	}
}

func (s *Server) updateIPAddress(c *gin.Context) {
	id := c.Param("id")

	existing, err := s.store.IPAddresses().Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, http.StatusNotFound, "IP address not found", err)
		return
	}

	var ip domain.IPAddress
	if err := c.ShouldBindJSON(&ip); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	ip.ID = existing.ID
	ip.CreatedAt = existing.CreatedAt
	ip.UpdatedAt = time.Now()

	if err := s.store.IPAddresses().Update(c.Request.Context(), &ip); err != nil {
		handleError(c, http.StatusInternalServerError, "failed to update IP address", err)
		return
	}

	c.JSON(http.StatusOK, ip)
}

func (s *Server) deleteIPAddress(c *gin.Context) {
	id := c.Param("id")

	if err := s.store.IPAddresses().Delete(c.Request.Context(), id); err != nil {
		handleError(c, http.StatusNotFound, "IP address not found", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "IP address deleted successfully"})
}

func (s *Server) assignIP(c *gin.Context) {
	var assignment domain.ComputeIP

	if err := c.ShouldBindJSON(&assignment); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Verify compute exists
	if _, err := s.store.Computes().Get(c.Request.Context(), assignment.ComputeID); err != nil {
		handleError(c, http.StatusNotFound, "compute not found", err)
		return
	}

	// Verify IP exists
	ip, err := s.store.IPAddresses().Get(c.Request.Context(), assignment.IPID)
	if err != nil {
		handleError(c, http.StatusNotFound, "IP address not found", err)
		return
	}

	// Check if assignment already exists (upsert)
	existing, err := s.store.ComputeIPs().GetByComputeAndIP(c.Request.Context(), assignment.ComputeID, assignment.IPID)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to check existing assignment", err)
		return
	}

	if existing != nil {
		// Update existing assignment's primary flag
		if err := s.store.ComputeIPs().UpdatePrimary(c.Request.Context(), existing.ID, assignment.IsPrimary); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to update assignment", err)
			return
		}

		// Return updated assignment with new updated_at
		existing.IsPrimary = assignment.IsPrimary
		existing.UpdatedAt = time.Now()
		c.JSON(http.StatusOK, existing)
	} else {
		// Create new assignment
		if assignment.ID == "" {
			assignment.ID = uuid.New().String()
		}

		now := time.Now()
		assignment.CreatedAt = now
		assignment.UpdatedAt = now

		// Update IP state to assigned
		ip.State = domain.IPStateAssigned
		ip.UpdatedAt = time.Now()
		if err := s.store.IPAddresses().Update(c.Request.Context(), ip); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to update IP state", err)
			return
		}

		if err := s.store.ComputeIPs().Assign(c.Request.Context(), &assignment); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to assign IP", err)
			return
		}

		c.JSON(http.StatusCreated, assignment)
	}
}

func (s *Server) unassignIP(c *gin.Context) {
	id := c.Param("id")

	if err := s.store.ComputeIPs().Unassign(c.Request.Context(), id); err != nil {
		handleError(c, http.StatusNotFound, "assignment not found", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "IP unassigned successfully"})
}

func (s *Server) listComputeIPs(c *gin.Context) {
	computeID := c.Query("compute_id")
	ipID := c.Query("ip_id")

	var assignments []*domain.ComputeIP
	var err error

	if computeID != "" {
		assignments, err = s.store.ComputeIPs().ListByCompute(c.Request.Context(), computeID)
	} else if ipID != "" {
		assignments, err = s.store.ComputeIPs().ListByIP(c.Request.Context(), ipID)
	} else {
		// If no filter, return all IP assignments
		assignments, err = s.store.ComputeIPs().List(c.Request.Context())
	}

	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to list assignments", err)
		return
	}

	if assignments == nil {
		assignments = []*domain.ComputeIP{}
	}

	c.JSON(http.StatusOK, assignments)
}
