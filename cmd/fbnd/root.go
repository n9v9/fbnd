package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Values for these variables get injected by the go linker for every release.
var (
	version  string = "development"
	commitID string
)

// Global flags that are set for all commands.
var (
	outputJSON = false
)

func versionString() string {
	if commitID == "" {
		return version
	}
	return version + " " + commitID
}

func cmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fbnd",
		Version: versionString(),
		Short:   "Timetables of FB03 inside your terminal",
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			noColor, err := cmd.Flags().GetBool("no-color")
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			color.NoColor = noColor
		},
	}

	cmd.PersistentFlags().BoolVar(&outputJSON, "json", false, "Enable printing results in JSON format")
	cmd.PersistentFlags().Bool("no-color", false, "Disable colorized output")

	cmd.AddCommand(cmdTime())
	cmd.AddCommand(cmdList())

	return cmd
}
