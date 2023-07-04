package ecs

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

//TaskDefinition

func ListTaskDefinition() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "taskdefinition",
		Aliases: []string{"td"},
		Short:   "List ECS task definition",
		Run: func(cmd *cobra.Command, args []string) {
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))
			client := ecs.New(sess)

			input := &ecs.ListTaskDefinitionFamiliesInput{
				Status: aws.String("ACTIVE"),
			}

			response, err := client.ListTaskDefinitionFamilies(input)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\t")

			for _, tasks := range response.Families {
				task := *tasks
				input := &ecs.DescribeTaskDefinitionInput{
					TaskDefinition: aws.String(task),
				}

				_, err := client.DescribeTaskDefinition(input)
				if err != nil {
					fmt.Println(err)
					os.Exit(0)
				}

				fmt.Fprintf(w, "%s\t\n", task)
			}

			w.Flush()

		},
	}

	return cmd

}
