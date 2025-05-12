package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/brianliurabbithole/gitlog/logger"
	"go.uber.org/zap"
)

func scan(path string) {
	// Scan the path for files
	fmt.Println("Scanning folder:", path)
	repositories := recursiveScanFolder(path)
	if len(repositories) == 0 {
		logger.GetLogger().Error("No repositories found.")
		return
	}

	// Get the path to the .dot file
	dotFilePath := getDotFilePath()
	if dotFilePath == "" {
		logger.GetLogger().Error("Error getting .dot file path.")
		return
	}

	// Add new elements to the .dot file
	addNewSliceElementsToFile(dotFilePath, repositories)
}

func recursiveScanFolder(path string) []string {
	return scanGitFolders([]string{}, path)
}

func getDotFilePath() string {
	user, err := os.UserHomeDir()
	if err != nil {
		logger.GetLogger().Error("Error getting user home directory", zap.String("error", err.Error()))
		return ".gogitlocalstats"
	}

	return user + "/.gogitlocalstats"
}

func addNewSliceElementsToFile(filePath string, slice []string) {
	existingRepos, err := parseFileLinesToSlice(filePath)
	if err != nil {
		logger.GetLogger().Error("Error parsing file lines to slice", zap.String("error", err.Error()))
		return
	}
	newRepos := joinSlice(slice, existingRepos)
	dumpStringSliceToFile(newRepos, filePath)
}

func parseFileLinesToSlice(filePath string) ([]string, error) {
	f, err := openFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			return nil, err
		}
	}

	return lines, nil
}

func openFile(filePath string) (*os.File, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		if os.IsNotExist(err) {
			return os.Create(filePath)
		}
		return nil, err
	}
	return file, nil
}

func joinSlice(new []string, existing []string) []string {
	for _, i := range new {
		if !sliceContains(existing, i) {
			existing = append(existing, i)
		}
	}

	return existing
}

func sliceContains(slice []string, item string) bool {
	for _, i := range slice {
		if i == item {
			return true
		}
	}
	return false
}

func dumpStringSliceToFile(repos []string, filePath string) {
	content := strings.Join(repos, "\n")
	os.WriteFile(filePath, []byte(content), 0755)
}

func scanGitFolders(folders []string, path string) []string {
	folder := strings.TrimSuffix(path, "/")

	f, err := os.Open(folder)
	if err != nil {
		logger.GetLogger().Error("Error opening folder", zap.String("folder", folder), zap.String("error", err.Error()))
		return folders
	}
	defer f.Close()

	files, err := f.Readdir(-1)
	if err != nil {
		logger.GetLogger().Error("Error reading folder", zap.String("folder", folder), zap.String("error", err.Error()))
		return folders
	}

	for _, file := range files {
		if !file.IsDir() ||
			file.Name() == "vendor" ||
			file.Name() == "node_modules" {
			continue
		}

		path := folder + "/" + file.Name()

		if file.Name() == ".git" {
			path = strings.TrimSuffix(path, "/.git")
			folders = append(folders, path)
			continue
		}

		folders = scanGitFolders(folders, path)
	}

	return folders
}
