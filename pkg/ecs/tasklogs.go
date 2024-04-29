package ecs

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

type Message struct {
	Timestamp string
	Message   string
}

func TaskLogs() *cobra.Command {
	var cluster string
	var limit int64
	var GetQueryResultOutput *cloudwatchlogs.GetQueryResultsOutput
	var logs []Message

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
			if container == "xray-daemon" {
				container = aws.StringValue(response.Tasks[0].Containers[1].Name)
			}
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

			query_string := fmt.Sprintf("filter @logStream = '%s' | fields @timestamp, @message | sort @timestamp desc | limit %d", log_stream, limit)

			StartQueryInput := &cloudwatchlogs.StartQueryInput{
				QueryString:  aws.String(query_string),
				LogGroupName: aws.String(log_group),
				EndTime:      aws.Int64(time.Now().UnixNano() / int64(time.Millisecond)),
				StartTime:    aws.Int64(5),
			}
			StartQueryOutput, err := client_cw.StartQuery(StartQueryInput)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
			query_id := aws.String(*StartQueryOutput.QueryId)

			GetQueryResultInput := &cloudwatchlogs.GetQueryResultsInput{
				QueryId: aws.String(*query_id),
			}

			for {
				GetQueryResultOutput, err = client_cw.GetQueryResults(GetQueryResultInput)
				if err != nil {
					fmt.Println(err)
					os.Exit(0)
				}
				status := aws.String(*GetQueryResultOutput.Status)
				if *status == "Complete" {
					break
				}
				time.Sleep(1 * time.Second)
			}

			for _, result := range GetQueryResultOutput.Results {
				timestamp := aws.String(*result[0].Value)
				message := aws.String(*result[1].Value)
				logs = append(logs, Message{
					Timestamp: *timestamp,
					Message:   *message,
				})

			}

			sort.SliceStable(logs, func(i, j int) bool {
				t1, _ := time.Parse(time.RFC3339Nano, logs[i].Timestamp)
				t2, _ := time.Parse(time.RFC3339Nano, logs[j].Timestamp)
				return t1.After(t2)
			})

			for i := len(logs) - 1; i >= 0; i-- {
				fmt.Printf("%s    %s\n", logs[i].Timestamp, logs[i].Message)
			}

		},
	}

	cmd.Flags().Int64VarP(&limit, "limit", "l", 200, "Logs max result")
	cmd.Flags().StringVarP(&cluster, "cluster", "c", "string", "ECS Cluster name")
	cmd.MarkFlagRequired("cluster")

	return cmd

}
