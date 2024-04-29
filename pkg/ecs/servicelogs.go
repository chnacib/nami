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

func ServiceLogs() *cobra.Command {
	var cluster string
	var limit int64
	var GetQueryResultOutput *cloudwatchlogs.GetQueryResultsOutput
	var logs []Message

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

			InputDescribeService := &ecs.DescribeServicesInput{
				Services: []*string{
					aws.String(service),
				},
				Cluster: aws.String(cluster),
			}

			OutputDescribeService, err := client.DescribeServices(InputDescribeService)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

			task_def := aws.String(*OutputDescribeService.Services[0].TaskDefinition)

			InputDescribeTaskDefinition := &ecs.DescribeTaskDefinitionInput{
				TaskDefinition: aws.String(*task_def),
			}

			OutputDescribeTaskDefinition, err := client.DescribeTaskDefinition(InputDescribeTaskDefinition)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

			log_group := aws.String(*OutputDescribeTaskDefinition.TaskDefinition.ContainerDefinitions[0].LogConfiguration.Options["awslogs-group"])

			query_string := fmt.Sprintf("fields @timestamp, @message | sort @timestamp desc | limit %d", limit)

			StartQueryInput := &cloudwatchlogs.StartQueryInput{
				QueryString:  aws.String(query_string),
				LogGroupName: aws.String(*log_group),
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

	cmd.Flags().StringVarP(&cluster, "cluster", "c", "string", "ECS Cluster name")
	cmd.Flags().Int64VarP(&limit, "limit", "l", 200, "Logs max result")
	cmd.MarkFlagRequired("cluster")

	return cmd

}
