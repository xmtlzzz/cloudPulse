package core

import "context"

// PlatformPlugin defines the interface that all platform plugins must implement
type PlatformPlugin interface {
	// Name returns the platform name (e.g., "Vercel", "Neon")
	Name() string

	// Description returns a short description of the platform
	Description() string

	// Icon returns the icon identifier for the platform
	Icon() string

	// AuthConfig returns the authentication configuration for this platform
	AuthConfig() AuthConfig

	// FetchUsage fetches usage data from the platform API
	FetchUsage(ctx context.Context, credentials map[string]string) (*UsageData, error)

	// FetchProjects fetches the list of projects on this platform
	FetchProjects(ctx context.Context, credentials map[string]string) ([]ProjectInfo, error)

	// ResourceCategories returns the list of resource categories this platform provides
	ResourceCategories() []ResourceCategory

	// ValidateCredentials checks if the provided credentials are valid
	ValidateCredentials(ctx context.Context, credentials map[string]string) error
}
