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

func ServiceLogs() *cobra.Command {
	var cluster string

	cmd := &cobra.Command{
		Use:     "service",
		Aliases: []string{"svc", "services"},
		Short:   "Get ECS service log events",
		Run: func(cmd *cobra.Command, args []string) {
			service := args[0]
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))
			client := ecs.New(sess)
			client_cw := cloudwatchlogs.New(sess)

			input_list_tasks := &ecs.ListTasksInput{
				ServiceName: aws.String(service),
				Cluster:     aws.String(cluster),
			}

			list_tasks, err := client.ListTasks(input_list_tasks)
			if err != nil {
				fmt.Println(err)
			}

			for _, tasks := range list_tasks.TaskArns {
				task := NameArn(*tasks)
				fmt.Println(task)
				input := &ecs.DescribeTasksInput{
					Tasks: []*string{
						aws.String(*tasks),
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
					Limit:         aws.Int64(100),
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

			}

		},
	}

	cmd.Flags().StringVarP(&cluster, "cluster", "c", "string", "ECS Cluster name")
	cmd.MarkFlagRequired("cluster")

	return cmd

}
