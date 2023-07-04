package ecs

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

//Tasks

func ListTasks() *cobra.Command {
	var cluster string

	cmd := &cobra.Command{
		Use:     "tasks",
		Aliases: []string{"tsk", "task"},
		Short:   "List ECS tasks from services",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("No service specified")
				return
			}

			service := args[0]
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))
			client := ecs.New(sess)

			input := &ecs.ListTasksInput{
				ServiceName: aws.String(service),
				Cluster:     aws.String(cluster),
			}

			response, err := client.ListTasks(input)
			if err != nil {
				fmt.Println(err)
				return
			}

			var taskArns []*string
			for _, taskArn := range response.TaskArns {
				taskArns = append(taskArns, taskArn)
			}

			describeTasksInput := &ecs.DescribeTasksInput{
				Tasks:   taskArns,
				Cluster: aws.String(cluster),
			}

			describeTasksOutput, err := client.DescribeTasks(describeTasksInput)
			if err != nil {
				fmt.Println(err)
				return
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			defer w.Flush()

			fmt.Fprintln(w, "NAME\tREVISION\tSTATUS\tCPU\tMEMORY\tNETWORK\tSTARTED")

			for _, task := range describeTasksOutput.Tasks {
				taskName := NameArn(aws.StringValue(task.TaskArn))
				revision := NameArn(aws.StringValue(task.TaskDefinitionArn))
				cpu := aws.StringValue(task.Cpu)
				memory := aws.StringValue(task.Memory)
				status := aws.StringValue(task.LastStatus)
				network := aws.StringValue(task.Containers[0].NetworkInterfaces[0].PrivateIpv4Address)
				timeStart := aws.TimeValue(task.StartedAt)
				formatTime := timeStart.Format("2006-01-02 15:04:05")

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", taskName, revision, status, cpu, memory, network, formatTime)
			}
		},
	}

	cmd.Flags().StringVarP(&cluster, "cluster", "c", "string", "ECS Cluster name")
	cmd.MarkFlagRequired("cluster")

	return cmd
}
