package supabase

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
	apiBase = "https://api.supabase.com/v1"
)

type Plugin struct {
	client *http.Client
}

func NewPlugin() *Plugin {
	return &Plugin{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *Plugin) Name() string        { return "Supabase" }
func (p *Plugin) Description() string { return "Open source Firebase alternative with Postgres database" }
func (p *Plugin) Icon() string        { return "supabase" }

func (p *Plugin) AuthConfig() core.AuthConfig {
	return core.AuthConfig{
		Fields: []core.AuthField{
			{
				ID:          "access_token",
				Name:        "Access Token",
				Description: "Supabase access token from Account > Access Tokens",
				Type:        "password",
				Required:    true,
			},
		},
	}
}

func (p *Plugin) ResourceCategories() []core.ResourceCategory {
	return []core.ResourceCategory{
		{ID: "database_size", Name: "数据库大小", Unit: "MB", Limit: 500},
		{ID: "bandwidth", Name: "带宽", Unit: "GB", Limit: 5},
		{ID: "storage", Name: "文件存储", Unit: "MB", Limit: 1024},
	}
}

func (p *Plugin) makeRequest(ctx context.Context, token, path string, result interface{}) error {
	url := apiBase + path
	fmt.Printf("[Supabase] Making request to: %s\n", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("[Supabase] Response status: %d\n", resp.StatusCode)

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid access token")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if len(body) > 500 {
		fmt.Printf("[Supabase] Response body (first 500 chars): %s...\n", string(body[:500]))
	} else {
		fmt.Printf("[Supabase] Response body: %s\n", string(body))
	}

	return json.Unmarshal(body, result)
}

// SupabaseProject represents a project from Supabase Management API
type SupabaseProject struct {
	ID              string `json:"id"`
	Ref             string `json:"ref"`
	Name            string `json:"name"`
	Region          string `json:"region"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`
	OrganizationID  string `json:"organization_id"`
	OrganizationSlug string `json:"organization_slug"`
	Database        struct {
		Host           string `json:"host"`
		Version        string `json:"version"`
		PostgresEngine string `json:"postgres_engine"`
		ReleaseChannel string `json:"release_channel"`
	} `json:"database"`
}

// SupabaseUsageResponse represents API usage counts
type SupabaseUsageResponse struct {
	Result []struct {
		Timestamp            string `json:"timestamp"`
		TotalAuthRequests    int    `json:"total_auth_requests"`
		TotalRealtimeRequests int   `json:"total_realtime_requests"`
		TotalRestRequests    int    `json:"total_rest_requests"`
		TotalStorageRequests int    `json:"total_storage_requests"`
	} `json:"result"`
}

func (p *Plugin) ValidateCredentials(ctx context.Context, credentials map[string]string) error {
	accessToken, ok := credentials["access_token"]
	if !ok || accessToken == "" {
		return fmt.Errorf("access token is required")
	}

	fmt.Printf("[Supabase] Validating access token...\n")

	var projects []SupabaseProject
	if err := p.makeRequest(ctx, accessToken, "/projects", &projects); err != nil {
		return fmt.Errorf("failed to validate access token: %w", err)
	}

	fmt.Printf("[Supabase] Access token valid, found %d projects\n", len(projects))
	return nil
}

func (p *Plugin) FetchUsage(ctx context.Context, credentials map[string]string) (*core.UsageData, error) {
	accessToken := credentials["access_token"]
	fmt.Printf("[Supabase] FetchUsage called\n")

	now := time.Now()
	resetDate := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())

	// First get projects to find refs
	var projects []SupabaseProject
	if err := p.makeRequest(ctx, accessToken, "/projects", &projects); err != nil {
		fmt.Printf("[Supabase] Failed to fetch projects for usage: %v\n", err)
		return p.buildDefaultUsage(now, resetDate), nil
	}

	if len(projects) == 0 {
		fmt.Printf("[Supabase] No projects found, returning default usage\n")
		return p.buildDefaultUsage(now, resetDate), nil
	}

	// Get usage for the first project (primary)
	ref := projects[0].Ref
	accountName := projects[0].Name
	plan := "Free Plan"

	usagePath := fmt.Sprintf("/projects/%s/analytics/endpoints/usage.api-counts?interval=1d", ref)
	var usageResp SupabaseUsageResponse
	if err := p.makeRequest(ctx, accessToken, usagePath, &usageResp); err != nil {
		fmt.Printf("[Supabase] Failed to fetch usage: %v, using defaults\n", err)
		return p.buildDefaultUsageWithName(now, resetDate, accountName, plan), nil
	}

	// Aggregate total requests
	var totalRequests float64
	for _, r := range usageResp.Result {
		totalRequests += float64(r.TotalAuthRequests + r.TotalRealtimeRequests + r.TotalRestRequests + r.TotalStorageRequests)
	}

	resources := []core.ResourceUsage{
		{
			CategoryID: "database_size",
			Name:       "数据库大小",
			Used:       0, // Not available via this endpoint
			Limit:      500,
			Unit:       "MB",
			ResetDate:  &resetDate,
		},
		{
			CategoryID: "bandwidth",
			Name:       "带宽",
			Used:       math.Round(totalRequests/1000*100) / 100, // Approximate GB from requests
			Limit:      5,
			Unit:       "GB",
			ResetDate:  &resetDate,
		},
		{
			CategoryID: "storage",
			Name:       "文件存储",
			Used:       0,
			Limit:      1024,
			Unit:       "MB",
			ResetDate:  &resetDate,
		},
	}

	status := core.StatusOK
	for _, res := range resources {
		if res.Limit > 0 && res.Used > 0 {
			pct := res.Used / res.Limit
			if pct > 0.9 {
				status = core.StatusDanger
			} else if pct > 0.8 && status != core.StatusDanger {
				status = core.StatusWarn
			}
		}
	}

	fmt.Printf("[Supabase] FetchUsage completed: totalRequests=%.0f\n", totalRequests)

	return &core.UsageData{
		Platform:    "Supabase",
		AccountName: accountName,
		Plan:        plan,
		Resources:   resources,
		LastUpdated: now,
		Status:      status,
	}, nil
}

func (p *Plugin) buildDefaultUsage(now, resetDate time.Time) *core.UsageData {
	return p.buildDefaultUsageWithName(now, resetDate, "Personal", "Free Plan")
}

func (p *Plugin) buildDefaultUsageWithName(now, resetDate time.Time, name, plan string) *core.UsageData {
	return &core.UsageData{
		Platform:    "Supabase",
		AccountName: name,
		Plan:        plan,
		Resources: []core.ResourceUsage{
			{CategoryID: "database_size", Name: "数据库大小", Used: 0, Limit: 500, Unit: "MB", ResetDate: &resetDate},
			{CategoryID: "bandwidth", Name: "带宽", Used: 0, Limit: 5, Unit: "GB", ResetDate: &resetDate},
			{CategoryID: "storage", Name: "文件存储", Used: 0, Limit: 1024, Unit: "MB", ResetDate: &resetDate},
		},
		LastUpdated: now,
		Status:      core.StatusOK,
	}
}

func (p *Plugin) FetchProjects(ctx context.Context, credentials map[string]string) ([]core.ProjectInfo, error) {
	accessToken := credentials["access_token"]
	fmt.Printf("[Supabase] FetchProjects called\n")

	var projects []SupabaseProject
	if err := p.makeRequest(ctx, accessToken, "/projects", &projects); err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}

	fmt.Printf("[Supabase] Found %d projects\n", len(projects))

	result := make([]core.ProjectInfo, len(projects))
	for i, proj := range projects {
		var createdAt time.Time
		if proj.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339Nano, proj.CreatedAt); err == nil {
				createdAt = t
			}
		}
		result[i] = core.ProjectInfo{
			ID:        proj.ID,
			Name:      proj.Name,
			URL:       fmt.Sprintf("https://supabase.com/dashboard/project/%s", proj.Ref),
			CreatedAt: createdAt,
		}
	}

	return result, nil
}
