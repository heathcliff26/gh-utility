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
	}

	keyPath := cmd.Flags().StringP(appKeyPathFlag, "k", "", "Path to the private key of the app")
	_ = cmd.MarkFlagFilename(appKeyPathFlag)
	_ = cmd.MarkFlagRequired(appKeyPathFlag)

	clientID := cmd.Flags().StringP(clientIDFlag, "c", "", "Client ID of the app")
	_ = cmd.MarkFlagRequired(clientIDFlag)

	installationID := cmd.Flags().StringP(installationIDFlag, "i", "", "Installation ID of the app")
	_ = cmd.MarkFlagRequired(installationIDFlag)

	output := cmd.Flags().StringP(outputFlag, "o", "", "Write the token to disk instead of outputting on console")
	_ = cmd.MarkFlagFilename(outputFlag)

	endpoint := cmd.Flags().String(endpointFlag, client.DefaultEndpoint, "GitHub API endpoint")

	cmd.Run = func(cmd *cobra.Command, _ []string) {
		err := run(cmd, *keyPath, *clientID, *installationID, *output, *endpoint)
		if err != nil {
			cmd.PrintErrln(cmd.ErrPrefix(), err)
			os.Exit(1)
		}
	}

	return cmd
}

func run(cmd *cobra.Command, keyPath, clientID, installationID, output, endpoint string) error {
	c := client.NewClient(endpoint)
	token, err := c.GetToken(keyPath, clientID, installationID)
	if err != nil {
		return err
	}

	if output == "" {
		cmd.Println(token)
	} else {
		// #nosec G306 -- File should be readable by CI systems, final file permissions are decided by umask.
		return os.WriteFile(output, []byte(token), 0666)
	}
	return nil
}
