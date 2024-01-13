package ecs

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

func UpdateRevision() *cobra.Command {
	var cluster string
	var service string
	var task string
	var revision string
	var force bool = false
	cmd := &cobra.Command{
		Use:     "revision",
		Aliases: []string{"rvs", "revisions"},
		Short:   "Update task definition revision",
		Run: func(cmd *cobra.Command, args []string) {
			service = args[0]
			revision = args[1]
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			client := ecs.New(sess)

			input := &ecs.DescribeServicesInput{
				Services: []*string{
					aws.String(service),
				},
				Cluster: aws.String(cluster),
			}

			response, err := client.DescribeServices(input)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

			task = aws.StringValue(response.Services[0].TaskDefinition)

			lastColonIndex := strings.LastIndex(task, ":")

			family := task[:lastColonIndex]

			UpdateServiceTaskDefinition(service, cluster, family, revision, force)

		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force new deployment(default false)")
	cmd.Flags().StringVarP(&cluster, "cluster", "c", "string", "ECS cluster name")
	cmd.MarkFlagRequired("cluster")

	return cmd
}

func UpdateServiceTaskDefinition(service string, cluster string, task string, revision string, force bool) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := ecs.New(sess)

	_, err := client.UpdateService(&ecs.UpdateServiceInput{
		Cluster:              aws.String(cluster),
		Service:              aws.String(service),
		EnableExecuteCommand: aws.Bool(true),
		ForceNewDeployment:   aws.Bool(force),
		TaskDefinition:       aws.String(fmt.Sprintf("%s:%s", task, revision)),
	})

	if err != nil {
		fmt.Println("Failed to update ECS service:", err)
		os.Exit(0)
	}

	fmt.Printf("Service %s updated\n", service)

}
