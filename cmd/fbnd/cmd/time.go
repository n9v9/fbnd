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
	Long: `Display the timetable for a specific degree program

This command expects the ID of the degree program for which to display the timetable.
If you do not know the ID, you can see all available ones by calling the list command.`,
	Args: cobra.ExactArgs(1),
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
	printlnWeekdayToday := color.New(color.FgYellow, color.Underline, color.Bold).PrintlnFunc()
	printlnCourse := color.New(color.FgBlue, color.Bold).PrintlnFunc()
	printlnNextCourse := color.New(color.FgBlue).PrintlnFunc()

	currentHour := time.Now().Hour()

	nextCourseIndexes := func(courses []fbnd.Course) map[int]struct{} {
		// Map all start hours to the corresponding index into courses.
		startHours := make(map[int][]int)
		for i, v := range courses {
			startHours[v.Time.HourStart] = append(startHours[v.Time.HourStart], i)
		}

		// Now we save the nearest next courses.
		var (
			currentMin  []int
			nextMinHour *int
		)
		for k, v := range startHours {
			// This is important because the outer k does not change, only the outer k's value.
			k := k
			if k > currentHour && (nextMinHour == nil || k < *nextMinHour) {
				nextMinHour = &k
				currentMin = v
			}
		}

		// Build the result of next courses where the key is the index into courses.
		res := make(map[int]struct{})
		for _, courseIndex := range currentMin {
			res[courseIndex] = struct{}{}
		}
		return res
	}

	for _, day := range timetable.Days {
		var (
			isToday = day.Weekday == time.Now().Weekday()
			next    map[int]struct{}
		)

		if isToday {
			printlnWeekdayToday(day.Weekday)
			next = nextCourseIndexes(day.Courses)
		} else {
			printlnWeekday(day.Weekday)
		}

		maxNameShort := internal.Max(day.Courses, func(v *fbnd.Course) int { return len(v.NameShort) })
		maxLesson := internal.Max(day.Courses, func(v *fbnd.Course) int { return len(v.Lesson.String()) })
		maxProfessorShort := internal.Max(day.Courses, func(v *fbnd.Course) int { return len(v.ProfessorShort) })

		for i, v := range day.Courses {
			line := fmt.Sprintf("%02d - %02d | %-*s | %-*s | %0-*s | %s",
				v.Time.HourStart, v.Time.HourEnd,
				maxNameShort, v.NameShort,
				maxLesson, v.Lesson,
				maxProfessorShort, v.ProfessorShort,
				v.Room)

			if isToday && currentHour >= v.Time.HourStart && currentHour < v.Time.HourEnd {
				// Highlight the current course.
				printlnCourse(line)
				continue
			} else if isToday {
				if _, ok := next[i]; ok {
					// Highlight the next course.
					printlnNextCourse(line)
					continue
				}
			}

			// This course is either not today or not one of the directly next ones,
			// so we do not highlight it.
			fmt.Println(line)
		}
	}

	return nil
}
