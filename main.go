package main

import (
	"fmt"
	"nami/pkg/ecs"
	"nami/pkg/pulumi"
	"os"

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
		Short: "Describe Resources",
	}
	rootCmd.AddCommand(Version())

	//login
	rootCmd.AddCommand(pulumi.Login())
	//cluster
	getCmd.AddCommand(ecs.ListClusters())
	describeCmd.AddCommand(ecs.DescribeCluster())

	//service
	getCmd.AddCommand(ecs.ListServices())
	describeCmd.AddCommand(ecs.DescribeService())
	logsCmd.AddCommand(ecs.ServiceLogs())

	//taskdefinition
	getCmd.AddCommand(ecs.ListTaskDefinition())
	describeCmd.AddCommand(ecs.DescribeTaskDefinition())

	//task
	getCmd.AddCommand(ecs.ListTasks())
	describeCmd.AddCommand(ecs.DescribeTask())
	logsCmd.AddCommand(ecs.TaskLogs())
	//root
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(describeCmd)
	rootCmd.AddCommand(logsCmd)

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
