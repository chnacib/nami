package ecs

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/spf13/cobra"
)

//Cluster operations

func ListClusters() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "clusters",
		Aliases: []string{"cluster"},
		Short:   "List ECS Clusters in default region",
		Run: func(cmd *cobra.Command, args []string) {
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			client := ecs.New(sess)

			input := &ecs.ListClustersInput{
				MaxResults: aws.Int64(100),
			}

			response, err := client.ListClusters(input)
			if err != nil {
				log.Fatal(err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\tARN")

			for _, clusterArn := range response.ClusterArns {
				clusterName := NameArn(*clusterArn)
				fmt.Fprintf(w, "%s\t%s\n", clusterName, *clusterArn)
			}

			w.Flush()

		},
	}

	return cmd
}
