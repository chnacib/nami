package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// Config holds all configuration for the application
type Config struct {
	// AWS configuration
	AwsConfig aws.Config
	// Any other configuration your app needs
	// e.g., LogLevel string
}

// LoadConfig loads configuration from environment variables and/or config files
func LoadConfig() (*Config, error) {
	ctx := context.Background()

	// Load AWS configuration from environment variables or ~/.aws/credentials
	awsConfig, err := config.LoadDefaultConfig(
		ctx,
		config.WithDefaultRegion("us-east-1"), // Default region, can be overridden by AWS_REGION env var
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Override region from environment variable if provided
	if region := os.Getenv("AWS_REGION"); region != "" {
		awsConfig.Region = region
	}

	// Look for AWS profiles and use if specified
	if profile := os.Getenv("AWS_PROFILE"); profile != "" {
		// Load config with specific profile
		profileConfig, err := config.LoadDefaultConfig(
			ctx,
			config.WithSharedConfigProfile(profile),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS profile '%s': %w", profile, err)
		}
		awsConfig = profileConfig
	}

	// You might also load app-specific configuration from a file
	// For example, from ~/.nami/config.yaml
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine user home directory: %w", err)
	}

	configPath := filepath.Join(userHomeDir, ".nami", "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		// Config file exists, you could parse it here
		// For example with:
		// configData, err := os.ReadFile(configPath)
		// if err != nil {
		//     return nil, fmt.Errorf("cannot read config file: %w", err)
		// }
		// Parse YAML, JSON, etc.
	}

	return &Config{
		AwsConfig: awsConfig,
		// Set other config values
	}, nil
}
