package main

import (
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
		Short: Name + " interact with the GitHub API as an app",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	rootCmd.AddCommand(
		version.NewCommand(Name),
	)

	return rootCmd
}
