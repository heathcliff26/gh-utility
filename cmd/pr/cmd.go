package pr

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
	messageFlag    = "commit-message"
	titleFlag      = "title"
	bodyFlag       = "body"
	labelFlag      = "label"
	endpointFlag   = "endpoint"
)

// Create a new clone command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Commits and (force) pushes current changes to a (new) branch and creates/updates a pull request",
		Run: func(cmd *cobra.Command, _ []string) {
			err := run(cmd)
			if err != nil {
				cmd.PrintErrln(cmd.ErrPrefix(), err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringP(repositoryFlag, "r", ".", "Path to the git repository.")

	cmd.Flags().StringP(branchFlag, "b", "", "Branch to push to, will be created if it does not exist")
	_ = cmd.MarkFlagRequired(branchFlag)

	cmd.Flags().String(messageFlag, "", "Commit message (default PR Title + Body)")

	cmd.Flags().StringP(titleFlag, "t", "", "Title of the PR")
	_ = cmd.MarkFlagRequired(titleFlag)

	cmd.Flags().StringP(bodyFlag, "m", "", "Description of the PR")

	cmd.Flags().StringArrayP(labelFlag, "l", []string{}, "Label(s) to add to the PR. Can be specified multiple times.")

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
	title, err := cmd.Flags().GetString(titleFlag)
	if err != nil {
		return err
	}
	body, err := cmd.Flags().GetString(bodyFlag)
	if err != nil {
		return err
	}
	// TODO: Add option to set labels for the PR
	labels, err := cmd.Flags().GetStringArray(labelFlag)
	if err != nil {
		return err
	}
	// TODO: Remove
	if len(labels) > 0 {
		cmd.PrintErr("Warning: Labels are currently not supported, they will be ignored\n")
	}
	endpoint, err := cmd.Flags().GetString(endpointFlag)
	if err != nil {
		return err
	}

	c := client.NewClient(endpoint)

	if msg == "" {
		msg = title
		if body != "" {
			msg += "\n\n" + body
		}
	}

	token := utils.GetToken()

	tree, commit, err := pullrequest.Commit(c, path, msg, branch, token)
	if tree == "" && commit == "" && err == nil {
		cmd.Println("No changes detected, exiting")
		return nil
	} else {
		cmd.Printf("Tree: %s\nCommit: %s\n\n", tree, commit)
	}
	if err != nil {
		return err
	}

	url, err := pullrequest.PullRequest(c, path, branch, title, body, token, labels)
	if err != nil {
		return err
	}
	cmd.Println(url)
	return nil
}
