package ecs

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

func ListAutoscaling() *cobra.Command {
	var cluster string

	cmd := &cobra.Command{
		Use:     "autoscaling",
		Aliases: []string{"as", "autoscale"},
		Short:   "list ECS services",
		Run: func(cmd *cobra.Command, args []string) {
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))
			cfg, err := config.LoadDefaultConfig(context.TODO())
			if err != nil {
				log.Fatal()
			}
			client_ecs := ecs.New(sess)

			client_app := applicationautoscaling.NewFromConfig(cfg)

			input := &ecs.ListServicesInput{
				Cluster:    aws.String(cluster),
				MaxResults: aws.Int64(100),
			}

			response, err := client_ecs.ListServices(input)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "SERVICE\tRUNNING\tDESIRED\tMIN\tMAX\tTARGETS")

			policyCache := make(map[string]string)
			var policyCacheMutex sync.Mutex

			var wg sync.WaitGroup
			for _, serviceArn := range response.ServiceArns {
				wg.Add(1)
				serviceArnCopy := *serviceArn
				go func(serviceArn string) {
					defer wg.Done()

					service := NameArn(serviceArn)

					input_target := &applicationautoscaling.DescribeScalableTargetsInput{
						ServiceNamespace: types.ServiceNamespaceEcs,
						MaxResults:       aws.Int32(100),
						ResourceIds: []string{
							*aws.String(fmt.Sprintf("service/%s/%s", cluster, service)),
						},
						ScalableDimension: types.ScalableDimensionECSServiceDesiredCount,
					}

					output_target, err := client_app.DescribeScalableTargets(context.TODO(), input_target)
					if err != nil {
						fmt.Println(err)
						os.Exit(0)
					}

					input_svc := &ecs.DescribeServicesInput{
						Services: []*string{
							aws.String(service),
						},
						Cluster: aws.String(cluster),
					}

					output_svc, err := client_ecs.DescribeServices(input_svc)
					if err != nil {
						fmt.Println(err)
						os.Exit(0)
					}

					desired := aws.Int64Value(output_svc.Services[0].DesiredCount)
					running := aws.Int64Value(output_svc.Services[0].RunningCount)

					var min_capacity, max_capacity int32
					for _, scalable_targets := range output_target.ScalableTargets {
						min_capacity = aws.Int32Value(scalable_targets.MinCapacity)
						max_capacity = aws.Int32Value(scalable_targets.MaxCapacity)
					}

					input_policy := &applicationautoscaling.DescribeScalingPoliciesInput{
						ServiceNamespace:  types.ServiceNamespaceEcs,
						ResourceId:        aws.String(fmt.Sprintf("service/%s/%s", cluster, service)),
						ScalableDimension: types.ScalableDimensionECSServiceDesiredCount,
					}
					output_policy, err := client_app.DescribeScalingPolicies(context.TODO(), input_policy)
					if err != nil {
						fmt.Println(err)
						os.Exit(0)
					}

					var policyResult strings.Builder
					for _, policies := range output_policy.ScalingPolicies {
						var policyValue string
						policy := *&policies.TargetTrackingScalingPolicyConfiguration.PredefinedMetricSpecification.PredefinedMetricType
						switch policy {
						case "ALBRequestCountPerTarget":
							policyValue = "REQUESTS"
						case "ECSServiceAverageCPUUtilization":
							policyValue = "CPU"
						case "ECSServiceAverageMemoryUtilization":
							policyValue = "MEMORY"
						}

						target_value := aws.Float64Value(policies.TargetTrackingScalingPolicyConfiguration.TargetValue)
						if policyResult.Len() > 0 {
							policyResult.WriteString(" | ")
						}

						policyResult.WriteString(fmt.Sprintf("%s:%s", policyValue, strconv.FormatFloat(target_value, 'f', -1, 64)))
					}

					policyCacheMutex.Lock()
					policyCache[serviceArn] = policyResult.String()
					policyCacheMutex.Unlock()

					fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%s\n", service, running, desired, min_capacity, max_capacity, policyCache[serviceArn])
				}(serviceArnCopy)
			}

			wg.Wait()
			w.Flush()
		},
	}

	cmd.Flags().StringVarP(&cluster, "cluster", "c", "string", "ECS Cluster name")
	cmd.MarkFlagRequired("cluster")

	return cmd
}
