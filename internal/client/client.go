package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/studiowebux/kubebuddy/internal/domain"
	"github.com/studiowebux/kubebuddy/internal/storage"
)

// Client is the HTTP client for KubeBuddy API
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New creates a new API client
func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request
func (c *Client) doRequest(ctx context.Context, method, path string, body, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			if errMsg, ok := errResp["error"].(string); ok {
				return fmt.Errorf("API error (%d): %s", resp.StatusCode, errMsg)
			}
		}
		return fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Compute methods
func (c *Client) ListComputes(ctx context.Context, filters storage.ComputeFilters) ([]*domain.Compute, error) {
	var computes []*domain.Compute
	path := "/api/v1/computes"
	// TODO: Add query parameters for filters
	err := c.doRequest(ctx, http.MethodGet, path, nil, &computes)
	return computes, err
}

func (c *Client) GetCompute(ctx context.Context, id string) (*domain.Compute, error) {
	var compute domain.Compute
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/computes/%s", id), nil, &compute)
	return &compute, err
}

func (c *Client) CreateCompute(ctx context.Context, compute *domain.Compute) (*domain.Compute, error) {
	var result domain.Compute
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/computes", compute, &result)
	return &result, err
}

func (c *Client) UpdateCompute(ctx context.Context, id string, compute *domain.Compute) (*domain.Compute, error) {
	var result domain.Compute
	err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/api/v1/computes/%s", id), compute, &result)
	return &result, err
}

func (c *Client) DeleteCompute(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/computes/%s", id), nil, nil)
}

// Service methods
func (c *Client) ListServices(ctx context.Context) ([]*domain.Service, error) {
	var services []*domain.Service
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/services", nil, &services)
	return services, err
}

func (c *Client) GetService(ctx context.Context, id string) (*domain.Service, error) {
	var service domain.Service
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/services/%s", id), nil, &service)
	return &service, err
}

func (c *Client) CreateService(ctx context.Context, service *domain.Service) (*domain.Service, error) {
	var result domain.Service
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/services", service, &result)
	return &result, err
}

func (c *Client) UpdateService(ctx context.Context, id string, service *domain.Service) (*domain.Service, error) {
	var result domain.Service
	err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/api/v1/services/%s", id), service, &result)
	return &result, err
}

func (c *Client) DeleteService(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/services/%s", id), nil, nil)
}

// Assignment methods
func (c *Client) ListAssignments(ctx context.Context, filters storage.AssignmentFilters) ([]*domain.Assignment, error) {
	var assignments []*domain.Assignment
	path := "/api/v1/assignments"
	if filters.ComputeID != "" {
		path += "?compute_id=" + filters.ComputeID
	} else if filters.ServiceID != "" {
		path += "?service_id=" + filters.ServiceID
	}
	err := c.doRequest(ctx, http.MethodGet, path, nil, &assignments)
	return assignments, err
}

func (c *Client) CreateAssignment(ctx context.Context, assignment *domain.Assignment, force bool) (*domain.Assignment, error) {
	var result domain.Assignment
	path := "/api/v1/assignments"
	if force {
		path += "?force=true"
	}
	err := c.doRequest(ctx, http.MethodPost, path, assignment, &result)
	return &result, err
}

func (c *Client) DeleteAssignment(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/assignments/%s", id), nil, nil)
}

// Capacity planning methods
func (c *Client) PlanCapacity(ctx context.Context, request domain.PlanRequest) (*domain.PlanResult, error) {
	var result domain.PlanResult
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/capacity/plan", request, &result)
	return &result, err
}

// Journal methods
func (c *Client) ListJournalEntries(ctx context.Context, filters storage.JournalFilters) ([]*domain.JournalEntry, error) {
	var entries []*domain.JournalEntry
	path := "/api/v1/journal"
	if filters.ComputeID != "" {
		path += "?compute_id=" + filters.ComputeID
	}
	err := c.doRequest(ctx, http.MethodGet, path, nil, &entries)
	return entries, err
}

func (c *Client) ListJournal(ctx context.Context, filters storage.JournalFilters) ([]*domain.JournalEntry, error) {
	return c.ListJournalEntries(ctx, filters)
}

func (c *Client) CreateJournalEntry(ctx context.Context, entry *domain.JournalEntry) (*domain.JournalEntry, error) {
	var result domain.JournalEntry
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/journal", entry, &result)
	return &result, err
}

// Admin methods
func (c *Client) ListAPIKeys(ctx context.Context) ([]*domain.APIKey, error) {
	var keys []*domain.APIKey
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/admin/apikeys", nil, &keys)
	return keys, err
}

type CreateAPIKeyRequest struct {
	Name        string              `json:"name"`
	Scope       domain.APIKeyScope  `json:"scope"`
	Description string              `json:"description"`
	ExpiresAt   *time.Time          `json:"expires_at"`
}

type CreateAPIKeyResponse struct {
	APIKey *domain.APIKey `json:"api_key"`
	Key    string         `json:"key"`
}

func (c *Client) CreateAPIKey(ctx context.Context, req CreateAPIKeyRequest) (*CreateAPIKeyResponse, error) {
	var result CreateAPIKeyResponse
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/admin/apikeys", req, &result)
	return &result, err
}

func (c *Client) DeleteAPIKey(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/admin/apikeys/%s", id), nil, nil)
}

// Component methods
func (c *Client) ListComponents(ctx context.Context, filters storage.ComponentFilters) ([]*domain.Component, error) {
	var components []*domain.Component
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/components", nil, &components)
	return components, err
}

func (c *Client) GetComponent(ctx context.Context, id string) (*domain.Component, error) {
	var component domain.Component
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/components/%s", id), nil, &component)
	return &component, err
}

func (c *Client) CreateComponent(ctx context.Context, component *domain.Component) (*domain.Component, error) {
	var result domain.Component
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/components", component, &result)
	return &result, err
}

func (c *Client) UpdateComponent(ctx context.Context, id string, component *domain.Component) (*domain.Component, error) {
	var result domain.Component
	err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/api/v1/components/%s", id), component, &result)
	return &result, err
}

func (c *Client) DeleteComponent(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/components/%s", id), nil, nil)
}

// Component assignment methods
func (c *Client) AssignComponent(ctx context.Context, assignment *domain.ComputeComponent) (*domain.ComputeComponent, error) {
	var result domain.ComputeComponent
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/component-assignments", assignment, &result)
	return &result, err
}

func (c *Client) UnassignComponent(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/component-assignments/%s", id), nil, nil)
}

func (c *Client) ListComponentAssignments(ctx context.Context, filters storage.ComputeComponentFilters) ([]*domain.ComputeComponent, error) {
	var assignments []*domain.ComputeComponent
	path := "/api/v1/component-assignments"
	if filters.ComputeID != "" {
		path += "?compute_id=" + filters.ComputeID
	} else if filters.ComponentID != "" {
		path += "?component_id=" + filters.ComponentID
	}
	err := c.doRequest(ctx, http.MethodGet, path, nil, &assignments)
	return assignments, err
}

// IP address methods
func (c *Client) ListIPAddresses(ctx context.Context, filters storage.IPAddressFilters) ([]*domain.IPAddress, error) {
	var ips []*domain.IPAddress
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/ips", nil, &ips)
	return ips, err
}

func (c *Client) GetIPAddress(ctx context.Context, id string) (*domain.IPAddress, error) {
	var ip domain.IPAddress
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/ips/%s", id), nil, &ip)
	return &ip, err
}

func (c *Client) CreateIPAddress(ctx context.Context, ip *domain.IPAddress) (*domain.IPAddress, error) {
	var result domain.IPAddress
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/ips", ip, &result)
	return &result, err
}

func (c *Client) UpdateIPAddress(ctx context.Context, id string, ip *domain.IPAddress) (*domain.IPAddress, error) {
	var result domain.IPAddress
	err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/api/v1/ips/%s", id), ip, &result)
	return &result, err
}

func (c *Client) DeleteIPAddress(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/ips/%s", id), nil, nil)
}

// IP assignment methods
func (c *Client) AssignIP(ctx context.Context, assignment *domain.ComputeIP) (*domain.ComputeIP, error) {
	var result domain.ComputeIP
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/ip-assignments", assignment, &result)
	return &result, err
}

func (c *Client) UnassignIP(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/ip-assignments/%s", id), nil, nil)
}

func (c *Client) ListIPAssignments(ctx context.Context, computeID, ipID string) ([]*domain.ComputeIP, error) {
	var assignments []*domain.ComputeIP
	path := "/api/v1/ip-assignments"
	if computeID != "" {
		path += "?compute_id=" + computeID
	} else if ipID != "" {
		path += "?ip_id=" + ipID
	}
	err := c.doRequest(ctx, http.MethodGet, path, nil, &assignments)
	return assignments, err
}

// DNS record methods
func (c *Client) ListDNSRecords(ctx context.Context, filters storage.DNSRecordFilters) ([]*domain.DNSRecord, error) {
	var records []*domain.DNSRecord
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/dns", nil, &records)
	return records, err
}

func (c *Client) GetDNSRecord(ctx context.Context, id string) (*domain.DNSRecord, error) {
	var record domain.DNSRecord
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/dns/%s", id), nil, &record)
	return &record, err
}

func (c *Client) CreateDNSRecord(ctx context.Context, record *domain.DNSRecord) (*domain.DNSRecord, error) {
	var result domain.DNSRecord
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/dns", record, &result)
	return &result, err
}

func (c *Client) UpdateDNSRecord(ctx context.Context, id string, record *domain.DNSRecord) (*domain.DNSRecord, error) {
	var result domain.DNSRecord
	err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/api/v1/dns/%s", id), record, &result)
	return &result, err
}

func (c *Client) DeleteDNSRecord(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/dns/%s", id), nil, nil)
}

// Port assignment methods
func (c *Client) ListPortAssignments(ctx context.Context, filters storage.PortAssignmentFilters) ([]*domain.PortAssignment, error) {
	url := "/api/v1/ports?"
	params := []string{}
	if filters.AssignmentID != "" {
		params = append(params, "assignment_id="+filters.AssignmentID)
	}
	if filters.IPID != "" {
		params = append(params, "ip_id="+filters.IPID)
	}
	if filters.Protocol != "" {
		params = append(params, "protocol="+filters.Protocol)
	}
	if len(params) > 0 {
		url += strings.Join(params, "&")
	}

	var assignments []*domain.PortAssignment
	err := c.doRequest(ctx, http.MethodGet, url, nil, &assignments)
	return assignments, err
}

func (c *Client) GetPortAssignment(ctx context.Context, id string) (*domain.PortAssignment, error) {
	var assignment domain.PortAssignment
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/ports/%s", id), nil, &assignment)
	return &assignment, err
}

func (c *Client) CreatePortAssignment(ctx context.Context, assignment *domain.PortAssignment) (*domain.PortAssignment, error) {
	var result domain.PortAssignment
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/ports", assignment, &result)
	return &result, err
}

func (c *Client) UpdatePortAssignment(ctx context.Context, id string, assignment *domain.PortAssignment) (*domain.PortAssignment, error) {
	var result domain.PortAssignment
	err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/api/v1/ports/%s", id), assignment, &result)
	return &result, err
}

func (c *Client) DeletePortAssignment(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/ports/%s", id), nil, nil)
}

// Firewall rule methods
func (c *Client) ListFirewallRules(ctx context.Context, filters storage.FirewallRuleFilters) ([]*domain.FirewallRule, error) {
	var rules []*domain.FirewallRule
	err := c.doRequest(ctx, http.MethodGet, "/api/v1/firewall-rules", nil, &rules)
	return rules, err
}

func (c *Client) GetFirewallRule(ctx context.Context, id string) (*domain.FirewallRule, error) {
	var rule domain.FirewallRule
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/firewall-rules/%s", id), nil, &rule)
	return &rule, err
}

func (c *Client) CreateFirewallRule(ctx context.Context, rule *domain.FirewallRule) (*domain.FirewallRule, error) {
	var result domain.FirewallRule
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/firewall-rules", rule, &result)
	return &result, err
}

func (c *Client) UpdateFirewallRule(ctx context.Context, id string, rule *domain.FirewallRule) (*domain.FirewallRule, error) {
	var result domain.FirewallRule
	err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/api/v1/firewall-rules/%s", id), rule, &result)
	return &result, err
}

func (c *Client) DeleteFirewallRule(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/firewall-rules/%s", id), nil, nil)
}

// Firewall assignment methods
func (c *Client) ListComputeFirewallRules(ctx context.Context, computeID, ruleID string) ([]*domain.ComputeFirewallRule, error) {
	var assignments []*domain.ComputeFirewallRule
	path := "/api/v1/firewall-assignments"
	if computeID != "" {
		path += "?compute_id=" + computeID
	} else if ruleID != "" {
		path += "?rule_id=" + ruleID
	}
	err := c.doRequest(ctx, http.MethodGet, path, nil, &assignments)
	return assignments, err
}

func (c *Client) AssignFirewallRule(ctx context.Context, assignment *domain.ComputeFirewallRule) (*domain.ComputeFirewallRule, error) {
	var result domain.ComputeFirewallRule
	err := c.doRequest(ctx, http.MethodPost, "/api/v1/firewall-assignments", assignment, &result)
	return &result, err
}

func (c *Client) UnassignFirewallRule(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/firewall-assignments/%s", id), nil, nil)
}
