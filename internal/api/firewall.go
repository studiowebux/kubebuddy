package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

func (s *Server) listFirewallRules(c *gin.Context) {
	filters := storage.FirewallRuleFilters{
		Action:   c.Query("action"),
		Protocol: c.Query("protocol"),
	}

	rules, err := s.store.FirewallRules().List(c.Request.Context(), filters)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to list firewall rules", err)
		return
	}

	if rules == nil {
		rules = []*domain.FirewallRule{}
	}

	c.JSON(http.StatusOK, rules)
}

func (s *Server) getFirewallRule(c *gin.Context) {
	id := c.Param("id")

	rule, err := s.store.FirewallRules().Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, http.StatusNotFound, "firewall rule not found", err)
		return
	}

	c.JSON(http.StatusOK, rule)
}

func (s *Server) createFirewallRule(c *gin.Context) {
	var rule domain.FirewallRule

	if err := c.ShouldBindJSON(&rule); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Check if firewall rule with same name already exists (upsert)
	existing, err := s.store.FirewallRules().GetByName(c.Request.Context(), rule.Name)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to check existing firewall rule", err)
		return
	}

	if existing != nil {
		// Update existing rule
		rule.ID = existing.ID
		rule.CreatedAt = existing.CreatedAt
		rule.UpdatedAt = time.Now()

		if err := s.store.FirewallRules().Update(c.Request.Context(), &rule); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to update firewall rule", err)
			return
		}

		c.JSON(http.StatusOK, rule)
	} else {
		// Create new rule
		if rule.ID == "" {
			rule.ID = uuid.New().String()
		}

		now := time.Now()
		rule.CreatedAt = now
		rule.UpdatedAt = now

		if rule.Priority == 0 {
			rule.Priority = 100 // Default priority
		}

		if err := s.store.FirewallRules().Create(c.Request.Context(), &rule); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to create firewall rule", err)
			return
		}

		c.JSON(http.StatusCreated, rule)
	}
}

func (s *Server) updateFirewallRule(c *gin.Context) {
	id := c.Param("id")

	existing, err := s.store.FirewallRules().Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, http.StatusNotFound, "firewall rule not found", err)
		return
	}

	var rule domain.FirewallRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	rule.ID = existing.ID
	rule.CreatedAt = existing.CreatedAt
	rule.UpdatedAt = time.Now()

	if err := s.store.FirewallRules().Update(c.Request.Context(), &rule); err != nil {
		handleError(c, http.StatusInternalServerError, "failed to update firewall rule", err)
		return
	}

	c.JSON(http.StatusOK, rule)
}

func (s *Server) deleteFirewallRule(c *gin.Context) {
	id := c.Param("id")

	if err := s.store.FirewallRules().Delete(c.Request.Context(), id); err != nil {
		handleError(c, http.StatusNotFound, "firewall rule not found", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "firewall rule deleted successfully"})
}

func (s *Server) listComputeFirewallRules(c *gin.Context) {
	computeID := c.Query("compute_id")
	ruleID := c.Query("rule_id")

	if computeID != "" {
		assignments, err := s.store.ComputeFirewallRules().ListByCompute(c.Request.Context(), computeID)
		if err != nil {
			handleError(c, http.StatusInternalServerError, "failed to list compute firewall rules", err)
			return
		}
		c.JSON(http.StatusOK, assignments)
		return
	}

	if ruleID != "" {
		assignments, err := s.store.ComputeFirewallRules().ListByRule(c.Request.Context(), ruleID)
		if err != nil {
			handleError(c, http.StatusInternalServerError, "failed to list compute firewall rules", err)
			return
		}
		c.JSON(http.StatusOK, assignments)
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "compute_id or rule_id required"})
}

func (s *Server) assignFirewallRule(c *gin.Context) {
	var assignment domain.ComputeFirewallRule

	if err := c.ShouldBindJSON(&assignment); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if assignment.ID == "" {
		assignment.ID = uuid.New().String()
	}

	assignment.CreatedAt = time.Now()

	if err := s.store.ComputeFirewallRules().Assign(c.Request.Context(), &assignment); err != nil {
		handleError(c, http.StatusInternalServerError, "failed to assign firewall rule", err)
		return
	}

	c.JSON(http.StatusCreated, assignment)
}

func (s *Server) unassignFirewallRule(c *gin.Context) {
	id := c.Param("id")

	if err := s.store.ComputeFirewallRules().Unassign(c.Request.Context(), id); err != nil {
		handleError(c, http.StatusNotFound, "firewall rule assignment not found", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "firewall rule unassigned successfully"})
}

func (s *Server) updateFirewallRuleEnabled(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if err := s.store.ComputeFirewallRules().UpdateEnabled(c.Request.Context(), id, req.Enabled); err != nil {
		handleError(c, http.StatusNotFound, "firewall rule assignment not found", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "firewall rule enabled status updated"})
}
