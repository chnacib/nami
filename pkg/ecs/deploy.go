package ecs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awsecs "github.com/aws/aws-sdk-go-v2/service/ecs"
	ectypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"

	"github.com/spf13/cobra"
)

type DeployOptions struct {
	Cluster       string
	Service       string
	ContainerName string // opcional; se vazio, usa índice 0
	Image         string
	Wait          bool
	Timeout       time.Duration
}

func Deploy() *cobra.Command {
	var (
		cluster       string
		service       string
		containerName string
		image         string
		wait          bool
		timeoutSec    int
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new image to an ECS service (CI friendly)",
		RunE: func(cmd *cobra.Command, args []string) error {
			// flags + env vars
			if cluster == "" {
				cluster = os.Getenv("NAMI_CLUSTER")
			}
			if service == "" {
				service = os.Getenv("NAMI_SERVICE")
			}
			if containerName == "" {
				containerName = os.Getenv("NAMI_CONTAINER")
			}
			if image == "" {
				image = os.Getenv("NAMI_IMAGE")
			}

			if cluster == "" {
				return fmt.Errorf("cluster is required (use --cluster or NAMI_CLUSTER)")
			}
			if service == "" {
				return fmt.Errorf("service is required (use --service or NAMI_SERVICE)")
			}
			if image == "" {
				return fmt.Errorf("image is required (use --image or NAMI_IMAGE)")
			}
			if timeoutSec <= 0 {
				timeoutSec = 300
			}

			opts := DeployOptions{
				Cluster:       cluster,
				Service:       service,
				ContainerName: containerName,
				Image:         image,
				Wait:          wait,
				Timeout:       time.Duration(timeoutSec) * time.Second,
			}

			tdArn, err := deployService(cmd.Context(), opts)
			if err != nil {
				return err
			}

			fmt.Println(tdArn)
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: false,
	}

	cmd.Flags().StringVarP(&cluster, "cluster", "c", "", "ECS cluster name (or NAMI_CLUSTER)")
	cmd.Flags().StringVarP(&service, "service", "s", "", "ECS service name (or NAMI_SERVICE)")
	cmd.Flags().StringVar(&containerName, "container-name", "", "Container name in task definition (or NAMI_CONTAINER)")
	cmd.Flags().StringVar(&image, "image", "", "Container image (or NAMI_IMAGE)")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait until service reaches steady state")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 300, "Timeout in seconds for wait")

	return cmd
}

func deployService(ctx context.Context, opts DeployOptions) (string, error) {
	if opts.Cluster == "" || opts.Service == "" || opts.Image == "" {
		return "", errors.New("cluster, service and image are required")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("loading AWS config: %w", err)
	}

	client := awsecs.NewFromConfig(cfg)

	svcOut, err := client.DescribeServices(ctx, &awsecs.DescribeServicesInput{
		Cluster:  aws.String(opts.Cluster),
		Services: []string{opts.Service},
	})
	if err != nil {
		return "", fmt.Errorf("describe service: %w", err)
	}
	if len(svcOut.Services) == 0 {
		return "", fmt.Errorf("service %q not found in cluster %q", opts.Service, opts.Cluster)
	}

	svc := svcOut.Services[0]
	if svc.TaskDefinition == nil {
		return "", fmt.Errorf("service %q has no task definition", opts.Service)
	}

	currentTDArn := aws.ToString(svc.TaskDefinition)

	tdOut, err := client.DescribeTaskDefinition(ctx, &awsecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(currentTDArn),
	})
	if err != nil {
		return "", fmt.Errorf("describe task definition %q: %w", currentTDArn, err)
	}

	td := tdOut.TaskDefinition
	if td == nil {
		return "", fmt.Errorf("task definition %q not found", currentTDArn)
	}

	containers := make([]ectypes.ContainerDefinition, len(td.ContainerDefinitions))
	copy(containers, td.ContainerDefinitions)

	if len(containers) == 0 {
		return "", fmt.Errorf("task definition %q has no container definitions", currentTDArn)
	}

	if opts.ContainerName == "" {
		// comportamento padrão: usa sempre o container de índice 0
		containers[0].Image = aws.String(opts.Image)
	} else {
		// se containerName foi informado, procura pelo nome
		found := false
		for i, c := range containers {
			if aws.ToString(c.Name) == opts.ContainerName {
				containers[i].Image = aws.String(opts.Image)
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("container %q not found in task definition %q", opts.ContainerName, currentTDArn)
		}
	}

	regIn := &awsecs.RegisterTaskDefinitionInput{
		Family:                  td.Family,
		TaskRoleArn:             td.TaskRoleArn,
		ExecutionRoleArn:        td.ExecutionRoleArn,
		NetworkMode:             td.NetworkMode,
		ContainerDefinitions:    containers,
		Volumes:                 td.Volumes,
		PlacementConstraints:    td.PlacementConstraints,
		RequiresCompatibilities: td.RequiresCompatibilities,
		Cpu:                     td.Cpu,
		Memory:                  td.Memory,
		EphemeralStorage:        td.EphemeralStorage,
		ProxyConfiguration:      td.ProxyConfiguration,
		InferenceAccelerators:   td.InferenceAccelerators,
		RuntimePlatform:         td.RuntimePlatform,
		IpcMode:                 td.IpcMode,
		PidMode:                 td.PidMode,
	}

	regOut, err := client.RegisterTaskDefinition(ctx, regIn)
	if err != nil {
		return "", fmt.Errorf("register task definition: %w", err)
	}
	if regOut.TaskDefinition == nil || regOut.TaskDefinition.TaskDefinitionArn == nil {
		return "", errors.New("register task definition returned empty task definition ARN")
	}

	newTDArn := aws.ToString(regOut.TaskDefinition.TaskDefinitionArn)

	_, err = client.UpdateService(ctx, &awsecs.UpdateServiceInput{
		Cluster:        aws.String(opts.Cluster),
		Service:        aws.String(opts.Service),
		TaskDefinition: aws.String(newTDArn),
	})
	if err != nil {
		return "", fmt.Errorf("update service to task definition %q: %w", newTDArn, err)
	}

	if !opts.Wait {
		return newTDArn, nil
	}

	waitCtx := ctx
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-waitCtx.Done():
			return "", fmt.Errorf("wait for deployment steady state timed out after %s", opts.Timeout)
		case <-ticker.C:
			dsOut, err := client.DescribeServices(waitCtx, &awsecs.DescribeServicesInput{
				Cluster:  aws.String(opts.Cluster),
				Services: []string{opts.Service},
			})
			if err != nil {
				return "", fmt.Errorf("describe service while waiting: %w", err)
			}

			if len(dsOut.Services) == 0 {
				return "", fmt.Errorf("service %q disappeared while waiting", opts.Service)
			}

			s := dsOut.Services[0]

			if len(s.Deployments) == 0 {
				continue
			}

			allStable := true
			hasPrimaryNewTD := false

			for _, d := range s.Deployments {
				status := aws.ToString(d.Status)

				if aws.ToString(d.TaskDefinition) == newTDArn && status == "PRIMARY" {
					hasPrimaryNewTD = true
				}

				if status == "IN_PROGRESS" {
					allStable = false
					break
				}
			}

			if !hasPrimaryNewTD {
				allStable = false
			}

			if allStable && s.RunningCount == s.DesiredCount {
				return newTDArn, nil
			}
		}
	}
}
