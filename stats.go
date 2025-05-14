package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/brianliurabbithole/gitlog/logger"
	"go.uber.org/zap"
	"gopkg.in/src-d/go-git.v4"
)

const (
	daysInLastSixMonths  = 183
	weeksInLastSixMonths = 26
	outOfRange           = 9999
)

type column []int

func stats(email string) {
	comments, err := processRepositories(email)
	if err != nil {
		logger.GetLogger().Error("Error processing repositories", zap.String("error", err.Error()))
		return
	}
	printCommitsStats(comments)
}

func processRepositories(email string) (map[int]int, error) {
	filePath := getDotFilePath()
	repositories, err := parseFileLinesToSlice(filePath)
	if err != nil {
		return nil, err
	}
	daysInMap := daysInLastSixMonths

	commits := make(map[int]int, daysInMap)

	for i := daysInMap; i > 0; i-- {
		commits[i] = 0
	}

	for _, repo := range repositories {
		commits = fillCommits(email, repo, commits)
	}

	return commits, nil
}

func fillCommits(email, path string, commits map[int]int) map[int]int {
	// Get the commits for the email
	repo, err := git.PlainOpen(path)
	if err != nil {
		logger.GetLogger().Error("Error opening repository", zap.String("path", path), zap.String("error", err.Error()))
		return nil
	}

	ref, err := repo.Head()
	if err != nil {
		logger.GetLogger().Error("Error getting repository head", zap.String("path", path), zap.String("error", err.Error()))
		return nil
	}

	// Get the commits for the email
	cIter, err := repo.Log(&git.LogOptions{
		From: ref.Hash(),
	})
	if err != nil {
		logger.GetLogger().Error("Error getting commits", zap.String("path", path), zap.String("error", err.Error()))
		return nil
	}
	// Iterate over the commits
	for {
		commit, err := cIter.Next()
		if err != nil {
			break
		}

		// Check if the commit author matches the email
		if commit.Author.Email != email {
			continue
		}

		// Get the commit date
		date := commit.Committer.When
		days := countDaysSinceDate(getBeginningOfDay(date))
		if days < 0 || days > daysInLastSixMonths {
			continue
		}
		// Increment the commit count for the day
		commits[days]++
	}
	// Return the commits map
	return commits
}

func getBeginningOfDay(date time.Time) time.Time {
	// Get the beginning of the day
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}

// countDaysSinceDate counts how many days passed since the passed `date`
func countDaysSinceDate(date time.Time) int {
	days := 0
	now := getBeginningOfDay(time.Now())
	for date.Before(now) {
		date = date.Add(time.Hour * 24)
		days++
		if days > daysInLastSixMonths {
			return outOfRange
		}
	}
	return days
}

func calcOffset() int {
	var offset int
	weekday := time.Now().Weekday()

	switch weekday {
	case time.Sunday:
		offset = 1
	case time.Monday:
		offset = 2
	case time.Tuesday:
		offset = 3
	case time.Wednesday:
		offset = 4
	case time.Thursday:
		offset = 5
	case time.Friday:
		offset = 6
	case time.Saturday:
		offset = 7
	default:
		offset = 0
	}

	return offset
}

func printCommitsStats(commits map[int]int) {
	keys := sortMapIntoSlice(commits)
	cols := buildCols(keys, commits)
	printCells(cols)
}

func sortMapIntoSlice(m map[int]int) []int {
	// Sort the map into a slice
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	// Sort the slice
	sort.Ints(keys)
	return keys
}

func buildCols(keys []int, commits map[int]int) map[int]column {
	// Build the columns
	cols := make(map[int]column)

	offset := calcOffset() - 1
	col := make(column, offset)
	for _, k := range keys {
		weekDay := k + offset
		week := weekDay / 7
		day := weekDay % 7
		if day == 0 {
			col = make(column, 0)
		}

		col = append(col, commits[k])
		if day == 6 {
			cols[week] = col
		}
	}

	return cols
}

// printCells prints the cells of the graph
func printCells(cols map[int]column) {
	printMonths()
	for j := 0; j < 7; j++ {
		for i := weeksInLastSixMonths + 1; i >= 0; i-- {
			if i == weeksInLastSixMonths+1 {
				printDayCol(j)
			}
			if col, ok := cols[i]; ok {
				//special case today
				if i == 0 && j == calcOffset()-1 {
					printCell(col[j], true)
					continue
				} else {
					if len(col) > j {
						printCell(col[j], false)
						continue
					}
				}
			}
			printCell(0, false)
		}
		fmt.Printf("\n")
	}
}

func printMonths() {
	week := getBeginningOfDay(time.Now()).Add(-(daysInLastSixMonths * time.Hour * 24))
	month := week.Month()
	fmt.Printf("         ")
	for {
		if week.Month() != month {
			fmt.Printf("%s ", week.Month().String()[:3])
			month = week.Month()
		} else {
			fmt.Printf("    ")
		}

		week = week.Add(time.Hour * 24 * 7)
		if week.After(time.Now()) {
			break
		}
	}
	fmt.Println()
}

func printDayCol(day int) {
	out := "     "
	switch day {
	case 5:
		out = " Mon "
	case 3:
		out = " Wed "
	case 1:
		out = " Fri "
	}
	fmt.Print(out)
}

// printCell given a cell value prints it with a different format
// based on the value amount, and on the `today` flag.
func printCell(val int, today bool) {
	escape := "\033[0;37;30m"
	switch {
	case val > 0 && val < 5:
		escape = "\033[1;30;47m"
	case val >= 5 && val < 10:
		escape = "\033[1;30;43m"
	case val >= 10:
		escape = "\033[1;30;42m"
	}

	if today {
		escape = "\033[1;37;45m"
	}

	if val == 0 {
		fmt.Print(escape + "  - " + "\033[0m")
		return
	}

	str := "  %d "
	switch {
	case val >= 10:
		str = " %d "
	case val >= 100:
		str = "%d "
	}

	fmt.Printf(escape+str+"\033[0m", val)
}
