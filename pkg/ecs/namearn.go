package ecs

import (
	"strings"
)

func NameArn(arn string) string {
	arnParts := strings.Split(arn, "/")
	serviceName := arnParts[len(arnParts)-1]
	return serviceName
}
