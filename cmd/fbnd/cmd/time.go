package cmd

import (
	"fmt"
	"os"

	"github.com/n9v9/fbnd"
	"github.com/n9v9/fbnd/cmd/fbnd/cmd/internal"
	"github.com/spf13/cobra"
)

var timeCmd = &cobra.Command{
	Use:   "time",
	Short: "Display the timetable for a specific degree program",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runTime(args[0]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(timeCmd)
}

func runTime(id string) error {
	timetable, err := fbnd.TimetableForDegreeProgram(fbnd.ID(id))
	if err != nil {
		return err
	}
	if len(timetable.Days) == 0 {
		return fmt.Errorf("could find no courses for degree program with id %s", id)
	}

	for _, day := range timetable.Days {
		fmt.Println(day.Weekday)

		maxNameShort := internal.Max(day.Courses, func(i int) int { return len(day.Courses[i].NameShort) })
		maxLesson := internal.Max(day.Courses, func(i int) int { return len(day.Courses[i].Lesson.String()) })
		maxProfessorShort := internal.Max(day.Courses, func(i int) int { return len(day.Courses[i].ProfessorShort) })

		for _, v := range day.Courses {
			fmt.Printf("%02d - %02d | %-*s | %-*s | %0-*s | %s\n",
				int(v.Time.HourStart.Hours()), int(v.Time.HourEnd.Hours()),
				maxNameShort, v.NameShort,
				maxLesson, v.Lesson,
				maxProfessorShort, v.ProfessorShort,
				v.Room)
		}
	}

	return nil
}
