package ecs

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

//TaskDefinition

func DescribeTaskDefinition() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "taskdefinition",
		Aliases: []string{"td"},
		Short:   "Describe ECS task definition",
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))
			client := ecs.New(sess)

			input := &ecs.DescribeTaskDefinitionInput{
				TaskDefinition: aws.String(name),
			}

			response, err := client.DescribeTaskDefinition(input)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
			fmt.Println(response)

		},
	}

	return cmd

}
