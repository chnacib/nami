package ecs

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

func Autoscaling() *cobra.Command {
	var cluster string
	var minimum int
	var maximum int
	var cpu int32
	var mem int32
	var request int32

	cmd := &cobra.Command{
		Use:     "autoscale",
		Aliases: []string{"as", "autoscaling", "scale"},
		Short:   "Register autoscaling config",
		Run: func(cmd *cobra.Command, args []string) {
			service := args[0]
			registerScalableTarget(cluster, service, int32(minimum), int32(maximum), cpu, mem, request)
		},
	}

	cmd.Flags().Int32VarP(&request, "request", "", 0, "Memory Average utilization")
	cmd.Flags().Int32VarP(&mem, "mem", "", 0, "Memory Average utilization")
	cmd.Flags().Int32VarP(&cpu, "cpu", "", 0, "CPU Average utilization")
	cmd.Flags().IntVarP(&maximum, "max", "", 1, "Maximum desired count")
	cmd.MarkFlagRequired("max")
	cmd.Flags().IntVarP(&minimum, "min", "", 1, "Minimum desired count")
	cmd.MarkFlagRequired("min")
	cmd.Flags().StringVarP(&cluster, "cluster", "c", "string", "ECS Cluster name")
	cmd.MarkFlagRequired("cluster")
	return cmd
}

func registerScalableTarget(clusterName, serviceName string, minCapacity, maxCapacity, cpu, mem, request int32) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to load AWS configuration: %v", err)
	}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	client_elb := elasticloadbalancingv2.NewFromConfig(cfg)

	client_ecs := ecs.New(sess)

	// Create an AWS Application Auto Scaling client
	client := applicationautoscaling.NewFromConfig(cfg)

	// Specify the resource ID of the scalable target
	resourceID := fmt.Sprintf("service/%s/%s", clusterName, serviceName)

	// Set the parameters for the scalable target registration
	input := &applicationautoscaling.RegisterScalableTargetInput{
		ServiceNamespace:  types.ServiceNamespaceEcs,
		ResourceId:        &resourceID,
		ScalableDimension: types.ScalableDimensionECSServiceDesiredCount,
		MinCapacity:       &minCapacity,
		MaxCapacity:       &maxCapacity,
		SuspendedState:    nil,
	}

	_, err = client.RegisterScalableTarget(context.TODO(), input)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	cpuScalingInput := &applicationautoscaling.PutScalingPolicyInput{
		PolicyName:        aws.String("CPUTrackingPolicy"),
		PolicyType:        types.PolicyTypeTargetTrackingScaling,
		ResourceId:        &resourceID,
		ScalableDimension: types.ScalableDimensionECSServiceDesiredCount,
		ServiceNamespace:  types.ServiceNamespaceEcs,
		TargetTrackingScalingPolicyConfiguration: &types.TargetTrackingScalingPolicyConfiguration{
			ScaleInCooldown:  aws.Int32(0),
			ScaleOutCooldown: aws.Int32(0),
			TargetValue:      aws.Float64(float64(cpu)),
			PredefinedMetricSpecification: &types.PredefinedMetricSpecification{
				PredefinedMetricType: types.MetricType("ECSServiceAverageCPUUtilization"),
			},
		},
	}

	memScalingInput := &applicationautoscaling.PutScalingPolicyInput{
		PolicyName:        aws.String("MemoryTrackingPolicy"),
		PolicyType:        types.PolicyTypeTargetTrackingScaling,
		ResourceId:        &resourceID,
		ScalableDimension: types.ScalableDimensionECSServiceDesiredCount,
		ServiceNamespace:  types.ServiceNamespaceEcs,
		TargetTrackingScalingPolicyConfiguration: &types.TargetTrackingScalingPolicyConfiguration{
			ScaleInCooldown:  aws.Int32(0),
			ScaleOutCooldown: aws.Int32(0),
			TargetValue:      aws.Float64(float64(mem)),
			PredefinedMetricSpecification: &types.PredefinedMetricSpecification{
				PredefinedMetricType: types.MetricType("ECSServiceAverageMemoryUtilization"),
			},
		},
	}

	if cpu != 0 {
		_, err = client.PutScalingPolicy(context.TODO(), cpuScalingInput)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}

	if mem != 0 {
		_, err = client.PutScalingPolicy(context.TODO(), memScalingInput)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}

	if request != 0 {
		input := &ecs.DescribeServicesInput{
			Services: []*string{
				aws.String(serviceName),
			},
			Cluster: aws.String(clusterName),
		}

		response, err := client_ecs.DescribeServices(input)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

		target_group := aws.StringValue(response.Services[0].LoadBalancers[0].TargetGroupArn)

		inputTG := &elasticloadbalancingv2.DescribeTargetGroupsInput{
			TargetGroupArns: []string{
				*aws.String(target_group),
			},
		}

		outputTG, err := client_elb.DescribeTargetGroups(context.TODO(), inputTG)
		if err != nil {
			log.Fatalf("Error describing target groups: %v", err)
		}

		loadbalancer := outputTG.TargetGroups[0].LoadBalancerArns[0]
		resource_label := extractLoadBalancerID(loadbalancer) + "/" + extractTargetGroupID(target_group)

		requestScalingInput := &applicationautoscaling.PutScalingPolicyInput{
			PolicyName:        aws.String("RequestTrackingPolicy"),
			PolicyType:        types.PolicyTypeTargetTrackingScaling,
			ResourceId:        &resourceID,
			ScalableDimension: types.ScalableDimensionECSServiceDesiredCount,
			ServiceNamespace:  types.ServiceNamespaceEcs,
			TargetTrackingScalingPolicyConfiguration: &types.TargetTrackingScalingPolicyConfiguration{
				ScaleInCooldown:  aws.Int32(0),
				ScaleOutCooldown: aws.Int32(0),
				TargetValue:      aws.Float64(float64(request)),
				PredefinedMetricSpecification: &types.PredefinedMetricSpecification{
					PredefinedMetricType: types.MetricType("ALBRequestCountPerTarget"),
					ResourceLabel:        aws.String(resource_label),
				},
			},
		}
		_, err = client.PutScalingPolicy(context.TODO(), requestScalingInput)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}

	fmt.Printf("autoscaling configured")

	return nil
}

func Replicas() *cobra.Command {
	var cluster string
	var desired int64
	cmd := &cobra.Command{
		Use:     "replicas",
		Aliases: []string{"replica"},
		Short:   "Register autoscaling config",
		Run: func(cmd *cobra.Command, args []string) {
			service := args[0]
			UpdateDesiredCount(service, cluster, desired)

		},
	}

	cmd.Flags().Int64VarP(&desired, "desired", "d", 0, "Desired count")
	cmd.MarkFlagRequired("desired")
	cmd.Flags().StringVarP(&cluster, "cluster", "c", "string", "ECS Cluster name")
	cmd.MarkFlagRequired("cluster")
	return cmd
}

func UpdateDesiredCount(service string, cluster string, desired int64) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := ecs.New(sess)

	_, err := client.UpdateService(&ecs.UpdateServiceInput{
		Cluster:      aws.String(cluster),
		Service:      aws.String(service),
		DesiredCount: aws.Int64(desired),
	})

	if err != nil {
		fmt.Println("Failed to set service desired count:", err)
		os.Exit(0)
	}

	fmt.Printf("%s desired count updated to %d\n", service, desired)

}
