package utils

import (
	"strings"
)

// GetResourceName extracts the resource name from an AWS ARN
// Examples:
// - arn:aws:ecs:us-east-1:123456789012:cluster/my-cluster -> my-cluster
// - arn:aws:ecs:us-west-2:123456789012:service/cluster-name/service-name -> service-name
// - arn:aws:ecs:us-east-1:123456789012:task-definition/task-def-name:1 -> task-def-name:1
func GetResourceName(arn string) string {
	// Simple ARN handling for common resources
	parts := strings.Split(arn, "/")
	if len(parts) < 2 {
		return arn // Return original string if not in expected format
	}

	// Handle different types of ARNs
	if strings.Contains(arn, "task-definition") {
		// Task definition ARNs include the version after a colon
		return parts[len(parts)-1]
	} else if strings.Contains(arn, "service") && len(parts) > 2 {
		// Service ARNs in newer format include the cluster name
		// Format: arn:aws:ecs:region:account:service/cluster-name/service-name
		return parts[len(parts)-1]
	} else {
		// Handle cluster ARNs and other simple resource ARNs
		return parts[len(parts)-1]
	}
}

// ParseTaskFamily extracts just the family name from a task definition ARN or name
// Example:
// - family:revision -> family
// - arn:aws:ecs:region:account:task-definition/family:revision -> family
func ParseTaskFamily(taskDef string) string {
	// First extract the task definition name from the ARN if needed
	name := GetResourceName(taskDef)

	// Then split by colon to remove the revision
	parts := strings.Split(name, ":")
	return parts[0]
}

// IsValidClusterName checks if a string is a valid ECS cluster name or ARN
func IsValidClusterName(cluster string) bool {
	// If it's an ARN, it should contain the expected format
	if strings.HasPrefix(cluster, "arn:aws:ecs:") && strings.Contains(cluster, ":cluster/") {
		return true
	}

	// If it's a name, it should follow ECS naming rules
	// (up to 255 letters, numbers, hyphens, and underscores)
	if len(cluster) > 255 {
		return false
	}

	for _, char := range cluster {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}

	return true
}

// FormatServiceName ensures a service name is in the expected format
func FormatServiceName(serviceName string) string {
	// Remove any trailing whitespace
	return strings.TrimSpace(serviceName)
}
