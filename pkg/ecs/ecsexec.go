package ecs

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/spf13/cobra"
)

func Exec() *cobra.Command {

	var command string
	var task string
	var cluster string
	var service string

	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Execute command in container",
		Run: func(cmd *cobra.Command, args []string) {
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))
			task = args[0]
			command = args[1]
			client := ecs.New(sess)

			task_input := &ecs.DescribeTasksInput{
				Tasks: []*string{
					aws.String(task),
				},
				Cluster: aws.String(cluster),
			}

			response, err := client.DescribeTasks(task_input)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
			container := aws.StringValue(response.Tasks[0].Containers[0].Name)
			status := aws.StringValue(response.Tasks[0].DesiredStatus)
			enable_execute_cmd := aws.BoolValue(response.Tasks[0].EnableExecuteCommand)
			service = Format(aws.StringValue(response.Tasks[0].Group))
			task_def := aws.StringValue(response.Tasks[0].TaskDefinitionArn)

			if enable_execute_cmd == false && status == "RUNNING" {
				fmt.Println("Execute command is disabled")
				input := &ecs.DescribeTaskDefinitionInput{
					TaskDefinition: aws.String(task_def),
				}

				response, err := client.DescribeTaskDefinition(input)
				if err != nil {
					fmt.Println(err)
					os.Exit(0)
				}
				role_arn := NameArn(aws.StringValue(response.TaskDefinition.ExecutionRoleArn))
				AttachPolicy(role_arn)
				UpdateService(service, cluster)

			}
			if enable_execute_cmd == true && status == "RUNNING" {
				ExecuteCommand(cluster, task, container, command)
			}
		},
	}
	cmd.Flags().StringVarP(&cluster, "cluster", "c", "string", "ECS Cluster name")
	cmd.MarkFlagRequired("cluster")

	return cmd
}

func Format(str string) string {
	var result string
	index := strings.Index(str, ":")
	if index != -1 {
		result = str[index+1:]
	}

	return result
}

func AttachPolicy(role_arn string) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := iam.New(sess)
	fmt.Println("Attaching inline policy...")
	// Specify the IAM role and policy
	policyDocument := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"ssmmessages:CreateControlChannel",
					"ssmmessages:CreateDataChannel",
					"ssmmessages:OpenControlChannel",
					"ssmmessages:OpenDataChannel"
				],
				"Resource": "*"
			}
		]
	}`

	_, err := client.PutRolePolicy(&iam.PutRolePolicyInput{
		PolicyDocument: aws.String(policyDocument),
		PolicyName:     aws.String("nami-ecs-exec"),
		RoleName:       aws.String(role_arn),
	})

	if err != nil {
		fmt.Println("Failed to attach inline policy to IAM role:", err)
		return
	}

	fmt.Println("Inline policy attached to IAM role successfully.")

}

func UpdateService(service string, cluster string) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := ecs.New(sess)

	_, err := client.UpdateService(&ecs.UpdateServiceInput{
		Cluster:              aws.String(cluster),
		Service:              aws.String(service),
		EnableExecuteCommand: aws.Bool(true),
		ForceNewDeployment:   aws.Bool(true),
	})

	if err != nil {
		fmt.Println("Failed to update ECS service:", err)
		return
	}

	fmt.Println("Execute command Enabled")
	fmt.Println("Service updated successfully")
	fmt.Println("Wait for new deployment and run again")

}

func ExecuteCommand(cluster, task, container, command string) {
	ecsCommand := fmt.Sprintf("aws ecs execute-command --cluster %s --task %s --container %s --command '%s' --interactive", cluster, task, container, command)

	output := exec.Command("sh", "-c", ecsCommand)
	output.Stdin = os.Stdin
	output.Stdout = os.Stdout

	err := output.Run()
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

}
