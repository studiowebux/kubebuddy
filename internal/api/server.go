package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

// Server represents the API server
type Server struct {
	store  storage.Storage
	router *gin.Engine
	addr   string
}

// NewServer creates a new API server
func NewServer(store storage.Storage, addr string) *Server {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware())

	s := &Server{
		store:  store,
		router: router,
		addr:   addr,
	}

	s.setupRoutes()

	return s
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health check (no auth required)
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	v1.Use(AuthMiddleware(s.store))

	// Compute routes
	computes := v1.Group("/computes")
	{
		computes.GET("", s.listComputes)
		computes.GET("/:id", s.getCompute)
		computes.POST("", RequireWrite(), s.createCompute)
		computes.PUT("/:id", RequireWrite(), s.updateCompute)
		computes.DELETE("/:id", RequireWrite(), s.deleteCompute)
	}

	// Service routes
	services := v1.Group("/services")
	{
		services.GET("", s.listServices)
		services.GET("/:id", s.getService)
		services.POST("", RequireWrite(), s.createService)
		services.PUT("/:id", RequireWrite(), s.updateService)
		services.DELETE("/:id", RequireWrite(), s.deleteService)
	}

	// Assignment routes
	assignments := v1.Group("/assignments")
	{
		assignments.GET("", s.listAssignments)
		assignments.GET("/:id", s.getAssignment)
		assignments.POST("", RequireWrite(), s.createAssignment)
		assignments.DELETE("/:id", RequireWrite(), s.deleteAssignment)
	}

	// Capacity planning routes
	capacity := v1.Group("/capacity")
	{
		capacity.POST("/plan", s.planCapacity)
		capacity.GET("/report", s.capacityReport)
	}

	// Journal routes
	journal := v1.Group("/journal")
	{
		journal.GET("", s.listJournalEntries)
		journal.GET("/:id", s.getJournalEntry)
		journal.POST("", RequireWrite(), s.createJournalEntry)
		journal.DELETE("/:id", RequireWrite(), s.deleteJournalEntry)
	}

	// Component routes
	components := v1.Group("/components")
	{
		components.GET("", s.listComponents)
		components.GET("/:id", s.getComponent)
		components.POST("", RequireWrite(), s.createComponent)
		components.PUT("/:id", RequireWrite(), s.updateComponent)
		components.DELETE("/:id", RequireWrite(), s.deleteComponent)
	}

	// Component assignment routes
	componentAssignments := v1.Group("/component-assignments")
	{
		componentAssignments.GET("", s.listComputeComponents)
		componentAssignments.POST("", RequireWrite(), s.assignComponent)
		componentAssignments.DELETE("/:id", RequireWrite(), s.unassignComponent)
	}

	// Admin routes (API key management)
	admin := v1.Group("/admin")
	admin.Use(RequireAdmin())
	{
		admin.GET("/apikeys", s.listAPIKeys)
		admin.GET("/apikeys/:id", s.getAPIKey)
		admin.POST("/apikeys", s.createAPIKey)
		admin.DELETE("/apikeys/:id", s.deleteAPIKey)
	}
}

// Start starts the API server
func (s *Server) Start() error {
	fmt.Printf("Starting KubeBuddy API server on %s\n", s.addr)
	return s.router.Run(s.addr)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	srv := &http.Server{
		Addr:    s.addr,
		Handler: s.router,
	}
	return srv.Shutdown(ctx)
}

// Helper to handle errors
func handleError(c *gin.Context, statusCode int, message string, err error) {
	if err != nil {
		c.JSON(statusCode, gin.H{
			"error":   message,
			"details": err.Error(),
		})
	} else {
		c.JSON(statusCode, gin.H{"error": message})
	}
}
