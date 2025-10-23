package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// ScanAllBots validates all bot files in a directory
func ScanAllBots(dir string, allowedAppID string) ([]string, error) {
	invalidBots := []string{}

	// Walk through the directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Only scan .html/.js files
		if filepath.Ext(path) != ".html" && filepath.Ext(path) != ".js" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNumber := 1
		re := regexp.MustCompile(`APP_ID\s*=\s*(\d+)`)

		for scanner.Scan() {
			line := scanner.Text()
			matches := re.FindStringSubmatch(line)
			if len(matches) == 2 && matches[1] != allowedAppID {
				invalidBots = append(invalidBots, fmt.Sprintf("%s (Line %d: %s)", path, lineNumber, matches[1]))
				break // stop scanning this file, already invalid
			}
			lineNumber++
		}
		return nil
	})

	return invalidBots, err
}
