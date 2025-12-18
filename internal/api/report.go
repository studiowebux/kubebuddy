package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

type ComputeReportResponse struct {
	Compute             interface{} `json:"compute"`
	ComponentAssignments interface{} `json:"component_assignments"`
	ServiceAssignments  interface{} `json:"service_assignments"`
	IPAssignments       interface{} `json:"ip_assignments"`
	JournalEntries      interface{} `json:"journal_entries"`
	Statistics          *ResourceStatistics `json:"statistics,omitempty"`
}

func (s *Server) getComputeReport(c *gin.Context) {
	computeID := c.Param("id")

	// Get compute
	compute, err := s.store.Computes().Get(c.Request.Context(), computeID)
	if err != nil {
		handleError(c, http.StatusNotFound, "compute not found", err)
		return
	}

	// Get component assignments
	componentAssignments, err := s.store.ComputeComponents().ListByCompute(c.Request.Context(), computeID)
	if err != nil {
		componentAssignments = nil
	}

	// Get service assignments
	serviceAssignments, err := s.store.Assignments().List(c.Request.Context(), storage.AssignmentFilters{
		ComputeID: computeID,
	})
	if err != nil {
		serviceAssignments = nil
	}

	// Get IP assignments
	ipAssignments, err := s.store.ComputeIPs().ListByCompute(c.Request.Context(), computeID)
	if err != nil {
		ipAssignments = nil
	}

	// Get journal entries
	journalEntries, err := s.store.Journal().List(c.Request.Context(), storage.JournalFilters{
		ComputeID: computeID,
	})
	if err != nil {
		journalEntries = nil
	}

	// Get all services for statistics calculation
	allServices, err := s.store.Services().List(c.Request.Context())
	if err != nil {
		allServices = nil
	}
	servicesMap := make(map[string]*domain.Service)
	for _, svc := range allServices {
		servicesMap[svc.ID] = svc
	}

	// Calculate statistics for this compute's assignments
	stats := calculateResourceStatistics(serviceAssignments, servicesMap)

	report := ComputeReportResponse{
		Compute:             compute,
		ComponentAssignments: componentAssignments,
		ServiceAssignments:  serviceAssignments,
		IPAssignments:       ipAssignments,
		JournalEntries:      journalEntries,
		Statistics:          stats,
	}

	c.JSON(http.StatusOK, report)
}
