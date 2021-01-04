package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/n9v9/fbnd"
	"github.com/n9v9/fbnd/cmd/fbnd/cmd/internal"
	"github.com/spf13/cobra"
)

var (
	summer, winter bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all degree programs for which timetables are available",
	Args: func(cmd *cobra.Command, args []string) error {
		if summer && winter {
			return errors.New("the flags summer and winter are mutually exclusive")
		}
		return cobra.NoArgs(cmd, args)
	},
	Run: func(cmd *cobra.Command, _ []string) {
		if err := runList(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&summer, "summer", "s", false, "List degree programs for summer semesters only")
	listCmd.Flags().BoolVarP(&winter, "winter", "w", false, "List degree programs for winter semesters only")
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

	maxID := internal.Max(programs, func(i int) int { return len(programs[i].ID) })
	maxCycle := internal.Max(programs, func(i int) int { return len(formatCycle(programs[i].Semester)) })
	maxSemester := internal.Max(programs, func(i int) int { return len(formatSemester(programs[i].Semester)) })
	maxDegree := internal.Max(programs, func(i int) int { return len(programs[i].Degree) })

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
