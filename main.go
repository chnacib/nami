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
	//login
	rootCmd.AddCommand(pulumi.Login())
	//cluster
	getCmd.AddCommand(ecs.ListClusters())
	describeCmd.AddCommand(ecs.DescribeCluster())

	//service
	getCmd.AddCommand(ecs.ListServices())
	describeCmd.AddCommand(ecs.DescribeService())

	//taskdefinition
	getCmd.AddCommand(ecs.ListTaskDefinition())
	describeCmd.AddCommand(ecs.DescribeTaskDefinition())

	//task
	getCmd.AddCommand(ecs.ListTasks())
	describeCmd.AddCommand(ecs.DescribeTask())
	//root
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(describeCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

}
