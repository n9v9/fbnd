package fbnd

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const timetableURL = "https://mpl-server.kr.hs-niederrhein.de/fb03/sp/stundenplan.php"

// Degree is used to describe a DegreeProgram, either Bachelor or Master.
type Degree string

const (
	Bachelor Degree = "Bachelor"
	Master   Degree = "Master"
)

// SemesterCycle represents the type of a Semester, either Summer or Winter.
type SemesterCycle string

const (
	Summer SemesterCycle = "Summer"
	Winter SemesterCycle = "Winter"
)

// Semester is used to describe the accompanying DegreeProgram.
type Semester struct {
	Cycle SemesterCycle
	Year  int
	Term  int
}

// ID is the internal ID of each DegreeProgram returned by DegreePrograms.
// It is needed to fetch the timetable for a given DegreeProgram.
type ID string

// DegreeProgram represents a degree program for which a timetable is available.
type DegreeProgram struct {
	ID       ID
	Name     string
	Degree   Degree
	Semester Semester
}

// Lesson is used to describe the type of the accompanying Course.
type Lesson string

func (l Lesson) String() string {
	switch l {
	case Lecture:
		return "Lecture"
	case Exercise:
		return "Exercise"
	case Internship:
		return "Internship"
	case Seminar:
		return "Seminar"
	case SeminarLecture:
		return "Seminar Lecture"
	case LanguageLecture:
		return "Language Lecture"
	case Tutorial:
		return "Tutorial"
	default:
		panic(fmt.Sprintf("invalid value %s for type Lesson", string(l)))
	}
}

const (
	Lecture         Lesson = "V"
	Exercise        Lesson = "U"
	Internship      Lesson = "P"
	Seminar         Lesson = "S"
	SeminarLecture  Lesson = "SL"
	LanguageLecture Lesson = "F"
	Tutorial        Lesson = "T"
)

// Time represents the day, start and end of a Course.
type Time struct {
	Weekday   time.Weekday
	HourStart time.Duration
	HourEnd   time.Duration
}

// Course represents a single course of a timetable for a DegreeProgram.
type Course struct {
	NameLong       string
	NameShort      string
	ProfessorLong  string
	ProfessorShort string
	Room           string
	Lesson         Lesson
	Time           Time
}

// TimetableDay contains all courses for a Weekday for an accompanying Timetable.
type TimetableDay struct {
	Weekday time.Weekday
	Courses []Course
}

// Timetable contains all days for which courses exist as well as a DegreeProgram
// for which the timetable is valid.
type Timetable struct {
	// Can be nil, use FillDegreeProgram to fill the value, or if you obtained
	// the right DegreeProgram through calling DegreePrograms prior to calling
	// TimetableForDegreeProgram you can set the instance yourself.
	DegreeProgram *DegreeProgram
	Days          []TimetableDay
	id            ID
	oldCycle      SemesterCycle
}

// FillDegreeProgram calls DegreePrograms at most one time to obtain the correct
// DegreeProgram.
// The reason that t.DegreeProgram can be nil is as follows:
// If the function TimetableForDegreeProgram is called there is no way to know
// which semester the passed in ID belongs to; the server defaults to Winter.
// Now if the ID belongs to Winter then the DegreeProgram can be found and parsed within
// one request but if it belongs to Summer then the response we get does not contain
// the DegreeProgram, only the timetable for it and another request has to be made.
func (t *Timetable) FillDegreeProgram() error {
	if t.DegreeProgram != nil {
		return nil
	}

	newCycle := Summer
	if t.oldCycle == Summer {
		newCycle = Winter
	}
	programs, err := DegreePrograms(newCycle)
	if err != nil {
		return err
	}

	for _, v := range programs {
		if v.ID == t.id {
			t.DegreeProgram = &v
			return nil
		}
	}

	// This should never happen unless the structure of the website changes.
	panic("could not find DegreeProgram after parsing sites for both summer and winter")
}

// DegreePrograms returns all degree programs for which timetables are available
// and that fall into the given cycle.
// If the HTML could not be parsed, an error is returned.
func DegreePrograms(cycle SemesterCycle) ([]DegreeProgram, error) {
	doc, err := degreeProgramsDoc(cycle)
	if err != nil {
		return nil, err
	}

	year, parsedCycle, err := parseSemesterYear(doc)
	if err != nil {
		return nil, err
	}
	if parsedCycle != cycle {
		// This should never happen unless the structure of the website changes.
		panic(fmt.Sprintf("expected parsed semester cycle %s but got %s", cycle, parsedCycle))
	}

	return parseDegreeProgramNames(doc, cycle, year)
}

// TimetableForDegreeProgram returns a Timetable that contains all courses for the given degree program.
// The days inside Timetable are sorted by their weekday and the courses inside each day are sorted
// by their start hour.
// The ID can be obtained by calling DegreePrograms.
func TimetableForDegreeProgram(id ID) (*Timetable, error) {
	id = ID(strings.ToUpper(string(id)))

	doc, err := timeTableDoc(id)
	if err != nil {
		return nil, err
	}

	hours, err := parseHours(doc)
	if err != nil {
		return nil, err
	}

	weekdays := map[string]time.Weekday{
		"Mo": time.Monday,
		"Di": time.Tuesday,
		"Mi": time.Wednesday,
		"Do": time.Thursday,
		"Fr": time.Friday,
		"Sa": time.Saturday,
	}

	var (
		courses        []Course
		currentWeekday time.Weekday
		errEach        error
	)

	doc.Find("tbody tr:not([style])").EachWithBreak(func(i int, s *goquery.Selection) bool {
		// Keep track of the column spans per course.
		var offset int
		s.Find("td").EachWithBreak(func(i int, s *goquery.Selection) bool {
			// Only the first element has this class so this means
			// it contains a weekday.
			if i == 0 && s.HasClass("text-center") {
				currentWeekday = weekdays[strings.TrimSpace(s.Text())]
				return true
			}

			// All other `td` elements that contain a course have the attribute `title`.
			title, ok := s.Attr("title")
			if !ok {
				return true
			}

			fields := strings.Split(title, "/")
			nameLong := strings.TrimSpace(fields[0])
			professorLong := strings.TrimSpace(fields[1])

			fields = strings.Fields(strings.TrimSpace(s.Text()))
			nameShort := fields[0]
			lessonType := fields[1]
			professorShort := fields[2]
			room := fields[3]

			span, err := strconv.Atoi(s.AttrOr("colspan", "1"))
			if err != nil {
				errEach = err
				return false
			}
			span--

			courses = append(courses, Course{
				NameLong:       nameLong,
				NameShort:      nameShort,
				ProfessorLong:  professorLong,
				ProfessorShort: professorShort,
				Room:           room,
				Lesson:         Lesson(lessonType),
				Time: Time{
					Weekday:   currentWeekday,
					HourStart: hours[i+offset].HourStart,
					HourEnd:   hours[i+offset+span].HourEnd,
				},
			})

			offset += span
			return true
		})
		return errEach == nil
	})

	// Sort the courses by weekdays and then by their start hour.
	weekdaysOrder := map[time.Weekday]int{
		time.Monday:    0,
		time.Tuesday:   1,
		time.Wednesday: 2,
		time.Thursday:  3,
		time.Friday:    4,
		time.Saturday:  5,
	}

	// Map each course to its weekday.
	timetable := make(map[time.Weekday][]Course)
	for _, v := range courses {
		timetable[v.Time.Weekday] = append(timetable[v.Time.Weekday], v)
	}

	days := make([]TimetableDay, 0, len(timetable))
	for weekday, courses := range timetable {
		// Sort the courses within each day by their start hour.
		sort.SliceStable(courses, func(i, j int) bool {
			return courses[i].Time.HourStart < courses[j].Time.HourStart
		})
		days = append(days, TimetableDay{
			Weekday: weekday,
			Courses: courses,
		})
	}

	// Sort the days by their weekday.
	sort.SliceStable(days, func(i, j int) bool {
		return weekdaysOrder[days[i].Weekday] < weekdaysOrder[days[j].Weekday]
	})

	year, cycle, err := parseSemesterYear(doc)
	if err != nil {
		return nil, err
	}

	// Try to find the DegreeProgram as it might not be possible, see FillDegreeProgram for more.
	var selected *DegreeProgram
	names, err := parseDegreeProgramNames(doc, cycle, year)
	if err != nil {
		return nil, err
	}
	for _, v := range names {
		if v.ID == id {
			selected = &v
		}
	}

	return &Timetable{
		DegreeProgram: selected,
		Days:          days,
		id:            id,
		oldCycle:      cycle,
	}, errEach
}

// parseHours returns a map that maps the index of each `th` element to its containing Time.
// This way, getting the Time for a `td` element can be done by indexing
// the map with the index of the `td` element.
func parseHours(doc *goquery.Document) (map[int]Time, error) {
	var (
		hours   = make(map[int]Time)
		errEach error
	)

	doc.Find("thead tr th:not(:first-child)").EachWithBreak(func(i int, s *goquery.Selection) bool {
		fields := strings.Split(strings.TrimSpace(s.Text()), "-")
		start, err := strconv.Atoi(fields[0])
		if err != nil {
			errEach = err
			return false
		}
		end, err := strconv.Atoi(fields[1])
		if err != nil {
			errEach = err
			return false
		}

		// We need i+1 instead of i because we skipped the first `th` element with `:not(:first-child)`.
		hours[i+1] = Time{
			HourStart: time.Hour * time.Duration(start),
			HourEnd:   time.Hour * time.Duration(end),
		}

		return true
	})

	return hours, errEach
}

func parseDegreeProgramNames(doc *goquery.Document, cycle SemesterCycle, year int) ([]DegreeProgram, error) {
	// All available degree programs are structured in the following way:
	// <select id="select_S">
	//     <optgroup label="Bachelor">
	//         <option value="<ID>">Bachelor <degree program name> (<semester term> Semester)</option>
	//         ...
	//     </optgroup>
	//         <option value="<ID>">Master <degree program name> (<semester term> Semester)</option>
	//         ...
	//     <optgroup label="Master">
	//     </optgroup>
	// </select>
	degreeSelector := `#select_S > optgroup[label="Bachelor"] option, #select_S > optgroup[label="Master"] option`

	r := regexp.MustCompile(`(Bachelor|Master) +(.*?) +\((\d+)`)

	var (
		programs []DegreeProgram
		errEach  error
	)

	doc.Find(degreeSelector).EachWithBreak(func(_ int, s *goquery.Selection) bool {
		id, exists := s.Attr("value")
		if !exists {
			// This should never happen unless the structure of the site changes.
			errEach = fmt.Errorf("could not get ID of degree program '%s'", s.Text())
			return false
		}

		groups := r.FindStringSubmatch(strings.TrimSpace(s.Text()))
		term, err := strconv.Atoi(groups[3])
		if err != nil {
			// This should never happen unless the structure of the site changes.
			errEach = fmt.Errorf("could not parse semester term %s as int: %w", groups[3], err)
			return false
		}

		var degree Degree
		if groups[1] == "Bachelor" {
			degree = Bachelor
		} else {
			degree = Master
		}

		programs = append(programs, DegreeProgram{
			ID:     ID(strings.ToUpper(id)),
			Name:   groups[2],
			Degree: degree,
			Semester: Semester{
				Cycle: cycle,
				Term:  term,
				Year:  year,
			},
		})

		return true
	})

	return programs, errEach
}

// parseSemesterYear looks for the radio box in doc that describes the semester,
// either winter or summer and that is checked.
// It returns the year of the semester, the SemesterCycle as well as any error that occurred.
func parseSemesterYear(doc *goquery.Document) (year int, cycle SemesterCycle, err error) {
	var yearText string

	if _, ok := doc.Find(`input[id="inlineWintersemester"]`).Attr("checked"); ok {
		yearText = doc.Find(`label[for="inlineWintersemester"]`).Text()
	} else if _, ok := doc.Find(`input[id="inlineSommersemester"]`).Attr("checked"); ok {
		yearText = doc.Find(`label[for="inlineSommersemester"]`).Text()
	} else {
		// We should never get here unless the structure of the website changes.
		panic("could not parse semester year")
	}

	r := regexp.MustCompile(`^(Winter|Sommer)semester (\d{4})`)
	groups := r.FindStringSubmatch(yearText)

	switch groups[1] {
	case "Winter":
		cycle = Winter
	case "Sommer":
		cycle = Summer
	}

	year, err = strconv.Atoi(groups[2])
	return
}

// degreeProgramsDoc fetches the HTML for the cycle and returns the parsed document.
// If the request failed or the response could not be parsed, an error is returned.
func degreeProgramsDoc(cycle SemesterCycle) (*goquery.Document, error) {
	var semester string
	switch cycle {
	case Summer:
		semester = "SS"
	case Winter:
		semester = "WS"
	}

	resp, err := http.PostForm(timetableURL, url.Values{
		"Lage":  []string{semester},
		"fkt":   []string{"SR"},
		"clear": []string{"false"},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return goquery.NewDocumentFromReader(resp.Body)
}

// timeTableDoc fetches the HTML for the id and returns the parsed document.
// If the request failed or the response could not be parsed, an error is returned.
func timeTableDoc(id ID) (*goquery.Document, error) {
	resp, err := http.PostForm(timetableURL, url.Values{
		"fkt":   []string{"SR"},
		"SR":    []string{string(id)},
		"mode":  []string{"SR"},
		"clear": []string{"false"},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return goquery.NewDocumentFromReader(resp.Body)
}
