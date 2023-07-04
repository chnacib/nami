package ecs

import (
	"fmt"
	"log"
	"nami/pkg/cw"
	"os"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

//Cluster operations

func ListClusters() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "clusters",
		Aliases: []string{"cluster"},
		Short:   "List ECS Clusters in default region",
		Run: func(cmd *cobra.Command, args []string) {
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			client := ecs.New(sess)

			input := &ecs.ListClustersInput{
				MaxResults: aws.Int64(100),
			}

			response, err := client.ListClusters(input)
			if err != nil {
				log.Fatal(err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\tARN")

			for _, clusterArn := range response.ClusterArns {
				clusterName := NameArn(*clusterArn)
				fmt.Fprintf(w, "%s\t%s\n", clusterName, *clusterArn)
			}

			w.Flush()

		},
	}

	return cmd
}

func DescribeCluster() *cobra.Command {

	var name string

	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Describe ECS Cluster",
		Run: func(cmd *cobra.Command, args []string) {
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			client := ecs.New(sess)

			input := &ecs.DescribeClustersInput{
				Clusters: []*string{
					aws.String(name),
				}}

			response, err := client.DescribeClusters(input)
			if err != nil {
				fmt.Println("Failed to describe ECS cluster: Cluster not found")
				os.Exit(0)
			}

			fmt.Println(response)

		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "string", "ECS Cluster name")
	cmd.MarkFlagRequired("name")

	return cmd
}

func ListServices() *cobra.Command {
	var cluster string

	cmd := &cobra.Command{
		Use:     "services",
		Aliases: []string{"svc", "service"},
		Short:   "list ECS services",
		Run: func(cmd *cobra.Command, args []string) {
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))
			client := ecs.New(sess)

			input := &ecs.ListServicesInput{
				Cluster:    aws.String(cluster),
				MaxResults: aws.Int64(100),
			}

			response, err := client.ListServices(input)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\tTASK DEFINITION\tRUNNING\tCPU\tMEMORY\tLAUNCH")

			utilizationCh := make(chan struct {
				CPU    float64
				Memory float64
			}, len(response.ServiceArns))

			var wg sync.WaitGroup

			for _, serviceArn := range response.ServiceArns {
				wg.Add(1)

				go func(serviceArn string) {
					defer wg.Done()

					serviceName := NameArn(serviceArn)
					input := &ecs.DescribeServicesInput{
						Services: []*string{
							aws.String(serviceName),
						},
						Cluster: aws.String(cluster),
					}

					serviceresponse, err := client.DescribeServices(input)
					if err != nil {
						fmt.Println(err)
						return
					}

					cpu := cw.CpuAverage(cluster, serviceName)
					memory := cw.MemoryAverage(cluster, serviceName)
					utilizationCh <- struct {
						CPU    float64
						Memory float64
					}{CPU: cpu, Memory: memory}

					service := serviceresponse.Services[0]
					taskdef := NameArn(aws.StringValue(service.TaskDefinition))
					running := aws.Int64Value(service.RunningCount)
					desired := aws.Int64Value(service.DesiredCount)
					launchtype := aws.StringValue(service.LaunchType)

					fmt.Fprintf(w, "%s\t%s\t%d/%d\t%.2f%%\t%.2f%%\t%s\n", serviceName, taskdef, running, desired, cpu, memory, launchtype)
				}(aws.StringValue(serviceArn))
			}

			go func() {
				wg.Wait()
				close(utilizationCh)
			}()

			for util := range utilizationCh {
				_ = util.CPU
				_ = util.Memory
			}

			w.Flush()
		},
	}

	cmd.Flags().StringVarP(&cluster, "cluster", "c", "string", "ECS Cluster name")
	cmd.MarkFlagRequired("cluster")

	return cmd
}

func DescribeService() *cobra.Command {
	var cluster string
	cmd := &cobra.Command{
		Use:     "service",
		Aliases: []string{"svc", "services"},
		Short:   "Describe ECS services",
		Run: func(cmd *cobra.Command, args []string) {
			service := args[0]
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))
			client := ecs.New(sess)

			input := &ecs.DescribeServicesInput{
				Services: []*string{
					aws.String(service),
				},
				Cluster: aws.String(cluster),
			}

			response, err := client.DescribeServices(input)
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

//TaskDefinition

func ListTaskDefinition() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "taskdefinition",
		Aliases: []string{"td"},
		Short:   "List ECS task definition",
		Run: func(cmd *cobra.Command, args []string) {
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))
			client := ecs.New(sess)

			input := &ecs.ListTaskDefinitionFamiliesInput{
				Status: aws.String("ACTIVE"),
			}

			response, err := client.ListTaskDefinitionFamilies(input)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\t")

			for _, tasks := range response.Families {
				task := *tasks
				input := &ecs.DescribeTaskDefinitionInput{
					TaskDefinition: aws.String(task),
				}

				_, err := client.DescribeTaskDefinition(input)
				if err != nil {
					fmt.Println(err)
					os.Exit(0)
				}

				fmt.Fprintf(w, "%s\t\n", task)
			}

			w.Flush()

		},
	}

	return cmd

}

func DescribeTaskDefinition() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "taskdefinition",
		Aliases: []string{"td"},
		Short:   "Describe ECS task definition",
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))
			client := ecs.New(sess)

			input := &ecs.DescribeTaskDefinitionInput{
				TaskDefinition: aws.String(name),
			}

			response, err := client.DescribeTaskDefinition(input)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
			fmt.Println(response)

		},
	}

	return cmd

}

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

func TaskLogs() *cobra.Command {
	var cluster string

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

		},
	}

	cmd.Flags().StringVarP(&cluster, "cluster", "c", "string", "ECS Cluster name")
	cmd.MarkFlagRequired("cluster")

	return cmd

}

func NameArn(arn string) string {
	arnParts := strings.Split(arn, "/")
	serviceName := arnParts[len(arnParts)-1]
	return serviceName
}
