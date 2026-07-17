package commit

import (
	"os"

	"github.com/heathcliff26/gh-utility/pkg/client"
	"github.com/heathcliff26/gh-utility/pkg/pullrequest"
	"github.com/heathcliff26/gh-utility/pkg/utils"
	"github.com/spf13/cobra"
)

const (
	repositoryFlag = "repository"
	branchFlag     = "branch"
	messageFlag    = "message"
	endpointFlag   = "endpoint"
)

// Create a new clone command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Push a new commit to the given branch using force push",
		Run: func(cmd *cobra.Command, _ []string) {
			err := run(cmd)
			if err != nil {
				cmd.PrintErrln(cmd.ErrPrefix(), err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringP(repositoryFlag, "r", ".", "Path to the git repository")

	cmd.Flags().StringP(branchFlag, "b", "", "Branch to push to, will be created if it does not exist")
	_ = cmd.MarkFlagRequired(branchFlag)

	cmd.Flags().StringP(messageFlag, "m", "", "Commit message")
	_ = cmd.MarkFlagRequired(messageFlag)

	cmd.Flags().String(endpointFlag, client.DefaultEndpoint, "GitHub API endpoint")

	return cmd
}

func run(cmd *cobra.Command) error {
	path, err := cmd.Flags().GetString(repositoryFlag)
	if err != nil {
		return err
	}
	branch, err := cmd.Flags().GetString(branchFlag)
	if err != nil {
		return err
	}
	msg, err := cmd.Flags().GetString(messageFlag)
	if err != nil {
		return err
	}

	endpoint, err := cmd.Flags().GetString(endpointFlag)
	if err != nil {
		return err
	}

	c := client.NewClient(endpoint)

	tree, commit, err := pullrequest.Commit(c, path, msg, branch, utils.GetToken())
	if tree == "" && commit == "" && err == nil {
		cmd.Println("No changes detected, exiting")
	} else {
		cmd.Printf("Tree: %s\nCommit: %s\n", tree, commit)
	}
	return err
}
