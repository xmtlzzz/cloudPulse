package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"cloudpulse/core"
)

const (
	apiBase = "https://api.cloudflare.com/client/v4"
)

type Plugin struct {
	client *http.Client
}

func NewPlugin() *Plugin {
	return &Plugin{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *Plugin) Name() string        { return "Cloudflare" }
func (p *Plugin) Description() string { return "CDN, DNS, and edge computing platform" }
func (p *Plugin) Icon() string        { return "cloudflare" }

func (p *Plugin) AuthConfig() core.AuthConfig {
	return core.AuthConfig{
		Fields: []core.AuthField{
			{
				ID:          "api_token",
				Name:        "API Token",
				Description: "Cloudflare API token from My Profile > API Tokens",
				Type:        "password",
				Required:    true,
			},
			{
				ID:          "account_id",
				Name:        "Account ID",
				Description: "Cloudflare Account ID from the dashboard URL",
				Type:        "text",
				Required:    true,
			},
		},
	}
}

func (p *Plugin) ResourceCategories() []core.ResourceCategory {
	return []core.ResourceCategory{
		{ID: "workers_requests", Name: "Workers 请求", Unit: "次", Limit: 100000},
		{ID: "r2_storage", Name: "R2 存储", Unit: "GB", Limit: 10},
		{ID: "d1_database", Name: "D1 数据库", Unit: "GB", Limit: 5},
		{ID: "pages_deployments", Name: "Pages 部署", Unit: "次", Limit: 500},
	}
}

// CloudflareResponse wraps the standard Cloudflare API response
type CloudflareResponse struct {
	Success bool            `json:"success"`
	Errors  []interface{}   `json:"errors"`
	Result  json.RawMessage `json:"result"`
}

// CloudflarePagesProject represents a Pages project
type CloudflarePagesProject struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Subdomain     string `json:"subdomain"`
	CreatedOn     string `json:"created_on"`
	ProductionDeployment struct {
		URL string `json:"url"`
	} `json:"production_deployment"`
}

// CloudflareR2Bucket represents an R2 bucket
type CloudflareR2Bucket struct {
	Name         string `json:"name"`
	CreationDate string `json:"creation_date"`
	StorageClass string `json:"storage_class"`
}

type CloudflareR2ListResponse struct {
	Buckets []CloudflareR2Bucket `json:"buckets"`
}

// CloudflareD1Database represents a D1 database
type CloudflareD1Database struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func (p *Plugin) makeRequest(ctx context.Context, token, path string, result interface{}) error {
	url := apiBase + path
	fmt.Printf("[Cloudflare] Making request to: %s\n", url)

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

	fmt.Printf("[Cloudflare] Response status: %d\n", resp.StatusCode)

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("invalid API token or insufficient permissions")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if len(body) > 500 {
		fmt.Printf("[Cloudflare] Response body (first 500 chars): %s...\n", string(body[:500]))
	} else {
		fmt.Printf("[Cloudflare] Response body: %s\n", string(body))
	}

	// For standard Cloudflare wrapped responses, unwrap the result
	if result != nil {
		var cfResp CloudflareResponse
		if err := json.Unmarshal(body, &cfResp); err == nil && cfResp.Success {
			if len(cfResp.Result) > 0 {
				return json.Unmarshal(cfResp.Result, result)
			}
			return nil
		}
		// If not a wrapped response, try direct unmarshal
		return json.Unmarshal(body, result)
	}

	return nil
}

func (p *Plugin) ValidateCredentials(ctx context.Context, credentials map[string]string) error {
	apiToken, ok := credentials["api_token"]
	if !ok || apiToken == "" {
		return fmt.Errorf("API token is required")
	}
	accountId, ok := credentials["account_id"]
	if !ok || accountId == "" {
		return fmt.Errorf("Account ID is required")
	}

	fmt.Printf("[Cloudflare] Validating credentials...\n")

	// Validate by fetching account details
	var account struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := p.makeRequest(ctx, apiToken, "/accounts/"+accountId, &account); err != nil {
		return fmt.Errorf("failed to validate credentials: %w", err)
	}

	fmt.Printf("[Cloudflare] Credentials valid, account: %s\n", account.Name)
	return nil
}

func (p *Plugin) FetchUsage(ctx context.Context, credentials map[string]string) (*core.UsageData, error) {
	apiToken := credentials["api_token"]
	accountId := credentials["account_id"]
	fmt.Printf("[Cloudflare] FetchUsage called\n")

	now := time.Now()
	resetDate := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())

	// Fetch R2 buckets count as a usage indicator
	var r2Resp CloudflareR2ListResponse
	r2Path := fmt.Sprintf("/accounts/%s/r2/buckets", accountId)
	r2Err := p.makeRequest(ctx, apiToken, r2Path, &r2Resp)
	r2Count := 0
	if r2Err == nil {
		r2Count = len(r2Resp.Buckets)
	} else {
		fmt.Printf("[Cloudflare] R2 fetch failed: %v\n", r2Err)
	}

	// Fetch D1 databases
	var d1Dbs []CloudflareD1Database
	d1Path := fmt.Sprintf("/accounts/%s/d1/database", accountId)
	d1Err := p.makeRequest(ctx, apiToken, d1Path, &d1Dbs)
	d1Count := 0
	if d1Err == nil {
		d1Count = len(d1Dbs)
	} else {
		fmt.Printf("[Cloudflare] D1 fetch failed: %v\n", d1Err)
	}

	resources := []core.ResourceUsage{
		{
			CategoryID: "r2_storage",
			Name:       "R2 存储",
			Used:       float64(r2Count),
			Limit:      10,
			Unit:       "GB",
			ResetDate:  &resetDate,
		},
		{
			CategoryID: "d1_database",
			Name:       "D1 数据库",
			Used:       float64(d1Count),
			Limit:      5,
			Unit:       "GB",
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

	fmt.Printf("[Cloudflare] FetchUsage completed: R2 buckets=%d, D1 databases=%d\n", r2Count, d1Count)

	return &core.UsageData{
		Platform:    "Cloudflare",
		AccountName: accountId,
		Plan:        "Free Plan",
		Resources:   resources,
		LastUpdated: now,
		Status:      status,
	}, nil
}

func (p *Plugin) FetchProjects(ctx context.Context, credentials map[string]string) ([]core.ProjectInfo, error) {
	apiToken := credentials["api_token"]
	accountId := credentials["account_id"]
	fmt.Printf("[Cloudflare] FetchProjects called\n")

	var projects []core.ProjectInfo

	// Fetch Pages projects
	var pagesProjects []CloudflarePagesProject
	pagesPath := fmt.Sprintf("/accounts/%s/pages/projects", accountId)
	if err := p.makeRequest(ctx, apiToken, pagesPath, &pagesProjects); err != nil {
		fmt.Printf("[Cloudflare] Failed to fetch Pages projects: %v\n", err)
	} else {
		for _, proj := range pagesProjects {
			url := proj.ProductionDeployment.URL
			if url == "" {
				url = fmt.Sprintf("https://%s.pages.dev", proj.Name)
			}
			projects = append(projects, core.ProjectInfo{
				ID:   proj.ID,
				Name: proj.Name,
				URL:  url,
			})
		}
	}

	// Fetch R2 buckets
	var r2Resp CloudflareR2ListResponse
	r2Path := fmt.Sprintf("/accounts/%s/r2/buckets", accountId)
	if err := p.makeRequest(ctx, apiToken, r2Path, &r2Resp); err != nil {
		fmt.Printf("[Cloudflare] Failed to fetch R2 buckets: %v\n", err)
	} else {
		for _, bucket := range r2Resp.Buckets {
			projects = append(projects, core.ProjectInfo{
				ID:   bucket.Name,
				Name: bucket.Name,
				URL:  fmt.Sprintf("https://%s.r2.dev", bucket.Name),
			})
		}
	}

	// Fetch D1 databases
	var d1Dbs []CloudflareD1Database
	d1Path := fmt.Sprintf("/accounts/%s/d1/database", accountId)
	if err := p.makeRequest(ctx, apiToken, d1Path, &d1Dbs); err != nil {
		fmt.Printf("[Cloudflare] Failed to fetch D1 databases: %v\n", err)
	} else {
		for _, db := range d1Dbs {
			projects = append(projects, core.ProjectInfo{
				ID:   db.UUID,
				Name: db.Name,
				URL:  fmt.Sprintf("https://dash.cloudflare.com/%s/workers/d1/database/%s", accountId, db.UUID),
			})
		}
	}

	fmt.Printf("[Cloudflare] Found %d projects total\n", len(projects))
	return projects, nil
}
