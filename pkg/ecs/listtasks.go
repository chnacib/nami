package ecs

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

//Tasks

func ListTasks() *cobra.Command {
	var cluster string
	var taskArns []*string
	var timeStart time.Time
	var formatTime string
	var network string

	cmd := &cobra.Command{
		Use:     "tasks",
		Aliases: []string{"tsk", "task"},
		Short:   "List ECS tasks from services",
		Run: func(cmd *cobra.Command, args []string) {
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))
			client := ecs.New(sess)
			if len(args) == 0 {
				listTaskInput := &ecs.ListTasksInput{
					Cluster: aws.String(cluster),
				}

				listTaskOutput, err := client.ListTasks(listTaskInput)
				if err != nil {
					fmt.Printf("Tasks not found for cluster %s", cluster)
					os.Exit(0)
				}

				taskArns = append(taskArns, listTaskOutput.TaskArns...)

				describeTasksInput := &ecs.DescribeTasksInput{
					Tasks:   taskArns,
					Cluster: aws.String(cluster),
				}

				describeTasksOutput, err := client.DescribeTasks(describeTasksInput)
				if err != nil {
					fmt.Printf("Tasks not found for cluster %s\n", cluster)
					os.Exit(0)
				}

				w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

				fmt.Fprintln(w, "NAME\tREVISION\tSTATUS\tCPU\tMEMORY\tNETWORK\tSTARTED")

				for _, task := range describeTasksOutput.Tasks {

					taskName := NameArn(aws.StringValue(task.TaskArn))
					revision := NameArn(aws.StringValue(task.TaskDefinitionArn))
					cpu := aws.StringValue(task.Cpu)
					memory := aws.StringValue(task.Memory)
					status := aws.StringValue(task.LastStatus)
					if len(task.Containers) > 0 {
						network = aws.StringValue(task.Containers[0].NetworkInterfaces[0].PrivateIpv4Address)
					} else {
						network = "-"
					}
					if task.StartedAt != nil {
						timeStart = aws.TimeValue(task.StartedAt)
						formatTime = timeStart.Format("2006-01-02 15:04:05")
					} else {
						formatTime = "-"
					}

					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", taskName, revision, status, cpu, memory, network, formatTime)
				}
				w.Flush()
				os.Exit(0)
			}
		},
	}

	cmd.Flags().StringVarP(&cluster, "cluster", "c", "string", "ECS Cluster name")
	cmd.MarkFlagRequired("cluster")

	return cmd
}
