package ecs

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

func ListNodes() *cobra.Command {
	var cluster string
	var container_instances []string

	cmd := &cobra.Command{
		Use:     "nodes",
		Aliases: []string{"node"},
		Short:   "List ECS instances in default region",
		Run: func(cmd *cobra.Command, args []string) {
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			client := ecs.New(sess)

			list_input := &ecs.ListContainerInstancesInput{
				Cluster:    aws.String(cluster),
				MaxResults: aws.Int64(100),
			}

			list_response, err := client.ListContainerInstances(list_input)
			if err != nil {
				log.Fatal(err)
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\tNETWORK\tSIZE\tREGISTERED CPU\tREGISTERED MEMORY\tREMAINING CPU\tREMAINING MEMORY\tSTATUS")

			for _, container_instance_arn := range list_response.ContainerInstanceArns {
				container_instance := NameArn(*container_instance_arn)
				container_instances = append(container_instances, container_instance)
			}

			describe_input := &ecs.DescribeContainerInstancesInput{
				Cluster:            aws.String(cluster),
				ContainerInstances: aws.StringSlice(container_instances),
			}

			describe_response, err := client.DescribeContainerInstances(describe_input)
			if err != nil {
				log.Fatal(err)
			}

			for _, output_container_instance := range describe_response.ContainerInstances {
				capacity_provider_name := *output_container_instance.CapacityProviderName
				remaining_cpu := *output_container_instance.RemainingResources[0].IntegerValue
				remaining_memory := *output_container_instance.RemainingResources[1].IntegerValue
				registered_cpu := *output_container_instance.RegisteredResources[0].IntegerValue
				registered_memory := *output_container_instance.RegisteredResources[1].IntegerValue
				status := *output_container_instance.Status
				private_ipv4 := *output_container_instance.Attachments[0].Details[4].Value
				var instance_type string
				for _, attribute := range output_container_instance.Attributes {
					if *attribute.Name == "ecs.instance-type" {
						instance_type = *attribute.Value
						break
					}
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%d\t%d\t%s\n", capacity_provider_name, private_ipv4, instance_type, registered_cpu, registered_memory, remaining_cpu, remaining_memory, status)

			}
			w.Flush()

		},
	}
	cmd.Flags().StringVarP(&cluster, "cluster", "c", "string", "ECS Cluster name")
	cmd.MarkFlagRequired("cluster")

	return cmd
}
