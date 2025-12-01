package main

import (
	"fmt"
	"os"

	"github.com/chnacib/nami/pkg/ecs"
	"github.com/chnacib/nami/pkg/pulumi"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "nami",
		Short: "Nami CLI",
	}
	//logs
	logsCmd := &cobra.Command{
		Use:     "logs",
		Aliases: []string{"log"},
		Short:   "Retrieve log events from service/task",
	}

	//get
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get resources",
	}
	//describe
	describeCmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe resources",
	}
	rootCmd.AddCommand(Version())

	//set
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set config",
	}

	//delete

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources",
	}

	//deploy

	//login
	rootCmd.AddCommand(pulumi.Login())
	//cluster
	getCmd.AddCommand(ecs.ListClusters())
	describeCmd.AddCommand(ecs.DescribeCluster())
	deleteCmd.AddCommand(ecs.DeleteCluster())

	//service
	getCmd.AddCommand(ecs.ListServices())
	describeCmd.AddCommand(ecs.DescribeService())
	logsCmd.AddCommand(ecs.ServiceLogs())
	deleteCmd.AddCommand(ecs.DeleteService())

	//nodes
	getCmd.AddCommand(ecs.ListNodes())

	//taskdefinition
	getCmd.AddCommand(ecs.ListTaskDefinition())
	describeCmd.AddCommand(ecs.DescribeTaskDefinition())

	//revision
	getCmd.AddCommand(ecs.ListTaskDefinitionRevision())
	setCmd.AddCommand(ecs.UpdateRevision())

	//task
	getCmd.AddCommand(ecs.ListTasks())
	describeCmd.AddCommand(ecs.DescribeTask())
	logsCmd.AddCommand(ecs.TaskLogs())
	deleteCmd.AddCommand(ecs.DeleteTask())

	//exec
	rootCmd.AddCommand(ecs.Exec())

	//replicas
	setCmd.AddCommand(ecs.Replicas())

	//autoscaling
	getCmd.AddCommand(ecs.ListAutoscaling())
	setCmd.AddCommand(ecs.Autoscaling())

	//root
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(describeCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(ecs.Deploy())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

}

func Version() *cobra.Command {
	var version string = "alpha"

	cmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Show version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)

		},
	}
	return cmd

}
