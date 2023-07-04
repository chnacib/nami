package ecs

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

func DescribeService() *cobra.Command {
	var cluster string
	cmd := &cobra.Command{
		Use:     "service",
		Aliases: []string{"svc", "services"},
		Short:   "Describe ECS services",
		Run: func(cmd *cobra.Command, args []string) {
			service := args[0]
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

			fmt.Println(response)

		},
	}
	cmd.Flags().StringVarP(&cluster, "cluster", "c", "string", "ECS Cluster name")
	cmd.MarkFlagRequired("cluster")

	return cmd
}
