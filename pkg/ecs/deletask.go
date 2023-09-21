package ecs

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

func DeleteTask() *cobra.Command {
	var cluster string
	var task string
	cmd := &cobra.Command{
		Use:     "task",
		Aliases: []string{"tsk", "tasks"},
		Short:   "Stop ECS running task",
		Run: func(cmd *cobra.Command, args []string) {
			task = args[0]
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			client := ecs.New(sess)

			input := &ecs.StopTaskInput{
				Task:    aws.String(task),
				Cluster: aws.String(cluster),
				Reason:  aws.String("Stopped by nami"),
			}

			_, err := client.StopTask(input)

			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

			fmt.Printf("Task %s stopped\n", task)

		},
	}
	cmd.Flags().StringVarP(&cluster, "cluster", "c", "string", "ECS task name")
	cmd.MarkFlagRequired("cluster")

	return cmd
}
