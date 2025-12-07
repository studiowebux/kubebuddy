package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

func (s *Server) listJournalEntries(c *gin.Context) {
	filters := storage.JournalFilters{
		ComputeID: c.Query("compute_id"),
		Category:  c.Query("category"),
	}

	// Parse from/to dates if provided
	if fromStr := c.Query("from"); fromStr != "" {
		if from, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filters.From = &from
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if to, err := time.Parse(time.RFC3339, toStr); err == nil {
			filters.To = &to
		}
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}

	entries, err := s.store.Journal().List(c.Request.Context(), filters)
	if err != nil {
		handleError(c, http.StatusInternalServerError, "failed to list journal entries", err)
		return
	}

	c.JSON(http.StatusOK, entries)
}

func (s *Server) getJournalEntry(c *gin.Context) {
	id := c.Param("id")

	entry, err := s.store.Journal().Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, http.StatusNotFound, "journal entry not found", err)
		return
	}

	c.JSON(http.StatusOK, entry)
}

func (s *Server) createJournalEntry(c *gin.Context) {
	var entry domain.JournalEntry

	if err := c.ShouldBindJSON(&entry); err != nil {
		handleError(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Generate ID if not provided
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}

	// Set created_by from authenticated API key
	if apiKey := GetAPIKey(c); apiKey != nil {
		entry.CreatedBy = apiKey.Name
	}

	// Verify compute exists
	if _, err := s.store.Computes().Get(c.Request.Context(), entry.ComputeID); err != nil {
		handleError(c, http.StatusBadRequest, "compute not found", err)
		return
	}

	if err := s.store.Journal().Create(c.Request.Context(), &entry); err != nil {
		handleError(c, http.StatusInternalServerError, "failed to create journal entry", err)
		return
	}

	c.JSON(http.StatusCreated, entry)
}

func (s *Server) deleteJournalEntry(c *gin.Context) {
	id := c.Param("id")

	if err := s.store.Journal().Delete(c.Request.Context(), id); err != nil {
		handleError(c, http.StatusNotFound, "journal entry not found", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "journal entry deleted successfully"})
}
