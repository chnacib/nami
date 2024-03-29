package ecs

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

func TaskLogs() *cobra.Command {
	var cluster string
	var limit int64

	cmd := &cobra.Command{
		Use:     "task",
		Aliases: []string{"tsk", "tasks"},
		Short:   "Get ECS task log events",
		Run: func(cmd *cobra.Command, args []string) {
			task := args[0]
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))
			client := ecs.New(sess)
			client_cw := cloudwatchlogs.New(sess)
			input := &ecs.DescribeTasksInput{
				Tasks: []*string{
					aws.String(task),
				},
				Cluster: aws.String(cluster),
			}

			response, err := client.DescribeTasks(input)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
			container := aws.StringValue(response.Tasks[0].Containers[0].Name)

			task_def := aws.StringValue(response.Tasks[0].TaskDefinitionArn)

			task_input := &ecs.DescribeTaskDefinitionInput{
				TaskDefinition: aws.String(task_def),
			}
			result, err_task := client.DescribeTaskDefinition(task_input)
			if err != nil {
				fmt.Println(err_task)
				os.Exit(0)
			}
			log_group := aws.StringValue(result.TaskDefinition.ContainerDefinitions[0].LogConfiguration.Options["awslogs-group"])
			log_prefix := aws.StringValue(result.TaskDefinition.ContainerDefinitions[0].LogConfiguration.Options["awslogs-stream-prefix"])
			log_stream := fmt.Sprintf("%s/%s/%s", log_prefix, container, task)

			log_input := &cloudwatchlogs.GetLogEventsInput{
				LogGroupName:  aws.String(log_group),
				LogStreamName: aws.String(log_stream),
				Limit:         aws.Int64(limit),
			}
			output, err := client_cw.GetLogEvents(log_input)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
			for _, outputs := range output.Events {
				message := *outputs.Message
				fmt.Println(message)
			}

		},
	}

	cmd.Flags().Int64VarP(&limit, "limit", "l", 100, "Logs max result")
	cmd.Flags().StringVarP(&cluster, "cluster", "c", "string", "ECS Cluster name")
	cmd.MarkFlagRequired("cluster")

	return cmd

}
