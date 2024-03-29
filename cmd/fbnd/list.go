package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/n9v9/fbnd"
	"github.com/spf13/cobra"
)

var summer, winter bool

func cmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all degree programs for which timetables are available",
		Args: func(cmd *cobra.Command, args []string) error {
			if summer && winter {
				return errors.New("the flags summer and winter are mutually exclusive")
			}
			return cobra.NoArgs(cmd, args)
		},
		Run: func(_ *cobra.Command, _ []string) {
			if err := runList(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().BoolVarP(&summer, "summer", "s", false, "List degree programs for summer semesters only")
	cmd.Flags().BoolVarP(&winter, "winter", "w", false, "List degree programs for winter semesters only")

	return cmd
}

func runList() error {
	if !summer && !winter {
		// By default we want to display all degree programs.
		// So if they were not passed in we can invert them for easier handling.
		summer = true
		winter = true
	}

	const maxFetchCalls = 2
	var (
		programsCh = make(chan []fbnd.DegreeProgram, maxFetchCalls)
		errCh      = make(chan error, maxFetchCalls)
	)

	fetch := func(cycle fbnd.SemesterCycle) {
		programs, err := fbnd.DegreePrograms(cycle)
		if err != nil {
			var semester string
			if cycle == fbnd.Summer {
				semester = "summer"
			} else {
				semester = "winter"
			}
			errCh <- fmt.Errorf("could not fetch degree programs for the %v semester: %v", semester, err)
		}
		programsCh <- programs
	}

	var (
		programs       []fbnd.DegreeProgram
		realFetchCalls = 1
	)

	if summer && winter {
		realFetchCalls = 2
		go fetch(fbnd.Summer)
		go fetch(fbnd.Winter)
	} else if summer {
		fetch(fbnd.Summer)
	} else {
		fetch(fbnd.Winter)
	}

	for i := 0; i < realFetchCalls; i++ {
		select {
		case err := <-errCh:
			return err
		case p := <-programsCh:
			programs = append(programs, p...)
		}
	}

	return printTable(programs)
}

func printTable(programs []fbnd.DegreeProgram) error {
	if outputJSON {
		return json.NewEncoder(os.Stdout).Encode(programs)
	}

	formatCycle := func(s fbnd.Semester) string { return fmt.Sprintf("%s %d", s.Cycle, s.Year) }
	formatSemester := func(s fbnd.Semester) string { return fmt.Sprintf("Semester %d", s.Term) }
	formatHeader := func(cell string) string { return color.New(color.FgWhite, color.Bold).Sprint(cell) }

	maxID := Max(programs, func(v *fbnd.DegreeProgram) int { return len(v.ID) })
	maxCycle := Max(programs, func(v *fbnd.DegreeProgram) int { return len(formatCycle(v.Semester)) })
	maxSemester := Max(programs, func(v *fbnd.DegreeProgram) int { return len(formatSemester(v.Semester)) })
	maxDegree := Max(programs, func(v *fbnd.DegreeProgram) int { return len(v.Degree) })

	// Print the header.
	// This has to be so cumbersome because specifying a width for ANSI colored strings somehow has no effect.
	// So we can not format it like we do for all other lines.
	fmt.Fprintf(color.Output, "%s%s | %s%s | %s%s | %s%s | %s\n",
		formatHeader("ID"), strings.Repeat(" ", maxID-len("ID")),
		formatHeader("Cycle"), strings.Repeat(" ", maxCycle-len("Cycle")),
		formatHeader("Semester"), strings.Repeat(" ", maxSemester-len("Semester")),
		formatHeader("Degree"), strings.Repeat(" ", maxDegree-len("Degree")),
		formatHeader("Name"),
	)

	for _, v := range programs {
		fmt.Printf("%-*s | %-*s | %-*s | %-*s | %s\n",
			maxID, v.ID,
			maxCycle, formatCycle(v.Semester),
			maxSemester, formatSemester(v.Semester),
			maxDegree, v.Degree,
			v.Name,
		)
	}

	return nil
}
