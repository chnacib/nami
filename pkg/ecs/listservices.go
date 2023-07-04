package ecs

import (
	"fmt"
	"nami/pkg/cw"
	"os"
	"sync"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

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
