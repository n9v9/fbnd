package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
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

	if outputJSON {
		// When we output JSON we want to get the accompanying DegreeProgram.
		if timetable.DegreeProgram == nil {
			if err := timetable.FillDegreeProgram(); err != nil {
				return err
			}
		}
		return json.NewEncoder(os.Stdout).Encode(timetable)
	}

	printlnWeekday := color.New(color.FgWhite, color.Underline, color.Bold).PrintlnFunc()
	printlnWeekdayToday := color.New(color.FgGreen, color.Underline, color.Bold).PrintlnFunc()

	for _, day := range timetable.Days {
		if day.Weekday == time.Now().Weekday() {
			printlnWeekdayToday(day.Weekday)
		} else {
			printlnWeekday(day.Weekday)
		}

		maxNameShort := internal.Max(day.Courses, func(i int) int { return len(day.Courses[i].NameShort) })
		maxLesson := internal.Max(day.Courses, func(i int) int { return len(day.Courses[i].Lesson.String()) })
		maxProfessorShort := internal.Max(day.Courses, func(i int) int { return len(day.Courses[i].ProfessorShort) })

		for _, v := range day.Courses {
			fmt.Printf("%02d - %02d | %-*s | %-*s | %0-*s | %s\n",
				v.Time.HourStart, v.Time.HourEnd,
				maxNameShort, v.NameShort,
				maxLesson, v.Lesson,
				maxProfessorShort, v.ProfessorShort,
				v.Room)
		}
	}

	return nil
}
