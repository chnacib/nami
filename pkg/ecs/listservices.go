package ecs

import (
	"context"
	"fmt"
	"os"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/chnacib/nami/pkg/config"
	"github.com/chnacib/nami/pkg/cw"
	"github.com/chnacib/nami/pkg/utils"
	"github.com/spf13/cobra"
)

type ServiceOptions struct {
	Cluster string
	Service string
}

type ListServicesOutput struct {
	Cluster            string
	Service            string
	LaunchType         string
	Status             string
	DesiredCount       int32
	RunningCount       int32
	PendingCount       int32
	CreatedAt          *time.Time
	TaskDefinition     string
	TaskDefinitionName string
}

// ListServices creates a cobra command for listing ECS services
func ListServices() *cobra.Command {
	var cluster string
	cmd := &cobra.Command{
		Use:     "services",
		Aliases: []string{"svc", "service"},
		Short:   "list ECS services",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			options := ServiceOptions{
				Cluster: cluster,
			}

			// If service name is provided as an argument
			if len(args) > 0 {
				options.Service = args[0]
			}

			services, err := GetECSServices(ctx, options)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}

			// Display the results in a table
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\tTASK DEFINITION\tRUNNING\tCPU\tMEMORY\tLAUNCH")

			// Process services and collect CloudWatch metrics in parallel
			utilizationCh := make(chan struct {
				Service string
				CPU     float64
				Memory  float64
			}, len(services))

			var wg sync.WaitGroup
			for _, service := range services {
				wg.Add(1)
				go func(svc ListServicesOutput) {
					defer wg.Done()

					// Get CloudWatch metrics
					cpu := cw.CpuAverage(svc.Cluster, svc.Service)
					memory := cw.MemoryAverage(svc.Cluster, svc.Service)

					utilizationCh <- struct {
						Service string
						CPU     float64
						Memory  float64
					}{
						Service: svc.Service,
						CPU:     cpu,
						Memory:  memory,
					}
				}(service)
			}

			// Wait for all goroutines to complete
			go func() {
				wg.Wait()
				close(utilizationCh)
			}()

			// Build a map of utilization data
			utilMap := make(map[string]struct {
				CPU    float64
				Memory float64
			})

			for util := range utilizationCh {
				utilMap[util.Service] = struct {
					CPU    float64
					Memory float64
				}{
					CPU:    util.CPU,
					Memory: util.Memory,
				}
			}

			// Print service information with utilization data
			for _, service := range services {
				util, ok := utilMap[service.Service]
				cpuUtil := 0.0
				memUtil := 0.0
				if ok {
					cpuUtil = util.CPU
					memUtil = util.Memory
				}

				fmt.Fprintf(w, "%s\t%s\t%d/%d\t%.2f%%\t%.2f%%\t%s\n",
					service.Service,
					service.TaskDefinitionName,
					service.RunningCount,
					service.DesiredCount,
					cpuUtil,
					memUtil,
					service.LaunchType)
			}

			w.Flush()
		},
	}

	cmd.Flags().StringVarP(&cluster, "cluster", "c", "", "ECS Cluster name")
	cmd.MarkFlagRequired("cluster")
	return cmd
}

func GetECSServices(ctx context.Context, options ServiceOptions) ([]ListServicesOutput, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	client := ecs.NewFromConfig(cfg.AwsConfig)
	var output []ListServicesOutput

	clusters := []string{}
	if options.Cluster != "" {
		clusters = append(clusters, options.Cluster)
	} else {
		listClustersResp, err := client.ListClusters(ctx, &ecs.ListClustersInput{})
		if err != nil {
			return nil, fmt.Errorf("failed to list ECS clusters: %w", err)
		}

		if len(listClustersResp.ClusterArns) == 0 {
			return nil, fmt.Errorf("no ECS clusters found in the account")
		}

		clusters = listClustersResp.ClusterArns
	}

	for _, cluster := range clusters {

		clusterName := utils.GetResourceName(cluster)

		var serviceArns []string
		var nextToken *string

		for {
			listServicesResp, err := client.ListServices(ctx, &ecs.ListServicesInput{
				Cluster:    aws.String(cluster),
				MaxResults: aws.Int32(100),
				NextToken:  nextToken,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list services for cluster %s: %w", clusterName, err)
			}

			serviceArns = append(serviceArns, listServicesResp.ServiceArns...)

			if listServicesResp.NextToken == nil {
				break
			}
			nextToken = listServicesResp.NextToken
		}

		if len(serviceArns) == 0 {
			continue
		}

		for i := 0; i < len(serviceArns); i += 10 {
			end := i + 10
			if end > len(serviceArns) {
				end = len(serviceArns)
			}

			batch := serviceArns[i:end]

			describeServicesResp, err := client.DescribeServices(ctx, &ecs.DescribeServicesInput{
				Cluster:  aws.String(cluster),
				Services: batch,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe services in cluster %s: %w", clusterName, err)
			}

			if len(describeServicesResp.Failures) > 0 {
				for _, failure := range describeServicesResp.Failures {
					fmt.Fprintf(os.Stderr, "Warning: Service %s failed to be described: %s (reason: %s)\n",
						aws.ToString(failure.Arn), aws.ToString(failure.Detail), aws.ToString(failure.Reason))
				}
			}

			// Process service details
			for _, service := range describeServicesResp.Services {
				serviceName := aws.ToString(service.ServiceName)

				// Skip if we're filtering by service name and this isn't a match
				if options.Service != "" && options.Service != serviceName {
					continue
				}

				// Get task definition name from ARN
				taskDefName := utils.GetResourceName(aws.ToString(service.TaskDefinition))

				outputService := ListServicesOutput{
					Cluster:            clusterName,
					Service:            serviceName,
					Status:             aws.ToString(service.Status),
					DesiredCount:       service.DesiredCount,
					RunningCount:       service.RunningCount,
					PendingCount:       service.PendingCount,
					CreatedAt:          service.CreatedAt,
					TaskDefinition:     aws.ToString(service.TaskDefinition),
					TaskDefinitionName: taskDefName,
				}

				if service.LaunchType != "" {
					// Use the explicit launch type if available
					outputService.LaunchType = string(service.LaunchType)
				} else if len(service.CapacityProviderStrategy) > 0 {
					// Check capacity providers to determine launch type
					for _, cp := range service.CapacityProviderStrategy {
						cpName := aws.ToString(cp.CapacityProvider)
						if cpName == "" {
							continue
						}
						if cpName != "" {
							outputService.LaunchType = fmt.Sprintf("%s", cpName)
							break

						}
					}
					if outputService.LaunchType == "" {
						outputService.LaunchType = "CAPACITY_PROVIDER"
					}
				} else if service.DeploymentConfiguration != nil &&
					service.DeploymentConfiguration.DeploymentCircuitBreaker != nil &&
					service.DeploymentConfiguration.DeploymentCircuitBreaker.Enable {
					// Try to determine launch type from deployment configuration
					outputService.LaunchType = "EC2/FARGATE"
				} else {
					outputService.LaunchType = "UNKNOWN"
				}

				output = append(output, outputService)
			}
		}
	}

	if len(output) == 0 {
		if options.Cluster != "" && options.Service != "" {
			return nil, fmt.Errorf("service '%s' not found in cluster '%s'", options.Service, options.Cluster)
		} else if options.Cluster != "" {
			return nil, fmt.Errorf("no services found in cluster '%s'", options.Cluster)
		} else if options.Service != "" {
			return nil, fmt.Errorf("service '%s' not found in any cluster", options.Service)
		}
		return nil, fmt.Errorf("no ECS services found in the account")
	}

	return output, nil
}
