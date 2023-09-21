package ecs

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

func DeleteCluster() *cobra.Command {
	var cluster string
	cmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"clusters"},
		Short:   "Delete ECS Cluster",
		Run: func(cmd *cobra.Command, args []string) {
			cluster = args[0]
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			client := ecs.New(sess)

			input := &ecs.DeleteClusterInput{
				Cluster: aws.String(cluster),
			}

			_, err := client.DeleteCluster(input)

			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

			fmt.Printf("Cluster %s deleted\n", cluster)

		},
	}

	return cmd
}
