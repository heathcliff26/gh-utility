package token

import (
	"os"

	"github.com/heathcliff26/gh-utility/pkg/client"
	"github.com/spf13/cobra"
)

const (
	appKeyPathFlag     = "key"
	clientIDFlag       = "client-id"
	installationIDFlag = "installation-id"
	outputFlag         = "output"
	endpointFlag       = "endpoint"
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

	cmd.Flags().StringP(outputFlag, "o", "", "Write the token to disk instead of outputting on console")
	_ = cmd.MarkFlagFilename(outputFlag)

	cmd.Flags().String(endpointFlag, client.DefaultEndpoint, "GitHub API endpoint")

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
	output, err := cmd.Flags().GetString(outputFlag)
	if err != nil {
		return err
	}
	endpoint, err := cmd.Flags().GetString(endpointFlag)
	if err != nil {
		return err
	}

	c := client.NewClient(endpoint)
	token, err := c.GetToken(keyPath, clientID, installationID)
	if err != nil {
		return err
	}

	if output == "" {
		cmd.Println(token)
	} else {
		return os.WriteFile(output, []byte(token), 0666)
	}
	return nil
}
