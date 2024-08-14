package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
)

func main() {
	logger := log.New(os.Stderr, "", 0)

	if len(os.Args) < 2 {
		logger.Fatal("Usage: go run main.go <regex>")
		return
	}

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		logger.Fatalf("unable to determine path of source file: %s", filename)
	}

	migrationsDir := filepath.Join(filepath.Dir(filename), "..")

	regexPattern := os.Args[1]

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		logger.Fatalf("error reading migrations directory %q: %s", migrationsDir, err)
		return
	}

	var matchedFiles []string
	for _, file := range files {
		file = filepath.Base(file)
		if match, _ := regexp.MatchString(regexPattern, file); match {
			matchedFiles = append(matchedFiles, file)
		}
	}

	if len(matchedFiles) != 2 {
		logger.Printf("conflict: found %d files matching the regex, expected 2.", len(matchedFiles))
		for _, file := range matchedFiles {
			logger.Println(file)
		}

		logger.Fatal()
		return
	}

	var migrationNumbers []int
	for _, file := range files {
		file = filepath.Base(file)

		fnum := file[0:6]
		num, err := strconv.Atoi(fnum)
		if err != nil {
			logger.Printf("failed to convert %q to int: %s", fnum, err)
			continue
		}

		migrationNumbers = append(migrationNumbers, num)
	}

	sort.Ints(migrationNumbers)
	highestMigration := migrationNumbers[len(migrationNumbers)-1]
	newMigrationNumber := fmt.Sprintf("%06d", highestMigration+1)

	for _, file := range matchedFiles {
		oldPath := filepath.Join(migrationsDir, file)
		newName := newMigrationNumber + file[6:]
		newPath := filepath.Join(migrationsDir, newName)

		cmd := exec.Command("git", "mv", oldPath, newPath)
		err := cmd.Run()
		if err != nil {
			logger.Fatalf("error running git mv: %s", err)
			return
		}

		logger.Printf("renamed %s to %s using git mv", filepath.Base(oldPath), filepath.Base(newPath))
	}
}
