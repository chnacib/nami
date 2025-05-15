package ecs

import (
	"context"
	"encoding/json"
	"fmt"

	"os"

	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/ecs"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/chnacib/nami/pkg/config"

	"github.com/spf13/cobra"
)

func DescribeTask() *cobra.Command {

	var cluster string

	cmd := &cobra.Command{

		Use: "task [task-id]",

		Aliases: []string{"td"},

		Short: "Describe ECS task",

		Args: cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {

			taskID := args[0]

			if strings.TrimSpace(taskID) == "" {

				fmt.Fprintf(os.Stderr, "Error: Task ID cannot be empty\n")

				os.Exit(0)

			}

			cfg, err := config.LoadConfig()

			if err != nil {

				fmt.Errorf("failed to load configuration: %w", err)

			}

			client := ecs.NewFromConfig(cfg.AwsConfig)

			input := &ecs.DescribeTasksInput{
				Tasks:   []string{taskID},
				Cluster: aws.String(cluster),
			}

			response, err := client.DescribeTasks(context.Background(), input)

			if err != nil {

				handleAWSError(err, taskID, cluster)

				os.Exit(0)

			}

			if len(response.Tasks) == 0 {

				fmt.Fprintf(os.Stderr, "Error: No task with ID '%s' found in cluster '%s'\n", taskID, cluster)

				os.Exit(0)

			}

			data, err := json.MarshalIndent(response.Tasks[0], "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting response: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(data))

		},
	}

	cmd.Flags().StringVarP(&cluster, "cluster", "c", "", "ECS Cluster name (required)")

	cmd.MarkFlagRequired("cluster")

	return cmd

}

func handleAWSError(err error, taskID, clusterName string) {

	if awsErr, ok := err.(awserr.Error); ok {

		switch awsErr.Code() {

		case "ClusterNotFoundException":

			fmt.Fprintf(os.Stderr, "Error: Cluster '%s' not found\n", clusterName)

		case "InvalidParameterException":

			fmt.Fprintf(os.Stderr, "Error: Invalid parameter: %s\n", awsErr.Message())

		case "AccessDeniedException":

			fmt.Fprintf(os.Stderr, "Error: Access denied: %s\n", awsErr.Message())

		default:

			fmt.Fprintf(os.Stderr, "Error: AWS Error (%s): %s\n", awsErr.Code(), awsErr.Message())

		}

	} else {

		fmt.Fprintf(os.Stderr, "Error: Failed to describe task '%s': %v\n", taskID, err)

	}

}
