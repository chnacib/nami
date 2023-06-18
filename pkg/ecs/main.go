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

			input := &ecs.ListClustersInput{}

			result, err := client.ListClusters(input)
			if err != nil {
				log.Fatal(err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\tARN")

			for _, clusterArn := range result.ClusterArns {
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

			result, err := client.DescribeClusters(input)
			if err != nil {
				fmt.Println("Failed to describe ECS cluster: Cluster not found")
				os.Exit(0)
			}

			fmt.Println(result)

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

			result, err := client.ListServices(input)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\tTASK DEFINITION\tRUNNING\tCPU\tMEMORY\tLAUNCH")

			utilizationCh := make(chan struct {
				CPU    float64
				Memory float64
			}, len(result.ServiceArns))

			var wg sync.WaitGroup

			for _, serviceArn := range result.ServiceArns {
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

					serviceResult, err := client.DescribeServices(input)
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

					service := serviceResult.Services[0]
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

			result, err := client.DescribeServices(input)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

			fmt.Println(result)

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

			result, err := client.ListTaskDefinitionFamilies(input)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\t")

			for _, tasks := range result.Families {
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

func ListTasks() *cobra.Command {
	var cluster string
	cmd := &cobra.Command{
		Use:     "tasks",
		Aliases: []string{"tsk", "task"},
		Short:   "Describe ECS task definition",
		Run: func(cmd *cobra.Command, args []string) {
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
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\tREVISION\tSTATUS\tCPU\tMEMORY\tNETWORK\tSTARTED")

			for _, tasks := range response.TaskArns {
				task_name := NameArn(*tasks)
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
				task := response.Tasks[0]
				revision := NameArn(aws.StringValue(task.TaskDefinitionArn))
				cpu := aws.StringValue(task.Cpu)
				memory := aws.StringValue(task.Memory)
				status := aws.StringValue(task.LastStatus)
				network := aws.StringValue(task.Containers[0].NetworkInterfaces[0].PrivateIpv4Address)
				time_start := aws.TimeValue(task.StartedAt)
				format_time := time_start.Format("2006-01-02 15:04:05")
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", task_name, revision, status, cpu, memory, network, format_time)
			}
			w.Flush()
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

func NameArn(arn string) string {
	arnParts := strings.Split(arn, "/")
	serviceName := arnParts[len(arnParts)-1]
	return serviceName
}
