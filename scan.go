package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func scan(path string) {
	// Scan the path for files
	fmt.Println("Scanning folder:", path)
	repositories := recursiveScanFolder(path)
	if len(repositories) == 0 {
		fmt.Println("No repositories found in the folder.")
		return
	}

	// Get the path to the .dot file
	dotFilePath := getDotFilePath()
	if dotFilePath == "" {
		fmt.Println("No .dot file found.")
		return
	}

	// Add new elements to the .dot file
	addNewSliceElementsToFile(dotFilePath, repositories)
	fmt.Println("Repositories added to .dot file.")
}

func recursiveScanFolder(path string) []string {
	return scanGitFolders([]string{}, path)
}

func getDotFilePath() string {
	user, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		return ".gogitlocalstats"
	}

	return user + "/.gogitlocalstats"
}

func addNewSliceElementsToFile(filePath string, slice []string) {
	existingRepos, err := parseFileLinesToSlice(filePath)
	if err != nil {
		log.Println("Error parsing file:", err)
		return
	}
	newRepos := joinSlice(slice, existingRepos)
	dumpStringSliceToFile(newRepos, filePath)
	fmt.Println("Repositories added to file:", filePath)
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
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0755)
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
		fmt.Println("Error opening folder:", err)
		return folders
	}
	defer f.Close()

	files, err := f.Readdir(-1)
	if err != nil {
		fmt.Println("Error reading folder:", err)
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
