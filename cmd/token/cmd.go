package token

import (
	"fmt"
	"os"

	"github.com/heathcliff26/gh-utility/pkg/client"
	"github.com/spf13/cobra"
)

const (
	appKeyPathFlag     = "key"
	clientIDFlag       = "client"
	installationIDFlag = "installation"
)

// Create a new token command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Create a new temporary access token for the app",
		Run: func(cmd *cobra.Command, _ []string) {
			err := run(cmd)
			if err != nil {
				cmd.PrintErrln(cmd.ErrPrefix(), err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringP(appKeyPathFlag, "k", "", "Path to the private key of the app")
	_ = cmd.MarkFlagFilename(appKeyPathFlag)
	_ = cmd.MarkFlagRequired(appKeyPathFlag)

	cmd.Flags().StringP(clientIDFlag, "c", "", "Client ID of the app")
	_ = cmd.MarkFlagRequired(clientIDFlag)

	cmd.Flags().StringP(installationIDFlag, "i", "", "Installation ID of the app")
	_ = cmd.MarkFlagRequired(installationIDFlag)

	return cmd
}

func run(cmd *cobra.Command) error {
	keyPath, err := cmd.Flags().GetString(appKeyPathFlag)
	if err != nil {
		return err
	}
	clientID, err := cmd.Flags().GetString(clientIDFlag)
	if err != nil {
		return err
	}
	installationID, err := cmd.Flags().GetString(installationIDFlag)
	if err != nil {
		return err
	}

	c := client.NewClient()
	token, err := c.GetToken(keyPath, clientID, installationID)
	if err != nil {
		return err
	}

	fmt.Println(token)

	return nil
}
