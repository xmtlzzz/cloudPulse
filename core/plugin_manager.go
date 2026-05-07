package core

import (
	"fmt"
	"sync"
)

// PluginManager manages all registered platform plugins
type PluginManager struct {
	plugins map[string]PlatformPlugin
	mu      sync.RWMutex
}

// NewPluginManager creates a new plugin manager
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]PlatformPlugin),
	}
}

// Register adds a plugin to the manager
func (pm *PluginManager) Register(plugin PlatformPlugin) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.plugins[plugin.Name()] = plugin
}

// Get returns a plugin by name
func (pm *PluginManager) Get(name string) (PlatformPlugin, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	plugin, ok := pm.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}
	return plugin, nil
}

// List returns all registered plugin names
func (pm *PluginManager) List() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	names := make([]string, 0, len(pm.plugins))
	for name := range pm.plugins {
		names = append(names, name)
	}
	return names
}

// GetAll returns all registered plugins
func (pm *PluginManager) GetAll() []PlatformPlugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	plugins := make([]PlatformPlugin, 0, len(pm.plugins))
	for _, p := range pm.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}
