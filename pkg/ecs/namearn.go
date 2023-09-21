package ecs

import (
	"fmt"
	"regexp"
	"strings"
)

func NameArn(arn string) string {
	arnParts := strings.Split(arn, "/")
	serviceName := arnParts[len(arnParts)-1]
	return serviceName
}

func extractLoadBalancerID(arn string) string {
	regex := regexp.MustCompile(`^arn:aws:elasticloadbalancing:[^:]+:\d+:loadbalancer/([^/]+)/([^/]+)/([^/]+)$`)
	match := regex.FindStringSubmatch(arn)

	if len(match) == 4 {
		return fmt.Sprintf("%s/%s/%s", match[1], match[2], match[3])
	}

	return ""
}

func extractTargetGroupID(arn string) string {
	regex := regexp.MustCompile(`^arn:aws:elasticloadbalancing:[^:]+:\d+:targetgroup/([^/]+)/([^/]+)$`)
	match := regex.FindStringSubmatch(arn)

	if len(match) == 3 {
		return fmt.Sprintf("targetgroup/%s/%s", match[1], match[2])
	}

	return ""
}
