package vercel

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

const (
	apiBase = "https://api.vercel.com"
)

// Plugin implements the Vercel platform plugin
type Plugin struct {
	client *http.Client
}

// NewPlugin creates a new Vercel plugin instance
func NewPlugin() *Plugin {
	return &Plugin{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *Plugin) Name() string {
	return "Vercel"
}

func (p *Plugin) Description() string {
	return "Frontend cloud platform for static sites and serverless functions"
}

func (p *Plugin) Icon() string {
	return "vercel"
}

func (p *Plugin) AuthConfig() core.AuthConfig {
	return core.AuthConfig{
		Fields: []core.AuthField{
			{
				ID:          "token",
				Name:        "API Token",
				Description: "Vercel API token from Account Settings > Tokens",
				Type:        "password",
				Required:    true,
			},
		},
	}
}

func (p *Plugin) ResourceCategories() []core.ResourceCategory {
	return []core.ResourceCategory{
		{ID: "bandwidth", Name: "带宽", Unit: "GB", Limit: 100},
		{ID: "builds", Name: "构建次数", Unit: "次", Limit: 6000},
		{ID: "serverless", Name: "Serverless 执行", Unit: "次", Limit: 100000},
		{ID: "edge_functions", Name: "Edge Function Invocations", Unit: "次", Limit: 1000000},
		{ID: "image_optimization", Name: "Image Optimization", Unit: "次", Limit: 5000},
	}
}

func (p *Plugin) ValidateCredentials(ctx context.Context, credentials map[string]string) error {
	token, ok := credentials["token"]
	if !ok || token == "" {
		return fmt.Errorf("API token is required")
	}

	fmt.Printf("[Vercel] Validating token...\n")

	// Validate by fetching user info
	var user VercelUser
	if err := p.makeRequest(ctx, token, "/v2/user", &user); err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}

	// Also check if token can list projects (warn if not)
	var resp VercelProjectsResponse
	projectsPath := "/v9/projects"
	if user.User.DefaultTeamId != "" {
		projectsPath += "?teamId=" + user.User.DefaultTeamId
	}
	if err := p.makeRequest(ctx, token, projectsPath, &resp); err != nil {
		fmt.Printf("[Vercel] Warning: cannot list projects (token may lack project:read scope): %v\n", err)
	} else if len(resp.Projects) == 0 {
		fmt.Printf("[Vercel] Warning: 0 projects found. If you have projects, recreate token with Full Account scope at https://vercel.com/account/tokens\n")
	}

	return nil
}

// VercelUser represents the user response from Vercel API
type VercelUser struct {
	User struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		Username      string `json:"username"`
		DefaultTeamId string `json:"defaultTeamId"`
	} `json:"user"`
}

// VercelUsageResponse represents the usage response from Vercel API
type VercelUsageResponse struct {
	Commercial struct {
		Bandwidth struct {
			Total float64 `json:"total"`
			Limit float64 `json:"limit"`
		} `json:"bandwidth"`
		Builds struct {
			Total float64 `json:"total"`
			Limit float64 `json:"limit"`
		} `json:"builds"`
		ServerlessFunctionExecution struct {
			Total float64 `json:"total"`
			Limit float64 `json:"limit"`
		} `json:"serverlessFunctionExecution"`
		EdgeFunctionInvocations struct {
			Total float64 `json:"total"`
			Limit float64 `json:"limit"`
		} `json:"edgeFunctionInvocations"`
		SourceImages struct {
			Total float64 `json:"total"`
			Limit float64 `json:"limit"`
		} `json:"sourceImages"`
	} `json:"commercial"`
}

// VercelProject represents a project from Vercel API
type VercelProject struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	UpdatedAt int64  `json:"updatedAt"` // Unix timestamp in milliseconds
	Link      struct {
		Type string `json:"type"`
		Repo string `json:"repo"`
	} `json:"link,omitempty"`
	LatestDeployments []struct {
		UID   string   `json:"uid"`
		URL   string   `json:"url"`
		Alias []string `json:"alias,omitempty"`
	} `json:"latestDeployments,omitempty"`
}

type VercelProjectsResponse struct {
	Projects []VercelProject `json:"projects"`
}

func (p *Plugin) makeRequest(ctx context.Context, token, path string, result interface{}) error {
	url := apiBase + path
	fmt.Printf("[Vercel] Making request to: %s\n", url)

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

	fmt.Printf("[Vercel] Response status: %d\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Print first 500 chars of response for debugging
	if len(body) > 500 {
		fmt.Printf("[Vercel] Response body (first 500 chars): %s...\n", string(body[:500]))
	} else {
		fmt.Printf("[Vercel] Response body: %s\n", string(body))
	}

	return json.Unmarshal(body, result)
}

func (p *Plugin) FetchUsage(ctx context.Context, credentials map[string]string) (*core.UsageData, error) {
	token := credentials["token"]
	fmt.Printf("[Vercel] FetchUsage called\n")

	// Get user info for account name
	var user VercelUser
	if err := p.makeRequest(ctx, token, "/v2/user", &user); err != nil {
		fmt.Printf("[Vercel] Failed to get user info: %v\n", err)
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Use username if name is empty
	accountName := user.User.Name
	if accountName == "" {
		accountName = user.User.Username
	}
	fmt.Printf("[Vercel] User: %s (%s)\n", accountName, user.User.Email)

	// Try to get usage data
	now := time.Now()
	resetDate := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())

	// Note: Vercel usage API (/v2/usage) might not be available for Hobby plan
	// The API returns "invalid_from_date" errors for all date formats
	// This is a known limitation of the Vercel API for free tier users
	fmt.Printf("[Vercel] Note: Usage API may not be available for Hobby plan\n")

	// Build resources with default limits (usage API not available for Hobby plan)
	resources := []core.ResourceUsage{}

	// Usage API not available, use default limits
	fmt.Printf("[Vercel] Using default limits for Hobby plan\n")
	resources = append(resources, core.ResourceUsage{
		CategoryID: "bandwidth",
		Name:       "带宽",
		Used:       0,
		Limit:      100,
		Unit:       "GB",
		ResetDate:  &resetDate,
	})
	resources = append(resources, core.ResourceUsage{
		CategoryID: "builds",
		Name:       "构建次数",
		Used:       0,
		Limit:      6000,
		Unit:       "次",
		ResetDate:  &resetDate,
	})
	resources = append(resources, core.ResourceUsage{
		CategoryID: "serverless",
		Name:       "Serverless 执行",
		Used:       0,
		Limit:      100000,
		Unit:       "次",
		ResetDate:  &resetDate,
	})

	// Calculate status based on usage
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

	fmt.Printf("[Vercel] FetchUsage completed successfully with %d resources\n", len(resources))

	return &core.UsageData{
		Platform:    "Vercel",
		AccountName: accountName,
		Plan:        "Hobby Plan",
		Resources:   resources,
		LastUpdated: now,
		Status:      status,
	}, nil
}

func (p *Plugin) FetchProjects(ctx context.Context, credentials map[string]string) ([]core.ProjectInfo, error) {
	token := credentials["token"]
	fmt.Printf("[Vercel] FetchProjects called\n")

	// Get user info for teamId
	var user VercelUser
	if err := p.makeRequest(ctx, token, "/v2/user", &user); err != nil {
		fmt.Printf("[Vercel] Failed to get user info for projects: %v\n", err)
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	projectsPath := "/v9/projects"
	if user.User.DefaultTeamId != "" {
		projectsPath += "?teamId=" + user.User.DefaultTeamId
		fmt.Printf("[Vercel] Using teamId: %s\n", user.User.DefaultTeamId)
	}

	var resp VercelProjectsResponse
	if err := p.makeRequest(ctx, token, projectsPath, &resp); err != nil {
		fmt.Printf("[Vercel] Failed to fetch projects: %v\n", err)
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}

	if len(resp.Projects) == 0 {
		fmt.Printf("[Vercel] No projects found. Token may lack project:read scope. Recreate token at https://vercel.com/account/tokens\n")
	}

	fmt.Printf("[Vercel] Found %d projects\n", len(resp.Projects))

	projects := make([]core.ProjectInfo, len(resp.Projects))
	for i, proj := range resp.Projects {
		updatedAt := time.UnixMilli(proj.UpdatedAt)

		// Use alias URL (e.g. v-invoice-mocha.vercel.app) if available, otherwise deployment URL, otherwise synthesized
		projectURL := fmt.Sprintf("https://%s.vercel.app", proj.Name)
		if len(proj.LatestDeployments) > 0 {
			dep := proj.LatestDeployments[0]
			// Prefer alias (clean URL like v-invoice-mocha.vercel.app)
			if len(dep.Alias) > 0 {
				projectURL = "https://" + dep.Alias[0]
			} else if dep.URL != "" {
				url := dep.URL
				if !strings.HasPrefix(url, "http") {
					url = "https://" + url
				}
				projectURL = url
			}
		}

		projects[i] = core.ProjectInfo{
			ID:        proj.ID,
			Name:      proj.Name,
			URL:       projectURL,
			CreatedAt: updatedAt,
		}
	}

	return projects, nil
}
