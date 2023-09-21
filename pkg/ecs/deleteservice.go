package ecs

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

func DeleteService() *cobra.Command {
	var cluster string
	var service string
	var force bool = false
	cmd := &cobra.Command{
		Use:     "service",
		Aliases: []string{"svc", "services"},
		Short:   "Delete ECS service",
		Run: func(cmd *cobra.Command, args []string) {
			service = args[0]
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			client := ecs.New(sess)

			input := &ecs.DeleteServiceInput{
				Service: aws.String(service),
				Cluster: aws.String(cluster),
				Force:   aws.Bool(force),
			}

			_, err := client.DeleteService(input)

			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

			fmt.Printf("Service %s deleted\n", service)

		},
	}
	cmd.Flags().StringVarP(&cluster, "cluster", "c", "string", "ECS task name")
	cmd.MarkFlagRequired("cluster")
	cmd.Flags().BoolVarP(&force, "force", "f", true, "Force delete service")

	return cmd
}
