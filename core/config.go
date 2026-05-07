package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// ConfigManager manages application configuration and credentials
type ConfigManager struct {
	configPath string
	config     *AppConfig
	mu         sync.RWMutex
}

// AppConfig holds all application configuration
type AppConfig struct {
	Credentials []PlatformCredentials `json:"credentials"`
	Settings    AppSettings           `json:"settings"`
}

// AppSettings holds user preferences
type AppSettings struct {
	Theme          string `json:"theme"`          // "light", "dark", "auto"
	SyncInterval   int    `json:"syncInterval"`   // in minutes
	AlertThreshold int    `json:"alertThreshold"` // percentage (e.g., 80)
}

// NewConfigManager creates a new config manager
func NewConfigManager() (*ConfigManager, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	cm := &ConfigManager{
		configPath: filepath.Join(configDir, "config.json"),
		config: &AppConfig{
			Credentials: []PlatformCredentials{},
			Settings: AppSettings{
				Theme:          "auto",
				SyncInterval:   15,
				AlertThreshold: 80,
			},
		},
	}

	// Load existing config if available
	if err := cm.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return cm, nil
}

func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".cloudpulse"), nil
}

func (cm *ConfigManager) load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, cm.config)
}

func (cm *ConfigManager) save() error {
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cm.configPath, data, 0644)
}

// GetCredentials returns credentials for a specific platform
func (cm *ConfigManager) GetCredentials(platform string) []PlatformCredentials {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var result []PlatformCredentials
	for _, cred := range cm.config.Credentials {
		if cred.Platform == platform {
			result = append(result, cred)
		}
	}
	return result
}

// GetAllCredentials returns all stored credentials
func (cm *ConfigManager) GetAllCredentials() []PlatformCredentials {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.Credentials
}

// AddCredential adds or updates credentials for a platform
func (cm *ConfigManager) AddCredential(cred PlatformCredentials) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if credential already exists for this platform/account
	for i, existing := range cm.config.Credentials {
		if existing.Platform == cred.Platform && existing.AccountName == cred.AccountName {
			cm.config.Credentials[i] = cred
			return cm.save()
		}
	}

	cm.config.Credentials = append(cm.config.Credentials, cred)
	return cm.save()
}

// RemoveCredential removes credentials for a platform/account
func (cm *ConfigManager) RemoveCredential(platform, accountName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for i, cred := range cm.config.Credentials {
		if cred.Platform == platform && cred.AccountName == accountName {
			cm.config.Credentials = append(cm.config.Credentials[:i], cm.config.Credentials[i+1:]...)
			return cm.save()
		}
	}

	return nil
}

// GetSettings returns the current application settings
func (cm *ConfigManager) GetSettings() AppSettings {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.Settings
}

// UpdateSettings updates the application settings
func (cm *ConfigManager) UpdateSettings(settings AppSettings) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config.Settings = settings
	return cm.save()
}
