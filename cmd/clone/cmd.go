package clone

import (
	"os"

	"github.com/heathcliff26/gh-utility/pkg/git"
	"github.com/spf13/cobra"
)

const (
	fetchDepthFlag = "depth"
)

// Create a new clone command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone [flags] <repository> [<directory>]",
		Short: "Clone a repository with a temporary installation token into the specified directory",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			err := run(cmd, args)
			if err != nil {
				cmd.PrintErrln(cmd.ErrPrefix(), err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().IntP(fetchDepthFlag, "d", 0, "Fetch depth for the clone operation, defaults to 0 (everything)")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	fetchDepth, err := cmd.Flags().GetInt(fetchDepthFlag)
	if err != nil {
		return err
	}

	url := args[0]
	directory := ""
	if len(args) == 2 {
		directory = args[1]
	}

	return git.CloneRepository(url, directory, fetchDepth)
}
