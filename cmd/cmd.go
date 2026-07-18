package main

import (
	"github.com/heathcliff26/gh-utility/cmd/clone"
	"github.com/heathcliff26/gh-utility/cmd/commit"
	"github.com/heathcliff26/gh-utility/cmd/pr"
	"github.com/heathcliff26/gh-utility/cmd/token"
	"github.com/heathcliff26/gh-utility/pkg/version"
	"github.com/spf13/cobra"
)

const Name = "gh-utility"

func NewRoot() *cobra.Command {
	cobra.AddTemplateFunc(
		"ProgramName", func() string {
			return Name
		},
	)

	rootCmd := &cobra.Command{
		Use:   Name,
		Short: Name + " to interact with the GitHub API as an app",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	rootCmd.AddCommand(
		clone.NewCommand(),
		commit.NewCommand(),
		pr.NewCommand(),
		token.NewCommand(),
		version.NewCommand(Name),
	)

	return rootCmd
}
