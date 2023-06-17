package pulumi

import (
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func Login() *cobra.Command {
	var token string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Pulumi Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			os.Setenv("PULUMI_ACCESS_TOKEN", token)
			login := exec.Command("pulumi", "login")
			login.Stdout = os.Stdout
			login.Stderr = os.Stderr

			err := login.Run()
			if err != nil {
				log.Fatal(err)
			}

		},
	}

	cmd.Flags().StringVarP(&token, "token", "t", "", "pulumi Access Token")

	return cmd
}
