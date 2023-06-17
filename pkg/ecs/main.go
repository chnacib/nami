package ecs

import (
	"fmt"
	"log"
	"os"
	"strings"
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
		Aliases: []string{"svc"},
		Short:   "list ECS services",
		Run: func(cmd *cobra.Command, args []string) {
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))
			client := ecs.New(sess)

			input := &ecs.ListServicesInput{
				Cluster: aws.String(cluster),
			}

			result, err := client.ListServices(input)

			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\tTASK DEFINITION\tRUNNING\tLAUNCH")

			for _, serviceArn := range result.ServiceArns {
				serviceName := NameArn(*serviceArn)
				input := &ecs.DescribeServicesInput{
					Services: []*string{
						aws.String(serviceName),
					},
					Cluster: aws.String(cluster),
				}
				result, err1 := client.DescribeServices(input)
				if err != nil {
					fmt.Println(err1)
					os.Exit(0)
				}
				service := result.Services[0]
				taskdef := NameArn(aws.StringValue(service.TaskDefinition))
				running := aws.Int64Value(service.RunningCount)
				desired := aws.Int64Value(service.DesiredCount)
				launchtype := aws.StringValue(service.LaunchType)

				fmt.Fprintf(w, "%s\t%s\t%d/%d\t%s\n", serviceName, taskdef, running, desired, launchtype)
			}

			w.Flush()

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
