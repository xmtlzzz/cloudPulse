package custom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cloudpulse/core"
)

// CustomServiceConfig represents the configuration for a custom service
type CustomServiceConfig struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	BaseURL     string            `json:"baseUrl"`
	Headers     map[string]string `json:"headers"`
	AuthType    string            `json:"authType"` // "bearer", "header", "query"
	AuthKey     string            `json:"authKey"`  // Header name or query param name
	Resources   []ResourceMapping `json:"resources"`
}

// ResourceMapping maps API response fields to resource usage
type ResourceMapping struct {
	Name       string `json:"name"`
	Path       string `json:"path"`       // JSON path to the value (e.g., "data.bandwidth.used")
	LimitPath  string `json:"limitPath"`  // JSON path to the limit value
	Unit       string `json:"unit"`
	DefaultLimit float64 `json:"defaultLimit"`
}

// Plugin implements a custom platform plugin
type Plugin struct {
	config CustomServiceConfig
	client *http.Client
}

// NewPlugin creates a new custom plugin instance
func NewPlugin(config CustomServiceConfig) *Plugin {
	return &Plugin{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *Plugin) Name() string {
	return p.config.Name
}

func (p *Plugin) Description() string {
	if p.config.Description != "" {
		return p.config.Description
	}
	return "Custom service integration"
}

func (p *Plugin) Icon() string {
	return "custom"
}

func (p *Plugin) AuthConfig() core.AuthConfig {
	return core.AuthConfig{
		Fields: []core.AuthField{
			{
				ID:          "api_key",
				Name:        "API Key",
				Description: "API key or token for authentication",
				Type:        "password",
				Required:    true,
			},
		},
	}
}

func (p *Plugin) ResourceCategories() []core.ResourceCategory {
	categories := make([]core.ResourceCategory, len(p.config.Resources))
	for i, r := range p.config.Resources {
		categories[i] = core.ResourceCategory{
			ID:    strings.ToLower(strings.ReplaceAll(r.Name, " ", "_")),
			Name:  r.Name,
			Unit:  r.Unit,
			Limit: r.DefaultLimit,
		}
	}
	return categories
}

func (p *Plugin) ValidateCredentials(ctx context.Context, credentials map[string]string) error {
	apiKey, ok := credentials["api_key"]
	if !ok || apiKey == "" {
		return fmt.Errorf("API key is required")
	}

	// Try to make a test request
	req, err := http.NewRequestWithContext(ctx, "GET", p.config.BaseURL, nil)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	p.setAuthHeaders(req, apiKey)
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("authentication failed (status %d)", resp.StatusCode)
	}

	return nil
}

func (p *Plugin) setAuthHeaders(req *http.Request, apiKey string) {
	// Set custom headers
	for k, v := range p.config.Headers {
		req.Header.Set(k, v)
	}

	// Set auth based on auth type
	switch p.config.AuthType {
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+apiKey)
	case "header":
		if p.config.AuthKey != "" {
			req.Header.Set(p.config.AuthKey, apiKey)
		}
	case "query":
		if p.config.AuthKey != "" {
			q := req.URL.Query()
			q.Set(p.config.AuthKey, apiKey)
			req.URL.RawQuery = q.Encode()
		}
	default:
		// Default to bearer token
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
}

func (p *Plugin) FetchUsage(ctx context.Context, credentials map[string]string) (*core.UsageData, error) {
	apiKey := credentials["api_key"]

	req, err := http.NewRequestWithContext(ctx, "GET", p.config.BaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	p.setAuthHeaders(req, apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract resources from response
	resources := make([]core.ResourceUsage, 0, len(p.config.Resources))
	for _, mapping := range p.config.Resources {
		used := extractValue(data, mapping.Path)
		limit := extractValue(data, mapping.LimitPath)
		if limit == 0 {
			limit = mapping.DefaultLimit
		}

		resources = append(resources, core.ResourceUsage{
			CategoryID: strings.ToLower(strings.ReplaceAll(mapping.Name, " ", "_")),
			Name:       mapping.Name,
			Used:       used,
			Limit:      limit,
			Unit:       mapping.Unit,
		})
	}

	// Determine status
	status := core.StatusOK
	for _, res := range resources {
		if res.Limit > 0 {
			pct := res.Used / res.Limit
			if pct > 0.9 {
				status = core.StatusDanger
			} else if pct > 0.8 && status != core.StatusDanger {
				status = core.StatusWarn
			}
		}
	}

	return &core.UsageData{
		Platform:    p.config.Name,
		AccountName: "Custom",
		Plan:        "Custom Plan",
		Resources:   resources,
		LastUpdated: time.Now(),
		Status:      status,
	}, nil
}

func (p *Plugin) FetchProjects(ctx context.Context, credentials map[string]string) ([]core.ProjectInfo, error) {
	// Custom services don't have a standard projects endpoint
	return nil, nil
}

// extractValue extracts a value from a nested JSON structure using a dot-separated path
func extractValue(data interface{}, path string) float64 {
	if path == "" {
		return 0
	}

	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part]
			if !ok {
				return 0
			}
			current = val
		default:
			return 0
		}
	}

	// Try to convert to float64
	switch v := current.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		// Try to parse string as number
		var f float64
		fmt.Sscanf(v, "%f", &f)
		return f
	default:
		return 0
	}
}
