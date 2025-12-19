package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

func (s *Server) listDNSRecords(c *gin.Context) {
	filters := storage.DNSRecordFilters{
		Type: c.Query("type"),
		Zone: c.Query("zone"),
		IPID: c.Query("ip_id"),
		Name: c.Query("name"),
	}

	records, err := s.store.DNSRecords().List(c.Request.Context(), filters)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to list DNS records", err)
		return
	}

	if records == nil {
		records = []*domain.DNSRecord{}
	}

	c.JSON(http.StatusOK, records)
}

func (s *Server) getDNSRecord(c *gin.Context) {
	id := c.Param("id")

	record, err := s.store.DNSRecords().Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, http.StatusNotFound, "DNS record not found", err)
		return
	}

	c.JSON(http.StatusOK, record)
}

func (s *Server) createDNSRecord(c *gin.Context) {
	var record domain.DNSRecord

	if err := c.ShouldBindJSON(&record); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Check if DNS record with same name+type+zone already exists (upsert)
	existing, err := s.store.DNSRecords().GetByNameTypeZone(c.Request.Context(), record.Name, string(record.Type), record.Zone)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to check existing DNS record", err)
		return
	}

	if existing != nil {
		// Update existing record
		record.ID = existing.ID
		record.CreatedAt = existing.CreatedAt
		record.UpdatedAt = time.Now()

		if err := s.store.DNSRecords().Update(c.Request.Context(), &record); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to update DNS record", err)
			return
		}

		c.JSON(http.StatusOK, record)
	} else {
		// Create new record
		if record.ID == "" {
			record.ID = uuid.New().String()
		}

		now := time.Now()
		record.CreatedAt = now
		record.UpdatedAt = now

		if record.TTL == 0 {
			record.TTL = 3600 // Default TTL
		}

		if err := s.store.DNSRecords().Create(c.Request.Context(), &record); err != nil {
			handleError(c, http.StatusInternalServerError, "failed to create DNS record", err)
			return
		}

		c.JSON(http.StatusCreated, record)
	}
}

func (s *Server) updateDNSRecord(c *gin.Context) {
	id := c.Param("id")

	existing, err := s.store.DNSRecords().Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, http.StatusNotFound, "DNS record not found", err)
		return
	}

	var record domain.DNSRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	record.ID = existing.ID
	record.CreatedAt = existing.CreatedAt
	record.UpdatedAt = time.Now()

	if err := s.store.DNSRecords().Update(c.Request.Context(), &record); err != nil {
		handleError(c, http.StatusInternalServerError, "failed to update DNS record", err)
		return
	}

	c.JSON(http.StatusOK, record)
}

func (s *Server) deleteDNSRecord(c *gin.Context) {
	id := c.Param("id")

	if err := s.store.DNSRecords().Delete(c.Request.Context(), id); err != nil {
		handleError(c, http.StatusNotFound, "DNS record not found", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "DNS record deleted successfully"})
}
