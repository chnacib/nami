package ecs

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

func DescribeCluster() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Describe ECS Cluster",
		Run: func(cmd *cobra.Command, args []string) {
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			cluster := args[0]

			client := ecs.New(sess)

			input := &ecs.DescribeClustersInput{
				Clusters: []*string{
					aws.String(cluster),
				}}

			response, err := client.DescribeClusters(input)
			if err != nil {
				fmt.Println("Failed to describe ECS cluster: Cluster not found")
				os.Exit(0)
			}

			fmt.Println(response)

		},
	}

	return cmd
}
