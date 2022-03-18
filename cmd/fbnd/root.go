package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Global flags that are set for all commands.
var (
	outputJSON = false
)

func cmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fbnd",
		Version: version(),
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
	cmd.SetVersionTemplate("{{.Version}}")

	cmd.PersistentFlags().BoolVar(&outputJSON, "json", false, "Enable printing results in JSON format")
	cmd.PersistentFlags().Bool("no-color", false, "Disable colorized output")

	cmd.AddCommand(cmdTime())
	cmd.AddCommand(cmdList())

	return cmd
}

func version() string {
	info, ok := debug.ReadBuildInfo()

	if !ok {
		return "unknown"
	}

	var sb strings.Builder

	sb.WriteString("Go version: ")
	sb.WriteString(info.GoVersion)
	sb.WriteByte('\n')
	sb.WriteString("Git commit: ")

	for _, v := range info.Settings {
		if v.Key == "vcs.revision" {
			sb.WriteString(v.Value)
		}
	}

	sb.WriteByte('\n')

	return sb.String()
}
