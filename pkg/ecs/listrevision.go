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

// TaskDefinition

func ListTaskDefinitionRevision() *cobra.Command {
	var taskdef string
	cmd := &cobra.Command{
		Use:     "revision",
		Aliases: []string{"revisions"},
		Short:   "List ECS task definition revisions",
		Run: func(cmd *cobra.Command, args []string) {
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))
			taskdef = args[0]
			client := ecs.New(sess)

			input := &ecs.ListTaskDefinitionsInput{
				Status:       aws.String("ACTIVE"),
				FamilyPrefix: aws.String(taskdef),
			}

			response, err := client.ListTaskDefinitions(input)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}

			// Create a tabwriter with custom formatting
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.TabIndent)
			defer w.Flush()

			// Define table headers
			fmt.Fprintln(w, "REVISION\tIMAGE\tCREATED")

			for _, arn := range response.TaskDefinitionArns {
				revision := NameArn(*arn)

				input := &ecs.DescribeTaskDefinitionInput{
					TaskDefinition: aws.String(revision),
				}
				descResp, err := client.DescribeTaskDefinition(input)
				if err != nil {
					fmt.Println(err)
					os.Exit(0)
				}

				image := aws.StringValue(descResp.TaskDefinition.ContainerDefinitions[0].Image)
				created := aws.TimeValue(descResp.TaskDefinition.RegisteredAt)

				// Print formatted table row
				fmt.Fprintf(w, "%s\t%s\t%s\n", revision, image, created.Format("2006-01-02 15:04:05 MST"))
			}
		},
	}

	return cmd
}
