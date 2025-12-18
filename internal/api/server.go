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

	// API routes
	api := s.router.Group("/api")
	api.Use(AuthMiddleware(s.store))

	// Compute routes
	computes := api.Group("/computes")
	{
		computes.GET("", s.listComputes)
		computes.GET("/:id", s.getCompute)
		computes.POST("", RequireWrite(), s.createCompute)
		computes.PUT("/:id", RequireWrite(), s.updateCompute)
		computes.DELETE("/:id", RequireWrite(), s.deleteCompute)
	}

	// Service routes
	services := api.Group("/services")
	{
		services.GET("", s.listServices)
		services.GET("/:id", s.getService)
		services.POST("", RequireWrite(), s.createService)
		services.PUT("/:id", RequireWrite(), s.updateService)
		services.DELETE("/:id", RequireWrite(), s.deleteService)
	}

	// Assignment routes
	assignments := api.Group("/assignments")
	{
		assignments.GET("", s.listAssignments)
		assignments.GET("/:id", s.getAssignment)
		assignments.POST("", RequireWrite(), s.createAssignment)
		assignments.DELETE("/:id", RequireWrite(), s.deleteAssignment)
	}

	// Capacity planning routes
	capacity := api.Group("/capacity")
	{
		capacity.POST("/plan", s.planCapacity)
		capacity.GET("/report", s.capacityReport)
	}

	// Report routes
	reports := api.Group("/reports")
	{
		reports.GET("/compute/:id", s.getComputeReport)
	}

	// Journal routes
	journal := api.Group("/journal")
	{
		journal.GET("", s.listJournalEntries)
		journal.GET("/:id", s.getJournalEntry)
		journal.POST("", RequireWrite(), s.createJournalEntry)
		journal.DELETE("/:id", RequireWrite(), s.deleteJournalEntry)
	}

	// Component routes
	components := api.Group("/components")
	{
		components.GET("", s.listComponents)
		components.GET("/:id", s.getComponent)
		components.POST("", RequireWrite(), s.createComponent)
		components.PUT("/:id", RequireWrite(), s.updateComponent)
		components.DELETE("/:id", RequireWrite(), s.deleteComponent)
	}

	// Component assignment routes
	componentAssignments := api.Group("/component-assignments")
	{
		componentAssignments.GET("", s.listComputeComponents)
		componentAssignments.POST("", RequireWrite(), s.assignComponent)
		componentAssignments.DELETE("/:id", RequireWrite(), s.unassignComponent)
	}

	// IP address routes
	ips := api.Group("/ips")
	{
		ips.GET("", s.listIPAddresses)
		ips.GET("/:id", s.getIPAddress)
		ips.POST("", RequireWrite(), s.createIPAddress)
		ips.PUT("/:id", RequireWrite(), s.updateIPAddress)
		ips.DELETE("/:id", RequireWrite(), s.deleteIPAddress)
	}

	// IP assignment routes
	ipAssignments := api.Group("/ip-assignments")
	{
		ipAssignments.GET("", s.listComputeIPs)
		ipAssignments.POST("", RequireWrite(), s.assignIP)
		ipAssignments.DELETE("/:id", RequireWrite(), s.unassignIP)
	}

	// DNS record routes
	dns := api.Group("/dns")
	{
		dns.GET("", s.listDNSRecords)
		dns.GET("/:id", s.getDNSRecord)
		dns.POST("", RequireWrite(), s.createDNSRecord)
		dns.PUT("/:id", RequireWrite(), s.updateDNSRecord)
		dns.DELETE("/:id", RequireWrite(), s.deleteDNSRecord)
	}

	// Port assignment routes
	ports := api.Group("/ports")
	{
		ports.GET("", s.listPortAssignments)
		ports.GET("/:id", s.getPortAssignment)
		ports.POST("", RequireWrite(), s.createPortAssignment)
		ports.PUT("/:id", RequireWrite(), s.updatePortAssignment)
		ports.DELETE("/:id", RequireWrite(), s.deletePortAssignment)
	}

	// Firewall rule routes
	firewallRules := api.Group("/firewall-rules")
	{
		firewallRules.GET("", s.listFirewallRules)
		firewallRules.GET("/:id", s.getFirewallRule)
		firewallRules.POST("", RequireWrite(), s.createFirewallRule)
		firewallRules.PUT("/:id", RequireWrite(), s.updateFirewallRule)
		firewallRules.DELETE("/:id", RequireWrite(), s.deleteFirewallRule)
	}

	// Firewall rule assignment routes
	firewallAssignments := api.Group("/firewall-assignments")
	{
		firewallAssignments.GET("", s.listComputeFirewallRules)
		firewallAssignments.POST("", RequireWrite(), s.assignFirewallRule)
		firewallAssignments.DELETE("/:id", RequireWrite(), s.unassignFirewallRule)
		firewallAssignments.PATCH("/:id/enabled", RequireWrite(), s.updateFirewallRuleEnabled)
	}

	// Admin routes (API key management)
	admin := api.Group("/admin")
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
