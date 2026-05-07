package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"cloudpulse/core"
	"cloudpulse/plugins/cloudflare"
	"cloudpulse/plugins/custom"
	"cloudpulse/plugins/neon"
	"cloudpulse/plugins/supabase"
	"cloudpulse/plugins/vercel"
)

// App struct - main application
type App struct {
	ctx           context.Context
	pluginManager *core.PluginManager
	configManager *core.ConfigManager
	cacheManager  *core.CacheManager
	mu            sync.RWMutex
	usageCache    map[string]*core.UsageData
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		pluginManager: core.NewPluginManager(),
		usageCache:    make(map[string]*core.UsageData),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize config manager
	configManager, err := core.NewConfigManager()
	if err != nil {
		fmt.Printf("Error initializing config manager: %v\n", err)
	} else {
		a.configManager = configManager
	}

	// Initialize cache manager
	cacheManager, err := core.NewCacheManager()
	if err != nil {
		fmt.Printf("Error initializing cache manager: %v\n", err)
	} else {
		a.cacheManager = cacheManager
	}

	// Register plugins
	a.registerPlugins()
}

func (a *App) registerPlugins() {
	a.pluginManager.Register(vercel.NewPlugin())
	a.pluginManager.Register(neon.NewPlugin())
	a.pluginManager.Register(supabase.NewPlugin())
	a.pluginManager.Register(cloudflare.NewPlugin())

	// Load custom services
	a.loadCustomServices()
}

func (a *App) loadCustomServices() {
	configDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	customServicesFile := filepath.Join(configDir, ".cloudpulse", "custom_services.json")
	data, err := os.ReadFile(customServicesFile)
	if err != nil {
		return
	}

	var configs []custom.CustomServiceConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return
	}

	for _, config := range configs {
		a.pluginManager.Register(custom.NewPlugin(config))
	}
}

func (a *App) saveCustomServices(configs []custom.CustomServiceConfig) error {
	configDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(configDir, ".cloudpulse")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "custom_services.json"), data, 0644)
}

func (a *App) getCustomServiceConfigs() []custom.CustomServiceConfig {
	configDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	customServicesFile := filepath.Join(configDir, ".cloudpulse", "custom_services.json")
	data, err := os.ReadFile(customServicesFile)
	if err != nil {
		return nil
	}

	var configs []custom.CustomServiceConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil
	}

	return configs
}

// GetPlatformList returns list of available platforms
func (a *App) GetPlatformList() []string {
	return a.pluginManager.List()
}

// GetAuthConfig returns auth config for a platform
func (a *App) GetAuthConfig(platform string) (*core.AuthConfig, error) {
	plugin, err := a.pluginManager.Get(platform)
	if err != nil {
		return nil, err
	}
	config := plugin.AuthConfig()
	return &config, nil
}

// AddCredential adds credentials for a platform
func (a *App) AddCredential(platform, accountName string, credentials map[string]string) error {
	if a.configManager == nil {
		return fmt.Errorf("config manager not initialized")
	}

	// Validate credentials
	plugin, err := a.pluginManager.Get(platform)
	if err != nil {
		return err
	}

	if err := plugin.ValidateCredentials(a.ctx, credentials); err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	cred := core.PlatformCredentials{
		Platform:    platform,
		AccountName: accountName,
		Credentials: credentials,
	}

	return a.configManager.AddCredential(cred)
}

// RemoveCredential removes credentials for a platform
func (a *App) RemoveCredential(platform, accountName string) error {
	if a.configManager == nil {
		return fmt.Errorf("config manager not initialized")
	}

	// Also clear cache
	if a.cacheManager != nil {
		a.cacheManager.Delete(platform, accountName)
	}

	return a.configManager.RemoveCredential(platform, accountName)
}

// GetAllCredentials returns all stored credentials (without sensitive data)
func (a *App) GetAllCredentials() []map[string]string {
	if a.configManager == nil {
		return nil
	}

	creds := a.configManager.GetAllCredentials()
	result := make([]map[string]string, len(creds))
	for i, cred := range creds {
		result[i] = map[string]string{
			"platform":    cred.Platform,
			"accountName": cred.AccountName,
		}
	}
	return result
}

// FetchUsage fetches usage data for a specific platform/account
func (a *App) FetchUsage(platform, accountName string) (*core.UsageData, error) {
	fmt.Printf("[App] FetchUsage called for %s/%s\n", platform, accountName)

	plugin, err := a.pluginManager.Get(platform)
	if err != nil {
		fmt.Printf("[App] Plugin not found: %v\n", err)
		return nil, err
	}

	// Check cache first (skip cache for now to debug)
	// if a.cacheManager != nil {
	// 	if cached := a.cacheManager.Get(platform, accountName); cached != nil {
	// 		fmt.Printf("[App] Returning cached data for %s/%s\n", platform, accountName)
	// 		return cached, nil
	// 	}
	// }

	// Get credentials
	creds := a.configManager.GetCredentials(platform)
	fmt.Printf("[App] Found %d credentials for %s\n", len(creds), platform)

	var cred *core.PlatformCredentials
	for _, c := range creds {
		fmt.Printf("[App] Checking credential: %s\n", c.AccountName)
		if c.AccountName == accountName {
			cred = &c
			break
		}
	}

	if cred == nil {
		fmt.Printf("[App] No credentials found for %s/%s\n", platform, accountName)
		return nil, fmt.Errorf("no credentials found for %s/%s", platform, accountName)
	}

	fmt.Printf("[App] Fetching from API for %s/%s\n", platform, accountName)

	// Fetch from API
	data, err := plugin.FetchUsage(a.ctx, cred.Credentials)
	if err != nil {
		fmt.Printf("[App] API error for %s: %v\n", platform, err)
		return nil, err
	}

	fmt.Printf("[App] Successfully fetched data for %s: %s\n", platform, data.AccountName)

	// Cache the result
	if a.cacheManager != nil {
		settings := a.configManager.GetSettings()
		ttl := time.Duration(settings.SyncInterval) * time.Minute
		a.cacheManager.Set(platform, accountName, data, ttl)
	}

	return data, nil
}

// FetchAllUsage fetches usage data for all configured platforms
func (a *App) FetchAllUsage() (*core.DashboardData, error) {
	fmt.Printf("[App] FetchAllUsage called\n")

	if a.configManager == nil {
		return nil, fmt.Errorf("config manager not initialized")
	}

	allCreds := a.configManager.GetAllCredentials()
	fmt.Printf("[App] Found %d credentials to fetch\n", len(allCreds))

	var services []core.UsageData
	var alerts []core.AlertInfo
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, cred := range allCreds {
		fmt.Printf("[App] Starting fetch for %s/%s\n", cred.Platform, cred.AccountName)
		wg.Add(1)
		go func(c core.PlatformCredentials) {
			defer wg.Done()
			fmt.Printf("[App] Goroutine started for %s/%s\n", c.Platform, c.AccountName)
			data, err := a.FetchUsage(c.Platform, c.AccountName)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				fmt.Printf("[App] Error fetching %s: %v\n", c.Platform, err)
				services = append(services, core.UsageData{
					Platform:    c.Platform,
					AccountName: c.AccountName,
					Status:      core.StatusError,
					Error:       err.Error(),
					LastUpdated: time.Now(),
				})
			} else {
				fmt.Printf("[App] Success fetching %s: %+v\n", c.Platform, data)
				services = append(services, *data)
				// Check for alerts
				for _, res := range data.Resources {
					if res.Limit > 0 {
						pct := (res.Used / res.Limit) * 100
						settings := a.configManager.GetSettings()
						if pct >= float64(settings.AlertThreshold) {
							status := "warn"
							if pct >= 95 {
								status = "critical"
							} else if pct >= 90 {
								status = "danger"
							}
							alerts = append(alerts, core.AlertInfo{
								Platform:   c.Platform,
								Resource:   res.Name,
								Percentage: pct,
								Time:       time.Now(),
								Status:     status,
							})
						}
					}
				}
			}
		}(cred)
	}

	wg.Wait()
	fmt.Printf("[App] All fetches completed. Services: %d\n", len(services))

	// Calculate summary
	connectedCount := len(services)
	healthyCount := 0
	warnCount := 0
	for _, s := range services {
		if s.Status == core.StatusOK {
			healthyCount++
		} else if s.Status == core.StatusWarn {
			warnCount++
		}
	}

	return &core.DashboardData{
		Services:       services,
		Alerts:         alerts,
		LastSync:       time.Now(),
		ConnectedCount: connectedCount,
		HealthyCount:   healthyCount,
		WarnCount:      warnCount,
	}, nil
}

// GetSettings returns current app settings
func (a *App) GetSettings() core.AppSettings {
	if a.configManager == nil {
		return core.AppSettings{
			Theme:          "auto",
			SyncInterval:   15,
			AlertThreshold: 80,
		}
	}
	return a.configManager.GetSettings()
}

// UpdateSettings updates app settings
func (a *App) UpdateSettings(settings core.AppSettings) error {
	if a.configManager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	return a.configManager.UpdateSettings(settings)
}

// FetchAllProjects fetches projects from all configured platforms
func (a *App) FetchAllProjects() (map[string][]core.ProjectInfo, error) {
	if a.configManager == nil {
		return nil, fmt.Errorf("config manager not initialized")
	}

	allCreds := a.configManager.GetAllCredentials()
	result := make(map[string][]core.ProjectInfo)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, cred := range allCreds {
		wg.Add(1)
		go func(c core.PlatformCredentials) {
			defer wg.Done()
			plugin, err := a.pluginManager.Get(c.Platform)
			if err != nil {
				return
			}

			projects, err := plugin.FetchProjects(a.ctx, c.Credentials)
			if err != nil {
				fmt.Printf("[App] Failed to fetch projects for %s: %v\n", c.Platform, err)
				return
			}

			mu.Lock()
			result[c.Platform] = projects
			mu.Unlock()
		}(cred)
	}

	wg.Wait()
	return result, nil
}

// GetCustomServices returns list of custom service configurations
func (a *App) GetCustomServices() []custom.CustomServiceConfig {
	return a.getCustomServiceConfigs()
}

// AddCustomService adds a new custom service configuration
func (a *App) AddCustomService(config custom.CustomServiceConfig) error {
	configs := a.getCustomServiceConfigs()

	// Check if service with same name already exists
	for _, c := range configs {
		if c.Name == config.Name {
			return fmt.Errorf("service with name '%s' already exists", config.Name)
		}
	}

	configs = append(configs, config)
	if err := a.saveCustomServices(configs); err != nil {
		return err
	}

	// Register the new plugin
	a.pluginManager.Register(custom.NewPlugin(config))
	return nil
}

// UpdateCustomService updates an existing custom service configuration
func (a *App) UpdateCustomService(name string, config custom.CustomServiceConfig) error {
	configs := a.getCustomServiceConfigs()

	found := false
	for i, c := range configs {
		if c.Name == name {
			configs[i] = config
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("service '%s' not found", name)
	}

	return a.saveCustomServices(configs)
}

// RemoveCustomService removes a custom service configuration
func (a *App) RemoveCustomService(name string) error {
	configs := a.getCustomServiceConfigs()

	newConfigs := make([]custom.CustomServiceConfig, 0, len(configs))
	found := false
	for _, c := range configs {
		if c.Name == name {
			found = true
			continue
		}
		newConfigs = append(newConfigs, c)
	}

	if !found {
		return fmt.Errorf("service '%s' not found", name)
	}

	return a.saveCustomServices(newConfigs)
}

// GetAvailablePlatforms returns all available platforms (built-in + custom)
func (a *App) GetAvailablePlatforms() []map[string]string {
	platforms := []map[string]string{
		{"name": "Vercel", "icon": "vercel", "description": "前端部署平台，提供 Serverless Functions 和 Edge Functions"},
		{"name": "Neon", "icon": "neon", "description": "无服务器 Postgres 数据库平台"},
		{"name": "Supabase", "icon": "supabase", "description": "开源 Firebase 替代方案，提供数据库、认证、存储等"},
		{"name": "Cloudflare", "icon": "cloudflare", "description": "CDN、DNS 和边缘计算平台，提供 Workers、R2、D1 等"},
	}

	// Add custom services
	customConfigs := a.getCustomServiceConfigs()
	for _, config := range customConfigs {
		platforms = append(platforms, map[string]string{
			"name":        config.Name,
			"icon":        "custom",
			"description": config.Description,
		})
	}

	return platforms
}

// TestConnection tests the connection to a platform API
func (a *App) TestConnection(platform string) (map[string]interface{}, error) {
	fmt.Printf("[App] TestConnection called for platform: %s\n", platform)

	plugin, err := a.pluginManager.Get(platform)
	if err != nil {
		return nil, fmt.Errorf("plugin not found: %s", platform)
	}

	// Get credentials for this platform
	creds := a.configManager.GetCredentials(platform)
	if len(creds) == 0 {
		return nil, fmt.Errorf("no credentials found for %s", platform)
	}

	cred := creds[0]
	fmt.Printf("[App] Testing connection for %s/%s\n", platform, cred.AccountName)

	// Test validation
	if err := plugin.ValidateCredentials(a.ctx, cred.Credentials); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	// Try to fetch usage
	data, err := plugin.FetchUsage(a.ctx, cred.Credentials)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"success":    true,
		"account":    data.AccountName,
		"plan":       data.Plan,
		"resources":  len(data.Resources),
		"lastUpdate": data.LastUpdated,
	}, nil
}
