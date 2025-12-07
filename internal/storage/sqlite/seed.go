package sqlite

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/studiowebux/kubebuddy/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// CreateAdminKey creates the admin API key
func (s *SQLiteStorage) CreateAdminKey(ctx context.Context, adminKey string) error {
	keyHash, err := bcrypt.GenerateFromPassword([]byte(adminKey), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin key: %w", err)
	}

	adminAPIKey := &domain.APIKey{
		ID:          uuid.New().String(),
		Name:        "admin",
		KeyHash:     string(keyHash),
		Scope:       domain.APIKeyScopeAdmin,
		Description: "Default admin API key",
	}

	if err := s.APIKeys().Create(ctx, adminAPIKey); err != nil {
		return fmt.Errorf("failed to create admin API key: %w", err)
	}

	return nil
}

// Seed populates the database with sample data
func (s *SQLiteStorage) Seed(ctx context.Context) error {

	// Sample computes
	computes := []*domain.Compute{
		{
			ID:       uuid.New().String(),
			Name:     "baremetal-prod-01",
			Type:     domain.ComputeTypeBaremetal,
			Provider: "OVH",
			Region:   "us-east-1",
			Tags: map[string]string{
				"env":  "prod",
				"zone": "us-east-1a",
				"type": "baremetal",
			},
			Resources: domain.Resources{
				"cpu":            32,
				"ram_gb":         128,
				"nvme_gb":        2000,
				"bandwidth_mbps": 10000,
			},
			State: domain.ComputeStateActive,
		},
		{
			ID:       uuid.New().String(),
			Name:     "vps-staging-01",
			Type:     domain.ComputeTypeVPS,
			Provider: "AWS",
			Region:   "us-west-2",
			Tags: map[string]string{
				"env":  "staging",
				"zone": "us-west-2b",
				"type": "vps",
			},
			Resources: domain.Resources{
				"cpu":            8,
				"ram_gb":         16,
				"ssd_gb":         200,
				"bandwidth_mbps": 1000,
			},
			State: domain.ComputeStateActive,
		},
		{
			ID:       uuid.New().String(),
			Name:     "vm-dev-01",
			Type:     domain.ComputeTypeVM,
			Provider: "GCP",
			Region:   "eu-central-1",
			Tags: map[string]string{
				"env":  "dev",
				"zone": "eu-central-1c",
				"type": "vm",
				"gpu":  "nvidia-t4",
			},
			Resources: domain.Resources{
				"cpu":            4,
				"ram_gb":         8,
				"ssd_gb":         100,
				"bandwidth_mbps": 500,
				"gpu":            1,
			},
			State: domain.ComputeStateActive,
		},
	}

	for _, compute := range computes {
		if err := s.Computes().Create(ctx, compute); err != nil {
			return fmt.Errorf("failed to create compute %s: %w", compute.Name, err)
		}
	}

	// Sample services
	services := []*domain.Service{
		{
			ID:   uuid.New().String(),
			Name: "nginx-ingress",
			MinSpec: domain.Resources{
				"cpu":    2,
				"ram_gb": 4,
			},
			MaxSpec: domain.Resources{
				"cpu":    4,
				"ram_gb": 8,
			},
			Placement: domain.PlacementRules{
				Affinity: []domain.TagSelector{
					{
						MatchLabels: map[string]string{
							"env": "prod",
						},
					},
				},
				SpreadMax: 1, // Max 1 instance per compute for redundancy
			},
		},
		{
			ID:   uuid.New().String(),
			Name: "postgres-db",
			MinSpec: domain.Resources{
				"cpu":       4,
				"ram_gb":    16,
				"nvme_gb":   100,
			},
			MaxSpec: domain.Resources{
				"cpu":       8,
				"ram_gb":    32,
				"nvme_gb":   500,
			},
			Placement: domain.PlacementRules{
				AntiAffinity: []domain.TagSelector{
					{
						MatchLabels: map[string]string{
							"env": "dev",
						},
					},
				},
				SpreadMax: 1, // Max 1 instance per compute for HA
			},
		},
		{
			ID:   uuid.New().String(),
			Name: "ml-worker",
			MinSpec: domain.Resources{
				"cpu":    2,
				"ram_gb": 8,
				"gpu":    1,
			},
			MaxSpec: domain.Resources{
				"cpu":    4,
				"ram_gb": 16,
				"gpu":    1,
			},
			Placement: domain.PlacementRules{
				Affinity: []domain.TagSelector{
					{
						MatchExpressions: []domain.Expression{
							{
								Key:      "gpu",
								Operator: domain.OperatorExists,
							},
						},
					},
				},
			},
		},
	}

	for _, service := range services {
		if err := s.Services().Create(ctx, service); err != nil {
			return fmt.Errorf("failed to create service %s: %w", service.Name, err)
		}
	}

	// Create sample assignment (nginx on baremetal)
	assignment := &domain.Assignment{
		ID:        uuid.New().String(),
		ServiceID: services[0].ID, // nginx-ingress
		ComputeID: computes[0].ID, // baremetal-prod-01
		Allocated: domain.Resources{
			"cpu":    2,
			"ram_gb": 4,
		},
	}

	if err := s.Assignments().Create(ctx, assignment); err != nil {
		return fmt.Errorf("failed to create assignment: %w", err)
	}

	// Sample components
	components := []*domain.Component{
		{
			ID:           uuid.New().String(),
			Name:         "Intel Xeon Gold 6258R",
			Type:         domain.ComponentTypeCPU,
			Manufacturer: "Intel",
			Model:        "Xeon Gold 6258R",
			Specs: map[string]interface{}{
				"cores":       28,
				"threads":     56,
				"base_ghz":    2.7,
				"boost_ghz":   4.0,
				"tdp_watts":   205,
				"socket":      "LGA 3647",
			},
			Notes: "High-performance server CPU",
		},
		{
			ID:           uuid.New().String(),
			Name:         "Samsung 64GB DDR4-3200 ECC",
			Type:         domain.ComponentTypeRAM,
			Manufacturer: "Samsung",
			Model:        "M393A8G40AB2-CWE",
			Specs: map[string]interface{}{
				"capacity_gb": 64,
				"type":        "DDR4",
				"speed_mhz":   3200,
				"ecc":         true,
				"registered":  true,
			},
		},
		{
			ID:           uuid.New().String(),
			Name:         "Samsung 2TB NVMe SSD",
			Type:         domain.ComponentTypeStorage,
			Manufacturer: "Samsung",
			Model:        "PM9A3",
			Specs: map[string]interface{}{
				"capacity_gb":   2000,
				"type":          "nvme",
				"interface":     "PCIe 4.0 x4",
				"read_mbps":     6800,
				"write_mbps":    5000,
			},
		},
		{
			ID:           uuid.New().String(),
			Name:         "NVIDIA T4 GPU",
			Type:         domain.ComponentTypeGPU,
			Manufacturer: "NVIDIA",
			Model:        "Tesla T4",
			Specs: map[string]interface{}{
				"vram_gb":       16,
				"cuda_cores":    2560,
				"tensor_cores":  320,
				"tdp_watts":     70,
				"interface":     "PCIe 3.0 x16",
			},
		},
		{
			ID:           uuid.New().String(),
			Name:         "Intel X710 Dual Port 10GbE",
			Type:         domain.ComponentTypeNIC,
			Manufacturer: "Intel",
			Model:        "X710-DA2",
			Specs: map[string]interface{}{
				"ports":       2,
				"speed_gbps":  10,
				"interface":   "PCIe 3.0 x8",
			},
		},
	}

	for _, component := range components {
		if err := s.Components().Create(ctx, component); err != nil {
			return fmt.Errorf("failed to create component %s: %w", component.Name, err)
		}
	}

	// Assign components to baremetal server
	componentAssignments := []*domain.ComputeComponent{
		{
			ID:          uuid.New().String(),
			ComputeID:   computes[0].ID, // baremetal-prod-01
			ComponentID: components[0].ID, // Intel Xeon
			Quantity:    2,
			Slot:        "CPU1, CPU2",
			Notes:       "Dual socket configuration",
		},
		{
			ID:          uuid.New().String(),
			ComputeID:   computes[0].ID,
			ComponentID: components[1].ID, // 64GB RAM
			Quantity:    8,
			Slot:        "DIMM0-7",
			Notes:       "512GB total RAM (8x64GB)",
		},
		{
			ID:          uuid.New().String(),
			ComputeID:   computes[0].ID,
			ComponentID: components[2].ID, // NVMe SSD
			Quantity:    2,
			Slot:        "M.2_1, M.2_2",
			Notes:       "4TB total NVMe storage",
		},
		{
			ID:          uuid.New().String(),
			ComputeID:   computes[0].ID,
			ComponentID: components[4].ID, // 10GbE NIC
			Quantity:    1,
			Slot:        "PCIe Slot 1",
		},
	}

	for _, assignment := range componentAssignments {
		if err := s.ComputeComponents().Assign(ctx, assignment); err != nil {
			return fmt.Errorf("failed to assign component: %w", err)
		}
	}

	// Assign GPU to dev VM
	gpuAssignment := &domain.ComputeComponent{
		ID:          uuid.New().String(),
		ComputeID:   computes[2].ID, // vm-dev-01
		ComponentID: components[3].ID, // NVIDIA T4
		Quantity:    1,
		Slot:        "Virtual GPU",
	}

	if err := s.ComputeComponents().Assign(ctx, gpuAssignment); err != nil {
		return fmt.Errorf("failed to assign GPU: %w", err)
	}

	// Create sample journal entry
	journalEntry := &domain.JournalEntry{
		ID:        uuid.New().String(),
		ComputeID: computes[0].ID,
		Category:  domain.JournalCategoryDeployment,
		Content:   "Deployed nginx-ingress v1.9.0\n\n## Configuration\n- SSL enabled\n- Rate limiting: 100 req/s\n- Backend: app-cluster",
	}

	if err := s.Journal().Create(ctx, journalEntry); err != nil {
		return fmt.Errorf("failed to create journal entry: %w", err)
	}

	return nil
}
