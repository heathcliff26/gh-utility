package status

import (
	"fmt"
	"os"
	"strings"

	"github.com/heathcliff26/gh-utility/pkg/client"
	"github.com/heathcliff26/gh-utility/pkg/utils"
	"github.com/spf13/cobra"
)

const (
	endpointFlag    = "endpoint"
	descriptionFlag = "description"
	urlFlag         = "url"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [flags] <repository> <commit> <check-name>=<status> [<check-name>=<status>...]",
		Short: "Set or update the status of GitHub check runs on commits",
		Args:  cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			err := run(cmd, args)
			if err != nil {
				cmd.PrintErrln(cmd.ErrPrefix(), err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringP(endpointFlag, "e", client.DefaultEndpoint, "GitHub API endpoint")
	cmd.Flags().StringP(descriptionFlag, "d", "", "Description for the status check(s)")
	cmd.Flags().StringP(urlFlag, "u", "", "URL for the status check(s)")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	endpoint, err := cmd.Flags().GetString(endpointFlag)
	if err != nil {
		return err
	}
	description, err := cmd.Flags().GetString(descriptionFlag)
	if err != nil {
		return err
	}
	detailsURL, err := cmd.Flags().GetString(urlFlag)
	if err != nil {
		return err
	}

	repo := args[0]
	commit := args[1]
	checkPairs := args[2:]

	token := utils.GetToken()
	c := client.NewClient(endpoint)

	checks := make(map[string]string, len(checkPairs))
	for _, check := range checkPairs {
		parts := strings.SplitN(check, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid check format '%s', expected <check-name>=<status>", check)
		}
		checks[parts[0]] = parts[1]
	}

	checkRuns, err := c.GetCheckRunsForCommit(token, repo, commit)
	if err != nil {
		return fmt.Errorf("failed to get check runs for commit: %w", err)
	}

	for name, status := range checks {
		err = c.SetCheckRunStatus(token, repo, name, commit, description, detailsURL, status, checkRuns)
		if err != nil {
			return fmt.Errorf("failed to set check run '%s': %w", name, err)
		}

		cmd.Printf("Check run '%s' set to '%s'\n", name, status)
	}

	return nil
}
