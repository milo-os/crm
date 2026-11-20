// SPDX-License-Identifier: AGPL-3.0-only

// Package cmd contains the cobra command definitions for the crm binary.
package cmd

import "github.com/spf13/cobra"

// BuildInfo carries version metadata injected at build time via -ldflags.
type BuildInfo struct {
	Version      string
	GitCommit    string
	GitTreeState string
	BuildDate    string
}

// NewRootCommand returns the crm root cobra command.
// It has no RunE — invoking 'crm' with no subcommand prints help.
func NewRootCommand(info BuildInfo) *cobra.Command {
	root := &cobra.Command{
		Use:   "crm",
		Short: "Milo CRM controller manager",
		// No RunE — 'crm' with no subcommand prints help.
	}
	root.AddCommand(newOperatorCommand(info))
	return root
}
