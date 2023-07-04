package ecs

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

func DescribeTask() *cobra.Command {
	var cluster string
	cmd := &cobra.Command{
		Use:     "task",
		Aliases: []string{"td"},
		Short:   "Describe ECS task",
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))
			client := ecs.New(sess)

			input := &ecs.DescribeTasksInput{
				Tasks: []*string{
					aws.String(name),
				},
				Cluster: aws.String(cluster),
			}

			response, err := client.DescribeTasks(input)
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
