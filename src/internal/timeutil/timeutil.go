package timeutil

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aliaksandrZh/worklog/src/internal/model"
)

var (
	hoursRe = regexp.MustCompile(`(\d+(?:\.\d+)?)h`)
	minsRe  = regexp.MustCompile(`(\d+(?:\.\d+)?)m`)
)

// ParseTime converts a time string like "1h 30m" to decimal hours.
func ParseTime(timeStr string) float64 {
	if timeStr == "" {
		return 0
	}
	var total float64
	for _, m := range hoursRe.FindAllStringSubmatch(timeStr, -1) {
		v, _ := strconv.ParseFloat(m[1], 64)
		total += v
	}
	for _, m := range minsRe.FindAllStringSubmatch(timeStr, -1) {
		v, _ := strconv.ParseFloat(m[1], 64)
		total += v / 60
	}
	return total
}

// ParseDate parses M/D/YYYY to a time.Time. Returns zero time and false if invalid.
func ParseDate(dateStr string) (time.Time, bool) {
	parts := strings.Split(dateStr, "/")
	if len(parts) != 3 {
		return time.Time{}, false
	}
	m, err1 := strconv.Atoi(parts[0])
	d, err2 := strconv.Atoi(parts[1])
	y, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return time.Time{}, false
	}
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.Local), true
}

// GetWeekBounds returns the Monday 00:00 and Sunday 23:59:59 for the week containing date.
func GetWeekBounds(date time.Time) (monday, sunday time.Time) {
	d := date
	weekday := int(d.Weekday())
	// Shift so Monday=0
	offset := (weekday + 6) % 7
	monday = time.Date(d.Year(), d.Month(), d.Day()-offset, 0, 0, 0, 0, d.Location())
	sunday = monday.AddDate(0, 0, 6)
	sunday = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 999999999, d.Location())
	return monday, sunday
}

// FormatDateShort formats a date as M/D/YYYY (no zero-padding).
func FormatDateShort(d time.Time) string {
	return fmt.Sprintf("%d/%d/%d", d.Month(), d.Day(), d.Year())
}

// TodayStr returns today's date as M/D/YYYY.
func TodayStr() string {
	return FormatDateShort(time.Now())
}

// DateGroup represents a group of tasks sharing a date.
type DateGroup struct {
	Key   string
	Tasks []model.IndexedTask
	Total float64
}

// GroupByDate groups indexed tasks by date, sorted most-recent first.
func GroupByDate(tasks []model.IndexedTask) []DateGroup {
	m := make(map[string][]model.IndexedTask)
	for _, t := range tasks {
		m[t.Date] = append(m[t.Date], t)
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		di, _ := ParseDate(keys[i])
		dj, _ := ParseDate(keys[j])
		return di.After(dj) // most recent first
	})

	groups := make([]DateGroup, 0, len(keys))
	for _, k := range keys {
		var total float64
		for _, t := range m[k] {
			total += ParseTime(t.TimeSpent)
		}
		groups = append(groups, DateGroup{Key: k, Tasks: m[k], Total: total})
	}
	return groups
}

// SortTasks sorts tasks by the given column and direction.
func SortTasks(tasks []model.IndexedTask, sortBy, sortDir string) []model.IndexedTask {
	if sortBy == "" {
		return tasks
	}
	result := make([]model.IndexedTask, len(tasks))
	copy(result, tasks)

	dir := 1
	if sortDir == "desc" {
		dir = -1
	}

	sort.SliceStable(result, func(i, j int) bool {
		var cmp int
		a, b := result[i], result[j]
		switch sortBy {
		case "date":
			da, _ := ParseDate(a.Date)
			db, _ := ParseDate(b.Date)
			if da.Before(db) {
				cmp = -1
			} else if da.After(db) {
				cmp = 1
			}
		case "timeSpent":
			ta := ParseTime(a.TimeSpent)
			tb := ParseTime(b.TimeSpent)
			if ta < tb {
				cmp = -1
			} else if ta > tb {
				cmp = 1
			}
		default:
			va := getField(a.Task, sortBy)
			vb := getField(b.Task, sortBy)
			cmp = strings.Compare(va, vb)
		}
		return cmp*dir < 0
	})
	return result
}

func getField(t model.Task, field string) string {
	switch field {
	case "date":
		return t.Date
	case "type":
		return t.Type
	case "number":
		return t.Number
	case "name":
		return t.Name
	case "timeSpent":
		return t.TimeSpent
	case "comments":
		return t.Comments
	default:
		return ""
	}
}

// TodayIndex returns the index of today's date in the groups slice, or -1 if not found.
func TodayIndex(groups []DateGroup) int {
	today := TodayStr()
	for i, g := range groups {
		if g.Key == today {
			return i
		}
	}
	return -1
}

// SumTimeForDate returns total hours logged for a given date string across all tasks.
func SumTimeForDate(tasks []model.Task, dateStr string) float64 {
	var total float64
	for _, t := range tasks {
		if t.Date == dateStr {
			total += ParseTime(t.TimeSpent)
		}
	}
	return total
}

// WeekResult holds filtered tasks for a week period.
type WeekResult struct {
	Label string
	Tasks []model.IndexedTask
	Total float64
}

// FilterWeekByOffset filters tasks by week offset (0 = current week).
func FilterWeekByOffset(tasks []model.IndexedTask, offset int) WeekResult {
	ref := time.Now().AddDate(0, 0, offset*7)
	monday, sunday := GetWeekBounds(ref)

	var filtered []model.IndexedTask
	for _, t := range tasks {
		d, ok := ParseDate(t.Date)
		if ok && !d.Before(monday) && !d.After(sunday) {
			filtered = append(filtered, t)
		}
	}

	var total float64
	for _, t := range filtered {
		total += ParseTime(t.TimeSpent)
	}

	label := fmt.Sprintf("%s – %s", FormatDateShort(monday), FormatDateShort(sunday))
	return WeekResult{Label: label, Tasks: filtered, Total: total}
}

// MonthResult holds filtered tasks for a month period.
type MonthResult struct {
	Label string
	Tasks []model.IndexedTask
	Total float64
}

// FilterMonthByOffset filters tasks by month offset (0 = current month, -1 = previous).
func FilterMonthByOffset(tasks []model.IndexedTask, offset int) MonthResult {
	now := time.Now()
	ref := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, offset, 0)
	first := time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, ref.Location())
	last := first.AddDate(0, 1, -1)
	last = time.Date(last.Year(), last.Month(), last.Day(), 23, 59, 59, 999999999, last.Location())

	var filtered []model.IndexedTask
	for _, t := range tasks {
		d, ok := ParseDate(t.Date)
		if ok && !d.Before(first) && !d.After(last) {
			filtered = append(filtered, t)
		}
	}

	var total float64
	for _, t := range filtered {
		total += ParseTime(t.TimeSpent)
	}

	label := ref.Format("January 2006")
	return MonthResult{Label: label, Tasks: filtered, Total: total}
}
