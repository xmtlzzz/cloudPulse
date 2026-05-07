package neon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"cloudpulse/core"
)

const (
	apiBase = "https://console.neon.tech/api/v2"
)

type Plugin struct {
	client *http.Client
}

func NewPlugin() *Plugin {
	return &Plugin{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *Plugin) Name() string        { return "Neon" }
func (p *Plugin) Description() string { return "Serverless Postgres database platform" }
func (p *Plugin) Icon() string        { return "neon" }

func (p *Plugin) AuthConfig() core.AuthConfig {
	return core.AuthConfig{
		Fields: []core.AuthField{
			{
				ID:          "api_key",
				Name:        "API Key",
				Description: "Neon API key from Account Settings > API Keys",
				Type:        "password",
				Required:    true,
			},
			{
				ID:          "org_id",
				Name:        "Organization ID",
				Description: "Organization ID (required for Personal API Key, e.g. org-xxx). Leave empty for Org API Key.",
				Type:        "text",
				Required:    false,
			},
		},
	}
}

func (p *Plugin) orgQueryParam(credentials map[string]string) string {
	if orgID, ok := credentials["org_id"]; ok && orgID != "" {
		return "?org_id=" + orgID
	}
	return ""
}

func (p *Plugin) ResourceCategories() []core.ResourceCategory {
	return []core.ResourceCategory{
		{ID: "storage", Name: "存储", Unit: "MB", Limit: 512},
		{ID: "compute_time", Name: "计算时间", Unit: "h", Limit: 191.9},
		{ID: "branches", Name: "分支数", Unit: "个", Limit: 10},
	}
}

func (p *Plugin) makeRequest(ctx context.Context, apiKey, path string, result interface{}) error {
	url := apiBase + path
	fmt.Printf("[Neon] Making request to: %s\n", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("[Neon] Response status: %d\n", resp.StatusCode)

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API key (if using Personal API Key, make sure to provide Organization ID)")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if len(body) > 500 {
		fmt.Printf("[Neon] Response body (first 500 chars): %s...\n", string(body[:500]))
	} else {
		fmt.Printf("[Neon] Response body: %s\n", string(body))
	}

	return json.Unmarshal(body, result)
}

// NeonProject represents a project from Neon API
type NeonProject struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
	CPUUsedSec           int64  `json:"cpu_used_sec"`
	SyntheticStorageSize int64  `json:"synthetic_storage_size"`
	ActiveTime           int64  `json:"active_time"`
}

type NeonProjectsResponse struct {
	Projects []NeonProject `json:"projects"`
}

func (p *Plugin) ValidateCredentials(ctx context.Context, credentials map[string]string) error {
	apiKey, ok := credentials["api_key"]
	if !ok || apiKey == "" {
		return fmt.Errorf("API key is required")
	}

	fmt.Printf("[Neon] Validating API key...\n")

	var resp NeonProjectsResponse
	path := "/projects" + p.orgQueryParam(credentials)
	if err := p.makeRequest(ctx, apiKey, path, &resp); err != nil {
		return fmt.Errorf("failed to validate API key: %w", err)
	}

	fmt.Printf("[Neon] API key valid, found %d projects\n", len(resp.Projects))
	return nil
}

func (p *Plugin) FetchUsage(ctx context.Context, credentials map[string]string) (*core.UsageData, error) {
	apiKey := credentials["api_key"]
	fmt.Printf("[Neon] FetchUsage called\n")

	now := time.Now()
	resetDate := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())

	// Use /projects endpoint which includes usage data (cpu_used_sec, synthetic_storage_size, active_time)
	var resp NeonProjectsResponse
	path := "/projects" + p.orgQueryParam(credentials)
	if err := p.makeRequest(ctx, apiKey, path, &resp); err != nil {
		return nil, fmt.Errorf("failed to fetch projects for usage: %w", err)
	}

	if len(resp.Projects) == 0 {
		return p.buildDefaultUsage(now, resetDate), nil
	}

	// Aggregate usage across all projects
	var totalCPUSec int64
	var totalStorageBytes int64
	accountName := resp.Projects[0].Name
	for _, proj := range resp.Projects {
		totalCPUSec += proj.CPUUsedSec
		totalStorageBytes += proj.SyntheticStorageSize
	}

	computeHours := math.Round(float64(totalCPUSec)/3600*100) / 100
	storageMB := math.Round(float64(totalStorageBytes)/(1024*1024)*100) / 100

	resources := []core.ResourceUsage{
		{
			CategoryID: "storage",
			Name:       "存储",
			Used:       storageMB,
			Limit:      512,
			Unit:       "MB",
			ResetDate:  &resetDate,
		},
		{
			CategoryID: "compute_time",
			Name:       "计算时间",
			Used:       computeHours,
			Limit:      191.9,
			Unit:       "h",
			ResetDate:  &resetDate,
		},
	}

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

	fmt.Printf("[Neon] FetchUsage completed: storage=%.1fMB, compute=%.1fh\n", storageMB, computeHours)

	return &core.UsageData{
		Platform:    "Neon",
		AccountName: accountName,
		Plan:        "Free Plan",
		Resources:   resources,
		LastUpdated: now,
		Status:      status,
	}, nil
}

func (p *Plugin) buildDefaultUsage(now, resetDate time.Time) *core.UsageData {
	return &core.UsageData{
		Platform:    "Neon",
		AccountName: "Personal",
		Plan:        "Free Plan",
		Resources: []core.ResourceUsage{
			{CategoryID: "storage", Name: "存储", Used: 0, Limit: 512, Unit: "MB", ResetDate: &resetDate},
			{CategoryID: "compute_time", Name: "计算时间", Used: 0, Limit: 191.9, Unit: "h", ResetDate: &resetDate},
		},
		LastUpdated: now,
		Status:      core.StatusOK,
	}
}

func (p *Plugin) FetchProjects(ctx context.Context, credentials map[string]string) ([]core.ProjectInfo, error) {
	apiKey := credentials["api_key"]
	fmt.Printf("[Neon] FetchProjects called\n")

	var resp NeonProjectsResponse
	path := "/projects" + p.orgQueryParam(credentials)
	if err := p.makeRequest(ctx, apiKey, path, &resp); err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}

	fmt.Printf("[Neon] Found %d projects\n", len(resp.Projects))

	projects := make([]core.ProjectInfo, len(resp.Projects))
	for i, proj := range resp.Projects {
		var createdAt time.Time
		if proj.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, proj.CreatedAt); err == nil {
				createdAt = t
			}
		}
		projects[i] = core.ProjectInfo{
			ID:        proj.ID,
			Name:      proj.Name,
			URL:       fmt.Sprintf("https://console.neon.tech/app/projects/%s", proj.ID),
			CreatedAt: createdAt,
		}
	}

	return projects, nil
}
